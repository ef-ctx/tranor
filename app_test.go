// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/tsuru/tsuru/cmd"
)

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

func TestDeleteAppsNoTarget(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	errs, err := deleteApps([]string{"app1", "app2"}, nil)
	if len(errs) != 0 {
		t.Errorf("unexpected non-empty error list: %#v", errs)
	}
	if err == nil {
		t.Error("unexpected <nil> error")
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

func TestLastDeployEmpty(t *testing.T) {
	fakeServer := newFakeServer(t)
	defer fakeServer.stop()
	fakeServer.prepareResponse(preparedResponse{
		method: "GET",
		path:   "/apps?limit=1&app=myapp",
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
