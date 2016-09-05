// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/tsuru/gnuflag"
	"github.com/tsuru/tsuru/cmd"
	tsuruerrors "github.com/tsuru/tsuru/errors"
)

type projectCreate struct {
	fs       *gnuflag.FlagSet
	name     string
	platform string
	team     string
	plan     string
	envs     commaSeparatedFlag
}

func (*projectCreate) Info() *cmd.Info {
	return &cmd.Info{
		Name: "project-create",
		Desc: "creates a remote project in the tranor server",
	}
}

func (c *projectCreate) Run(ctx *cmd.Context, client *cmd.Client) error {
	if c.name == "" || c.platform == "" {
		return errors.New("please provide the name and the platform")
	}
	config, err := loadConfigFile()
	if err != nil {
		return errors.New("unable to load environments file, please make sure that tranor is properly configured")
	}
	err = c.envs.validate(config.envNames())
	if err != nil {
		return fmt.Errorf("failed to load environments: %s", err)
	}
	envs := c.filterEnvironments(config.Environments, c.envs.Values())
	apps, err := c.createApps(envs, client)
	if err != nil {
		return err
	}
	err = c.setCNames(apps, client)
	if err != nil {
		appNames := make([]string, len(apps))
		for i, app := range apps {
			appNames[i] = app["name"]
		}
		deleteApps(appNames, client)
		return fmt.Errorf("failed to configure project %q: %s", c.name, err)
	}
	fmt.Fprintf(ctx.Stdout, "successfully created the project %q!\n", c.name)
	if gitRepo := apps[0]["repository_url"]; gitRepo != "" {
		fmt.Fprintf(ctx.Stdout, "Git repository: %s\n", gitRepo)
	}
	return nil
}

func (c *projectCreate) Flags() *gnuflag.FlagSet {
	if c.fs == nil {
		c.fs = gnuflag.NewFlagSet("project-create", gnuflag.ExitOnError)
		c.fs.StringVar(&c.name, "name", "", "name of the project")
		c.fs.StringVar(&c.name, "n", "", "name of the project")
		c.fs.StringVar(&c.platform, "platform", "", "platform of the project")
		c.fs.StringVar(&c.platform, "l", "", "platform of the project")
		c.fs.StringVar(&c.team, "team", "", "team that owns the project")
		c.fs.StringVar(&c.team, "t", "", "team that owns the project")
		c.fs.StringVar(&c.plan, "plan", "", "plan to use for the project")
		c.fs.StringVar(&c.plan, "p", "", "plan to use for the project")
		c.fs.Var(&c.envs, "envs", "comma-separated list of environments to use (defaults to dev,qa,stage,production)")
		c.fs.Var(&c.envs, "e", "comma-separated list of environments to use (defaults to env,qa,stage,production)")
		c.envs.Set("dev,qa,stage,production")
	}
	return c.fs
}

func (c *projectCreate) createApps(envs []Environment, client *cmd.Client) ([]map[string]string, error) {
	createdApps := make([]map[string]string, 0, len(envs))
	appNames := make([]string, 0, len(envs))
	for _, env := range envs {
		appName := fmt.Sprintf("%s-%s", c.name, env.Name)
		app, err := createApp(client, createAppOptions{
			name:     appName,
			plan:     c.plan,
			platform: c.platform,
			pool:     env.poolName(),
			team:     c.team,
		})
		if err != nil {
			deleteApps(appNames, client)
			return nil, fmt.Errorf("failed to create the project in env %q: %s", env.Name, err)
		}
		app["name"] = appName
		app["dnsSuffix"] = env.DNSSuffix
		createdApps = append(createdApps, app)
		appNames = append(appNames, appName)
	}
	return createdApps, nil
}

