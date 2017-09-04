#!/usr/bin/env bash

pushd $GOPATH/src/github.com/appscode/swanc/hack/gendocs
go run main.go
popd
