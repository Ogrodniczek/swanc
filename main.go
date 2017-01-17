package main

import (
	"os"
	"os/exec"
	"time"

	"github.com/appscode/go/flags"
	"github.com/appscode/log"
	logs "github.com/appscode/log/golog"
	"github.com/spf13/pflag"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/cache"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/pkg/watch"
)

type syncer struct {
	master     string
	kubeconfig string
	client     *clientset.Clientset
	ttl        time.Duration

	hostIP string
	vpnPSK string // PSK: Pre Shared Key

	// This flag is only to write test driven code. Default value false.
	FakeServer bool

	reload chan struct{}
}

// Blocks caller. Intended to be called as a Go routine.
func (s *syncer) WatchNodes() {
	config, err := clientcmd.BuildConfigFromFlags(s.master, s.kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	s.client, err = clientset.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("started watching for peer endpoints")
	lw := &cache.ListWatch{
		ListFunc: func(opts kapi.ListOptions) (runtime.Object, error) {
			return s.client.Nodes().List(kapi.ListOptions{
				LabelSelector: labels.Everything(),
			})
		},
		WatchFunc: func(options kapi.ListOptions) (watch.Interface, error) {
			return s.client.Nodes().Watch(kapi.ListOptions{
				LabelSelector: labels.Everything(),
			})
		},
	}
	// kCachePopulated(k, events.Pod, &kapi.Pod{}, nil)
	_, controller := cache.NewInformer(lw,
		&kapi.Node{},
		s.ttl,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				log.Infoln("got one added event")
				s.reload <- struct{}{}
			},
			DeleteFunc: func(obj interface{}) {
				log.Infoln("got one deleted event", obj.(kapi.Node).Name)
				s.reload <- struct{}{}
			},
		},
	)
	controller.Run(wait.NeverStop)
}

func (s *syncer) Validate() {
	if s.hostIP == "" {
		log.Fatalln("Set HOST_IP environment variable to ip used for intra-cluster communication.")
	}
	if s.vpnPSK == "" {
		log.Fatalln("Set VPN_PSK environment variable to encrypt intra-cluster traffic.")
	}
}

// Blocks caller. Intended to be called as a Go routine.
func (s *syncer) SyncLoop() {
	for {
		select {
		case <-s.reload:
			s.reloadVPN()
		}
	}
}

func (s *syncer) reloadVPN() {
	nodes, err := s.client.Core().Nodes().List(kapi.ListOptions{LabelSelector: labels.Everything()})
	if err != nil {
		log.Fatalln(err)
	}

	nodeIPs := make([]string, len(nodes.Items))
	for i, node := range nodes.Items {
		for _, addr := range node.Status.Addresses {
			if addr.Type == kapi.NodeInternalIP {
				nodeIPs[i] = addr.Address
				break
			}
		}
		if nodeIPs[i] == "" {
			for _, addr := range node.Status.Addresses {
				if addr.Type == kapi.NodeExternalIP {
					nodeIPs[i] = addr.Address
					break
				}
			}
		}
	}

	f, err := os.OpenFile(confPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, confPerm)
	if err != nil {
		log.Fatalln(err)
	}

	err = cfgTemplate.Execute(f, struct {
		HostIP  string
		NodeIPs []string
	}{
		s.hostIP,
		nodeIPs,
	})
	if err := f.Close(); err != nil {
		log.Fatalln(err)
	}

	if err = exec.Command("/usr/sbin/ipsec", "update").Run(); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	defer logs.FlushLogs()

	s := &syncer{
		reload: make(chan struct{}),
	}
	pflag.StringVar(&s.master, "master", "", "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	pflag.StringVar(&s.kubeconfig, "kubeconfig", "", "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	pflag.DurationVar(&s.ttl, "peer-ttl", 10*time.Second, "The TTL for this node change watcher")
	pflag.BoolVar(&s.FakeServer, "fake", false, "runs a fake server only for testings")

	pflag.StringVar(&s.hostIP, "host-ip", os.Getenv("HOST_IP"), "IP used by host for intra-cluster communication")
	pflag.StringVar(&s.vpnPSK, "vpn-psk", os.Getenv("VPN_PSK"), "Pre shared secret used to encrypt VPN traffic")

	flags.InitFlags()
	logs.InitLogs()
	flags.DumpAll()

	s.Validate()
	s.reloadVPN() // initial loading
	go s.SyncLoop()
	s.WatchNodes()
}
