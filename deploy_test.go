// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/tsuru/gnuflag"
	"github.com/tsuru/tsuru-client/tsuru/client"
	"github.com/tsuru/tsuru/cmd"
)

func TestProjectDeployPromoteFlags(t *testing.T) {
	fakeServer := newFakeServer(t)
	defer fakeServer.stop()
	fakeServer.prepareResponse(preparedResponse{
		method:  http.MethodGet,
		path:    "/apps?name=" + url.QueryEscape("^proj1"),
		code:    http.StatusOK,
		payload: []byte(listOfApps),
	})
	envNames := []string{"dev", "qa", "stage", "prod"}
	for _, envName := range envNames {
		fakeServer.prepareResponse(preparedResponse{
			method:  http.MethodGet,
			path:    "/apps/proj1-" + envName,
			code:    http.StatusOK,
			payload: []byte(appInfo1),
		})
	}
	fakeServer.prepareResponse(preparedResponse{
		method:  http.MethodGet,
		path:    "/apps/proj1-dev",
		code:    http.StatusOK,
		payload: []byte(appInfo1),
	})
	fakeServer.prepareResponse(preparedResponse{
		method:  http.MethodGet,
		path:    "/deploys?limit=1&app=proj1-dev",
		code:    http.StatusOK,
		payload: []byte(deployments),
	})
	cleanup, err := setupFakeConfig(fakeServer.url(), "")
	if err != nil {
		t.Fatal(err)
	}
	oldCommand := tsuruDeployCommand
	fakeCommand := fakeTsuruCommand{FlaggedCommand: &client.AppDeploy{}}
	tsuruDeployCommand = &fakeCommand
	defer func() {
		tsuruDeployCommand = oldCommand
		cleanup()
	}()
	var c projectDeploy
	ctx := cmd.Context{}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	c.Flags().Parse(true, []string{"-n", "proj1", "-e", "stg", "-p", "dev"})
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	flags := fakeCommand.inputFlags()
	expectedFlags := map[string]string{
		"a":     "proj1-stg",
		"app":   "proj1-stg",
		"i":     "docker-registry.example.com/tsuru/app-proj1-dev:v938",
		"image": "docker-registry.example.com/tsuru/app-proj1-dev:v938",
	}
	if !reflect.DeepEqual(flags, expectedFlags) {
		t.Errorf("wrong flags used\nwant %#v\ngot  %#v", expectedFlags, flags)
	}
}

func TestProjectDeployPromoteFailToGetLastDeploy(t *testing.T) {
	fakeServer := newFakeServer(t)
	defer fakeServer.stop()
	fakeServer.prepareResponse(preparedResponse{
		method:  http.MethodGet,
		path:    "/apps?name=" + url.QueryEscape("^proj1"),
		code:    http.StatusOK,
		payload: []byte(listOfApps),
	})
	envNames := []string{"dev", "qa", "stage", "prod"}
	for _, envName := range envNames {
		fakeServer.prepareResponse(preparedResponse{
			method:  http.MethodGet,
			path:    "/apps/proj1-" + envName,
			code:    http.StatusOK,
			payload: []byte(appInfo1),
		})
	}
	fakeServer.prepareResponse(preparedResponse{
		method:  http.MethodGet,
		path:    "/apps/proj1-dev",
		code:    http.StatusOK,
		payload: []byte(appInfo1),
	})
	fakeServer.prepareResponse(preparedResponse{
		method:  http.MethodGet,
		path:    "/deploys?limit=1&app=proj1-dev",
		code:    http.StatusInternalServerError,
		payload: []byte("something went wrong"),
	})
	cleanup, err := setupFakeConfig(fakeServer.url(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectDeploy
	ctx := cmd.Context{}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	c.Flags().Parse(true, []string{"-n", "proj1", "-e", "stg", "-p", "dev"})
	err = c.Run(&ctx, client)
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
}

func TestProjectDeployListFlags(t *testing.T) {
	oldCommand := tsuruDeployListCommand
	defer func() { tsuruDeployListCommand = oldCommand }()
	fakeCommand := fakeTsuruCommand{FlaggedCommand: &client.AppDeployList{}}
	tsuruDeployListCommand = &fakeCommand
	var c projectDeployList
	ctx := cmd.Context{}
	c.Flags().Parse(true, []string{"-n", "myproj", "-e", "dev"})
	err := c.Run(&ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	expectedFlags := map[string]string{
		"a":   "myproj-dev",
		"app": "myproj-dev",
	}
	if gotFlags := fakeCommand.inputFlags(); !reflect.DeepEqual(gotFlags, expectedFlags) {
		t.Errorf("wrong flags sent to app-deploy\ngot  %#v\nwant %#v", gotFlags, expectedFlags)
	}
}

func TestProjectDeployListMissingParams(t *testing.T) {
	var tests = []struct {
		testCase string
		flags    []string
		args     []string
		errMsg   string
	}{
		{
			"missing project name",
			nil,
			nil,
			"please provide the project name and the environment",
		},
		{
			"missing env name",
			[]string{"-n", "myproj"},
			nil,
			"please provide the project name and the environment",
		},
	}
	for _, test := range tests {
		t.Run(test.testCase, func(t *testing.T) {
			var c projectDeployList
			ctx := cmd.Context{Args: test.args}
			c.Flags().Parse(true, test.flags)
			err := c.Run(&ctx, nil)
			if err == nil {
				t.Fatal("unexpected <nil> error")
			}
			if err.Error() != test.errMsg {
				t.Errorf("wrong error message\ngot  %q\nwant %q", err.Error(), test.errMsg)
			}
		})
	}
}

type fakeTsuruCommand struct {
	flags  map[string]gnuflag.Value
	ctx    *cmd.Context
	client *cmd.Client
	cmd.FlaggedCommand
	called bool
}

func (c *fakeTsuruCommand) Run(ctx *cmd.Context, cli *cmd.Client) error {
	c.client = cli
	c.ctx = ctx
	c.called = true
	return nil
}

func (c *fakeTsuruCommand) Flags() *gnuflag.FlagSet {
	c.flags = make(map[string]gnuflag.Value)
	fs := gnuflag.NewFlagSet("app-deploy", gnuflag.ExitOnError)
	c.FlaggedCommand.Flags().VisitAll(func(f *gnuflag.Flag) {
		fs.Var(f.Value, f.Name, f.Usage)
		c.flags[f.Name] = f.Value
	})
	return fs
}

func (c *fakeTsuruCommand) inputFlags() map[string]string {
	r := make(map[string]string, len(c.flags))
	for n, v := range c.flags {
		if value := v.String(); value != "" {
			r[n] = v.String()
		}
	}
	return r
}
