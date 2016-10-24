// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"reflect"
	"testing"

	"github.com/tsuru/gnuflag"
	"github.com/tsuru/tsuru-client/tsuru/client"
	"github.com/tsuru/tsuru/cmd"
)

func TestProjectDeployFlags(t *testing.T) {
	oldCommand := tsuruDeployCommand
	defer func() { tsuruDeployCommand = oldCommand }()
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
			c.Flags().Parse(true, test.flags)
			err := c.Run(&ctx, nil)
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
			"missing deploy options",
			[]string{"-n", "myproj", "-e", "dev"},
			nil,
			"please specify either the image, version or the list of files/directories to upload",
		},
		{
			"specifying all options",
			[]string{"-n", "myproj", "-e", "dev", "-i", "someimage", "-v", "someversion"},
			[]string{"."},
			"please specify only one of the image, version or the list of files/directories to upload",
		},
		{
			"specifying image and version",
			[]string{"-n", "myproj", "-e", "dev", "-i", "someimage", "-v", "someversion"},
			nil,
			"please specify only one of the image, version or the list of files/directories to upload",
		},
	}
	for _, test := range tests {
		t.Run(test.testCase, func(t *testing.T) {
			var c projectDeploy
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