func (c *projectCreate) setCNames(apps []map[string]string, client *cmd.Client) error {
	for _, app := range apps {
		reqURL, err := cmd.GetURL(fmt.Sprintf("/apps/%s/cname", app["name"]))
		if err != nil {
			return err
		}
		cname := fmt.Sprintf("%s.%s", c.name, app["dnsSuffix"])
		v := make(url.Values)
		v.Set("cname", cname)
		req, err := http.NewRequest("POST", reqURL, strings.NewReader(v.Encode()))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()
	}
	return nil
}

func (c *projectCreate) filterEnvironments(envs []Environment, names []string) []Environment {
	var filtered []Environment
	for _, e := range envs {
		for _, name := range names {
			if e.Name == name {
				filtered = append(filtered, e)
				break
			}
		}
	}
	return filtered
}

type commaSeparatedFlag struct {
	values []string
}

func (f *commaSeparatedFlag) Values() []string {
	return f.values
}

func (f *commaSeparatedFlag) String() string {
	return strings.Join(f.values, ",")
}

func (f *commaSeparatedFlag) Set(v string) error {
	f.values = strings.Split(v, ",")
	return nil
}

func (f *commaSeparatedFlag) validate(validValues []string) error {
	var invalidValues []string
	for _, cv := range f.values {
		var found bool
		for _, validValue := range validValues {
			if validValue == cv {
				found = true
				break
			}
		}
		if !found {
			invalidValues = append(invalidValues, cv)
		}
	}
	if len(invalidValues) > 0 {
		return fmt.Errorf("invalid values: %s (valid options are: %s)", strings.Join(invalidValues, ", "), strings.Join(validValues, ", "))
	}
	return nil
}

type projectRemove struct {
	cmd.ConfirmationCommand
	name string
	fs   *gnuflag.FlagSet
}

func (c *projectRemove) Info() *cmd.Info {
	return &cmd.Info{
		Name: "project-remove",
		Desc: "removes the given project",
	}
}

func (c *projectRemove) Run(ctx *cmd.Context, client *cmd.Client) error {
	if c.name == "" {
		return errors.New("please provide the name of the project")
	}
	config, err := loadConfigFile()
	if err != nil {
		return errors.New("unable to load environments file, please make sure that tranor is properly configured")
	}
	if !c.Confirm(ctx, fmt.Sprintf("Are you sure you want to remove the project %q?", c.name)) {
		return nil
	}
	envNames := config.envNames()
	appNames := make([]string, len(envNames))
	for i, envName := range envNames {
		appNames[i] = fmt.Sprintf("%s-%s", c.name, envName)
	}
	errs, err := deleteApps(appNames, client)
	if err != nil {
		return err
	}
	var notFound int
	for _, err := range errs {
		if err != nil {
			if e, ok := err.(*tsuruerrors.HTTP); ok && e.Code == http.StatusNotFound {
				notFound++
			}
		}
	}
	if notFound == len(errs) {
		return errors.New("project not found")
	}
	fmt.Fprintln(ctx.Stdout, "Project successfully removed!")
	return nil
}

func (c *projectRemove) Flags() *gnuflag.FlagSet {
	if c.fs == nil {
		c.fs = c.ConfirmationCommand.Flags()
		c.fs.StringVar(&c.name, "name", "", "name of the project to b remove")
		c.fs.StringVar(&c.name, "n", "", "name of the project to remove")
	}
	return c.fs
}

type projectInfo struct {
	name string
	fs   *gnuflag.FlagSet
}

func (c *projectInfo) Info() *cmd.Info {
	return &cmd.Info{
		Name: "project-info",
	}
}

func (c *projectInfo) Run(ctx *cmd.Context, client *cmd.Client) error {
	if c.name == "" {
		return errors.New("please provide the name of the project")
	}
	apps, err := c.projectApps(client)
	if err != nil {
		return err
	}
	fmt.Fprintf(ctx.Stdout, "Project name: %s\n\n", c.name)
	var envs cmd.Table
	envs.Headers = cmd.Row{"Environment", "Address", "Version", "Deploy date"}
	for _, app := range apps {
		row := cmd.Row{app.Env.Name, app.Addr, "", ""}
		if deploy, err := lastDeploy(client, app.Name); err == nil && deploy.Image != "" {
			row[2] = deploy.Image
			row[3] = deploy.Timestamp.Format(time.RFC1123)
		}
		envs.AddRow(row)
	}
	ctx.Stdout.Write(envs.Bytes())
	return nil
}

