package main

import (
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
	ttl        time.Duration

	// This flag is only to write test driven code. Default value false.
	FakeServer bool

	added   chan kapi.Node
	removed chan kapi.Node
}

// Blocks caller. Intended to be called as a Go routine.
func (s *syncer) WatchNodes() {
	config, err := clientcmd.BuildConfigFromFlags(s.master, s.kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	client, err := clientset.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("started watching for peer endpoints")
	lw := &cache.ListWatch{
		ListFunc: func(opts kapi.ListOptions) (runtime.Object, error) {
			return client.Nodes().List(kapi.ListOptions{
				LabelSelector: labels.Everything(),
			})
		},
		WatchFunc: func(options kapi.ListOptions) (watch.Interface, error) {
			return client.Nodes().Watch(kapi.ListOptions{
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
				s.added <- obj.(kapi.Node)
			},
			DeleteFunc: func(obj interface{}) {
				log.Infoln("got one deleted event")
				s.removed <- obj.(kapi.Node)
			},
		},
	)
	controller.Run(wait.NeverStop)
}

// Blocks caller. Intended to be called as a Go routine.
func (s *syncer) SyncLoop() {
	for {
		select {
		case node := <-s.added:
			ip, name := s.detect(node)
			s.dostuff(ip, name, true)
		case node := <-s.removed:
			ip, name := s.detect(node)
			s.dostuff(ip, name, false)
		}
	}
}

const newline = "\n"

func (s *syncer) detect(node kapi.Node) (string, string) {
	var ip string
	for _, addr := range node.Status.Addresses {
		if addr.Type == kapi.NodeInternalIP {
			ip = addr.Address
		}
	}
	if ip == "" {
		log.Fatalln("No internal ip found for node", node)
	}
	return ip, node.GetName()
}

func (s *syncer) dostuff(nodeIP, nodename string, add bool) {
}

func main() {
	defer logs.FlushLogs()

	s := &syncer{
		added:   make(chan kapi.Node),
		removed: make(chan kapi.Node),
	}

	pflag.StringVar(&s.master, "master", "", "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	pflag.StringVar(&s.kubeconfig, "kubeconfig", "", "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	pflag.DurationVar(&s.ttl, "peer-ttl", 10*time.Second, "The TTL for this node change watcher")
	pflag.BoolVar(&s.FakeServer, "fake", false, "runs a fake server only for testings")

	flags.InitFlags()
	logs.InitLogs()
	flags.DumpAll()

	go s.SyncLoop()
	s.WatchNodes()
}
