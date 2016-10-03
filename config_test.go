// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/tsuru/tsuru/cmd"
)

func TestProjectConfigSetInfo(t *testing.T) {
	info := (&projectConfigSet{}).Info()
	if info == nil {
		t.Fatal("unexpected <nil> info")
	}
	if info.Name != "config-set" {
		t.Errorf("wrong name. want %q. got %q", "config-set", info.Name)
	}
	if info.MinArgs != 1 {
		t.Errorf("wrong min args. want 1. got %d", info.MinArgs)
	}
}

func TestProjectConfigSet(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	appNames := []string{"proj1-dev", "proj1-stage", "proj1-prod"}
	for _, appName := range appNames {
		server.prepareResponse(preparedResponse{
			method:  "POST",
			path:    "/apps/" + appName + "/env",
			code:    http.StatusOK,
			payload: []byte("{}"),
		})
	}
	cleanup, err := setupFakeTarget(server.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectConfigSet
	err = c.Flags().Parse(true, []string{
		"-n", "proj1",
		"-e", "dev,stage,prod",
		"-p", "--no-restart",
	})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{"USER_NAME=root", "USER_PASSWORD=r00t", `PREFERRED_TEAM="some nice team"`},
	}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := `setting config vars in environment "dev"... ok
setting config vars in environment "stage"... ok
setting config vars in environment "prod"... ok
`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output\nwant:\n%s\ngot:\n%s", expectedOutput, stdout.String())
	}
}

func TestProjectConfigSetMissingName(t *testing.T) {
	var c projectConfigSet
	err := c.Flags().Parse(true, []string{"--no-restart"})
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
	expectedMsg := "please provide the name of the project"
	if err.Error() != expectedMsg {
		t.Errorf("wrong error message\nwant %q\ngot  %q", expectedMsg, err.Error())
	}
}

func TestProjectConfigSetInvalidFormat(t *testing.T) {
	var c projectConfigSet
	err := c.Flags().Parse(true, []string{"-n", "proj1"})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr, Args: []string{"ENV"}}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = c.Run(&ctx, client)
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
	expectedMsg := "configuration vars must be specified in the form NAME=value"
	if err.Error() != expectedMsg {
		t.Errorf("wrong error message\nwant %q\ngot  %q", expectedMsg, err.Error())
	}
}

func TestProjectConfigSetErrorInOneOfTheApps(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	server.prepareResponse(preparedResponse{
		method:  "POST",
		path:    "/apps/proj1-stage/env",
		code:    http.StatusInternalServerError,
		payload: []byte("something went wrong"),
	})
	appNames := []string{"proj1-dev", "proj1-prod"}
	for _, appName := range appNames {
		server.prepareResponse(preparedResponse{
			method:  "POST",
			path:    "/apps/" + appName + "/env",
			code:    http.StatusOK,
			payload: []byte("{}"),
		})
	}
	cleanup, err := setupFakeTarget(server.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectConfigSet
	err = c.Flags().Parse(true, []string{
		"-n", "proj1",
		"-e", "dev,stage,prod",
		"-p", "--no-restart",
	})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{"USER_NAME=root", "USER_PASSWORD=r00t", `PREFERRED_TEAM="some nice team"`},
	}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = c.Run(&ctx, client)
	if err == nil {
		t.Fatal("got unexpected <nil> error")
	}
	expectedOutput := `setting config vars in environment "dev"... ok
setting config vars in environment "stage"... failed
`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output\nwant:\n%s\ngot:\n%s", expectedOutput, stdout.String())
	}
}

func TestProjectConfigSetAppNotFound(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	appNames := []string{"proj1-dev", "proj1-prod"}
	for _, appName := range appNames {
		server.prepareResponse(preparedResponse{
			method:  "POST",
			path:    "/apps/" + appName + "/env",
			code:    http.StatusOK,
			payload: []byte("{}"),
		})
	}
	cleanup, err := setupFakeTarget(server.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectConfigSet
	err = c.Flags().Parse(true, []string{
		"-n", "proj1",
		"-e", "dev,stage,prod",
		"-p", "--no-restart",
	})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{"USER_NAME=root", "USER_PASSWORD=r00t", `PREFERRED_TEAM="some nice team"`},
	}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := `setting config vars in environment "dev"... ok
setting config vars in environment "stage"... not found
setting config vars in environment "prod"... ok
`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output\nwant:\n%s\ngot:\n%s", expectedOutput, stdout.String())
	}
}

