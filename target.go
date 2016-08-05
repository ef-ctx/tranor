package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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
	err = downloadTsuruTarget(ctx.Args[0])
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
	return writeConfigFile(resp.Body)
}

func downloadTsuruTarget(server string) error {
	url := strings.TrimRight(server, "/") + "/tsuru-target"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to download server information: %d - %s", resp.StatusCode, data)
	}
	target, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	os.MkdirAll(cmd.JoinWithUserDir(".tsuru"), 0755)
	targetFile, err := os.Create(cmd.JoinWithUserDir(".tsuru", "target"))
	if err != nil {
		return err
	}
	defer targetFile.Close()
	targetFile.Write(target)
	targetsFile, err := os.Create(cmd.JoinWithUserDir(".tsuru", "targets"))
	if err != nil {
		return err
	}
	defer targetsFile.Close()
	targets := bytes.Join([][]byte{[]byte("tranor"), target}, []byte{'\t'})
	targetsFile.Write(targets)
	return nil
}
