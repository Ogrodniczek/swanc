package cmds

import (
	"os"
	"time"

	"github.com/appscode/go/log"
	"github.com/appscode/swanc/lib"
	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func NewCmdRun() *cobra.Command {
	var (
		masterURL      string
		kubeconfigPath string
		initOnly       bool
		opts           = lib.Options{
			NodeName:             os.Getenv("NODE_NAME"),
			PreferredAddressType: string(core.NodeInternalIP),
			ResyncPeriod:         5 * time.Minute,
			MaxNumRequeues:       5,
			// ref: https://github.com/kubernetes/ingress-nginx/blob/e4d53786e771cc6bdd55f180674b79f5b692e552/pkg/ingress/controller/launch.go#L252-L259
			// High enough QPS to fit all expected use cases. QPS=0 is not set here, because client code is overriding it.
			QPS: 1e6,
			// High enough Burst to fit all expected use cases. Burst=0 is not set here, because client code is overriding it.
			Burst: 1e6,
		}
	)
	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Run controller",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
			if err != nil {
				log.Fatalf("Could not get Kubernetes config: %s", err)
			}
			config.Burst = opts.Burst
			config.QPS = opts.QPS
			client := kubernetes.NewForConfigOrDie(config)

			ctrl := lib.New(client, opts)
			err = ctrl.Setup()
			if err != nil {
				log.Fatalln(err)
			}
			if initOnly {
				os.Exit(0)
			}

			// Now let's start the controller
			stop := make(chan struct{})
			defer close(stop)
			go ctrl.Run(1, stop)

			// Wait forever
			select {}
		},
	}

	cmd.Flags().StringVar(&masterURL, "master", masterURL, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	cmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", kubeconfigPath, "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	cmd.Flags().BoolVar(&initOnly, "init-only", initOnly, "If true, exits after initial config mount")

	cmd.Flags().StringVar(&opts.NodeName, "node-name", opts.NodeName, "Name used by kubernetes to identify host")
	cmd.Flags().StringVar(&opts.PreferredAddressType, "preferred-address-type", opts.PreferredAddressType, "Preferred address type used for inter-node communication")
	cmd.Flags().Float32Var(&opts.QPS, "qps", opts.QPS, "The maximum QPS to the master from this client")
	cmd.Flags().IntVar(&opts.Burst, "burst", opts.Burst, "The maximum burst for throttle")
	cmd.Flags().DurationVar(&opts.ResyncPeriod, "resync-period", opts.ResyncPeriod, "If non-zero, will re-list this often. Otherwise, re-list will be delayed aslong as possible (until the upstream source closes the watch or times out.")

	return cmd
}
