#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GOPATH=$(go env GOPATH)
SRC=$GOPATH/src
BIN=$GOPATH/bin
ROOT=$GOPATH
REPO_ROOT=$GOPATH/src/github.com/pharmer/swanc

source "$REPO_ROOT/hack/libbuild/common/pharmer_image.sh"

APPSCODE_ENV=${APPSCODE_ENV:-dev}
IMG=swanc

DIST=$GOPATH/src/github.com/pharmer/swanc/dist
mkdir -p $DIST
if [ -f "$DIST/.tag" ]; then
	export $(cat $DIST/.tag | xargs)
fi

clean() {
    pushd $REPO_ROOT/hack/docker
    rm -f swanc
    popd
}

build_binary() {
    pushd $REPO_ROOT
    ./hack/builddeps.sh
    ./hack/make.py build swanc
    detect_tag $DIST/.tag
    popd
}

build_docker() {
    pushd $REPO_ROOT/hack/docker
    cp $DIST/swanc/swanc-linux-amd64 swanc
    chmod 755 swanc

    local cmd="docker build -t pharmer/$IMG:$TAG ."
    echo $cmd; $cmd

    rm swanc
    popd
}

build() {
    build_binary
    build_docker
}

docker_push() {
    if [ "$APPSCODE_ENV" = "prod" ]; then
        echo "Nothing to do in prod env. Are you trying to 'release' binaries to prod?"
        exit 1
    fi
    if [ "$TAG_STRATEGY" = "git_tag" ]; then
        echo "Are you trying to 'release' binaries to prod?"
        exit 1
    fi
    hub_canary
}

docker_release() {
    if [ "$APPSCODE_ENV" != "prod" ]; then
        echo "'release' only works in PROD env."
        exit 1
    fi
    if [ "$TAG_STRATEGY" != "git_tag" ]; then
        echo "'apply_tag' to release binaries and/or docker images."
        exit 1
    fi
    hub_up
}

source_repo $@
