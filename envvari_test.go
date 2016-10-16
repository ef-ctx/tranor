// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/tsuru/tsuru/cmd"
)

func TestProjectEnvVarSet(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeConfig(tsuruServer.url(), tsuruServer.token())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{"USER_NAME=root", "USER_PASSWORD=r00t", `PREFERRED_TEAM="some nice team"`},
	}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	appMaps, err := createApps([]Environment{
		{Name: "dev", DNSSuffix: "dev.example.com"},
		{Name: "stage", DNSSuffix: "stage.example.com"},
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
	var c projectEnvVarSet
	err = c.Flags().Parse(true, []string{
		"-n", "myproj",
		"-e", "dev,stage,prod",
		"--no-restart",
		"--private",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := `setting variables in environment "dev"... ok
setting variables in environment "stage"... ok
setting variables in environment "prod"... ok
`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output\nwant:\n%s\ngot:\n%s", expectedOutput, stdout.String())
	}
}

func TestProjectEnvVarSetDefaultEnvs(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeConfig(tsuruServer.url(), tsuruServer.token())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{"USER_NAME=root", "USER_PASSWORD=r00t", `PREFERRED_TEAM="some nice team"`},
	}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	config, _ := loadConfigFile()
	appMaps, err := createApps(config.Environments, client, "myproj", createAppOptions{
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
	var c projectEnvVarSet
	err = c.Flags().Parse(true, []string{
		"-n", "myproj",
		"--no-restart",
		"--private",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := `setting variables in environment "dev"... ok
setting variables in environment "qa"... ok
setting variables in environment "stage"... ok
setting variables in environment "prod"... ok
`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output\nwant:\n%s\ngot:\n%s", expectedOutput, stdout.String())
	}
}

func TestProjectEnvVarSetAppNotFound(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeConfig(tsuruServer.url(), tsuruServer.token())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{"USER_NAME=root", "USER_PASSWORD=r00t", `PREFERRED_TEAM="some nice team"`},
	}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	_, err = createApps([]Environment{
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
	var c projectEnvVarSet
	err = c.Flags().Parse(true, []string{
		"-n", "myproj",
		"-e", "dev,stage,prod",
		"--no-restart",
		"-p",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := `setting variables in environment "dev"... ok
setting variables in environment "stage"... not found
setting variables in environment "prod"... ok
`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output\nwant:\n%s\ngot:\n%s", expectedOutput, stdout.String())
	}
}

func TestProjectEnvVarGet(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeConfig(tsuruServer.url(), tsuruServer.token())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	config, _ := loadConfigFile()
	_, err = createApps(config.Environments, client, "myproj", createAppOptions{
		Plan:        "medium",
		Description: "my nice project",
		Team:        "myteam",
		Platform:    "python",
	})
	if err != nil {
		t.Fatal(err)
	}
	var c projectEnvVarGet
	err = c.Flags().Parse(true, []string{
		"-n", "myproj",
		"-e", "dev,stage,prod",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := `variables in "dev":

 TSURU_APPDIR=*** (private variable)
 TSURU_APPNAME=*** (private variable)
 TSURU_APP_TOKEN=*** (private variable)


variables in "stage":

 TSURU_APPDIR=*** (private variable)
 TSURU_APPNAME=*** (private variable)
 TSURU_APP_TOKEN=*** (private variable)


variables in "prod":

 TSURU_APPDIR=*** (private variable)
 TSURU_APPNAME=*** (private variable)
 TSURU_APP_TOKEN=*** (private variable)


`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output\nwant:\n%q\ngot:\n%q", expectedOutput, stdout.String())
	}
}

func TestProjectEnvVarGetDefaultEnvs(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeConfig(tsuruServer.url(), tsuruServer.token())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	config, _ := loadConfigFile()
	_, err = createApps(config.Environments, client, "myproj", createAppOptions{
		Plan:        "medium",
		Description: "my nice project",
		Team:        "myteam",
		Platform:    "python",
	})
	if err != nil {
		t.Fatal(err)
	}
	var c projectEnvVarGet
	err = c.Flags().Parse(true, []string{"-n", "myproj"})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := `variables in "dev":

 TSURU_APPDIR=*** (private variable)
 TSURU_APPNAME=*** (private variable)
 TSURU_APP_TOKEN=*** (private variable)


variables in "qa":

 TSURU_APPDIR=*** (private variable)
 TSURU_APPNAME=*** (private variable)
 TSURU_APP_TOKEN=*** (private variable)


variables in "stage":

 TSURU_APPDIR=*** (private variable)
 TSURU_APPNAME=*** (private variable)
 TSURU_APP_TOKEN=*** (private variable)


variables in "prod":

 TSURU_APPDIR=*** (private variable)
 TSURU_APPNAME=*** (private variable)
 TSURU_APP_TOKEN=*** (private variable)


`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output\nwant:\n%q\ngot:\n%q", expectedOutput, stdout.String())
	}
}

func TestProjectEnvVarGetAppNotFound(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeConfig(tsuruServer.url(), tsuruServer.token())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	_, err = createApps([]Environment{
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
	var c projectEnvVarGet
	err = c.Flags().Parse(true, []string{
		"-n", "myproj",
		"-e", "dev,stage,prod",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := `variables in "dev":

 TSURU_APPDIR=*** (private variable)
 TSURU_APPNAME=*** (private variable)
 TSURU_APP_TOKEN=*** (private variable)


variables in "prod":

 TSURU_APPDIR=*** (private variable)
 TSURU_APPNAME=*** (private variable)
 TSURU_APP_TOKEN=*** (private variable)


`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output\nwant:\n%q\ngot:\n%q", expectedOutput, stdout.String())
	}
	expectedStderr := `WARNING: project not found in environment "stage"` + "\n"
	if stderr.String() != expectedStderr {
		t.Errorf("wrong error output\nwant:\n%q\ngot:\n%q", expectedStderr, stderr.String())
	}
}

func TestProjectEnvVarUnset(t *testing.T) {
	tsuruServer.reset()
	server := newFakeServer(t)
	defer server.stop()
	cleanup, err := setupFakeConfig(tsuruServer.url(), tsuruServer.token())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{"USER_NAME", "USER_PASSWORD", "PREFERRED_TEAM"},
	}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	config, _ := loadConfigFile()
	_, err = createApps(config.Environments, client, "myproj", createAppOptions{
		Plan:        "medium",
		Description: "my nice project",
		Team:        "myteam",
		Platform:    "python",
	})
	var c projectEnvVarUnset
	err = c.Flags().Parse(true, []string{
		"-n", "myproj",
		"-e", "dev,stage,prod",
		"--no-restart",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := `unsetting variables from environment "dev"... ok
unsetting variables from environment "stage"... ok
unsetting variables from environment "prod"... ok
`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output\nwant:\n%s\ngot:\n%s", expectedOutput, stdout.String())
	}
}

func TestProjectEnvVarUnsetDefaultEnvs(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeConfig(tsuruServer.url(), tsuruServer.token())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{"USER_NAME", "USER_PASSWORD", "PREFERRED_TEAM"},
	}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	var c projectEnvVarUnset
	config, _ := loadConfigFile()
	_, err = createApps(config.Environments, client, "myproj", createAppOptions{
		Plan:        "medium",
		Description: "my nice project",
		Team:        "myteam",
		Platform:    "python",
	})
	err = c.Flags().Parse(true, []string{
		"-n", "myproj",
		"--no-restart",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := `unsetting variables from environment "dev"... ok
unsetting variables from environment "qa"... ok
unsetting variables from environment "stage"... ok
unsetting variables from environment "prod"... ok
`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output\nwant:\n%s\ngot:\n%s", expectedOutput, stdout.String())
	}
}

func TestProjectEnvVarUnsetAppNotFound(t *testing.T) {
	tsuruServer.reset()
	cleanup, err := setupFakeConfig(tsuruServer.url(), tsuruServer.token())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{"USER_NAME", "USER_PASSWORD", "PREFERRED_TEAM"},
	}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	_, err = createApps([]Environment{
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
	var c projectEnvVarUnset
	err = c.Flags().Parse(true, []string{
		"-n", "myproj",
		"-e", "dev,stage,prod",
		"--no-restart",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := `unsetting variables from environment "dev"... ok
unsetting variables from environment "stage"... not found
unsetting variables from environment "prod"... ok
`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output\nwant:\n%s\ngot:\n%s", expectedOutput, stdout.String())
	}
}
