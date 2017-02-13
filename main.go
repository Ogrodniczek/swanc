package main

import (
	"os"
	"os/exec"
	"reflect"
	"sort"
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

	nodeName string
	nodeIP   string
	vpnPSK   string // PSK: Pre Shared Key

	// This flag is only to write test driven code. Default value false.
	FakeServer bool

	reload chan struct{}
}

// Blocks caller. Intended to be called as a Go routine.
func (s *syncer) WatchNodes() {
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
				log.Infoln("got one added node")
				s.reload <- struct{}{}
			},
			DeleteFunc: func(obj interface{}) {
				log.Infoln("got one deleted node", obj.(kapi.Node).Name)
				s.reload <- struct{}{}
			},
			UpdateFunc: func(old, new interface{}) {
				if !reflect.DeepEqual(old, new) {
					log.Infoln("got one updated node", new.(kapi.Node).Name)
					s.reload <- struct{}{}
				}
			},
		},
	)
	controller.Run(wait.NeverStop)
}

func (s *syncer) Validate() {
	if s.nodeIP == "" {
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
	config, err := clientcmd.BuildConfigFromFlags(s.master, s.kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	s.client, err = clientset.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	nodes, err := s.client.Core().Nodes().List(kapi.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"net.beta.appscode.com/vpn": "true",
		}),
	})
	if err != nil {
		log.Fatalln(err)
	}

	hasLabel := false

	nodeIPs := make([]string, len(nodes.Items))
	i := 0
	for _, node := range nodes.Items {
		var ip string
		for _, addr := range node.Status.Addresses {
			if addr.Type == kapi.NodeInternalIP {
				ip = addr.Address
				break
			}
		}
		if ip == "" {
			for _, addr := range node.Status.Addresses {
				if addr.Type == kapi.NodeExternalIP {
					ip = addr.Address
					break
				}
			}
		}
		if ip != s.nodeIP {
			nodeIPs[i] = ip
			i++
		} else {
			hasLabel = true
		}
	}

	if !hasLabel {
		node, err := s.client.Core().Nodes().Get(s.nodeName)
		if err != nil {
			log.Fatalln(err)
		}
		node.Labels["net.beta.appscode.com/vpn"] = "true"
		_, err = s.client.Core().Nodes().Update(node)
		if err != nil {
			log.Fatalln(err)
		}
	}

	nodeIPs = nodeIPs[:i]
	if i > 0 {
		sort.Strings(nodeIPs)

		f, err := os.OpenFile(confPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, confPerm)
		if err != nil {
			log.Fatalln(err)
		}

		err = cfgTemplate.Execute(f, struct {
			HostIP  string
			NodeIPs []string
		}{
			s.nodeIP,
			nodeIPs,
		})
		if err := f.Close(); err != nil {
			log.Fatalln(err)
		}

		if err = exec.Command("/usr/sbin/ipsec", "update").Run(); err != nil {
			log.Fatalln(err)
		}
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

	pflag.StringVar(&s.nodeName, "node-name", os.Getenv("NODE_NAME"), "Name used by kubernetes to identify host")
	pflag.StringVar(&s.nodeIP, "node-ip", os.Getenv("NODE_IP"), "IP used by host for intra-cluster communication")
	pflag.StringVar(&s.vpnPSK, "vpn-psk", os.Getenv("VPN_PSK"), "Pre shared secret used to encrypt VPN traffic")

	flags.InitFlags()
	logs.InitLogs()
	flags.DumpAll()

	s.Validate()
	s.reloadVPN() // initial loading
	go s.SyncLoop()
	s.WatchNodes()
}
