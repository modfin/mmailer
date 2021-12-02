#!/usr/bin/env bash

export NAME=mmailer
export MAIN_PATH=mmailer/cmd/mmailerd/mmailerd.go
export BINARY_PATH=/mmailerd

# change directory to go/src/mfn (build script requires it)
cd ..

../../docker/build-push.base.sh "$@"
