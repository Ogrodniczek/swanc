package lib

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"sync/atomic"
	"time"

	ioutilz "github.com/appscode/go/ioutil"
	"github.com/appscode/go/log"
	core_util "github.com/appscode/kutil/core/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	core_listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	k8sClient kubernetes.Interface
	opts      Options
	recorder  record.EventRecorder
	writer    *ioutilz.AtomicWriter

	// Node
	nQueue    workqueue.RateLimitingInterface
	nIndexer  cache.Indexer
	nInformer cache.Controller
	nLister   core_listers.NodeLister
}

func New(client kubernetes.Interface, opts Options) *Controller {
	return &Controller{
		k8sClient: client,
		opts:      opts,
	}
}

func (c *Controller) Setup() error {
	if c.opts.NodeName == "" {
		return errors.New("set NODE_NAME environment variable to `nodeName`")
	}
	if c.opts.PreferredAddressType != string(core.NodeInternalIP) &&
		c.opts.PreferredAddressType != string(core.NodeExternalIP) {
		return fmt.Errorf("--preferred-address-type must be set to either %s or %s", core.NodeInternalIP, core.NodeExternalIP)
	}

	var err error
	c.writer, err = ioutilz.NewAtomicWriter(confDir)
	if err != nil {
		return err
	}

	c.initNodeWatcher()
	err = c.initNodeCache()
	if err != nil {
		return err
	}
	return c.mount(false)
}

func (c *Controller) initNodeCache() error {
	nodes, err := c.k8sClient.CoreV1().Nodes().List(metav1.ListOptions{
		LabelSelector: labels.Everything().String(),
	})
	if err != nil {
		return err
	}
	for i := range nodes.Items {
		c.nIndexer.Add(&nodes.Items[i])
	}
	return nil
}

func (c *Controller) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.nQueue.ShutDown()
	log.Info("Starting Stash controller")

	go c.nInformer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.nInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runNodeWatcher, time.Second, stopCh)
	}

	<-stopCh
	log.Info("Stopping Stash controller")
}

func (c *Controller) getPreferredAddress(node *core.Node) (string, bool) {
	for _, addr := range node.Status.Addresses {
		if addr.Type == core.NodeAddressType(c.opts.PreferredAddressType) {
			return addr.Address, true
		}
	}
	return "", false
}

func (c *Controller) isAnnotated(node *core.Node) bool {
	if node == nil || node.Annotations == nil {
		return false
	}
	_, found := node.Annotations[nodeKey]
	return found
}

func (c *Controller) mount(reload bool) error {
	nodes, err := c.nLister.List(nodeSelector)
	if err != nil {
		return err
	}

	var (
		td      TemplateData
		curNode *core.Node
	)
	for i := range nodes {
		node := nodes[i]
		if core_util.IsMaster(*node) {
			continue
		}
		if node.Name == c.opts.NodeName {
			if addr, found := c.getPreferredAddress(node); found {
				td.HostIP = addr
				curNode = node
			}
		} else {
			if addr, found := c.getPreferredAddress(node); found {
				td.NodeIPs = append(td.NodeIPs, addr)
			}
		}
	}
	sort.Strings(td.NodeIPs)

	if curNode != nil && !c.isAnnotated(curNode) {
		_, err = core_util.PatchNode(c.k8sClient, curNode, func(in *core.Node) *core.Node {
			if in.Annotations == nil {
				in.Annotations = map[string]string{}
			}
			in.Annotations[nodeKey] = ""
			return in
		})
		if err != nil {
			return err
		}
	}

	// Should generate the default ipsec.conf at least.
	var buf bytes.Buffer
	err = cfgTemplate.Execute(&buf, td)
	if err != nil {
		return err
	}
	changed, err := c.writer.Write(map[string]ioutilz.FileProjection{
		confFile: {Mode: 0644, Data: buf.Bytes()},
	})
	if err != nil {
		return err
	}
	if changed && reload {
		return runCmd()
	}
	return nil
}

var mountPerformed uint64

func incMountCounter() {
	atomic.AddUint64(&mountPerformed, 1)
	log.Infoln("Mount Performed:", atomic.LoadUint64(&mountPerformed))
}

func runCmd() error {
	if err := exec.Command("/usr/sbin/ipsec", "update").Run(); err != nil {
		return err
	}
	incMountCounter()
	return nil
}