func (c *projectInfo) projectApps(client *cmd.Client) ([]app, error) {
	config, err := loadConfigFile()
	if err != nil {
		return nil, errors.New("unable to load environments file, please make sure that tranor is properly configured")
	}
	apps, err := listApps(client, map[string]string{"name": c.name})
	if err != nil {
		return nil, err
	}
	var projectApps []app
	for _, env := range config.Environments {
		for _, app := range apps {
			if len(app.CName) != 1 {
				continue
			}
			nameParts := env.nameRegexp().FindStringSubmatch(app.Name)
			dnsParts := env.dnsRegexp().FindStringSubmatch(app.CName[0])
			if len(nameParts) == 2 && len(dnsParts) == 2 && nameParts[1] == dnsParts[1] && nameParts[1] == c.name {
				app.Addr = app.CName[0]
				app.Env = env
				projectApps = append(projectApps, app)
				continue
			}
		}
	}
	return projectApps, nil
}

func (c *projectInfo) Flags() *gnuflag.FlagSet {
	if c.fs == nil {
		c.fs = gnuflag.NewFlagSet("project-info", gnuflag.ExitOnError)
		c.fs.StringVar(&c.name, "name", "", "Name of the project")
		c.fs.StringVar(&c.name, "n", "", "Name of the project")
	}
	return c.fs
}

type projectList struct{}

func (c *projectList) Info() *cmd.Info {
	return &cmd.Info{
		Name: "project-list",
	}
}

func (c *projectList) Run(ctx *cmd.Context, client *cmd.Client) error {
	config, err := loadConfigFile()
	if err != nil {
		return errors.New("unable to load environments file, please make sure that tranor is properly configured")
	}
	apps, err := listApps(client, nil)
	if err != nil {
		return err
	}
	projects := make(map[string][]app)
	for _, env := range config.Environments {
		for _, app := range apps {
			if len(app.CName) != 1 {
				continue
			}
			partsName := env.nameRegexp().FindStringSubmatch(app.Name)
			partsDNS := env.dnsRegexp().FindStringSubmatch(app.CName[0])
			if len(partsName) == 2 && len(partsDNS) == 2 && partsName[1] == partsDNS[1] {
				projectName := partsDNS[1]
				app.Env = env
				app.Addr = app.CName[0]
				projects[projectName] = append(projects[projectName], app)
			}
		}
	}
	c.render(ctx.Stdout, projects)
	return nil
}

func (c *projectList) render(w io.Writer, projects map[string][]app) {
	var list projectSlice
	for name, apps := range projects {
		list = append(list, struct {
			Name string
			Apps []app
		}{Name: name, Apps: apps})
	}
	sort.Sort(list)
	var table cmd.Table
	table.LineSeparator = true
	table.Headers = cmd.Row{"Project", "Environments", "Address"}
	for _, project := range list {
		var (
			envNames  []string
			addresses []string
		)
		for _, app := range project.Apps {
			envNames = append(envNames, app.Env.Name)
			addresses = append(addresses, app.Addr)
		}
		table.AddRow(cmd.Row{project.Name, strings.Join(envNames, "\n"), strings.Join(addresses, "\n")})
	}
	w.Write(table.Bytes())
}

type projectSlice []struct {
	Name string
	Apps []app
}

func (l projectSlice) Len() int {
	return len(l)
}

func (l projectSlice) Less(i, j int) bool {
	return l[i].Name < l[j].Name
}

func (l projectSlice) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}