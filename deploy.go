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
	promoteFrom string
	image       string
}

func (c *projectDeploy) Info() *cmd.Info {
	return &cmd.Info{
		Name:  "project-deploy",
		Usage: "tranor project-deploy -n/--project-name <projectname> -e/--env <environment> [-i/--image dockerimage] [-p/--promote parent-env] [content]",
		Desc: `deploys a new version of a project. Also used to promote a version from one environment to another

Can deploy the project using one of the following strategies:

 - Content upload: just provide the list of files/directories to deploy as argument
 - Docker image: use the flags -i/--image
 - Promoting from other environment: use the flags -p/--promote
`,
	}
}

func (c *projectDeploy) Run(ctx *cmd.Context, cli *cmd.Client) error {
	ctx.RawOutput()
	if c.projectName == "" || c.envName == "" {
		return errors.New("please provide the project name and the environment")
	}
	// fail soon if project doesn't exist
	apps, err := projectApps(cli, c.projectName)
	if err != nil {
		return err
	}

	appName := fmt.Sprintf("%s-%s", c.projectName, c.envName)
	flags := []string{"-a", appName}
	checkEnv := true
	if len(ctx.Args) > 0 {
		if c.image != "" || c.promoteFrom != "" {
			return errors.New("please specify only one of the image, parent env or the list of files/directories to upload")
		}
	} else if c.image != "" {
		if c.promoteFrom != "" {
			return errors.New("please specify only one of the image, parent env or the list of files/directories to upload")
		}
		flags = append(flags, "-i", c.image)
	} else if c.promoteFrom != "" {
		promoteFlags, err := c.promoteFlags(c.projectName, c.promoteFrom, cli)
		if err != nil {
			return err
		}
		flags = append(flags, promoteFlags...)
		checkEnv = false
	} else {
		return errors.New("please specify either the image, parent env or the list of files/directories to upload")
	}

	if checkEnv && apps[0].Env.Name != c.envName {
		return fmt.Errorf("can only deploy directly to %q, use promote to deploy to other environments", apps[0].Env.Name)
	}
	tsuruDeployCommand.Flags().Parse(true, flags)
	return tsuruDeployCommand.Run(ctx, cli)
}

func (c *projectDeploy) promoteFlags(projectName, fromEnv string, cli *cmd.Client) ([]string, error) {
	config, _ := loadConfigFile()
	originApp := fmt.Sprintf("%s-%s", projectName, fromEnv)
	d, err := lastDeploy(cli, originApp)
	if err != nil {
		return nil, err
	}
	if d.Image == "" {
		return nil, fmt.Errorf("no version running in %q", fromEnv)
	}
	return []string{"-i", config.imageApp(originApp, d.Image)}, nil
}

func (c *projectDeploy) Flags() *gnuflag.FlagSet {
	if c.fs == nil {
		c.fs = gnuflag.NewFlagSet("project-deploy", gnuflag.ExitOnError)
		c.fs.StringVar(&c.projectName, "project-name", "", "name of the project to deploy to")
		c.fs.StringVar(&c.projectName, "n", "", "name of the project to deploy to")
		c.fs.StringVar(&c.envName, "env", "", "environment to deploy to")
		c.fs.StringVar(&c.envName, "e", "", "environment to deploy to")
		c.fs.StringVar(&c.promoteFrom, "promote", "", "promote version from the given environment")
		c.fs.StringVar(&c.promoteFrom, "p", "", "promote version from the given environment")
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
