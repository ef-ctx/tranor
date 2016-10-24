// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"

	"github.com/tsuru/gnuflag"
	"github.com/tsuru/tsuru-client/tsuru/client"
	"github.com/tsuru/tsuru/cmd"
)

var (
	tsuruDeployCommand     cmd.FlaggedCommand = &client.AppDeploy{}
	tsuruDeployListCommand cmd.FlaggedCommand = &client.AppDeployList{}
)

type projectDeploy struct {
	fs          *gnuflag.FlagSet
	projectName string
	envName     string
	version     string
	image       string
}

func (c *projectDeploy) Info() *cmd.Info {
	return &cmd.Info{
		Name:  "project-deploy",
		Usage: "tranor project-deploy -n/--project-name <projectname> -e/--env <environment> [-i/--image dockerimage] [-v/version vNN] [content]",
		Desc: `deploys a new version of a project. Also used to promote a version from one environment to another

Can deploy the project using one of the following strategies:

 - Content upload: just provide the list of files/directories to deploy as argument
 - Docker image: use the flags -i/--image
 - Version (also used for rollback): use the flags -v/--version
`,
	}
}

func (c *projectDeploy) Run(ctx *cmd.Context, cli *cmd.Client) error {
	ctx.RawOutput()
	if c.projectName == "" || c.envName == "" {
		return errors.New("please provide the project name and the environment")
	}
	appName := fmt.Sprintf("%s-%s", c.projectName, c.envName)
	flags := []string{"-a", appName}
	if len(ctx.Args) > 0 {
		if c.image != "" || c.version != "" {
			return errors.New("please specify only one of the image, version or the list of files/directories to upload")
		}
	} else if c.image != "" {
		if c.version != "" {
			return errors.New("please specify only one of the image, version or the list of files/directories to upload")
		}
		flags = append(flags, "-i", c.image)
	} else if c.version != "" {
		// TODO(fss): implement version-based deployment.
	} else {
		return errors.New("please specify either the image, version or the list of files/directories to upload")
	}
	err := tsuruDeployCommand.Flags().Parse(true, flags)
	if err != nil {
		return err
	}
	return tsuruDeployCommand.Run(ctx, cli)
}

func (c *projectDeploy) Flags() *gnuflag.FlagSet {
	if c.fs == nil {
		c.fs = gnuflag.NewFlagSet("project-deploy", gnuflag.ExitOnError)
		c.fs.StringVar(&c.projectName, "project-name", "", "name of the project to deploy to")
		c.fs.StringVar(&c.projectName, "n", "", "name of the project to deploy to")
		c.fs.StringVar(&c.envName, "env", "", "environment to deploy to")
		c.fs.StringVar(&c.envName, "e", "", "environment to deploy to")
		c.fs.StringVar(&c.version, "version", "", "version to deploy (when promoting from one environment to another")
		c.fs.StringVar(&c.version, "v", "", "version to deploy (when promoting from one environment to another")
		c.fs.StringVar(&c.image, "image", "", "Docker image to deploy")
		c.fs.StringVar(&c.image, "i", "", "Docker image to deploy")
	}
	return c.fs
}

type projectDeployList struct {
	fs          *gnuflag.FlagSet
	projectName string
	envName     string
}

func (c *projectDeployList) Info() *cmd.Info {
	return &cmd.Info{
		Name: "project-deploy-list",
		Desc: "lists all deployments of the project in the given environment",
	}
}

func (c *projectDeployList) Run(ctx *cmd.Context, cli *cmd.Client) error {
	if c.projectName == "" || c.envName == "" {
		return errors.New("please provide the project name and the environment")
	}
	appName := fmt.Sprintf("%s-%s", c.projectName, c.envName)
	tsuruDeployListCommand.Flags().Parse(true, []string{"-a", appName})
	return tsuruDeployListCommand.Run(ctx, cli)
}

func (c *projectDeployList) Flags() *gnuflag.FlagSet {
	if c.fs == nil {
		c.fs = gnuflag.NewFlagSet("project-deploy", gnuflag.ExitOnError)
		c.fs.StringVar(&c.projectName, "project-name", "", "name of the project to deploy to")
		c.fs.StringVar(&c.projectName, "n", "", "name of the project to deploy to")
		c.fs.StringVar(&c.envName, "env", "", "environment to deploy to")
		c.fs.StringVar(&c.envName, "e", "", "environment to deploy to")
	}
	return c.fs
}
