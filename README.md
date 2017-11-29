[![Go Report Card](https://goreportcard.com/badge/github.com/appscode/swanc)](https://goreportcard.com/report/github.com/appscode/swanc)

# SwanC
StrongSwan based VPN Controller for Kubernetes. This is intended to be used with [Pharmer by AppsCode](https://appscode.com/products/pharmer).

## Supported Versions
Kubernetes 1.8+

## Installation
To install Swanc, please follow the guide [here](/docs/install.md).

## Contribution guidelines
Want to help improve Swanc? Please start [here](/CONTRIBUTING.md).

## Versioning Policy
Swanc __does not follow semver__, rather the _major_ version of operator points to the
Kubernetes [client-go](https://github.com/kubernetes/client-go#branches-and-tags) version. You can verify this
from the `glide.yaml` file. This means there might be breaking changes between point releases of the operator.

## Acknowledgement
 - [Strongswan project](https://www.strongswan.org/)

## Support
If you have any questions, [file an issue](https://github.com/appscode/pharmer/issues/new) or talk to us on the [Kubernetes Slack team](http://slack.kubernetes.io/) channel `#pharmer`.

---

**The swanc operator collects anonymous usage statistics to help us learn
how the software is being used and how we can improve it. To disable stats collection,
run the operator with the flag** `--analytics=false`.

---
