// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tsuru/gnuflag"
	"github.com/tsuru/tsuru-client/tsuru/client"
	"github.com/tsuru/tsuru/api"
	"github.com/tsuru/tsuru/cmd"
	tsuruerrors "github.com/tsuru/tsuru/errors"
)

type projectCreate struct {
	fs          *gnuflag.FlagSet
	name        string
	platform    string
	team        string
	plan        string
	description string
	envs        commaSeparatedFlag
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
	envs := getEnvironmentsByName(config.Environments, c.envs.Values())
	apps, err := createApps(envs, client, c.name, createAppOptions{
		Description: c.description,
		Plan:        c.plan,
		Platform:    c.platform,
		Team:        c.team,
	})
	if err != nil {
		return err
	}
	err = setCNames(apps, client, c.name)
	if err != nil {
		appObjs := make([]app, len(apps))
		for i, a := range apps {
			appObjs[i] = app{Name: a["name"]}
		}
		deleteApps(appObjs, client, ioutil.Discard)
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
		c.fs.StringVar(&c.description, "description", "", "description of the project")
		c.fs.StringVar(&c.description, "d", "", "description of the project")
		c.fs.StringVar(&c.platform, "platform", "", "platform of the project")
		c.fs.StringVar(&c.platform, "l", "", "platform of the project")
		c.fs.StringVar(&c.team, "team", "", "team that owns the project")
		c.fs.StringVar(&c.team, "t", "", "team that owns the project")
		c.fs.StringVar(&c.plan, "plan", "", "plan to use for the project")
		c.fs.StringVar(&c.plan, "p", "", "plan to use for the project")
		c.fs.Var(&c.envs, "envs", "comma-separated list of environments to use")
		c.fs.Var(&c.envs, "e", "comma-separated list of environments to use")
		c.envs.Set(c.defaultEnvs())
	}
	return c.fs
}

func (c *projectCreate) defaultEnvs() string {
	var envNames []string
	if config, err := loadConfigFile(); err == nil {
		envNames = config.envNames()
	}
	return strings.Join(envNames, ",")
}

type projectUpdate struct {
	fs          *gnuflag.FlagSet
	name        string
	team        string
	plan        string
	description string
	addEnvs     commaSeparatedFlag
	removeEnvs  commaSeparatedFlag
}

func (c *projectUpdate) Info() *cmd.Info {
	return &cmd.Info{
		Name: "project-update",
		Desc: "Updates the given project",
	}
}

