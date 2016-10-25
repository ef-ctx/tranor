// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/tsuru/gnuflag"
	"github.com/tsuru/tsuru/cmd"
)

func TestCommaSeparatedFlag(t *testing.T) {
	var value gnuflag.Value
	value = &commaSeparatedFlag{}
	input := "dev,qa,staging,prod"
	err := value.Set(input)
	if err != nil {
		t.Fatal(err)
	}
	output := value.String()
	if output != input {
		t.Errorf("wrong output.\nwant %q\ngot  %q", input, output)
	}
	values := value.(*commaSeparatedFlag).Values()
	expectedValues := []string{"dev", "qa", "staging", "prod"}
	if !reflect.DeepEqual(values, expectedValues) {
		t.Errorf("wrong value list\nwant %#v\ngot  %#v", expectedValues, values)
	}
}

func TestCommaSeparatedFlagValidate(t *testing.T) {
	var tests = []struct {
		testCase      string
		flagInput     string
		validValues   []string
		expectedError string
	}{
		{
			"valid values",
			"dev,qa,staging,prod",
			[]string{"dev", "qa", "staging", "demo", "prod"},
			"",
		},
		{
			"one invalid value",
			"dev,qa,staging,prod",
			[]string{"dev", "qa", "demo", "prod"},
			"invalid values: staging (valid options are: dev, qa, demo, prod)",
		},
		{
			"all invalid values",
			"dev,qa,staging,prod",
			nil,
			"invalid values: dev, qa, staging, prod (valid options are: )",
		},
	}
	for _, test := range tests {
		var value commaSeparatedFlag
		err := value.Set(test.flagInput)
		if err != nil {
			t.Errorf("%s: %s", test.testCase, err)
			continue
		}
		err = value.validate(test.validValues)
		if err == nil {
			err = errors.New("")
		}
		if err.Error() != test.expectedError {
			t.Errorf("wrong error message\nwant %q\ngot  %q", test.expectedError, err.Error())
		}
	}
}

func TestProjectCreateInfo(t *testing.T) {
	info := (&projectCreate{}).Info()
	if info == nil {
		t.Fatal("unexpected <nil> info")
	}
	if info.Name != "project-create" {
		t.Errorf("wrong name. Want %q. Got %q", "project-create", info.Name)
	}
}

