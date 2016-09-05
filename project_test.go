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
	input := "dev,qa,staging,production"
	err := value.Set(input)
	if err != nil {
		t.Fatal(err)
	}
	output := value.String()
	if output != input {
		t.Errorf("wrong output.\nwant %q\ngot  %q", input, output)
	}
	values := value.(*commaSeparatedFlag).Values()
	expectedValues := []string{"dev", "qa", "staging", "production"}
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
			"dev,qa,staging,production",
			[]string{"dev", "qa", "staging", "demo", "production"},
			"",
		},
		{
			"one invalid value",
			"dev,qa,staging,production",
			[]string{"dev", "qa", "demo", "production"},
			"invalid values: staging (valid options are: dev, qa, demo, production)",
		},
		{
			"all invalid values",
			"dev,qa,staging,production",
			nil,
			"invalid values: dev, qa, staging, production (valid options are: )",
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

func TestProjectCreateDefaultEnvs(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	server.prepareResponse(preparedResponse{
		method:  "POST",
		path:    "/apps",
		code:    http.StatusCreated,
		payload: []byte(`{"repository_url":"git@git.example.com:myproj-dev.git"}`),
	})
	appNames := []string{"myproj-dev", "myproj-qa", "myproj-stage", "myproj-production"}
	for _, appName := range appNames {
		server.prepareResponse(preparedResponse{
			method:  "POST",
			path:    "/apps/" + appName + "/cname",
			code:    http.StatusOK,
			payload: []byte(`{}`),
		})
	}
	cleanup, err := setupFakeTarget(server.url())
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
Git repository: git@git.example.com:myproj-dev.git
`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output returned\nWant: %s\nGot:  %s", expectedOutput, stdout.String())
	}
	if len(server.reqs) != 8 {
		t.Fatalf("wrong number of requests sent to the server. Want 8. Got %d", len(server.reqs))
	}
	expectedPaths := []string{
		"/1.0/apps", "/1.0/apps", "/1.0/apps", "/1.0/apps",
		"/1.0/apps/myproj-dev/cname", "/1.0/apps/myproj-qa/cname",
		"/1.0/apps/myproj-stage/cname", "/1.0/apps/myproj-production/cname",
	}
	for i, req := range server.reqs {
		if req.Method != "POST" {
			t.Errorf("wrong method. Want %q. Got %q", "POST", req.Method)
		}
		if req.URL.Path != expectedPaths[i] {
			t.Errorf("wrong path. Want %q. Got %q", expectedPaths[i], req.URL.Path)
		}
	}
	expectedPayloads := []url.Values{
		{
			"name":      []string{"myproj-dev"},
			"platform":  []string{"python"},
			"plan":      []string{"medium"},
			"teamOwner": []string{"myteam"},
			"pool":      []string{"dev/dev.example.com"},
		},
		{
			"name":      []string{"myproj-qa"},
			"platform":  []string{"python"},
			"plan":      []string{"medium"},
			"teamOwner": []string{"myteam"},
			"pool":      []string{"qa/qa.example.com"},
		},
		{
			"name":      []string{"myproj-stage"},
			"platform":  []string{"python"},
			"plan":      []string{"medium"},
			"teamOwner": []string{"myteam"},
			"pool":      []string{"stage/stage.example.com"},
		},
		{
			"name":      []string{"myproj-production"},
			"platform":  []string{"python"},
			"plan":      []string{"medium"},
			"teamOwner": []string{"myteam"},
			"pool":      []string{"production/example.com"},
		},
		{"cname": []string{"myproj.dev.example.com"}},
		{"cname": []string{"myproj.qa.example.com"}},
		{"cname": []string{"myproj.stage.example.com"}},
		{"cname": []string{"myproj.example.com"}},
	}
	if len(server.payloads) != len(expectedPayloads) {
		t.Errorf("wrong number of payload. Want %d. Got %d", len(expectedPayloads), len(server.payloads))
	}
	for i, payload := range server.payloads {
		values, err := url.ParseQuery(string(payload))
		if err != nil {
			t.Errorf("invalid payload: %s", payload)
		}
		if !reflect.DeepEqual(values, expectedPayloads[i]) {
			t.Errorf("wrong payload\nWant %#v\nGot  %#v", expectedPayloads[i], values)
		}
	}
}

func TestProjectCreateSpecifyEnvs(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	server.prepareResponse(preparedResponse{
		method:  "POST",
		path:    "/apps",
		code:    http.StatusCreated,
		payload: []byte(`{"repository_url":"git@git.example.com:superproj-dev.git"}`),
	})
	appNames := []string{"superproj-dev", "superproj-production"}
	for _, appName := range appNames {
		server.prepareResponse(preparedResponse{
			method:  "POST",
			path:    "/apps/" + appName + "/cname",
			code:    http.StatusOK,
			payload: []byte(`{}`),
		})
	}
	cleanup, err := setupFakeTarget(server.url())
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
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = c.Run(&ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := `successfully created the project "superproj"!
Git repository: git@git.example.com:superproj-dev.git
`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output returned\nWant: %s\nGot:  %s", expectedOutput, stdout.String())
	}
	expectedPaths := []string{
		"/1.0/apps", "/1.0/apps",
		"/1.0/apps/superproj-dev/cname",
		"/1.0/apps/superproj-production/cname",
	}
	if len(server.reqs) != len(expectedPaths) {
		t.Fatalf("wrong number of requests sent to the server. Want %d. Got %d", len(expectedPaths), len(server.reqs))
	}
	for i, req := range server.reqs {
		if req.Method != "POST" {
			t.Errorf("wrong method. Want %q. Got %q", "POST", req.Method)
		}
		if req.URL.Path != expectedPaths[i] {
			t.Errorf("wrong path. Want %q. Got %q", expectedPaths[i], req.URL.Path)
		}
	}
	expectedPayloads := []url.Values{
		{
			"name":      []string{"superproj-dev"},
			"platform":  []string{"python"},
			"plan":      []string{"medium"},
			"teamOwner": []string{"myteam"},
			"pool":      []string{"dev/dev.example.com"},
		},
		{
			"name":      []string{"superproj-production"},
			"platform":  []string{"python"},
			"plan":      []string{"medium"},
			"teamOwner": []string{"myteam"},
			"pool":      []string{"production/example.com"},
		},
		{"cname": []string{"superproj.dev.example.com"}},
		{"cname": []string{"superproj.example.com"}},
	}
	if len(server.payloads) != len(expectedPayloads) {
		t.Errorf("wrong number of payload. Want %d. Got %d", len(expectedPayloads), len(server.payloads))
	}
	for i, payload := range server.payloads {
		values, err := url.ParseQuery(string(payload))
		if err != nil {
			t.Errorf("invalid payload: %s", payload)
		}
		if !reflect.DeepEqual(values, expectedPayloads[i]) {
			t.Errorf("wrong payload\nWant %#v\nGot  %#v", expectedPayloads[i], values)
		}
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
	appNames := []string{"superproj-dev", "superproj-production"}
	for _, appName := range appNames {
		server.prepareResponse(preparedResponse{
			method:  "POST",
			path:    "/apps/" + appName + "/cname",
			code:    http.StatusOK,
			payload: []byte(`{}`),
		})
	}
	cleanup, err := setupFakeTarget(server.url())
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
	cleanup, err := setupFakeTarget("http://localhost:8080")
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
	cleanup, err := setupFakeTarget("http://localhost:8080")
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

func TestProjectCreateFailToCreateApp(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	server.prepareResponse(preparedResponse{
		method:  "POST",
		path:    "/apps",
		code:    http.StatusInternalServerError,
		payload: []byte("something went wrong"),
	})
	cleanup, err := setupFakeTarget(server.url())
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
	err = c.Run(&ctx, cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{}))
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
		path:   "/apps/superproj-production",
		code:   http.StatusOK,
	})
	server.prepareResponse(preparedResponse{
		method:  "POST",
		path:    "/apps/superproj-production/cname",
		code:    http.StatusInternalServerError,
		payload: []byte(`{}`),
	})
	cleanup, err := setupFakeTarget(server.url())
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
	err = c.Run(&ctx, cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{}))
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
	expectedMethods := []string{"POST", "POST", "POST", "POST", "DELETE", "DELETE"}
	expectedPaths := []string{
		"/1.0/apps", "/1.0/apps",
		"/1.0/apps/superproj-dev/cname",
		"/1.0/apps/superproj-production/cname",
		"/1.0/apps/superproj-dev",
		"/1.0/apps/superproj-production",
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

func TestRemoveProject(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	appNames := []string{"myproj-dev", "myproj-qa", "myproj-stage", "myproj-production"}
	expectedPaths := make([]string, len(appNames))
	for i, appName := range appNames {
		path := "/apps/" + appName
		expectedPaths[i] = path
		server.prepareResponse(preparedResponse{
			method: "DELETE",
			path:   path,
			code:   http.StatusOK,
		})
	}
	cleanup, err := setupFakeTarget(server.url())
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
	if err != nil {
		t.Fatal(err)
	}
	if len(server.reqs) != len(expectedPaths) {
		t.Fatalf("wrong number of requests. Want %d. Got %d", len(expectedPaths), len(server.reqs))
	}
	for i, req := range server.reqs {
		if req.Method != "DELETE" {
			t.Errorf("wrong method. Want DELETE. Got %s", req.Method)
		}
		if p := strings.Replace(req.URL.Path, "/1.0", "", 1); p != expectedPaths[i] {
			t.Errorf("wrong path\nwant %q\ngot  %q", expectedPaths[i], p)
		}
	}
}

func TestRemoveProjectSomeEnvironments(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	appNames := []string{"myproj-dev", "myproj-qa", "myproj-stage", "myproj-production"}
	expectedPaths := make([]string, len(appNames))
	for i, appName := range appNames {
		code := http.StatusOK
		if i%2 == 0 {
			code = http.StatusNotFound
		}
		path := "/apps/" + appName
		expectedPaths[i] = path
		server.prepareResponse(preparedResponse{
			method: "DELETE",
			path:   path,
			code:   code,
		})
	}
	cleanup, err := setupFakeTarget(server.url())
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
	if err != nil {
		t.Fatal(err)
	}
	if len(server.reqs) != len(expectedPaths) {
		t.Fatalf("wrong number of requests. Want %d. Got %d", len(expectedPaths), len(server.reqs))
	}
	for i, req := range server.reqs {
		if req.Method != "DELETE" {
			t.Errorf("wrong method. Want DELETE. Got %s", req.Method)
		}
		if p := strings.Replace(req.URL.Path, "/1.0", "", 1); p != expectedPaths[i] {
			t.Errorf("wrong path\nwant %q\ngot  %q", expectedPaths[i], p)
		}
	}
}

func TestRemoveProjectNoConfirmation(t *testing.T) {
	cleanup, err := setupFakeTarget("http://localhost:8080")
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

func TestRemoveProjectNotFound(t *testing.T) {
	server := newFakeServer(t)
	defer server.stop()
	appNames := []string{"myproj-dev", "myproj-qa", "myproj-stage", "myproj-production"}
	for _, appName := range appNames {
		path := "/apps/" + appName
		server.prepareResponse(preparedResponse{
			method: "DELETE",
			path:   path,
			code:   http.StatusNotFound,
		})
	}
	cleanup, err := setupFakeTarget(server.url())
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

func setupFakeTarget(target string) (func(), error) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	os.Setenv("HOME", dir)
	config := Config{
		Target: target,
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
				Name:      "production",
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
