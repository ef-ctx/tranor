// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
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
	Target       string        `json:"target"`
	Environments []Environment `json:"envs"`
}

func (c *Config) envNames() []string {
	names := make([]string, len(c.Environments))
	for i, env := range c.Environments {
		names[i] = env.Name
	}
	return names
}

func (c *Config) writeTarget() error {
	cmd.WriteOnTargetList("tranor", c.Target)
	return cmd.WriteTarget(c.Target)
}

// Environment represents an environment for deploying projects.
type Environment struct {
	Name      string `json:"name"`
	DNSSuffix string `json:"dnsSuffix"`
}

func (e *Environment) poolName() string {
	return fmt.Sprintf("%s/%s", e.Name, e.DNSSuffix)
}

func loadConfigFile() (*Config, error) {
	filePath := cmd.JoinWithUserDir(".tranor", "config.json")
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseConfig(f)
}

func parseConfig(r io.Reader) (*Config, error) {
	var config Config
	err := json.NewDecoder(r).Decode(&config)
	return &config, err
}

func writeConfigFile(c *Config) error {
	dir := cmd.JoinWithUserDir(".tranor")
	err := os.Mkdir(dir, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}
	f, err := os.Create(filepath.Join(dir, "config.json"))
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(c)
}
