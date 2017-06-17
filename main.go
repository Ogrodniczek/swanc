package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/appscode/go/flags"
	"github.com/appscode/log"
	logs "github.com/appscode/log/golog"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	clientset "k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type syncer struct {
	master     string
	kubeconfig string
	client     *clientset.Clientset
	ttl        time.Duration

	nodeName string
	nodeIP   string

	// This flag is only to write test driven code. Default value false.
	FakeServer bool

	reload chan struct{}
}

const nodeKey = "net.beta.appscode.com/vpn"

var nodeSelector = labels.SelectorFromSet(map[string]string{
	nodeKey: "true",
})

// Blocks caller. Intended to be called as a Go routine.
func (s *syncer) WatchNodes() {
	log.Info("started watching for peer endpoints")
	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return s.client.Nodes().List(metav1.ListOptions{
				LabelSelector: nodeSelector.String(),
			})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return s.client.Nodes().Watch(metav1.ListOptions{
				LabelSelector: nodeSelector.String(),
			})
		},
	}
	// kCachePopulated(k, events.Pod, &apiv1.Pod{}, nil)
	_, controller := cache.NewInformer(lw,
		&apiv1.Node{},
		s.ttl,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				log.Infoln("got one added node")
				s.reload <- struct{}{}
			},
			DeleteFunc: func(obj interface{}) {
				log.Infoln("got one deleted node", obj.(*apiv1.Node).Name)
				s.reload <- struct{}{}
			},
			UpdateFunc: func(old, new interface{}) {
				oldNode, ok := old.(*apiv1.Node)
				if !ok {
					return
				}
				newNode, ok := new.(*apiv1.Node)
				if !ok {
					return
				}
				if oldNode.Labels[nodeKey] != newNode.Labels[nodeKey] ||
					isNodeReady(oldNode) != isNodeReady(newNode) {
					log.Infoln("got one updated node", new.(*apiv1.Node).Name)
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

func (s *syncer) init() {
	d := filepath.Dir(confPath)
	if _, err := os.Stat(d); os.IsNotExist(err) {
		err = os.MkdirAll(d, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	config, err := clientcmd.BuildConfigFromFlags(s.master, s.kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	s.client, err = clientset.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
}

func isNodeReady(n *apiv1.Node) bool {
	for _, cond := range n.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "True" {
			return true
		}
	}
	return false
}

func (s *syncer) reloadVPN() {
	nodes, err := s.client.CoreV1().Nodes().List(metav1.ListOptions{
		LabelSelector: nodeSelector.String(),
	})
	if err != nil {
		log.Fatalln(err)
	}

	hasLabel := false

	nodeIPs := make([]string, len(nodes.Items))
	i := 0
	for _, node := range nodes.Items {
		if !isNodeReady(&node) {
			continue
		}

		var ip string
		for _, addr := range node.Status.Addresses {
			if addr.Type == apiv1.NodeInternalIP {
				ip = addr.Address
				break
			}
		}
		if ip == "" {
			for _, addr := range node.Status.Addresses {
				if addr.Type == apiv1.NodeExternalIP {
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
		node, err := s.client.CoreV1().Nodes().Get(s.nodeName, metav1.GetOptions{})
		if err != nil {
			log.Fatalln(err)
		}
		node.Labels["net.beta.appscode.com/vpn"] = "true"
		_, err = s.client.CoreV1().Nodes().Update(node)
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

		cmd := strings.Split(reloadCmd, " ")
		if err = exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
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

	flags.InitFlags()
	logs.InitLogs()
	flags.DumpAll()

	s.init()
	s.Validate()
	s.reloadVPN() // initial loading
	go s.SyncLoop()
	s.WatchNodes()
}
