package cmds

import (
	"os"
	"time"

	"github.com/appscode/swanc/lib"
	"github.com/spf13/cobra"
)

func NewCmdRun() *cobra.Command {
	op := &lib.Operator{
		Reload: make(chan struct{}),
	}
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run operator",
		Run: func(cmd *cobra.Command, args []string) {
			op.Init()
			op.Validate()
			op.ReloadVPN() // initial loading
			go op.SyncLoop()
			op.WatchNodes()
		},
	}

	cmd.Flags().StringVar(&op.Master, "master", "", "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	cmd.Flags().StringVar(&op.Kubeconfig, "kubeconfig", "", "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	cmd.Flags().DurationVar(&op.Ttl, "peer-ttl", 10*time.Second, "The TTL for this node change watcher")

	cmd.Flags().StringVar(&op.NodeName, "node-name", os.Getenv("NODE_NAME"), "Name used by kubernetes to identify host")
	cmd.Flags().StringVar(&op.NodeIP, "node-ip", os.Getenv("NODE_IP"), "IP used by host for intra-cluster communication")

	return cmd
}
