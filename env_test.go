// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/tsuru/tsuru/cmd"
)

func TestEnvListIsATsuruCommand(t *testing.T) {
	var _ cmd.Command = envList{}
}

func TestEnvListInfo(t *testing.T) {
	info := envList{}.Info()
	if info == nil {
		t.Fatal("unexpected <nil> info")
	}
	if info.Name != "env-list" {
		t.Errorf("wrong command name. Want %q. Got %q", "env-list", info.Name)
	}
}

func TestEnvListRun(t *testing.T) {
	p, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatal(err)
	}
	os.Setenv("HOME", p)
	var e envList
	var stdout, stderr bytes.Buffer
	ctx := cmd.Context{Stdout: &stdout, Stderr: &stderr}
	err = e.Run(&ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := `+-------------+-------------------+
| Environment | DNS Suffix        |
+-------------+-------------------+
| dev         | dev.example.com   |
| qa          | qa.example.com    |
| stage       | stage.example.com |
| production  | example.com       |
+-------------+-------------------+
`
	if stdout.String() != expectedOutput {
		t.Errorf("wrong output.\nWANT:\n%s\nGOT:\n%s", expectedOutput, stdout.String())
	}
	if stderr.String() != "" {
		t.Errorf("got unexpected non-empty stderr: %q", stderr.String())
	}
}

func TestEnvListRunNoFile(t *testing.T) {
	p, err := filepath.Abs("testdata/not-found")
	if err != nil {
		t.Fatal(err)
	}
	os.Setenv("HOME", p)
	var (
		e   envList
		ctx cmd.Context
	)
	err = e.Run(&ctx, nil)
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
	expectedErrMsg := "unable to load environments file, please make sure that tranor is properly configured"
	if err.Error() != expectedErrMsg {
		t.Errorf("wrong error message.\nWant %q\nGot  %q", expectedErrMsg, err.Error())
	}
}

func TestLoadConfigFile(t *testing.T) {
	p, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatal(err)
	}
	os.Setenv("HOME", p)
	config, err := loadConfigFile()
	if err != nil {
		t.Fatal(err)
	}
	expectedConfig := Config{
		Environments: []Environment{
			{Name: "dev", DNSSuffix: "dev.example.com"},
			{Name: "qa", DNSSuffix: "qa.example.com"},
			{Name: "stage", DNSSuffix: "stage.example.com"},
			{Name: "production", DNSSuffix: "example.com"},
		},
	}
	if !reflect.DeepEqual(*config, expectedConfig) {
		t.Errorf("wrong list of environments.\nWant %#v\nGot  %#v", expectedConfig, *config)
	}
}

func TestLoadConfigFileNotConfigured(t *testing.T) {
	p, err := filepath.Abs("testdata/not-found")
	if err != nil {
		t.Fatal(err)
	}
	os.Setenv("HOME", p)
	config, err := loadConfigFile()
	if err == nil {
		t.Fatal("got unexpected <nil> error")
	}
	if config != nil {
		t.Errorf("got unexpected non-nil config: %#v", config)
	}
	if !os.IsNotExist(err) {
		t.Errorf("got unexpected error: %#v", err)
	}
}
