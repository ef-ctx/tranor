// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/tsuru/tsuru-client/tsuru/client"
	"github.com/tsuru/tsuru/cmd"
)

func TestProjectCreateDefaultEnvs(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeTarget(tsuruServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectCreate
	err = c.Flags().Parse(true, []string{
		"-n", "myproj",
		"-l", "python",
		"-t", "myteam",
		"-p", "medium",
		"-d", "my nice project",
	})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := `successfully created the project "myproj"!
Git repository: git@gandalf.example.com:myproj-dev.git
`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output returned\nWant: %s\nGot:  %s", expectedOutput, stdout.String())
	}
	expectedApp := app{
		Description: "my nice project",
		Platform:    "python",
		TeamOwner:   "myteam",
		Plan: struct {
			Name string `json:"name"`
		}{Name: "medium"},
	}
	envNames := []string{"dev", "qa", "stage", "prod"}
	dnsSuffixes := []string{"dev.example.com", "qa.example.com", "stage.example.com", "example.com"}
	apps, err := listApps(client, map[string]string{"name": "^myproj"})
	if err != nil {
		t.Fatal(err)
	}
	for i, a := range apps {
		expectedApp.Name = "myproj-" + envNames[i]
		expectedApp.Pool = fmt.Sprintf(`%s\%s`, envNames[i], dnsSuffixes[i])
		expectedApp.CName = []string{"myproj." + dnsSuffixes[i]}
		expectedApp.RepositoryURL = "git@gandalf.example.com:" + expectedApp.Name + ".git"

		// we don't care about the value of the fields below
		expectedApp.Addr = a.Addr
		expectedApp.Owner = a.Owner
		expectedApp.Teams = a.Teams

		if !reflect.DeepEqual(a, expectedApp) {
			t.Errorf("wrong app in env %q:\nwant %#v\ngot  %#v", envNames[i], expectedApp, a)
		}
	}
}

func TestProjectCreateSpecifyEnvs(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeTarget(tsuruServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectCreate
	err = c.Flags().Parse(true, []string{
		"-n", "superproj",
		"-d", "super project, just dev and prod needed",
		"-l", "python",
		"-t", "myteam",
		"-p", "medium",
		"-e", "dev,prod",
	})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := `successfully created the project "superproj"!
Git repository: git@gandalf.example.com:superproj-dev.git
`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output returned\nWant: %s\nGot:  %s", expectedOutput, stdout.String())
	}
	expectedApp := app{
		Description: "super project, just dev and prod needed",
		Platform:    "python",
		TeamOwner:   "myteam",
		Plan: struct {
			Name string `json:"name"`
		}{Name: "medium"},
	}
	envNames := []string{"dev", "prod"}
	dnsSuffixes := []string{"dev.example.com", "example.com"}
	apps, err := listApps(client, map[string]string{"name": "^superproj"})
	if err != nil {
		t.Fatal(err)
	}
	for i, a := range apps {
		expectedApp.Name = "superproj-" + envNames[i]
		expectedApp.Pool = fmt.Sprintf(`%s\%s`, envNames[i], dnsSuffixes[i])
		expectedApp.CName = []string{"superproj." + dnsSuffixes[i]}
		expectedApp.RepositoryURL = "git@gandalf.example.com:" + expectedApp.Name + ".git"

		// we don't care about the value of the fields below
		expectedApp.Addr = a.Addr
		expectedApp.Owner = a.Owner
		expectedApp.Teams = a.Teams

		if !reflect.DeepEqual(a, expectedApp) {
			t.Errorf("wrong app in env %q:\nwant %#v\ngot  %#v", envNames[i], expectedApp, a)
		}
	}
}

