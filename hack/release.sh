#!/bin/bash
set -xeou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT="$GOPATH/src/github.com/appscode/swanc"

pushd $REPO_ROOT

rm -rf dist
./hack/docker/setup.sh
env APPSCODE_ENV=prod ./hack/docker/setup.sh release
rm dist/.tag

popd
