package lib

import (
	"os"
	"testing"
)

func TestIPSecConf(t *testing.T) {
	hostIP := "10.0.0.1"
	nodeIPs := []string{
		"10.0.0.2",
		"10.0.0.3",
	}

	err := cfgTemplate.Execute(os.Stdout, struct {
		HostIP  string
		NodeIPs []string
	}{
		hostIP,
		nodeIPs,
	})
	if err != nil {
		t.Fatal(err)
	}
}
