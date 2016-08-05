// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/tsuru/tsuru/cmd"
)

type envList struct{}

func (envList) Info() *cmd.Info {
	return &cmd.Info{
		Name:  "env-list",
		Usage: "env-list",
		Desc:  "list currently available environments",
	}
}

func (envList) Run(ctx *cmd.Context, _ *cmd.Client) error {
	config, err := loadConfigFile()
	if err != nil {
		return errors.New("unable to load environments file, please make sure that tranor is properly configured")
	}
	table := cmd.NewTable()
	table.Headers = cmd.Row{"Environment", "DNS Suffix"}
	for _, env := range config.Environments {
		table.AddRow(cmd.Row{env.Name, env.DNSSuffix})
	}
	ctx.Stdout.Write(table.Bytes())
	return nil
}

// Config represents the configuration for the tranor command line.
type Config struct {
	Environments []Environment `json:"envs"`
}

// Environment represents an environment for deploying projects.
type Environment struct {
	Name      string `json:"name"`
	DNSSuffix string `json:"dnsSuffix"`
}

func loadConfigFile() (*Config, error) {
	filePath := filepath.Join(os.Getenv("HOME"), ".tranor", "config.json")
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var config Config
	err = json.NewDecoder(f).Decode(&config)
	return &config, err
}

func writeConfigFile(r io.Reader) error {
	dir := filepath.Join(os.Getenv("HOME"), ".tranor")
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(dir, "config.json"))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}
