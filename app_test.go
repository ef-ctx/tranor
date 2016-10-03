// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/tsuru/tsuru/api"
	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/errors"
)

func TestCreateApp(t *testing.T) {
	fakeServer := newFakeServer(t)
	defer fakeServer.stop()
	fakeServer.prepareResponse(preparedResponse{
		method:  "POST",
		path:    "/apps",
		code:    http.StatusOK,
		payload: []byte(`{"repository_url":"git@example.com:app.git"}`),
	})
	cleanup, err := setupFakeTarget(fakeServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	app, err := createApp(client, createAppOptions{
		name:        "app",
		description: "my nice app",
		plan:        "medium",
		platform:    "python",
		pool:        "mypool",
		team:        "admin",
	})
	if err != nil {
		t.Fatal(err)
	}
	expectedApp := map[string]string{"repository_url": "git@example.com:app.git"}
	if !reflect.DeepEqual(app, expectedApp) {
		t.Errorf("wrong app map returned\nwant %#v\ngot  %#v", expectedApp, app)
	}
	req := fakeServer.reqs[0]
	if req.Method != "POST" {
		t.Errorf("wrong method. Want POST. Got %s", req.Method)
	}
	if req.URL.Path != "/1.0/apps" {
		t.Errorf("wrong path. Want /1.0/apps. Got %s", req.URL.Path)
	}
	expectedParams := url.Values(map[string][]string{
		"name":        {"app"},
		"description": {"my nice app"},
		"plan":        {"medium"},
		"platform":    {"python"},
		"pool":        {"mypool"},
		"teamOwner":   {"admin"},
	})
	gotParams, err := url.ParseQuery(string(fakeServer.payloads[0]))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotParams, expectedParams) {
		t.Errorf("wrong params in body\nwant %#v\ngot  %#v", expectedParams, gotParams)
	}
}

func TestCreateAppNoTarget(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	app, err := createApp(nil, createAppOptions{})
	if app != nil {
		t.Errorf("unexpected non-nil app: %#v", app)
	}
	if err == nil {
		t.Error("unexpected <nil> error")
	}
}

func TestUpdateApp(t *testing.T) {
	fakeServer := newFakeServer(t)
	defer fakeServer.stop()
	fakeServer.prepareResponse(preparedResponse{
		method:  "PUT",
		path:    "/apps/myapp",
		code:    http.StatusOK,
		payload: []byte(`{"repository_url":"git@example.com:app.git"}`),
	})
	cleanup, err := setupFakeTarget(fakeServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = updateApp(client, createAppOptions{
		name:        "myapp",
		description: "my nice app - updated!",
		plan:        "medium",
		platform:    "",
		pool:        "mypool",
		team:        "admin",
	})
	if err != nil {
		t.Fatal(err)
	}
	expectedParams := url.Values(map[string][]string{
		"name":        {""},
		"description": {"my nice app - updated!"},
		"plan":        {"medium"},
		"platform":    {""},
		"pool":        {"mypool"},
		"teamOwner":   {"admin"},
	})
	gotParams, err := url.ParseQuery(string(fakeServer.payloads[0]))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotParams, expectedParams) {
		t.Errorf("wrong params in body\nwant %#v\ngot  %#v", expectedParams, gotParams)
	}
}

