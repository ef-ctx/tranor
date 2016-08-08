package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/tsuru/tsuru/cmd"
)

type targetSet struct{}

func (targetSet) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "target-set",
		Usage:   "target-set <target>",
		Desc:    "sets the remote tranor server",
		MinArgs: 1,
	}
}

func (targetSet) Run(ctx *cmd.Context, _ *cmd.Client) error {
	err := downloadConfiguration(ctx.Args[0])
	if err != nil {
		return err
	}
	fmt.Fprintln(ctx.Stdout, "Target successfully defined!")
	return nil
}

func downloadConfiguration(server string) error {
	url := strings.TrimRight(server, "/") + "/config.json"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to download config file: %d - %s", resp.StatusCode, data)
	}
	config, err := parseConfig(resp.Body)
	if err != nil {
		return fmt.Errorf("invalid configuration returned by the remote target: %s", err)
	}
	err = writeConfigFile(config)
	if err != nil {
		return err
	}
	return config.writeTarget()
}
