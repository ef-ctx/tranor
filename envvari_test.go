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
	cleanup, err := setupFakeTarget(tsuruServer.url())
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
	cleanup, err := setupFakeTarget(tsuruServer.url())
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
	cleanup, err := setupFakeTarget(tsuruServer.url())
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
