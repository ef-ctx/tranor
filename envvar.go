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

type projectEnvVarSet struct {
	projectName string
	envs        commaSeparatedFlag
	private     bool
	noRestart   bool
	fs          *gnuflag.FlagSet
}

func (c *projectEnvVarSet) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "envvar-set",
		Desc:    "defines environment variables for a given project",
		Usage:   "envvar-set <NAME=value> [NAME=value]... <-n/--project-name projectname> [-p/--private] [--no-restart]",
		MinArgs: 1,
	}
}

func (c *projectEnvVarSet) Run(ctx *cmd.Context, client *cmd.Client) error {
	ctx.RawOutput()
	if c.projectName == "" {
		return errors.New("please provide the name of the project")
	}
	config, err := loadConfigFile()
	if err != nil {
		return errors.New("unable to load environments file, please make sure that tranor is properly configured")
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
	envNames := c.envs.Values()
	if len(envNames) == 0 {
		envNames = config.envNames()
	}
	var cmdErr error
	for _, envName := range envNames {
		appName := fmt.Sprintf("%s-%s", c.projectName, envName)
		fmt.Fprintf(ctx.Stdout, "setting variables in environment %q... ", envName)
		err := setEnvVars(client, appName, &envVars)
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

func (c *projectEnvVarSet) Flags() *gnuflag.FlagSet {
	if c.fs == nil {
		c.fs = gnuflag.NewFlagSet("envvar-set", gnuflag.ExitOnError)
		c.fs.StringVar(&c.projectName, "project-name", "", "name of the project")
		c.fs.StringVar(&c.projectName, "n", "", "name of the project")
		c.fs.Var(&c.envs, "envs", "comma-separated list of environments to set the variables")
		c.fs.Var(&c.envs, "e", "comma-separated list of environments to set the variables")
		c.fs.BoolVar(&c.private, "private", false, "set the variables to private (not visible through command line)")
		c.fs.BoolVar(&c.private, "p", false, "set the variables to private (not visible through command line)")
		c.fs.BoolVar(&c.noRestart, "no-restart", false, "set the environment variables without restarting the application process")
	}
	return c.fs
}

type projectEnvVarGet struct {
	projectName string
	envs        commaSeparatedFlag
	fs          *gnuflag.FlagSet
}

func (c *projectEnvVarGet) Info() *cmd.Info {
	return &cmd.Info{
		Name: "envvar-get",
		Desc: "gets environment variables of the project in the given environments",
	}
}

func (c *projectEnvVarGet) Run(ctx *cmd.Context, client *cmd.Client) error {
	if c.projectName == "" {
		return errors.New("please provide the name of the project")
	}
	config, err := loadConfigFile()
	if err != nil {
		return errors.New("unable to load environments file, please make sure that tranor is properly configured")
	}
	envNames := c.envs.Values()
	if len(envNames) == 0 {
		envNames = config.envNames()
	}
	for _, envName := range envNames {
		appName := fmt.Sprintf("%s-%s", c.projectName, envName)
		envVars, err := getEnvVars(client, appName)
		if err != nil {
			if e, ok := err.(*tsuruerrors.HTTP); ok && e.Code == http.StatusNotFound {
				fmt.Fprintf(ctx.Stderr, "WARNING: project not found in environment %q\n", envName)
				continue
			}
			return err
		}
		fmt.Fprintf(ctx.Stdout, "variables in %q:\n\n", envName)
		for _, evar := range envVars {
			fmt.Fprintf(ctx.Stdout, " %s\n", &evar)
		}
		fmt.Fprint(ctx.Stdout, "\n\n")
	}
	return nil
}

func (c *projectEnvVarGet) Flags() *gnuflag.FlagSet {
	if c.fs == nil {
		c.fs = gnuflag.NewFlagSet("envvar-get", gnuflag.ExitOnError)
		c.fs.StringVar(&c.projectName, "project-name", "", "name of the project")
		c.fs.StringVar(&c.projectName, "n", "", "name of the project")
		c.fs.Var(&c.envs, "envs", "comma-separated list of environments to get the variables")
		c.fs.Var(&c.envs, "e", "comma-separated list of environments to get the variables")
	}
	return c.fs
}

type projectEnvVarUnset struct {
	projectName string
	noRestart   bool
	envs        commaSeparatedFlag
	fs          *gnuflag.FlagSet
}

func (c *projectEnvVarUnset) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "envvar-unset",
		Desc:    "unset environment variables of the project in the given environments",
		MinArgs: 1,
	}
}

func (c *projectEnvVarUnset) Run(ctx *cmd.Context, client *cmd.Client) error {
	if c.projectName == "" {
		return errors.New("please provide the name of the project")
	}
	var cmdErr error
	for _, envName := range c.envs.Values() {
		appName := fmt.Sprintf("%s-%s", c.projectName, envName)
		fmt.Fprintf(ctx.Stdout, "unsetting variables from environment %q... ", envName)
		err := unsetEnvVars(client, appName, c.noRestart, ctx.Args)
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

func (c *projectEnvVarUnset) Flags() *gnuflag.FlagSet {
	if c.fs == nil {
		c.fs = gnuflag.NewFlagSet("envvar-get", gnuflag.ExitOnError)
		c.fs.StringVar(&c.projectName, "project-name", "", "name of the project")
		c.fs.StringVar(&c.projectName, "n", "", "name of the project")
		c.fs.Var(&c.envs, "envs", "comma-separated list of environments to set the variables")
		c.fs.Var(&c.envs, "e", "comma-separated list of environments to set the variables")
		c.fs.BoolVar(&c.noRestart, "no-restart", false, "unset environment variables without restarting the application process")
	}
	return c.fs
}