func (c *projectUpdate) Run(ctx *cmd.Context, client *cmd.Client) error {
	if c.name == "" {
		return errors.New("please provide the name of the project")
	}
	config, err := loadConfigFile()
	if err != nil {
		return errors.New("unable to load environments file, please make sure that tranor is properly configured")
	}
	apps, err := projectApps(client, c.name)
	if err != nil {
		return err
	}
	err = c.validateEnvsToAdd(apps, config)
	if err != nil {
		return err
	}
	appsToUpdate, appsToRemove, err := c.validateEnvsToRemove(apps)
	if err != nil {
		return err
	}
	opts := c.baseOpts(apps[0])
	envsToAdd := getEnvironmentsByName(config.Environments, c.addEnvs.Values())
	fmt.Fprintln(ctx.Stdout, "adding new environments...")
	appMaps, err := createApps(envsToAdd, client, c.name, opts)
	if err != nil {
		return err
	}
	err = setCNames(appMaps, client, c.name)
	if err != nil {
		return err
	}
	fmt.Fprintln(ctx.Stdout, "removing old environments...")
	_, err = deleteApps(appsToRemove, client, ctx.Stdout)
	if err != nil {
		return err
	}
	if c.plan != "" || c.team != "" || c.description != "" {
		for _, a := range appsToUpdate {
			opts.Name = a.Name
			opts.Pool = a.Pool
			err = updateApp(client, opts)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *projectUpdate) baseOpts(a app) createAppOptions {
	opts := createAppOptions{
		Description: a.Description,
		Platform:    a.Platform,
		Pool:        a.Pool,
		Team:        a.TeamOwner,
	}
	if a.Plan.Name != "autogenerated" {
		opts.Plan = a.Plan.Name
	}
	if c.description != "" {
		opts.Description = c.description
	}
	if c.plan != "" {
		opts.Plan = c.plan
	}
	if c.team != "" {
		opts.Team = c.team
	}
	return opts
}

func (c *projectUpdate) validateEnvsToAdd(apps []app, config *Config) error {
	err := c.addEnvs.validate(config.envNames())
	if err != nil {
		return err
	}
	for _, envName := range c.addEnvs.Values() {
		for _, a := range apps {
			if a.Env.Name == envName {
				return fmt.Errorf("env %q is already defined in this project", envName)
			}
		}
	}
	return nil
}

func (c *projectUpdate) validateEnvsToRemove(apps []app) ([]app, []app, error) {
	var (
		appsToUpdate    []app
		appsToRemove    []app
		indexesToRemove []int
	)
	for _, envName := range c.removeEnvs.Values() {
		oldLen := len(appsToRemove)
		for i, a := range apps {
			if a.Env.Name == envName {
				appsToRemove = append(appsToRemove, a)
				indexesToRemove = append(indexesToRemove, i)
				break
			}
		}
		if len(appsToRemove) == oldLen {
			return nil, nil, fmt.Errorf("env %q is not defined in this project", envName)
		}
	}
	for i := range apps {
		var found bool
		for _, index := range indexesToRemove {
			if index == i {
				found = true
				break
			}
		}
		if !found {
			appsToUpdate = append(appsToUpdate, apps[i])
		}
	}
	return appsToUpdate, appsToRemove, nil
}

func (c *projectUpdate) Flags() *gnuflag.FlagSet {
	if c.fs == nil {
		c.fs = gnuflag.NewFlagSet("project-create", gnuflag.ExitOnError)
		c.fs.StringVar(&c.name, "name", "", "name of the project")
		c.fs.StringVar(&c.name, "n", "", "name of the project")
		c.fs.StringVar(&c.description, "description", "", "description of the project")
		c.fs.StringVar(&c.description, "d", "", "description of the project")
		c.fs.StringVar(&c.team, "team", "", "team that owns the project")
		c.fs.StringVar(&c.team, "t", "", "team that owns the project")
		c.fs.StringVar(&c.plan, "plan", "", "plan to use for the project")
		c.fs.StringVar(&c.plan, "p", "", "plan to use for the project")
		c.fs.Var(&c.addEnvs, "add-envs", "comma-separated list of environments to add to the project")
		c.fs.Var(&c.removeEnvs, "remove-envs", "comma-separated list of environments to remove from the project")
	}
	return c.fs
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
	ctx.RawOutput()
	config, err := loadConfigFile()
	if err != nil {
		return errors.New("unable to load environments file, please make sure that tranor is properly configured")
	}
	if !c.Confirm(ctx, fmt.Sprintf("Are you sure you want to remove the project %q?", c.name)) {
		return nil
	}
	apps := make([]app, len(config.Environments))
	for i, env := range config.Environments {
		apps[i] = app{Name: fmt.Sprintf("%s-%s", c.name, env.Name), Env: env}
	}
	errs, err := deleteApps(apps, client, ctx.Stdout)
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
		Desc: "Retrieves and displays information about the given project",
	}
}

func (c *projectInfo) Run(ctx *cmd.Context, client *cmd.Client) error {
	if c.name == "" {
		return errors.New("please provide the name of the project")
	}
	apps, err := projectApps(client, c.name)
	if err != nil {
		return err
	}
	fmt.Fprintf(ctx.Stdout, "Project name: %s\n", c.name)
	fmt.Fprintf(ctx.Stdout, "Description: %s\n", apps[0].Description)
	fmt.Fprintf(ctx.Stdout, "Repository: %s\n", apps[0].RepositoryURL)
	fmt.Fprintf(ctx.Stdout, "Platform: %s\n", apps[0].Platform)
	fmt.Fprintf(ctx.Stdout, "Teams: %s\n", strings.Join(apps[0].Teams, ", "))
	fmt.Fprintf(ctx.Stdout, "Owner: %s\n", apps[0].Owner)
	fmt.Fprintf(ctx.Stdout, "Team owner: %s\n", apps[0].TeamOwner)
	var envs cmd.Table
	envs.Headers = cmd.Row{"Environment", "Address", "Image", "Git hash/tag", "Deploy date", "Units"}
	for _, app := range apps {
		row := cmd.Row{app.Env.Name, app.Addr, "", "", "", strconv.Itoa(len(app.Units))}
		if appDeploy, err := lastDeploy(client, app.Name); err == nil && appDeploy.Image != "" {
			row[2] = appDeploy.Image
			row[4] = appDeploy.Timestamp.Format(time.RFC1123)
			if appDeploy.Commit != "" {
				row[3] = fmt.Sprintf("(git) %s", appDeploy.Commit)
			}
		}
		envs.AddRow(row)
	}
	fmt.Fprintln(ctx.Stdout)
	ctx.Stdout.Write(envs.Bytes())
	return nil
}

func (c *projectInfo) Flags() *gnuflag.FlagSet {
	if c.fs == nil {
		c.fs = gnuflag.NewFlagSet("project-info", gnuflag.ExitOnError)
		c.fs.StringVar(&c.name, "name", "", "Name of the project")
		c.fs.StringVar(&c.name, "n", "", "Name of the project")
	}
	return c.fs
}

type projectEnvInfo struct {
	envName string
	name    string
	fs      *gnuflag.FlagSet
}

func (c *projectEnvInfo) Info() *cmd.Info {
	return &cmd.Info{
		Name: "project-env-info",
		Desc: "Displays information about a project in a specific environment",
	}
}

func (c *projectEnvInfo) Run(ctx *cmd.Context, cli *cmd.Client) error {
	if c.name == "" || c.envName == "" {
		return errors.New("please provide the project name and the environment")
	}
	appName := fmt.Sprintf("%s-%s", c.name, c.envName)
	var appInfo client.AppInfo
	appInfo.Flags().Parse(true, []string{"--app", appName})
	return appInfo.Run(ctx, cli)
}

func (c *projectEnvInfo) Flags() *gnuflag.FlagSet {
	if c.fs == nil {
		c.fs = gnuflag.NewFlagSet("project-env-info", gnuflag.ExitOnError)
		c.fs.StringVar(&c.name, "name", "", "name of the project")
		c.fs.StringVar(&c.name, "n", "", "name of the project")
		c.fs.StringVar(&c.envName, "env", "", "name of the environment")
		c.fs.StringVar(&c.envName, "e", "", "name of the environment")
	}
	return c.fs
}

type projectList struct{}

func (c *projectList) Info() *cmd.Info {
	return &cmd.Info{
		Name: "project-list",
		Desc: "List the projects on tranor that you has access to",
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
			if len(app.CName) < 1 {
				continue
			}
			if projectName, cname, err := extractProjectName(app, env); err == nil {
				app.Env = env
				app.Addr = cname
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

func createApps(envs []Environment, client *cmd.Client, projectName string, opts createAppOptions) ([]map[string]string, error) {
	createdApps := make([]map[string]string, 0, len(envs))
	apps := make([]app, 0, len(envs))
	for _, env := range envs {
		opts.Name = fmt.Sprintf("%s-%s", projectName, env.Name)
		opts.Pool = env.poolName()
		a, err := createApp(client, opts)
		if err != nil {
			deleteApps(apps, client, ioutil.Discard)
			return nil, fmt.Errorf("failed to create the project in env %q: %s", env.Name, err)
		}
		a["name"] = opts.Name
		a["dnsSuffix"] = env.DNSSuffix
		setEnvVars(client, opts.Name, &api.Envs{
			Envs: []struct {
				Name  string
				Value string
			}{
				{Name: "TRANOR_ENV_NAME", Value: env.Name},
			},
		})
		createdApps = append(createdApps, a)
		apps = append(apps, app{Name: opts.Name})
	}
	return createdApps, nil
}

func setCNames(apps []map[string]string, client *cmd.Client, projectName string) error {
	for _, app := range apps {
		err := setCName(app["name"], fmt.Sprintf("%s.%s", projectName, app["dnsSuffix"]), client)
		if err != nil {
			return err
		}
	}
	return nil
}

func setCName(appName, cname string, client *cmd.Client) error {
	reqURL, err := cmd.GetURL(fmt.Sprintf("/apps/%s/cname", appName))
	if err != nil {
		return err
	}
	v := make(url.Values)
	v.Set("cname", cname)
	req, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func projectApps(client *cmd.Client, name string) ([]app, error) {
	config, err := loadConfigFile()
	if err != nil {
		return nil, errors.New("unable to load environments file, please make sure that tranor is properly configured")
	}
	apps, err := listApps(client, map[string]string{"name": "^" + name})
	if err != nil {
		return nil, err
	}
	var projectApps []app
	for _, env := range config.Environments {
		for _, app := range apps {
			if len(app.CName) < 1 {
				continue
			}
			projectName, cname, err := extractProjectName(app, env)
			if err != nil {
				continue
			}
			if projectName == name {
				app, err = getApp(client, app.Name)
				if err != nil {
					return nil, err
				}
				app.Addr = cname
				app.Env = env
				projectApps = append(projectApps, app)
			}
		}
	}
	if len(projectApps) == 0 {
		return nil, errors.New("project not found")
	}
	return projectApps, nil
}

func getEnvironmentsByName(envs []Environment, names []string) []Environment {
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

func extractProjectName(a app, env Environment) (projectName string, cname string, err error) {
	partsName := env.nameRegexp().FindStringSubmatch(a.Name)
	cname, err = findCName(a, env)
	if err != nil {
		return "", "", err
	}
	partsDNS := env.dnsRegexp().FindStringSubmatch(cname)
	if len(partsName) == 2 && len(partsDNS) == 2 && partsName[1] == partsDNS[1] {
		return partsDNS[1], cname, nil
	}
	return "", "", errors.New("not a tranor project")
}

func findCName(a app, env Environment) (string, error) {
	for _, cname := range a.CName {
		if env.dnsRegexp().MatchString(cname) {
			return cname, nil
		}
	}
	return "", errors.New("cname not defined")
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
