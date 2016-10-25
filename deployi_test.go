// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/tsuru/tsuru-client/tsuru/client"
	"github.com/tsuru/tsuru/cmd"
)

func TestProjectDeployFlags(t *testing.T) {
	cleanup := createTestProject("myproj", t)
	oldCommand := tsuruDeployCommand
	defer func() {
		tsuruDeployCommand = oldCommand
		cleanup()
	}()
	var tests = []struct {
		testCase      string
		flags         []string
		args          []string
		expectedFlags map[string]string
		expectedArgs  []string
	}{
		{
			"upload-based deploy",
			[]string{"-n", "myproj", "-e", "dev"},
			[]string{"."},
			map[string]string{
				"app": "myproj-dev",
				"a":   "myproj-dev",
			},
			[]string{"."},
		},
		{
			"image-based deploy",
			[]string{"-n", "myproj", "-e", "dev", "-i", "some/image"},
			nil,
			map[string]string{
				"a":     "myproj-dev",
				"app":   "myproj-dev",
				"i":     "some/image",
				"image": "some/image",
			},
			nil,
		},
	}
	for _, test := range tests {
		t.Run(test.testCase, func(t *testing.T) {
			fakeCommand := fakeTsuruCommand{FlaggedCommand: &client.AppDeploy{}}
			tsuruDeployCommand = &fakeCommand
			var c projectDeploy
			ctx := cmd.Context{Args: test.args}
			client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
			c.Flags().Parse(true, test.flags)
			err := c.Run(&ctx, client)
			if err != nil {
				t.Fatal(err)
			}
			if gotFlags := fakeCommand.inputFlags(); !reflect.DeepEqual(gotFlags, test.expectedFlags) {
				t.Errorf("wrong flags sent to app-deploy\ngot  %#v\nwant %#v", gotFlags, test.expectedFlags)
			}
		})
	}
}

func TestProjectDeployErrors(t *testing.T) {
	cleanup := createTestProject("myproj", t)
	defer cleanup()
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
		{
			"project not found",
			[]string{"-n", "someproj", "-e", "dev"},
			nil,
			"project not found",
		},
		{
			"missing deploy options",
			[]string{"-n", "myproj", "-e", "dev"},
			nil,
			"please specify either the image, parent env or the list of files/directories to upload",
		},
		{
			"specifying all options",
			[]string{"-n", "myproj", "-e", "stage", "-i", "someimage", "-p", "env"},
			[]string{"."},
			"please specify only one of the image, parent env or the list of files/directories to upload",
		},
		{
			"specifying image and version",
			[]string{"-n", "myproj", "-e", "stage", "-i", "someimage", "-p", "env"},
			nil,
			"please specify only one of the image, parent env or the list of files/directories to upload",
		},
		{
			"invalid direct deploy",
			[]string{"-n", "myproj", "-e", "stage"},
			[]string{"target/debug"},
			`can only deploy directly to "dev", use promote to deploy to other environments`,
		},
		{
			"promote from environment that hasn't been deployed",
			[]string{"-n", "myproj", "-e", "stage", "-p", "dev"},
			nil,
			`no version running in "dev"`,
		},
	}
	for _, test := range tests {
		t.Run(test.testCase, func(t *testing.T) {
			var c projectDeploy
			ctx := cmd.Context{Args: test.args}
			client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
			c.Flags().Parse(true, test.flags)
			err := c.Run(&ctx, client)
			if err == nil {
				t.Fatal("unexpected <nil> error")
			}
			if err.Error() != test.errMsg {
				t.Errorf("wrong error message\ngot  %q\nwant %q", err.Error(), test.errMsg)
			}
		})
	}
}

func createTestProject(name string, t *testing.T) func() {
	tsuruServer.reset()
	cleanup, err := setupFakeConfig(tsuruServer.url(), tsuruServer.token())
	if err != nil {
		t.Fatal(err)
	}
	config, err := loadConfigFile()
	if err != nil {
		t.Fatal(err)
	}
	client := cmd.NewClient(http.DefaultClient, &cmd.Context{}, &cmd.Manager{})
	apps, err := createApps(config.Environments, client, name, createAppOptions{
		Plan:        "medium",
		Description: "some project",
		Team:        "myteam",
		Platform:    "python",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = setCNames(apps, client, name)
	if err != nil {
		t.Fatal(err)
	}
	return cleanup
}
