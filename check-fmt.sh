#!/bin/bash -e

# Copyright 2016 EF CTX. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

pkgs=$(go list ./... | grep -v vendor | sed 's;github.com/ef-ctx/tranor;.;')

test -z "$(gofmt -s -l $pkgs | grep -v vendor/ | tee /dev/stderr)"

go vet $pkgs
go get github.com/golang/lint/golint
golint $pkgs

go get github.com/remyoudompheng/go-misc/deadcode
deadcode $pkgs
