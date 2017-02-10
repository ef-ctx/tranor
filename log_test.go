package main

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/tsuru/tsuru/cmd"
)

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
