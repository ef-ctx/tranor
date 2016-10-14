// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"
	"testing"

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
		Teams:       []string{"myteam"},
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
		Teams:       []string{"myteam"},
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
