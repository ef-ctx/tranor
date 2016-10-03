// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/tsuru/gnuflag"
	"github.com/tsuru/tsuru/api"
	"github.com/tsuru/tsuru/cmd"
	tsuruerrors "github.com/tsuru/tsuru/errors"
)

type projectConfigSet struct {
	projectName string
	envs        commaSeparatedFlag
	private     bool
	noRestart   bool
	fs          *gnuflag.FlagSet
}

func (c *projectConfigSet) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "config-set",
		Desc:    "defines configuration (environment variables) for a given project",
		Usage:   "config-set <NAME=value> [NAME=value]... <-n/--project-name projectname> [-p/--private] [--no-restart]",
		MinArgs: 1,
	}
}

func (c *projectConfigSet) Run(ctx *cmd.Context, client *cmd.Client) error {
	ctx.RawOutput()
	if c.projectName == "" {
		return errors.New("please provide the name of the project")
	}
	raw := strings.Join(ctx.Args, "\n")
	regex := regexp.MustCompile(`(\w+=[^\n]+)(\n|$)`)
	decls := regex.FindAllStringSubmatch(raw, -1)
	if len(decls) < 1 || len(decls) != len(ctx.Args) {
		return errors.New("configuration vars must be specified in the form NAME=value")
	}
	envVars := api.Envs{
		NoRestart: c.noRestart,
		Private:   c.private,
	}
	for _, decl := range decls {
		parts := strings.SplitN(decl[1], "=", 2)
		envVars.Envs = append(envVars.Envs, struct{ Name, Value string }{Name: parts[0], Value: parts[1]})
	}
	var cmdErr error
	for _, envName := range c.envs.Values() {
		appName := fmt.Sprintf("%s-%s", c.projectName, envName)
		fmt.Fprintf(ctx.Stdout, "setting config vars in environment %q... ", envName)
		err := setConfig(client, appName, &envVars)
		status := "ok"
		if err != nil {
			if e, ok := err.(*tsuruerrors.HTTP); ok && e.Code == http.StatusNotFound {
				status = "not found"
			} else {
				status = "failed"
				cmdErr = err
			}
		}
		fmt.Fprintln(ctx.Stdout, status)
	}
	return cmdErr
}

func (c *projectConfigSet) Flags() *gnuflag.FlagSet {
	if c.fs == nil {
		c.fs = gnuflag.NewFlagSet("config-set", gnuflag.ExitOnError)
		c.fs.StringVar(&c.projectName, "project-name", "", "name of the project")
		c.fs.StringVar(&c.projectName, "n", "", "name of the project")
		c.fs.Var(&c.envs, "envs", "comma-separated list of environments to set the configuration")
		c.fs.Var(&c.envs, "e", "comma-separated list of environments to set the configuration")
		c.fs.BoolVar(&c.private, "private", false, "set the configuration as private environment variables (not visible through command line)")
		c.fs.BoolVar(&c.private, "p", false, "set the configuration as private environment variables (not visible through command line)")
		c.fs.BoolVar(&c.noRestart, "no-restart", false, "set the configuration without restarting the application process")
	}
	return c.fs
}

type projectConfigGet struct {
	projectName string
	envs        commaSeparatedFlag
	fs          *gnuflag.FlagSet
}

func (c *projectConfigGet) Info() *cmd.Info {
	return &cmd.Info{
		Name: "config-get",
		Desc: "gets the configuration (environment variables) of the project in the given environments",
	}
}

func (c *projectConfigGet) Run(ctx *cmd.Context, client *cmd.Client) error {
	if c.projectName == "" {
		return errors.New("please provide the name of the project")
	}
	for _, envName := range c.envs.Values() {
		appName := fmt.Sprintf("%s-%s", c.projectName, envName)
		envVars, err := getConfig(client, appName)
		if err != nil {
			if e, ok := err.(*tsuruerrors.HTTP); ok && e.Code == http.StatusNotFound {
				fmt.Fprintf(ctx.Stderr, "WARNING: project not found in environment %q\n", envName)
				continue
			}
			return err
		}
		fmt.Fprintf(ctx.Stdout, "config vars in %q:\n\n", envName)
		for _, evar := range envVars {
			fmt.Fprintf(ctx.Stdout, " %s\n", &evar)
		}
		fmt.Fprint(ctx.Stdout, "\n\n")
	}
	return nil
}

func (c *projectConfigGet) Flags() *gnuflag.FlagSet {
	if c.fs == nil {
		c.fs = gnuflag.NewFlagSet("config-get", gnuflag.ExitOnError)
		c.fs.StringVar(&c.projectName, "project-name", "", "name of the project")
		c.fs.StringVar(&c.projectName, "n", "", "name of the project")
		c.fs.Var(&c.envs, "envs", "comma-separated list of environments to set the configuration")
		c.fs.Var(&c.envs, "e", "comma-separated list of environments to set the configuration")
	}
	return c.fs
}

type projectConfigUnset struct {
	projectName string
	noRestart   bool
	envs        commaSeparatedFlag
	fs          *gnuflag.FlagSet
}

func (c *projectConfigUnset) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "config-unset",
		Desc:    "undefine configuration params (environment variables) of the project in the given environments",
		MinArgs: 1,
	}
}

func (c *projectConfigUnset) Run(ctx *cmd.Context, client *cmd.Client) error {
	if c.projectName == "" {
		return errors.New("please provide the name of the project")
	}
	var cmdErr error
	for _, envName := range c.envs.Values() {
		appName := fmt.Sprintf("%s-%s", c.projectName, envName)
		fmt.Fprintf(ctx.Stdout, "unsetting config vars from environment %q... ", envName)
		err := unsetConfig(client, appName, c.noRestart, ctx.Args)
		status := "ok"
		if err != nil {
			if e, ok := err.(*tsuruerrors.HTTP); ok && e.Code == http.StatusNotFound {
				status = "not found"
			} else {
				status = "failed"
				cmdErr = err
			}
		}
		fmt.Fprintln(ctx.Stdout, status)
	}
	return cmdErr
}

func (c *projectConfigUnset) Flags() *gnuflag.FlagSet {
	if c.fs == nil {
		c.fs = gnuflag.NewFlagSet("config-get", gnuflag.ExitOnError)
		c.fs.StringVar(&c.projectName, "project-name", "", "name of the project")
		c.fs.StringVar(&c.projectName, "n", "", "name of the project")
		c.fs.Var(&c.envs, "envs", "comma-separated list of environments to set the configuration")
		c.fs.Var(&c.envs, "e", "comma-separated list of environments to set the configuration")
		c.fs.BoolVar(&c.noRestart, "no-restart", false, "unset configuration without restarting the application process")
	}
	return c.fs
}