func TestProjectConfigGetInfo(t *testing.T) {
	info := (&projectConfigGet{}).Info()
	if info == nil {
		t.Fatal("unexpected <nil> info")
	}
	if info.Name != "config-get" {
		t.Errorf("wrong name. want %q. got %q", "config-get", info.Name)
	}
}

func TestProjectConfigGet(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	appNames := []string{"proj1-dev", "proj1-stage", "proj1-prod"}
	for _, appName := range appNames {
		rawPayload := []map[string]interface{}{
			{
				"name":   "APP_NAME",
				"value":  appName,
				"public": true,
			},
			{
				"name":   "USER_NAME",
				"value":  "root",
				"public": true,
			},
			{
				"name":   "USER_PASSWORD",
				"value":  "r00t",
				"public": false,
			},
		}
		payload, _ := json.Marshal(rawPayload)
		server.prepareResponse(preparedResponse{
			method:  "GET",
			path:    "/apps/" + appName + "/env",
			code:    http.StatusOK,
			payload: payload,
		})
	}
	cleanup, err := setupFakeTarget(server.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectConfigGet
	err = c.Flags().Parse(true, []string{
		"-n", "proj1",
		"-e", "dev,stage,prod",
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
	expectedOutput := `config vars in "dev":

 APP_NAME=proj1-dev
 USER_NAME=root
 USER_PASSWORD=*** (private config)


config vars in "stage":

 APP_NAME=proj1-stage
 USER_NAME=root
 USER_PASSWORD=*** (private config)


config vars in "prod":

 APP_NAME=proj1-prod
 USER_NAME=root
 USER_PASSWORD=*** (private config)


`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output\nwant:\n%q\ngot:\n%q", expectedOutput, stdout.String())
	}
}

func TestProjectConfigGetMissingName(t *testing.T) {
	var c projectConfigGet
	err := c.Flags().Parse(true, []string{"-e", "dev,stage,prod"})
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
	expectedMsg := "please provide the name of the project"
	if err.Error() != expectedMsg {
		t.Errorf("wrong error message\nwant %q\ngot  %q", expectedMsg, err.Error())
	}
}

func TestProjectConfigGetAppNotFound(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	appNames := []string{"proj1-dev", "proj1-prod"}
	for _, appName := range appNames {
		rawPayload := []map[string]interface{}{
			{
				"name":   "APP_NAME",
				"value":  appName,
				"public": true,
			},
			{
				"name":   "USER_NAME",
				"value":  "root",
				"public": true,
			},
			{
				"name":   "USER_PASSWORD",
				"value":  "r00t",
				"public": false,
			},
		}
		payload, _ := json.Marshal(rawPayload)
		server.prepareResponse(preparedResponse{
			method:  "GET",
			path:    "/apps/" + appName + "/env",
			code:    http.StatusOK,
			payload: payload,
		})
	}
	cleanup, err := setupFakeTarget(server.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectConfigGet
	err = c.Flags().Parse(true, []string{
		"-n", "proj1",
		"-e", "dev,stage,prod",
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
	expectedOutput := `config vars in "dev":

 APP_NAME=proj1-dev
 USER_NAME=root
 USER_PASSWORD=*** (private config)


config vars in "prod":

 APP_NAME=proj1-prod
 USER_NAME=root
 USER_PASSWORD=*** (private config)


`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output\nwant:\n%q\ngot:\n%q", expectedOutput, stdout.String())
	}
	expectedStderr := `WARNING: project not found in environment "stage"` + "\n"
	if stderr.String() != expectedStderr {
		t.Errorf("wrong error output\nwant:\n%q\ngot:\n%q", expectedStderr, stderr.String())
	}
}

func TestProjectConfigGetErrorInOneOfTheApps(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	server.prepareResponse(preparedResponse{
		method:  "GET",
		path:    "/apps/proj1-stage/env",
		code:    http.StatusInternalServerError,
		payload: []byte("something went wrong"),
	})
	appNames := []string{"proj1-dev", "proj1-prod"}
	for _, appName := range appNames {
		rawPayload := []map[string]interface{}{
			{
				"name":   "APP_NAME",
				"value":  appName,
				"public": true,
			},
			{
				"name":   "USER_NAME",
				"value":  "root",
				"public": true,
			},
			{
				"name":   "USER_PASSWORD",
				"value":  "r00t",
				"public": false,
			},
		}
		payload, _ := json.Marshal(rawPayload)
		server.prepareResponse(preparedResponse{
			method:  "GET",
			path:    "/apps/" + appName + "/env",
			code:    http.StatusOK,
			payload: payload,
		})
	}
	cleanup, err := setupFakeTarget(server.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectConfigGet
	err = c.Flags().Parse(true, []string{
		"-n", "proj1",
		"-e", "dev,stage,prod",
	})
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
	expectedOutput := `config vars in "dev":

 APP_NAME=proj1-dev
 USER_NAME=root
 USER_PASSWORD=*** (private config)


`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output\nwant:\n%q\ngot:\n%q", expectedOutput, stdout.String())
	}
}

func TestProjectConfigUnsetInfo(t *testing.T) {
	info := (&projectConfigUnset{}).Info()
	if info == nil {
		t.Fatal("unexpected <nil> info")
	}
	if info.Name != "config-unset" {
		t.Errorf("wrong name. want %q. got %q", "config-unset", info.Name)
	}
	if info.MinArgs != 1 {
		t.Errorf("wrong min args. want 1. got %d", info.MinArgs)
	}
}

func TestProjectConfigUnset(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	appNames := []string{"proj1-dev", "proj1-stage", "proj1-prod"}
	for _, appName := range appNames {
		server.prepareResponse(preparedResponse{
			method:   "DELETE",
			path:     "/apps/" + appName + "/env",
			code:     http.StatusOK,
			payload:  []byte("{}"),
			ignoreQS: true,
		})
	}
	cleanup, err := setupFakeTarget(server.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectConfigUnset
	err = c.Flags().Parse(true, []string{
		"-n", "proj1",
		"-e", "dev,stage,prod",
		"--no-restart",
	})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{"USER_NAME", "USER_PASSWORD", "PREFERRED_TEAM"},
	}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := `unsetting config vars from environment "dev"... ok
unsetting config vars from environment "stage"... ok
unsetting config vars from environment "prod"... ok
`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output\nwant:\n%s\ngot:\n%s", expectedOutput, stdout.String())
	}
}

func TestProjectConfigUnsetMissingName(t *testing.T) {
	var c projectConfigUnset
	err := c.Flags().Parse(true, []string{"-e", "dev,stage,prod"})
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
	expectedMsg := "please provide the name of the project"
	if err.Error() != expectedMsg {
		t.Errorf("wrong error message\nwant %q\ngot  %q", expectedMsg, err.Error())
	}
}

func TestProjectConfigUnsetAppNotFound(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	appNames := []string{"proj1-dev", "proj1-prod"}
	for _, appName := range appNames {
		server.prepareResponse(preparedResponse{
			method:   "DELETE",
			path:     "/apps/" + appName + "/env",
			code:     http.StatusOK,
			payload:  []byte("{}"),
			ignoreQS: true,
		})
	}
	cleanup, err := setupFakeTarget(server.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectConfigUnset
	err = c.Flags().Parse(true, []string{
		"-n", "proj1",
		"-e", "dev,stage,prod",
		"--no-restart",
	})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{"USER_NAME", "USER_PASSWORD", "PREFERRED_TEAM"},
	}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := `unsetting config vars from environment "dev"... ok
unsetting config vars from environment "stage"... not found
unsetting config vars from environment "prod"... ok
`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output\nwant:\n%s\ngot:\n%s", expectedOutput, stdout.String())
	}
}

func TestProjectConfigError(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	server.prepareResponse(preparedResponse{
		method:   "DELETE",
		path:     "/apps/proj1-stage/env",
		code:     http.StatusInternalServerError,
		payload:  []byte("something went wrong"),
		ignoreQS: true,
	})
	appNames := []string{"proj1-dev", "proj1-prod"}
	for _, appName := range appNames {
		server.prepareResponse(preparedResponse{
			method:   "DELETE",
			path:     "/apps/" + appName + "/env",
			code:     http.StatusOK,
			payload:  []byte("{}"),
			ignoreQS: true,
		})
	}
	cleanup, err := setupFakeTarget(server.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectConfigUnset
	err = c.Flags().Parse(true, []string{
		"-n", "proj1",
		"-e", "dev,stage,prod",
		"--no-restart",
	})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{"USER_NAME", "USER_PASSWORD", "PREFERRED_TEAM"},
	}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = c.Run(&ctx, client)
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
	expectedOutput := `unsetting config vars from environment "dev"... ok
unsetting config vars from environment "stage"... failed
`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output\nwant:\n%s\ngot:\n%s", expectedOutput, stdout.String())
	}
}