func TestProjectCreateFailToCreateApp(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeTarget(tsuruServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	createApp(client, createAppOptions{
		Name:     "superproj-dev",
		Platform: "python",
		Team:     "myteam",
	})
	var c projectCreate
	err = c.Flags().Parse(true, []string{
		"-n", "superproj",
		"-l", "python",
		"-t", "myteam",
		"-p", "medium",
		"-e", "dev,prod",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Run(&ctx, client)
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
}

func TestProjectUpdate(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeTarget(tsuruServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	appMaps, err := createApps([]Environment{
		{Name: "dev", DNSSuffix: "dev.example.com"},
		{Name: "prod", DNSSuffix: "example.com"},
	}, client, "myproj", createAppOptions{
		Plan:        "medium",
		Description: "my nice project",
		Team:        "myteam",
		Platform:    "python",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = setCNames(appMaps, client, "myproj")
	if err != nil {
		t.Fatal(err)
	}
	var c projectUpdate
	err = c.Flags().Parse(true, []string{
		"-n", "myproj",
		"-d", "updated project description",
		"-t", "superteam",
		"-p", "huge",
		"--add-envs", "qa,stage",
		"--remove-envs", "dev",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	envNames := []string{"prod", "qa", "stage"}
	dnsSuffixes := []string{"example.com", "qa.example.com", "stage.example.com"}
	apps, err := listApps(client, map[string]string{"name": "^myproj"})
	if err != nil {
		t.Fatal(err)
	}
	expectedApp := app{
		Description: "updated project description",
		Platform:    "python",
		TeamOwner:   "superteam",
		Plan: struct {
			Name string `json:"name"`
		}{Name: "huge"},
	}
	alist := appList(apps)
	sort.Sort(alist)
	for i, a := range alist {
		expectedApp.Name = "myproj-" + envNames[i]
		expectedApp.Pool = fmt.Sprintf(`%s\%s`, envNames[i], dnsSuffixes[i])
		expectedApp.CName = []string{"myproj." + dnsSuffixes[i]}
		expectedApp.RepositoryURL = "git@gandalf.example.com:" + expectedApp.Name + ".git"

		// we don't care about the value of the fields below
		expectedApp.Addr = a.Addr
		expectedApp.Owner = a.Owner
		expectedApp.Teams = a.Teams

		if !reflect.DeepEqual(a, expectedApp) {
			t.Errorf("wrong app in env %q:\nwant %#v\ngot  %#v", envNames[i], expectedApp, a)
		}
	}
}

func TestProjectUpdateAutogenerated(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeTarget(tsuruServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	appMaps, err := createApps([]Environment{
		{Name: "dev", DNSSuffix: "dev.example.com"},
		{Name: "prod", DNSSuffix: "example.com"},
	}, client, "myproj", createAppOptions{
		Description: "my nice project",
		Team:        "myteam",
		Platform:    "python",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = setCNames(appMaps, client, "myproj")
	if err != nil {
		t.Fatal(err)
	}
	var c projectUpdate
	err = c.Flags().Parse(true, []string{
		"-n", "myproj",
		"-d", "updated project description",
		"-t", "superteam",
		"--add-envs", "qa,stage",
		"--remove-envs", "dev",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	envNames := []string{"prod", "qa", "stage"}
	dnsSuffixes := []string{"example.com", "qa.example.com", "stage.example.com"}
	apps, err := listApps(client, map[string]string{"name": "^myproj"})
	if err != nil {
		t.Fatal(err)
	}
	expectedApp := app{
		Description: "updated project description",
		Platform:    "python",
		TeamOwner:   "superteam",
		Plan: struct {
			Name string `json:"name"`
		}{Name: "autogenerated"},
	}
	alist := appList(apps)
	sort.Sort(alist)
	for i, a := range alist {
		expectedApp.Name = "myproj-" + envNames[i]
		expectedApp.Pool = fmt.Sprintf(`%s\%s`, envNames[i], dnsSuffixes[i])
		expectedApp.CName = []string{"myproj." + dnsSuffixes[i]}
		expectedApp.RepositoryURL = "git@gandalf.example.com:" + expectedApp.Name + ".git"

		// we don't care about the value of the fields below
		expectedApp.Addr = a.Addr
		expectedApp.Owner = a.Owner
		expectedApp.Teams = a.Teams

		if !reflect.DeepEqual(a, expectedApp) {
			t.Errorf("wrong app in env %q:\nwant %#v\ngot  %#v", envNames[i], expectedApp, a)
		}
	}
}

func TestProjectUpdateNoNewEnvs(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeTarget(tsuruServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	appMaps, err := createApps([]Environment{
		{Name: "dev", DNSSuffix: "dev.example.com"},
		{Name: "prod", DNSSuffix: "example.com"},
	}, client, "myproj", createAppOptions{
		Description: "my nice project",
		Team:        "myteam",
		Platform:    "python",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = setCNames(appMaps, client, "myproj")
	if err != nil {
		t.Fatal(err)
	}
	var c projectUpdate
	err = c.Flags().Parse(true, []string{
		"-n", "myproj",
		"-d", "updated project description",
		"-t", "superteam",
		"-p", "small",
		"--remove-envs", "dev",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	apps, err := listApps(client, map[string]string{"name": "^myproj"})
	if err != nil {
		t.Fatal(err)
	}
	if len(apps) != 1 {
		t.Errorf("too many apps: %#v", apps)
	}
	expectedApp := app{
		Name:          "myproj-prod",
		CName:         []string{"myproj.example.com"},
		Pool:          `prod\example.com`,
		Description:   "updated project description",
		Platform:      "python",
		RepositoryURL: "git@gandalf.example.com:myproj-prod.git",
		TeamOwner:     "superteam",
		Addr:          apps[0].Addr,
		Teams:         apps[0].Teams,
		Owner:         apps[0].Owner,
		Plan: struct {
			Name string `json:"name"`
		}{Name: "small"},
	}
	if !reflect.DeepEqual(apps[0], expectedApp) {
		t.Errorf("wrong app returned\nwant %#v\ngot  %#v", expectedApp, apps[0])
	}
}

func TestProjectUpdateOnlyEnvs(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeTarget(tsuruServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	appMaps, err := createApps([]Environment{
		{Name: "dev", DNSSuffix: "dev.example.com"},
		{Name: "prod", DNSSuffix: "example.com"},
	}, client, "myproj", createAppOptions{
		Description: "my nice project",
		Team:        "myteam",
		Plan:        "medium",
		Platform:    "python",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = setCNames(appMaps, client, "myproj")
	if err != nil {
		t.Fatal(err)
	}
	var c projectUpdate
	err = c.Flags().Parse(true, []string{
		"-n", "myproj",
		"--add-envs", "qa,stage",
		"--remove-envs", "dev",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Run(&ctx, client)
	if err != nil {
		t.Error(err)
	}
	envNames := []string{"prod", "qa", "stage"}
	dnsSuffixes := []string{"example.com", "qa.example.com", "stage.example.com"}
	apps, err := listApps(client, map[string]string{"name": "^myproj"})
	if err != nil {
		t.Fatal(err)
	}
	expectedApp := app{
		Description: "my nice project",
		Platform:    "python",
		TeamOwner:   "myteam",
		Plan: struct {
			Name string `json:"name"`
		}{Name: "medium"},
	}
	alist := appList(apps)
	sort.Sort(alist)
	for i, a := range alist {
		expectedApp.Name = "myproj-" + envNames[i]
		expectedApp.Pool = fmt.Sprintf(`%s\%s`, envNames[i], dnsSuffixes[i])
		expectedApp.CName = []string{"myproj." + dnsSuffixes[i]}
		expectedApp.RepositoryURL = "git@gandalf.example.com:" + expectedApp.Name + ".git"

		// we don't care about the value of the fields below
		expectedApp.Addr = a.Addr
		expectedApp.Owner = a.Owner
		expectedApp.Teams = a.Teams

		if !reflect.DeepEqual(a, expectedApp) {
			t.Errorf("wrong app in env %q:\nwant %#v\ngot  %#v", envNames[i], expectedApp, a)
		}
	}
}

func TestProjectRemove(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeTarget(tsuruServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	appMaps, err := createApps([]Environment{
		{Name: "dev", DNSSuffix: "dev.example.com"},
		{Name: "qa", DNSSuffix: "qa.example.com"},
		{Name: "stage", DNSSuffix: "stage.example.com"},
		{Name: "prod", DNSSuffix: "example.com"},
	}, client, "myproj", createAppOptions{
		Description: "my nice project",
		Team:        "myteam",
		Plan:        "medium",
		Platform:    "python",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = setCNames(appMaps, client, "myproj")
	if err != nil {
		t.Fatal(err)
	}
	var c projectRemove
	err = c.Flags().Parse(true, []string{"-yn", "myproj"})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := `Deleting from env "dev"... ok
Deleting from env "qa"... ok
Deleting from env "stage"... ok
Deleting from env "prod"... ok
`
	if stdout.String() != expectedOutput {
		t.Errorf("Wrong output\nWant:\n%s\nGot:\n%s", expectedOutput, stdout.String())
	}
}

func TestProjectRemoveSomeEnvironments(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeTarget(tsuruServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	appMaps, err := createApps([]Environment{
		{Name: "dev", DNSSuffix: "dev.example.com"},
		{Name: "prod", DNSSuffix: "stage.example.com"},
	}, client, "myproj", createAppOptions{
		Description: "my nice project",
		Team:        "myteam",
		Plan:        "medium",
		Platform:    "python",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = setCNames(appMaps, client, "myproj")
	if err != nil {
		t.Fatal(err)
	}
	var c projectRemove
	err = c.Flags().Parse(true, []string{"-yn", "myproj"})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := `Deleting from env "dev"... ok
Deleting from env "qa"... ok
Deleting from env "stage"... ok
Deleting from env "prod"... ok
`
	if stdout.String() != expectedOutput {
		t.Errorf("Wrong output\nWant:\n%s\nGot:\n%s", expectedOutput, stdout.String())
	}
}

func TestProjectRemoveNotFound(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeTarget(tsuruServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectRemove
	err = c.Flags().Parse(true, []string{"-yn", "myproj"})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = c.Run(&ctx, client)
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
	expectedMessage := "project not found"
	if err.Error() != expectedMessage {
		t.Errorf("wrong error message\nwant %q\ngot  %q", expectedMessage, err.Error())
	}
}

func TestProjectInfo(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeTarget(tsuruServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	config, _ := loadConfigFile()
	appMaps, err := createApps(config.Environments, client, "proj1", createAppOptions{
		Plan:        "medium",
		Description: "my nice project",
		Team:        "myteam",
		Platform:    "python",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = setCNames(appMaps, client, "proj1")
	if err != nil {
		t.Fatal(err)
	}
	var c projectInfo
	err = c.Flags().Parse(true, []string{"-n", "proj1"})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := getApp(client, "proj1-dev")
	table := cmd.Table{Headers: cmd.Row([]string{"Environment", "Address", "Image", "Git hash/tag", "Deploy date", "Units"})}
	expectedOutput := fmt.Sprintf(`Project name: proj1
Description: my nice project
Repository: git@gandalf.example.com:proj1-dev.git
Platform: python
Teams: %s
Owner: %s
Team owner: myteam`+"\n\n", strings.Join(a.Teams, ", "), a.Owner)
	rows := []cmd.Row{
		{"dev", "proj1.dev.example.com", "", "", "", "0"},
		{"qa", "proj1.qa.example.com", "", "", "", "0"},
		{"stage", "proj1.stage.example.com", "", "", "", "0"},
		{"prod", "proj1.example.com", "", "", "", "0"},
	}
	for _, row := range rows {
		table.AddRow(row)
	}
	expectedOutput += table.String()
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output\nWant:\n%s\nGot:\n%s", expectedOutput, stdout.String())
	}
}

func TestProjectInfoNotFound(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeTarget(tsuruServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectInfo
	err = c.Flags().Parse(true, []string{"-n", "proj1"})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = c.Run(&ctx, client)
	if err == nil {
		t.Fatal(err)
	}
	expectedMsg := "project not found"
	if err.Error() != expectedMsg {
		t.Errorf("wrong error message\nwant %q\ngot  %q", expectedMsg, err.Error())
	}
}

func TestProjectEnvInfo(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeTarget(tsuruServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	cli := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	_, err = createApp(cli, createAppOptions{
		Name:     "proj1-prod",
		Platform: "python",
		Team:     "myteam",
	})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	var appInfoCmd client.AppInfo
	err = appInfoCmd.Flags().Parse(true, []string{"-a", "proj1-prod"})
	if err != nil {
		t.Fatal(err)
	}
	err = appInfoCmd.Run(&cmd.Context{Stdout: &buf}, cli)
	if err != nil {
		t.Fatal(err)
	}
	var c projectEnvInfo
	err = c.Flags().Parse(true, []string{"-n", "proj1", "-e", "prod"})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Run(&ctx, cli)
	if err != nil {
		t.Fatal(err)
	}
	if stdout.String() != buf.String() {
		t.Errorf("Wrong output\nWant:\n%s\nGot:\n%s", &buf, &stdout)
	}
}

func TestProjectList(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeTarget(tsuruServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	config, _ := loadConfigFile()
	for _, projName := range []string{"myproj1", "myproj2"} {
		appMaps, innerErr := createApps(config.Environments, client, projName, createAppOptions{
			Plan:        "medium",
			Description: "my nice project",
			Team:        "myteam",
			Platform:    "python",
		})
		innerErr = setCNames(appMaps, client, projName)
		if innerErr != nil {
			t.Fatal(innerErr)
		}
	}
	appMaps, err := createApps([]Environment{
		{Name: "dev", DNSSuffix: "dev.example.com"},
		{Name: "prod", DNSSuffix: "example.com"},
	}, client, "myproj3", createAppOptions{
		Plan:        "medium",
		Description: "my nice project",
		Team:        "myteam",
		Platform:    "python",
	})
	err = setCNames(appMaps, client, "myproj3")
	if err != nil {
		t.Fatal(err)
	}
	var c projectList
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := `+---------+--------------+---------------------------+
| Project | Environments | Address                   |
+---------+--------------+---------------------------+
| myproj1 | dev          | myproj1.dev.example.com   |
|         | qa           | myproj1.qa.example.com    |
|         | stage        | myproj1.stage.example.com |
|         | prod         | myproj1.example.com       |
+---------+--------------+---------------------------+
| myproj2 | dev          | myproj2.dev.example.com   |
|         | qa           | myproj2.qa.example.com    |
|         | stage        | myproj2.stage.example.com |
|         | prod         | myproj2.example.com       |
+---------+--------------+---------------------------+
| myproj3 | dev          | myproj3.dev.example.com   |
|         | prod         | myproj3.example.com       |
+---------+--------------+---------------------------+
`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output\nWant:\n%s\nGot:\n%s", expectedOutput, stdout.String())
	}
}

type appList []app

func (s appList) Len() int {
	return len(s)
}

func (s appList) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s appList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
