## swanc run

Run controller

### Synopsis


Run controller

```
swanc run [flags]
```

### Options

```
      --burst int                       The maximum burst for throttle (default 1000000)
  -h, --help                            help for run
      --init-only                       If true, exits after initial config mount
      --kubeconfig string               Path to kubeconfig file with authorization information (the master location is set by the master flag).
      --master string                   The address of the Kubernetes API server (overrides any value in kubeconfig)
      --node-name string                Name used by kubernetes to identify host
      --preferred-address-type string   Preferred address type used for inter-node communication (default "InternalIP")
      --qps float32                     The maximum QPS to the master from this client (default 1e+06)
      --resync-period duration          If non-zero, will re-list this often. Otherwise, re-list will be delayed aslong as possible (until the upstream source closes the watch or times out. (default 5m0s)
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --analytics                        Send analytical events to Google Analytics (default true)
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO
* [swanc](swanc.md)	 - Swanc - StrongSwan based VPN Controller for Kubernetes by AppsCode

