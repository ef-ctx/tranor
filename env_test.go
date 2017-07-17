// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
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
| prod        | example.com       |
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
		Target:   "http://tsuru-api.example.com",
		Registry: "localhost:5000",
		Environments: []Environment{
			{Name: "dev", DNSSuffix: "dev.example.com"},
			{Name: "qa", DNSSuffix: "qa.example.com"},
			{Name: "stage", DNSSuffix: "stage.example.com"},
			{Name: "prod", DNSSuffix: "example.com"},
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

func TestWriteConfigFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	config := Config{
		Target: "http://mytsuru.example.com",
		Environments: []Environment{
			{
				Name:      "dev",
				DNSSuffix: "dev.example.com",
			},
		},
	}
	expectedContent, _ := json.Marshal(config)
	err = writeConfigFile(&config)
	if err != nil {
		t.Fatal(err)
	}
	fullPath := cmd.JoinWithUserDir(".tranor", "config.json")
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		t.Fatal(err)
	}
	expectedContent = append(expectedContent, '\n')
	if string(content) != string(expectedContent) {
		t.Errorf("wrong content in the config file.\nWant %s\nGot  %s", expectedContent, content)
	}
}

func TestWriteConfigFileDirAlreadyExists(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	config := Config{
		Target: "http://mytsuru.example.com",
		Environments: []Environment{
			{
				Name:      "dev",
				DNSSuffix: "dev.example.com",
			},
		},
	}
	os.Mkdir(filepath.Join(dir, ".tranor"), 0755)
	expectedContent, _ := json.Marshal(config)
	err = writeConfigFile(&config)
	if err != nil {
		t.Fatal(err)
	}
	fullPath := cmd.JoinWithUserDir(".tranor", "config.json")
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		t.Fatal(err)
	}
	expectedContent = append(expectedContent, '\n')
	if string(content) != string(expectedContent) {
		t.Errorf("wrong content in the config file.\nWant %s\nGot  %s", expectedContent, content)
	}
}

func TestWriteConfigFileErrorToCreateFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	os.MkdirAll(cmd.JoinWithUserDir(".tranor", "config.json", "this-is-a-dir"), 0755)
	config := Config{Target: "something"}
	err = writeConfigFile(&config)
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
}

func TestWriteConfigFileErrorToCreateDirectory(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	err = os.Chmod(dir, 0555)
	if err != nil {
		t.Fatal(err)
	}
	err = writeConfigFile(&Config{})
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
}

func TestConfigEnvNames(t *testing.T) {
	c := Config{
		Environments: []Environment{
			{Name: "dev", DNSSuffix: "dev.example.com"},
			{Name: "qa", DNSSuffix: "qa.example.com"},
			{Name: "prod", DNSSuffix: "example.com"},
		},
	}
	expectedNames := []string{"dev", "qa", "prod"}
	gotNames := c.envNames()
	if !reflect.DeepEqual(gotNames, expectedNames) {
		t.Errorf("wrong env names returned\nwant %#v\ngot  %#v", expectedNames, gotNames)
	}
}

func TestConfigImageApp(t *testing.T) {
	var tests = []struct {
		testCase      string
		config        Config
		expectedImage string
	}{
		{
			"with registry",
			Config{Registry: "localhost:4040"},
			"localhost:4040/tsuru/app-myapp:v10",
		},
		{
			"without registry",
			Config{},
			"tsuru/app-myapp:v10",
		},
	}
	for _, test := range tests {
		t.Run(test.testCase, func(t *testing.T) {
			gotImage := test.config.imageApp("myapp", "v10")
			if gotImage != test.expectedImage {
				t.Errorf("wrong image returned\nwant %q\ngot  %q", test.expectedImage, gotImage)
			}
		})
	}
}

func TestEnvironmentPoolName(t *testing.T) {
	var tests = []struct {
		input  Environment
		output string
	}{
		{
			Environment{Name: "dev", DNSSuffix: "dev.example.com"},
			`dev,dev.example.com`,
		},
		{
			Environment{Name: "qa", DNSSuffix: "whatever"},
			`qa,whatever`,
		},
	}
	for _, test := range tests {
		got := test.input.poolName()
		if got != test.output {
			t.Errorf("wrong pool name\nWant %q\nGot  %q", test.output, got)
		}
	}
}
