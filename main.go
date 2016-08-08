// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"

	"github.com/tsuru/tsuru/cmd"
)

const version = "0.1"

var baseCommandsToRemove = []string{"target-add", "target-list", "target-remove", "target-set"}

func buildManager(name string) *cmd.Manager {
	mngr := cmd.BuildBaseManager(name, version, "", nil)
	for _, c := range baseCommandsToRemove {
		delete(mngr.Commands, c)
	}
	mngr.Register(envList{})
	mngr.Register(targetSet{})
	return mngr
}

func main() {
	name := cmd.ExtractProgramName(os.Args[0])
	buildManager(name).Run(os.Args[1:])
}