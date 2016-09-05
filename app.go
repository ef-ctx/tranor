// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tsuru/tsuru/cmd"
)

type createAppOptions struct {
	name     string
	platform string
	team     string
	plan     string
	pool     string
}

func (o *createAppOptions) encode() string {
	values := make(url.Values)
	values.Set("name", o.name)
	values.Set("platform", o.platform)
	values.Set("plan", o.plan)
	values.Set("teamOwner", o.team)
	values.Set("pool", o.pool)
	return values.Encode()
}

func createApp(client *cmd.Client, opts createAppOptions) (map[string]string, error) {
	url, err := cmd.GetURL("/apps")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, strings.NewReader(opts.encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var app map[string]string
	err = json.NewDecoder(resp.Body).Decode(&app)
	return app, err
}

func deleteApps(apps []string, client *cmd.Client) ([]error, error) {
	var errs []error
	for _, app := range apps {
		url, err := cmd.GetURL("/apps/" + app)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
		resp.Body.Close()
		errs = append(errs, err)
	}
	return errs, nil
}

func listApps(client *cmd.Client, filters map[string]string) ([]app, error) {
	qs := make(url.Values)
	for k, v := range filters {
		qs.Set(k, v)
	}
	url, err := cmd.GetURL("/apps?" + qs.Encode())
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}
	var apps []app
	err = json.NewDecoder(resp.Body).Decode(&apps)
	return apps, err
}

func lastDeploy(client *cmd.Client, appName string) (deploy, error) {
	var d deploy
	url, err := cmd.GetURL("/deploys?limit=1&app=" + appName)
	if err != nil {
		return d, err
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return d, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return d, err
	}
	defer resp.Body.Close()
	var deploys []deploy
	if resp.StatusCode == http.StatusNoContent {
		return d, nil
	}
	err = json.NewDecoder(resp.Body).Decode(&deploys)
	if err != nil {
		return d, err
	}
	if len(deploys) > 0 {
		d = deploys[0]
	}
	return d, nil
}

type app struct {
	Name  string   `json:"name"`
	CName []string `json:"cname"`
	Env   Environment
	Addr  string
}

type deploy struct {
	ID        string
	Timestamp time.Time
	Image     string
}
