package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/tsuru/tsuru-client/tsuru/client"
	"github.com/tsuru/tsuru/cmd"
)

func TestProjectLog(t *testing.T) {
	tsuruServer := newFakeServer(t)
	defer tsuruServer.stop()
	cleanup, err := setupFakeConfig(tsuruServer.url(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	now := time.Now().UTC()
	logs := []apiLog{
		{Date: now, Message: "creating app lost", Source: "tsuru"},
		{Date: now.Add(2 * time.Hour), Message: "app lost successfully created", Source: "app", Unit: "abcdef"},
	}
	result, err := json.Marshal(logs)
	if err != nil {
		t.Fatal(err)
	}
	tsuruServer.prepareResponse(preparedResponse{
		code:    http.StatusOK,
		payload: []byte(result),
		method:  "GET",
		path:    "/apps/myapp-dev/log?lines=10",
	})

	var appLogCmd client.AppLog
	err = appLogCmd.Flags().Parse(true, []string{"--app", "myapp-dev", "--lines", "10"})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
	}
	cli := cmd.NewClient(http.DefaultClient, nil, &cmd.Manager{})
	err = appLogCmd.Run(&context, cli)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := stdout.String()
	stdout.Reset()
	stderr.Reset()

	var command projectLog
	err = command.Flags().Parse(true, []string{"-n", "myapp", "-e", "dev", "-l", "10"})
	if err != nil {
		t.Fatal(err)
	}
	err = command.Run(&context, cli)
	if err != nil {
		t.Fatal(err)
	}
	if got := stdout.String(); got != expectedOutput {
		t.Errorf("wrong output\nwant: %q\ngot:  %q", expectedOutput, got)
	}
}

func TestProjectLogMissingName(t *testing.T) {
	var c projectLog
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

func TestProjectLogMissingEnv(t *testing.T) {
	var c projectLog
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

type apiLog struct {
	Date    time.Time
	Message string
	Source  string
	Unit    string
}
