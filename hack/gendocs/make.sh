#!/usr/bin/env bash

pushd $GOPATH/src/github.com/pharmer/swanc/hack/gendocs
go run main.go
popd