func TestProjectCreateNoRepo(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	server.prepareResponse(preparedResponse{
		method:  "POST",
		path:    "/apps",
		code:    http.StatusCreated,
		payload: []byte(`{}`),
	})
	appNames := []string{"superproj-dev", "superproj-prod"}
	for _, appName := range appNames {
		server.prepareResponse(preparedResponse{
			method:  "POST",
			path:    "/apps/" + appName + "/cname",
			code:    http.StatusOK,
			payload: []byte(`{}`),
		})
	}
	cleanup, err := setupFakeConfig(server.url(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
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
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := `successfully created the project "superproj"!
`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output returned\nWant: %s\nGot:  %s", expectedOutput, stdout.String())
	}
}

func TestProjectCreateMissingParams(t *testing.T) {
	var tests = []struct {
		testCase string
		flags    []string
	}{
		{
			"missing name",
			[]string{"-l", "python"},
		},
		{
			"missing platform",
			[]string{"-n", "superproj"},
		},
		{
			"missing all flags",
			[]string{},
		},
	}
	cleanup, err := setupFakeConfig("http://localhost:8080", "")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	for _, test := range tests {
		var c projectCreate
		err = c.Flags().Parse(true, test.flags)
		if err != nil {
			t.Errorf("%s: %s", test.testCase, err)
			continue
		}
		var stdout, stderr bytes.Buffer
		ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
		client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
		err = c.Run(&ctx, client)
		if err == nil {
			t.Errorf("%s: unexpected <nil> error", test.testCase)
			continue
		}
		expectedMsg := "please provide the name and the platform"
		if err.Error() != expectedMsg {
			t.Errorf("%s: wrong error message\nwant %q\ngot  %q", test.testCase, expectedMsg, err.Error())
		}
	}
}

func TestProjectCreateInvalidEnv(t *testing.T) {
	cleanup, err := setupFakeConfig("http://localhost:8080", "")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectCreate
	err = c.Flags().Parse(true, []string{
		"-n", "superproj",
		"-l", "python",
		"-t", "myteam",
		"-p", "medium",
		"-e", "dev,production",
	})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	err = c.Run(&ctx, nil)
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
}

func TestProjectCreateFailToLoadConfig(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	var c projectCreate
	err = c.Flags().Parse(true, []string{
		"-n", "superproj",
		"-l", "python",
		"-t", "myteam",
		"-p", "medium",
	})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	err = c.Run(&ctx, nil)
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
}

func TestProjectCreateFailToSetCNames(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	server.prepareResponse(preparedResponse{
		method:  "POST",
		path:    "/apps",
		code:    http.StatusCreated,
		payload: []byte(`{"repository_url":"git@git.example.com:superproj-dev.git"}`),
	})
	server.prepareResponse(preparedResponse{
		method:  "POST",
		path:    "/apps/superproj-dev/cname",
		code:    http.StatusOK,
		payload: []byte(`{}`),
	})
	server.prepareResponse(preparedResponse{
		method: "DELETE",
		path:   "/apps/superproj-prod",
		code:   http.StatusOK,
	})
	server.prepareResponse(preparedResponse{
		method:  "POST",
		path:    "/apps/superproj-prod/cname",
		code:    http.StatusInternalServerError,
		payload: []byte(`{}`),
	})
	cleanup, err := setupFakeConfig(server.url(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
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
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	err = c.Run(&ctx, cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{}))
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
	expectedMethods := []string{"POST", "POST", "POST", "POST", "POST", "POST", "DELETE", "DELETE"}
	expectedPaths := []string{
		"/1.0/apps", "/1.0/apps/superproj-dev/env",
		"/1.0/apps", "/1.0/apps/superproj-prod/env",
		"/1.0/apps/superproj-dev/cname",
		"/1.0/apps/superproj-prod/cname",
		"/1.0/apps/superproj-dev",
		"/1.0/apps/superproj-prod",
	}
	if len(server.reqs) != len(expectedPaths) {
		t.Fatalf("wrong number of requests sent to the server. Want %d. Got %d", len(expectedPaths), len(server.reqs))
	}
	for i, req := range server.reqs {
		if req.Method != expectedMethods[i] {
			t.Errorf("wrong method. Want %q. Got %q", expectedMethods[i], req.Method)
		}
		if req.URL.Path != expectedPaths[i] {
			t.Errorf("wrong path. Want %q. Got %q", expectedPaths[i], req.URL.Path)
		}
	}
}

func TestProjectUpdateMissingName(t *testing.T) {
	var c projectUpdate
	err := c.Flags().Parse(true, []string{
		"--add-envs", "qa,stage",
		"--remove-envs", "dev",
	})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = c.Run(&ctx, client)
	if err == nil {
		t.Fatal("got unexpected <nil> error")
	}
	expectedMsg := "please provide the name of the project"
	if err.Error() != expectedMsg {
		t.Errorf("wrong error message\nwant %q\ngot  %q", expectedMsg, err.Error())
	}
}

func TestProjectUpdateNotFound(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	server.prepareResponse(preparedResponse{
		method: "GET",
		path:   "/apps?name=" + url.QueryEscape("^proj3"),
		code:   http.StatusNoContent,
	})
	cleanup, err := setupFakeConfig(server.url(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectUpdate
	err = c.Flags().Parse(true, []string{
		"-n", "proj3",
		"-d", "updated project description",
		"-t", "superteam",
		"-p", "huge",
		"--add-envs", "qa,stage",
		"--remove-envs", "dev",
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
	expectedMsg := "project not found"
	if err.Error() != expectedMsg {
		t.Errorf("wrong error message\nwant %q\ngot  %q", expectedMsg, err.Error())
	}
}

func TestProjectUpdateFailToCreateNewApps(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	server.prepareResponse(preparedResponse{
		method:  "GET",
		path:    "/apps?name=" + url.QueryEscape("^proj3"),
		code:    http.StatusOK,
		payload: []byte(listOfApps),
	})
	appRespMap := map[string][]byte{
		"proj3-dev":  []byte(appInfo1),
		"proj3-prod": []byte(appInfo2),
	}
	for appName, payload := range appRespMap {
		server.prepareResponse(preparedResponse{
			method:  "GET",
			path:    "/apps/" + appName,
			code:    http.StatusOK,
			payload: payload,
		})
	}
	server.prepareResponse(preparedResponse{
		method:  "POST",
		path:    "/apps",
		code:    http.StatusInternalServerError,
		payload: []byte(`{}`),
	})
	cleanup, err := setupFakeConfig(server.url(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectUpdate
	err = c.Flags().Parse(true, []string{
		"-n", "proj3",
		"-d", "updated project description",
		"-t", "superteam",
		"-p", "huge",
		"--add-envs", "qa,stage",
		"--remove-envs", "dev",
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
}

func TestProjectUpdateFailToSetCNames(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	server.prepareResponse(preparedResponse{
		method:  "GET",
		path:    "/apps?name=" + url.QueryEscape("^proj3"),
		code:    http.StatusOK,
		payload: []byte(listOfApps),
	})
	appRespMap := map[string][]byte{
		"proj3-dev":  []byte(appInfo1),
		"proj3-prod": []byte(appInfo2),
	}
	for appName, payload := range appRespMap {
		server.prepareResponse(preparedResponse{
			method:  "GET",
			path:    "/apps/" + appName,
			code:    http.StatusOK,
			payload: payload,
		})
	}
	server.prepareResponse(preparedResponse{
		method:  "POST",
		path:    "/apps",
		code:    http.StatusCreated,
		payload: []byte(`{}`),
	})
	server.prepareResponse(preparedResponse{
		method: "POST",
		path:   "/apps/proj3-qa/cname",
		code:   http.StatusOK,
	})
	server.prepareResponse(preparedResponse{
		method: "POST",
		path:   "/apps/proj3-stage/cname",
		code:   http.StatusInternalServerError,
	})
	cleanup, err := setupFakeConfig(server.url(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectUpdate
	err = c.Flags().Parse(true, []string{
		"-n", "proj3",
		"-d", "updated project description",
		"-t", "superteam",
		"-p", "huge",
		"--add-envs", "qa,stage",
		"--remove-envs", "dev",
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
}

func TestProjectUpdateFailToUpdate(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	server.prepareResponse(preparedResponse{
		method:  "GET",
		path:    "/apps?name=" + url.QueryEscape("^proj3"),
		code:    http.StatusOK,
		payload: []byte(listOfApps),
	})
	appRespMap := map[string][]byte{
		"proj3-dev":  []byte(appInfo1),
		"proj3-prod": []byte(appInfo2),
	}
	for appName, payload := range appRespMap {
		server.prepareResponse(preparedResponse{
			method:  "GET",
			path:    "/apps/" + appName,
			code:    http.StatusOK,
			payload: payload,
		})
	}
	server.prepareResponse(preparedResponse{
		method: "DELETE",
		path:   "/apps/proj3-dev",
		code:   http.StatusOK,
	})
	server.prepareResponse(preparedResponse{
		method: "PUT",
		path:   "/apps/proj3-prod",
		code:   http.StatusInternalServerError,
	})
	server.prepareResponse(preparedResponse{
		method:  "POST",
		path:    "/apps",
		code:    http.StatusCreated,
		payload: []byte(`{}`),
	})
	server.prepareResponse(preparedResponse{
		method: "POST",
		path:   "/apps/proj3-qa/cname",
		code:   http.StatusOK,
	})
	server.prepareResponse(preparedResponse{
		method: "POST",
		path:   "/apps/proj3-stage/cname",
		code:   http.StatusOK,
	})
	cleanup, err := setupFakeConfig(server.url(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectUpdate
	err = c.Flags().Parse(true, []string{
		"-n", "proj3",
		"-d", "updated project description",
		"-t", "superteam",
		"-p", "huge",
		"--add-envs", "qa,stage",
		"--remove-envs", "dev",
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
}

func TestProjectUpdateInvalidNewEnv(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	server.prepareResponse(preparedResponse{
		method:  "GET",
		path:    "/apps?name=" + url.QueryEscape("^proj3"),
		code:    http.StatusOK,
		payload: []byte(listOfApps),
	})
	appRespMap := map[string][]byte{
		"proj3-dev":  []byte(appInfo1),
		"proj3-prod": []byte(appInfo2),
	}
	for appName, payload := range appRespMap {
		server.prepareResponse(preparedResponse{
			method:  "GET",
			path:    "/apps/" + appName,
			code:    http.StatusOK,
			payload: payload,
		})
	}
	cleanup, err := setupFakeConfig(server.url(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectUpdate
	err = c.Flags().Parse(true, []string{
		"-n", "proj3",
		"--add-envs", "qa2,stg",
		"--remove-envs", "dev",
	})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = c.Run(&ctx, client)
	if err == nil {
		t.Fatal("unexpected <nil> errror")
	}
}

func TestProjectUpdateDuplicateEnv(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	server.prepareResponse(preparedResponse{
		method:  "GET",
		path:    "/apps?name=" + url.QueryEscape("^proj3"),
		code:    http.StatusOK,
		payload: []byte(listOfApps),
	})
	appRespMap := map[string][]byte{
		"proj3-dev":  []byte(appInfo1),
		"proj3-prod": []byte(appInfo2),
	}
	for appName, payload := range appRespMap {
		server.prepareResponse(preparedResponse{
			method:  "GET",
			path:    "/apps/" + appName,
			code:    http.StatusOK,
			payload: payload,
		})
	}
	cleanup, err := setupFakeConfig(server.url(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectUpdate
	err = c.Flags().Parse(true, []string{
		"-n", "proj3",
		"--add-envs", "qa,prod",
		"--remove-envs", "dev",
	})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = c.Run(&ctx, client)
	if err == nil {
		t.Fatal("unexpected <nil> errror")
	}
	expectedMsg := `env "prod" is already defined in this project`
	if err.Error() != expectedMsg {
		t.Errorf("wrong error message\nwant %q\ngot  %q", expectedMsg, err.Error())
	}
}

func TestProjectUpdateInvalidRemoveEnvs(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	server.prepareResponse(preparedResponse{
		method:  "GET",
		path:    "/apps?name=" + url.QueryEscape("^proj3"),
		code:    http.StatusOK,
		payload: []byte(listOfApps),
	})
	appRespMap := map[string][]byte{
		"proj3-dev":  []byte(appInfo1),
		"proj3-prod": []byte(appInfo2),
	}
	for appName, payload := range appRespMap {
		server.prepareResponse(preparedResponse{
			method:  "GET",
			path:    "/apps/" + appName,
			code:    http.StatusOK,
			payload: payload,
		})
	}
	cleanup, err := setupFakeConfig(server.url(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectUpdate
	err = c.Flags().Parse(true, []string{
		"-n", "proj3",
		"--add-envs", "qa",
		"--remove-envs", "stage",
	})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = c.Run(&ctx, client)
	if err == nil {
		t.Fatal("unexpected <nil> errror")
	}
	expectedMsg := `env "stage" is not defined in this project`
	if err.Error() != expectedMsg {
		t.Errorf("wrong error message\nwant %q\ngot  %q", expectedMsg, err.Error())
	}
}

func TestProjectUpdateFailToLoadConfig(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	var c projectUpdate
	err = c.Flags().Parse(true, []string{
		"-n", "superproj",
		"-t", "myteam",
		"-p", "medium",
	})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	err = c.Run(&ctx, nil)
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
}

func TestRemoveProjectNoConfirmation(t *testing.T) {
	cleanup, err := setupFakeConfig("http://localhost:8080", "")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectRemove
	err = c.Flags().Parse(true, []string{"-n", "myproj"})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr, Stdin: strings.NewReader("n\n")}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRemoveProjectValidation(t *testing.T) {
	var c projectRemove
	err := c.Flags().Parse(true, []string{})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr, Stdin: strings.NewReader("n\n")}
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

func TestRemoveProjectConfigurationIssue(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	var c projectRemove
	err = c.Flags().Parse(true, []string{"-n", "superproj"})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	err = c.Run(&ctx, nil)
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
}

func TestProjectInfoErrorToListApps(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	server.prepareResponse(preparedResponse{
		method:  "GET",
		path:    "/apps?name=proj1",
		code:    http.StatusInternalServerError,
		payload: []byte("something went wrong"),
	})
	cleanup, err := setupFakeConfig(server.url(), "")
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
		t.Fatal("unexpected <nil> error")
	}
}

func TestProjectInfoConfigIssue(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	var c projectInfo
	err = c.Flags().Parse(true, []string{"-n", "proj1"})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	err = c.Run(&ctx, nil)
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
}

func TestProjectInfoMissingName(t *testing.T) {
	var c projectInfo
	err := c.Flags().Parse(true, []string{})
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

func TestProjectEnvInfoMissingName(t *testing.T) {
	var c projectEnvInfo
	err := c.Flags().Parse(true, []string{"-e", "prod"})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	cli := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = c.Run(&ctx, cli)
	if err == nil {
		t.Fatal("got unexpected <nil> error")
	}
}

func TestProjectEnvInfoMissingEnv(t *testing.T) {
	var c projectEnvInfo
	err := c.Flags().Parse(true, []string{"-n", "proj1"})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	cli := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = c.Run(&ctx, cli)
	if err == nil {
		t.Fatal("got unexpected <nil> error")
	}
}

func TestProjectListErrorToListApps(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	server.prepareResponse(preparedResponse{
		method:  "GET",
		path:    "/apps",
		code:    http.StatusInternalServerError,
		payload: []byte("something went wrong"),
	})
	cleanup, err := setupFakeConfig(server.url(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var c projectList
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = c.Run(&ctx, client)
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
}

func TestProjectListNoConfiguration(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	var c projectList
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	err = c.Run(&ctx, nil)
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
}

func setupFakeConfig(target, token string) (func(), error) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	os.Setenv("HOME", dir)
	os.Unsetenv("TSURU_HOST")
	os.Setenv("TSURU_TOKEN", token)
	config := Config{
		Target:   target,
		Registry: "docker-registry.example.com",
		Environments: []Environment{
			{
				Name:      "dev",
				DNSSuffix: "dev.example.com",
			},
			{
				Name:      "qa",
				DNSSuffix: "qa.example.com",
			},
			{
				Name:      "stage",
				DNSSuffix: "stage.example.com",
			},
			{
				Name:      "prod",
				DNSSuffix: "example.com",
			},
		},
	}
	err = config.writeTarget()
	if err != nil {
		os.RemoveAll(dir)
		return nil, err
	}
	return func() { os.RemoveAll(dir) }, writeConfigFile(&config)
}
