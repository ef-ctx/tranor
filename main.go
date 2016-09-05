// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"

	"github.com/tsuru/tsuru-client/tsuru/admin"
	"github.com/tsuru/tsuru-client/tsuru/client"
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
	mngr.Register(admin.PlatformList{})
	mngr.Register(&client.TeamList{})
	mngr.Register(&client.TeamCreate{})
	mngr.Register(&client.TeamRemove{})
	mngr.Register(&client.PlanList{})
	mngr.Register(&projectCreate{})
	mngr.Register(&projectRemove{})
	mngr.Register(&projectList{})
	return mngr
}

func main() {
	name := cmd.ExtractProgramName(os.Args[0])
	buildManager(name).Run(os.Args[1:])
}