func TestUpdateAppNotFound(t *testing.T) {
	fakeServer := newFakeServer(t)
	defer fakeServer.stop()
	fakeServer.prepareResponse(preparedResponse{
		method:  "PUT",
		path:    "/apps/myapp",
		code:    http.StatusNotFound,
		payload: []byte("app not found"),
	})
	cleanup, err := setupFakeTarget(fakeServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = updateApp(client, createAppOptions{
		name:        "myapp",
		description: "my nice app - updated!",
		plan:        "medium",
		platform:    "",
		pool:        "mypool",
		team:        "admin",
	})
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
}

func TestUpdateAppNoTarget(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	err = updateApp(nil, createAppOptions{})
	if err == nil {
		t.Error("unexpected <nil> error")
	}
}

func TestDeleteApps(t *testing.T) {
	fakeServer := newFakeServer(t)
	defer fakeServer.stop()
	fakeServer.prepareResponse(preparedResponse{
		method: "DELETE",
		path:   "/apps/proj1-dev",
		code:   http.StatusOK,
	})
	fakeServer.prepareResponse(preparedResponse{
		method: "DELETE",
		path:   "/apps/proj1-prod",
		code:   http.StatusOK,
	})
	cleanup, err := setupFakeTarget(fakeServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	errs, err := deleteApps([]app{
		{Name: "proj1-dev", Env: Environment{Name: "dev"}},
		{Name: "proj1-qa", Env: Environment{Name: "qa"}},
		{Name: "proj1-prod", Env: Environment{Name: "prod"}},
	}, client, &stdout)
	if err != nil {
		t.Fatal(err)
	}
	expectedErrs := []error{nil, &errors.HTTP{Code: http.StatusNotFound, Message: "not found\n"}, nil}
	if !reflect.DeepEqual(errs, expectedErrs) {
		t.Errorf("wrong error list\nwant %#v\ngot  %#v", expectedErrs, errs)
	}
	paths := []string{"/1.0/apps/proj1-dev", "/1.0/apps/proj1-qa", "/1.0/apps/proj1-prod"}
	for i, req := range fakeServer.reqs {
		if req.Method != "DELETE" {
			t.Errorf("wrong method. Want DELETE. Got %s", req.Method)
		}
		if req.URL.Path != paths[i] {
			t.Errorf("wrong path\nwant %q\ngot  %q", paths[i], req.URL.Path)
		}
	}
	expectedOutput := `Deleting from env "dev"... ok
Deleting from env "qa"... ok
Deleting from env "prod"... ok
`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output\nWant:\n%s\nGot\n%s", expectedOutput, stdout.String())
	}
}

func TestDeleteAppsNoTarget(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	errs, err := deleteApps([]app{{Name: "app1"}, {Name: "app2"}}, nil, ioutil.Discard)
	if len(errs) != 0 {
		t.Errorf("unexpected non-empty error list: %#v", errs)
	}
	if err == nil {
		t.Error("unexpected <nil> error")
	}
}

func TestListApps(t *testing.T) {
	fakeServer := newFakeServer(t)
	defer fakeServer.stop()
	fakeServer.prepareResponse(preparedResponse{
		method:  "GET",
		path:    "/apps?",
		code:    http.StatusOK,
		payload: []byte(listOfApps),
	})
	fakeServer.prepareResponse(preparedResponse{
		method:  "GET",
		path:    "/apps",
		code:    http.StatusOK,
		payload: []byte(listOfApps),
	})
	cleanup, err := setupFakeTarget(fakeServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	apps, err := listApps(client, nil)
	if err != nil {
		t.Error(err)
	}
	expectedApps := []app{
		{Name: "tsuru-dashboard", CName: []string{}},
		{Name: "proj2-dev", CName: []string{"proj2.dev.example.com"}},
		{Name: "proj2-qa", CName: []string{"proj2.qa.example.com"}},
		{Name: "proj2-stage", CName: []string{"proj2.stage.example.com"}},
		{Name: "proj2-prod", CName: []string{"proj2.example.com"}},
		{Name: "myblog-qa", CName: []string{"myblog.qa.example.com", "myblog.qa2.example.com"}},
		{Name: "proj1-dev", CName: []string{"proj1.dev.example.com"}},
		{Name: "proj1-qa", CName: []string{"proj1.qa.example.com"}},
		{Name: "proj1-stage", CName: []string{"proj1.stage.example.com"}},
		{Name: "myblog-dev", CName: []string{}},
		{Name: "proj1-prod", CName: []string{"proj1.example.com"}},
		{Name: "proj3-dev", CName: []string{"proj3.dev.example.com"}},
		{Name: "proj3-prod", CName: []string{"proj3.example.com"}},
	}
	if !reflect.DeepEqual(apps, expectedApps) {
		t.Errorf("wrong list of apps\nwant %#v\ngot  %#v", expectedApps, apps)
	}
}

func TestListAppsEmpty(t *testing.T) {
	fakeServer := newFakeServer(t)
	defer fakeServer.stop()
	fakeServer.prepareResponse(preparedResponse{
		method: "GET",
		path:   "/apps?",
		code:   http.StatusNoContent,
	})
	fakeServer.prepareResponse(preparedResponse{
		method: "GET",
		path:   "/apps",
		code:   http.StatusNoContent,
	})
	cleanup, err := setupFakeTarget(fakeServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	apps, err := listApps(client, nil)
	if err != nil {
		t.Error(err)
	}
	if len(apps) != 0 {
		t.Errorf("got unexpected non-empty app list: %#v", apps)
	}
}

func TestListAppsNoTarget(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	apps, err := listApps(nil, nil)
	if len(apps) != 0 {
		t.Errorf("unexpected non-empty app list: %#v", apps)
	}
	if err == nil {
		t.Error("unexpected <nil> error")
	}
}

func TestLastDeploy(t *testing.T) {
	fakeServer := newFakeServer(t)
	defer fakeServer.stop()
	fakeServer.prepareResponse(preparedResponse{
		method:  "GET",
		path:    "/deploys?limit=1&app=myapp",
		code:    http.StatusOK,
		payload: []byte(deployments),
	})
	cleanup, err := setupFakeTarget(fakeServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	d, err := lastDeploy(client, "myapp")
	if err != nil {
		t.Fatal(err)
	}
	expectedDeployment := deploy{
		ID:        "57ccc9490640fd3def98b157",
		Commit:    "40244ff2866eba7e2da6eee8a6fc51464c9f604f",
		Image:     "v938",
		Timestamp: time.Date(2016, 9, 5, 1, 24, 25, 706e6, time.UTC),
	}
	if !reflect.DeepEqual(d, expectedDeployment) {
		t.Errorf("wrong deploy\nwant %#v\ngot  %#v", expectedDeployment, d)
	}
}

func TestLastDeployEmpty(t *testing.T) {
	fakeServer := newFakeServer(t)
	defer fakeServer.stop()
	fakeServer.prepareResponse(preparedResponse{
		method: "GET",
		path:   "/deploys?limit=1&app=myapp",
		code:   http.StatusNoContent,
	})
	cleanup, err := setupFakeTarget(fakeServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	d, err := lastDeploy(client, "myapp")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(d, deploy{}) {
		t.Errorf("expected an empty deploy, got %#v", d)
	}
}

func TestLastDeployNoTarget(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	_, err = lastDeploy(nil, "myapp")
	if err == nil {
		t.Error("unexpected <nil> error")
	}
}

func TestGetApp(t *testing.T) {
	fakeServer := newFakeServer(t)
	defer fakeServer.stop()
	fakeServer.prepareResponse(preparedResponse{
		method:  "GET",
		path:    "/apps/proj3-prod",
		code:    http.StatusOK,
		payload: []byte(appInfo6),
	})
	cleanup, err := setupFakeTarget(fakeServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	a, err := getApp(client, "proj3-prod")
	if err != nil {
		t.Fatal(err)
	}
	var expectedApp app
	err = json.Unmarshal([]byte(appInfo6), &expectedApp)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, expectedApp) {
		t.Errorf("wrong app returned\nwant %#v\ngot  %#v", expectedApp, a)
	}
}

func TestGetAppNotFound(t *testing.T) {
	fakeServer := newFakeServer(t)
	defer fakeServer.stop()
	fakeServer.prepareResponse(preparedResponse{
		method: "GET",
		path:   "/apps/proj1-prod",
		code:   http.StatusNotFound,
	})
	cleanup, err := setupFakeTarget(fakeServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	a, err := getApp(client, "proj1-prod")
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
	if !reflect.DeepEqual(a, app{}) {
		t.Errorf("go non-empty app: %#v", a)
	}
}

func TestGetAppNoTarget(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	_, err = getApp(nil, "proj1-prod")
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
}

func TestSetEnvs(t *testing.T) {
	fakeServer := newFakeServer(t)
	defer fakeServer.stop()
	fakeServer.prepareResponse(preparedResponse{
		code:    http.StatusOK,
		method:  "POST",
		path:    "/apps/proj1-prod/env",
		payload: []byte("{}"),
	})
	cleanup, err := setupFakeTarget(fakeServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = setEnvs(client, "proj1-prod", &api.Envs{
		Private:   true,
		NoRestart: true,
		Envs: []struct {
			Name  string
			Value string
		}{
			{Name: "USER_NAME", Value: "root"},
			{Name: "USER_PASSWORD", Value: "r00t"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	expectedPayload := url.Values{
		"NoRestart":    []string{"true"},
		"Private":      []string{"true"},
		"Envs.0.Name":  []string{"USER_NAME"},
		"Envs.0.Value": []string{"root"},
		"Envs.1.Name":  []string{"USER_PASSWORD"},
		"Envs.1.Value": []string{"r00t"},
	}
	payload, err := url.ParseQuery(string(fakeServer.payloads[0]))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(payload, expectedPayload) {
		t.Errorf("wrong payload\nwant %#v\ngot  %#v", expectedPayload, payload)
	}
}

func TestSetEnvsNotFound(t *testing.T) {
	fakeServer := newFakeServer(t)
	defer fakeServer.stop()
	cleanup, err := setupFakeTarget(fakeServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = setEnvs(client, "proj1-prod", nil)
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
	e := err.(*errors.HTTP)
	if e.Code != http.StatusNotFound {
		t.Errorf("wrong error code. Want %d. Got %d", http.StatusNotFound, e.Code)
	}
}

func TestSetEnvsNoTarget(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	err = setEnvs(nil, "proj1-prod", nil)
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
}

func TestGetEnvs(t *testing.T) {
	fakeServer := newFakeServer(t)
	defer fakeServer.stop()
	fakeServer.prepareResponse(preparedResponse{
		code:    http.StatusOK,
		method:  "GET",
		path:    "/apps/proj1-prod/env",
		payload: []byte(`[{"name":"USER_NAME","value":"root","public":true},{"name":"USER_PASSWORD","value":"r00t","public":false}]`),
	})
	cleanup, err := setupFakeTarget(fakeServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	evars, err := getEnvs(client, "proj1-prod")
	if err != nil {
		t.Fatal(err)
	}
	expectedVars := []envVar{
		{
			Name:   "USER_NAME",
			Value:  "root",
			Public: true,
		},
		{
			Name:  "USER_PASSWORD",
			Value: "r00t",
		},
	}
	if !reflect.DeepEqual(evars, expectedVars) {
		t.Errorf("wrong list of vars\nwant %#v\ngot  %#v", expectedVars, evars)
	}
}

func TestGetEnvsNotFound(t *testing.T) {
	fakeServer := newFakeServer(t)
	defer fakeServer.stop()
	cleanup, err := setupFakeTarget(fakeServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	evars, err := getEnvs(client, "proj1-prod")
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
	if evars != nil {
		t.Fatalf("unexpected non-nil vars: %#v", evars)
	}
	e := err.(*errors.HTTP)
	if e.Code != http.StatusNotFound {
		t.Errorf("wrong error code. Want %d. Got %d", http.StatusNotFound, e.Code)
	}
}

func TestGetEnvsNoTarget(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	evars, err := getEnvs(nil, "proj1-prod")
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
	if evars != nil {
		t.Fatalf("unexpected non-nil vars: %#v", evars)
	}
}

func TestUnsetEnvs(t *testing.T) {
	fakeServer := newFakeServer(t)
	defer fakeServer.stop()
	fakeServer.prepareResponse(preparedResponse{
		code:     http.StatusOK,
		method:   "DELETE",
		path:     "/apps/proj1-prod/env",
		payload:  []byte("{}"),
		ignoreQS: true,
	})
	cleanup, err := setupFakeTarget(fakeServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = unsetEnvs(client, "proj1-prod", true, []string{"USER_NAME", "USER_PASSWORD", "PASSWORD_HINT"})
	if err != nil {
		t.Fatal(err)
	}
	expectedQS := url.Values{
		"noRestart": []string{"true"},
		"env":       []string{"USER_NAME", "USER_PASSWORD", "PASSWORD_HINT"},
	}
	qs := fakeServer.reqs[0].URL.Query()
	if !reflect.DeepEqual(qs, expectedQS) {
		t.Errorf("wrong querystring\nwant %#v\ngot  %#v", expectedQS, qs)
	}
}

func TestUnsetEnvsNotFound(t *testing.T) {
	fakeServer := newFakeServer(t)
	defer fakeServer.stop()
	cleanup, err := setupFakeTarget(fakeServer.url())
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	client := cmd.NewClient(http.DefaultClient, &ctx, &cmd.Manager{})
	err = unsetEnvs(client, "proj1-prod", false, nil)
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
	e := err.(*errors.HTTP)
	if e.Code != http.StatusNotFound {
		t.Errorf("wrong error code. Want %d. Got %d", http.StatusNotFound, e.Code)
	}
}

func TestUnsetEnvsNoTarget(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	err = unsetEnvs(nil, "proj1-prod", false, nil)
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
}

func TestEnvVarStringRepr(t *testing.T) {
	var tests = []struct {
		testCase string
		input    envVar
		expected string
	}{
		{
			"public variable",
			envVar{
				Name:   "USER_NAME",
				Value:  "root",
				Public: true,
			},
			"USER_NAME=root",
		},
		{
			"private variable",
			envVar{
				Name:  "USER_PASSWORD",
				Value: "r00t",
			},
			"USER_PASSWORD=*** (private config)",
		},
	}
	for _, test := range tests {
		repr := test.input.String()
		if repr != test.expected {
			t.Errorf("%s: wrong representation\nwant %q\ngot  %q", test.testCase, test.expected, repr)
		}
	}
}
