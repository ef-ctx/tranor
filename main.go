// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"

	"github.com/tsuru/tsuru/cmd"
)

const version = "0.1"

func buildManager(name string) *cmd.Manager {
	mngr := cmd.BuildBaseManager(name, version, "", nil)
	mngr.Register(envList{})
	return mngr
}

func main() {
	name := cmd.ExtractProgramName(os.Args[0])
	buildManager(name).Run(os.Args[1:])
}
