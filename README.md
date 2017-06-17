[![Go Report Card](https://goreportcard.com/badge/github.com/appscode/swanc)](https://goreportcard.com/report/github.com/appscode/swanc)

[Website](https://appscode.com) • [Slack](https://slack.appscode.com) • [Forum](https://discuss.appscode.com) • [Twitter](https://twitter.com/AppsCodeHQ)

# SwanC
StrongSwan based VPN Controller for Kubernetes

## Versioning Policy
Swanc __does not follow semver__, rather the _major_ version of operator points to the
Kubernetes [client-go](https://github.com/kubernetes/client-go#branches-and-tags) version. You can verify this
from the `glide.yaml` file. This means there might be breaking changes between point releases of the operator.

---

**The swanc operator collects anonymous usage statistics to help us learn
how the software is being used and how we can improve it. To disable stats collection,
run the operator with the flag** `--analytics=false`.

---

## Acknowledgement
 - strongswan https://www.strongswan.org/

## Support
If you have any questions, you can reach out to us.
* [Slack](https://slack.appscode.com)
* [Forum](https://discuss.appscode.com)
* [Twitter](https://twitter.com/AppsCodeHQ)
* [Website](https://appscode.com)
