// Copyright 2017 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/tsuru/gnuflag"
	"github.com/tsuru/tsuru-client/tsuru/client"
	"github.com/tsuru/tsuru/cmd"
)

type projectLog struct {
	fs         *gnuflag.FlagSet
	name       string
	envName    string
	lines      int
	follow     bool
	omitDate   bool
	omitSource bool
}

func (c *projectLog) Info() *cmd.Info {
	return &cmd.Info{
		Name: "project-log",
		Desc: "Display logs of the given project and environment",
	}
}

func (c *projectLog) Run(ctx *cmd.Context, cli *cmd.Client) error {
	if c.name == "" || c.envName == "" {
		return errors.New("please provide the project name and the environment")
	}
	appName := fmt.Sprintf("%s-%s", c.name, c.envName)
	var appLog client.AppLog
	appLog.Flags().Parse(true, []string{
		"--app=" + appName,
		"--lines=" + strconv.Itoa(c.lines),
		"--follow=" + strconv.FormatBool(c.follow),
		"--no-date=" + strconv.FormatBool(c.omitDate),
		"--no-source=" + strconv.FormatBool(c.omitSource),
	})
	return appLog.Run(ctx, cli)
}

func (c *projectLog) Flags() *gnuflag.FlagSet {
	if c.fs == nil {
		c.fs = gnuflag.NewFlagSet("project-log", gnuflag.ExitOnError)
		c.fs.StringVar(&c.name, "name", "", "name of the project")
		c.fs.StringVar(&c.name, "n", "", "name of the project")
		c.fs.StringVar(&c.envName, "env", "", "name of the environment")
		c.fs.StringVar(&c.envName, "e", "", "name of the environment")
		c.fs.IntVar(&c.lines, "lines", 10, "number of log lines to display")
		c.fs.IntVar(&c.lines, "l", 10, "number of log lines to display")
		c.fs.BoolVar(&c.follow, "follow", false, "follow logs")
		c.fs.BoolVar(&c.follow, "f", false, "follow logs")
		c.fs.BoolVar(&c.omitDate, "no-date", false, "follow logs")
		c.fs.BoolVar(&c.omitSource, "no-source", false, "follow logs")
	}
	return c.fs
}
