[![Go Report Card](https://goreportcard.com/badge/github.com/pharmer/swanc)](https://goreportcard.com/report/github.com/pharmer/swanc)
[![Build Status](https://travis-ci.org/pharmer/swanc.svg?branch=master)](https://travis-ci.org/pharmer/swanc)
[![codecov](https://codecov.io/gh/pharmer/swanc/branch/master/graph/badge.svg)](https://codecov.io/gh/pharmer/swanc)
[![Docker Pulls](https://img.shields.io/docker/pulls/pharmer/swanc.svg)](https://hub.docker.com/r/pharmer/swanc/)
[![Slack](http://slack.kubernetes.io/badge.svg)](http://slack.kubernetes.io)
[![Twitter](https://img.shields.io/twitter/follow/appscodehq.svg?style=social&logo=twitter&label=Follow)](https://twitter.com/intent/follow?screen_name=AppsCodeHQ)

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
