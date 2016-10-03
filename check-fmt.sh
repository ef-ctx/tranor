#!/bin/bash -e

# Copyright 2016 EF CTX. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

go get github.com/alecthomas/gometalinter honnef.co/go/unused/cmd/unused
gometalinter --install --vendored-linters
go install ./...
gometalinter -j 4 --enable=gofmt --enable=unused --disable=dupl --disable=errcheck --disable=gas --disable=interfacer --disable=gocyclo --deadline=10m --tests --vendor ./...
