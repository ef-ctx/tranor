package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/tsuru/tsuru/cmd"
)

func TestTargetSetIsACommand(t *testing.T) {
	var _ cmd.Command = targetSet{}
}

func TestTargetSetInfo(t *testing.T) {
	info := targetSet{}.Info()
	if info == nil {
		t.Fatal("unexpected <nil> info")
	}
	if info.Name != "target-set" {
		t.Errorf("wrong command name. Want %q. Got %q", "target-set", info.Name)
	}
	if info.MinArgs != 1 {
		t.Errorf("wrong min args. Want 1. Got %d", info.MinArgs)
	}
}

func TestTargetSetRun(t *testing.T) {
	os.Unsetenv("TSURU_TARGET")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/config.json":
			w.Write([]byte(`{"target":"mytarget","envs":[]}`))
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer server.Close()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{
		Args:   []string{server.URL},
		Stdout: &stdout,
		Stderr: &stderr,
	}
	err = targetSet{}.Run(&ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	expectedMsg := "Target successfully defined!\n"
	expectedTarget := "mytarget"
	expectedTargets := "tranor\tmytarget\n"
	expectedConfig := `{"target":"mytarget","envs":[]}` + "\n"
	if stdout.String() != expectedMsg {
		t.Errorf("wrong stdout msg.\nWant %q\nGot  %q", expectedMsg, stdout.String())
	}
	target, err := cmd.ReadTarget()
	if err != nil {
		t.Fatal(err)
	}
	if string(target) != expectedTarget {
		t.Errorf("wrong target. Want %q. Got %q", expectedTarget, string(target))
	}
	targets, err := ioutil.ReadFile(cmd.JoinWithUserDir(".tsuru", "targets"))
	if err != nil {
		t.Fatal(err)
	}
	if string(targets) != expectedTargets {
		t.Errorf("wrong targets file. Want %q. Got %q", expectedTargets, string(targets))
	}
	config, err := ioutil.ReadFile(cmd.JoinWithUserDir(".tranor", "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(config) != expectedConfig {
		t.Errorf("wrong config. Want %q. Got %q", expectedConfig, string(config))
	}
}

func TestDownloadConfiguration(t *testing.T) {
	config := `{"target":"http://mytarget.example.com","envs":[]}`
	var req http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = *r
		w.Write([]byte(config))
	}))
	defer server.Close()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	err = downloadConfiguration(server.URL)
	if err != nil {
		t.Error(err)
	}
	content, err := ioutil.ReadFile(cmd.JoinWithUserDir(".tranor", "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	config += "\n"
	if string(content) != config {
		t.Errorf("wrong config written. Want %q. Got %q", config, string(content))
	}
	if req.Method != "GET" {
		t.Errorf("wrong method in request. Want %q. Got %q", "GET", req.Method)
	}
	if req.URL.Path != "/config.json" {
		t.Errorf("wrong path in request. Want %q. Got %q", "/config.json", req.URL.Path)
	}
}

func TestDownloadConfigurationHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("something went wrong"))
	}))
	defer server.Close()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	err = downloadConfiguration(server.URL)
	expectedMsg := "failed to download config file: 404 - something went wrong"
	if err.Error() != expectedMsg {
		t.Errorf("wrong error message\nWant %q\nGot  %q", expectedMsg, err.Error())
	}
}

func TestDownloadConfigurationNetworkError(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	err = downloadConfiguration("http://192.0.2.12:66000")
	if err == nil {
		t.Error("unexpected <nil> error")
	}
}
