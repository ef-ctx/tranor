// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cezarsa/form"
	"github.com/tsuru/tsuru/api"
	"github.com/tsuru/tsuru/cmd"
)

type createAppOptions struct {
	name        string
	description string
	platform    string
	team        string
	plan        string
	pool        string
}

func (o *createAppOptions) encode() string {
	values := make(url.Values)
	values.Set("name", o.name)
	values.Set("description", o.description)
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

func updateApp(client *cmd.Client, opts createAppOptions) error {
	url, err := cmd.GetURL("/apps/" + opts.name)
	if err != nil {
		return err
	}
	opts.name = ""
	opts.platform = ""
	req, err := http.NewRequest("PUT", url, strings.NewReader(opts.encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	cmd.StreamJSONResponse(ioutil.Discard, resp)
	return nil
}

func deleteApps(apps []app, client *cmd.Client, w io.Writer) ([]error, error) {
	var errs []error
	for _, app := range apps {
		fmt.Fprintf(w, "Deleting from env %q... ", app.Env.Name)
		url, err := cmd.GetURL("/apps/" + app.Name)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
		errs = append(errs, err)
		cmd.StreamJSONResponse(ioutil.Discard, resp)
		fmt.Fprintln(w, "ok")
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
	resp, err := doReq(client, "/deploys?limit=1&app="+appName)
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

func getApp(client *cmd.Client, appName string) (app, error) {
	var a app
	resp, err := doReq(client, "/apps/"+appName)
	if err != nil {
		return a, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&a)
	return a, err
}

func setEnvs(client *cmd.Client, appName string, envVars *api.Envs) error {
	url, err := cmd.GetURL("/apps/" + appName + "/env")
	if err != nil {
		return err
	}
	v, err := form.EncodeToValues(envVars)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	return cmd.StreamJSONResponse(ioutil.Discard, resp)
}

func getEnvs(client *cmd.Client, appName string) ([]envVar, error) {
	var vars []envVar
	resp, err := doReq(client, "/apps/"+appName+"/env")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&vars)
	return vars, err
}

func unsetEnvs(client *cmd.Client, appName string, noRestart bool, envVars []string) error {
	qs := make(url.Values)
	qs.Set("noRestart", strconv.FormatBool(noRestart))
	for _, e := range envVars {
		qs.Add("env", e)
	}
	url, err := cmd.GetURL("/apps/" + appName + "/env?" + qs.Encode())
	if err != nil {
		return err
	}
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	return cmd.StreamJSONResponse(ioutil.Discard, resp)
}

func doReq(client *cmd.Client, path string) (*http.Response, error) {
	url, err := cmd.GetURL(path)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

type envVar struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Public bool   `json:"public"`
}

func (e *envVar) String() string {
	value := "*** (private config)"
	if e.Public {
		value = e.Value
	}
	return fmt.Sprintf("%s=%s", e.Name, value)
}

type app struct {
	Name          string        `json:"name"`
	CName         []string      `json:"cname"`
	Description   string        `json:"description"`
	RepositoryURL string        `json:"repository"`
	Platform      string        `json:"platform"`
	Teams         []string      `json:"teams"`
	Owner         string        `json:"owner"`
	Pool          string        `json:"pool"`
	TeamOwner     string        `json:"teamowner"`
	Units         []interface{} `json:"units"`
	Env           Environment
	Addr          string
	Plan          struct {
		Name string `json:"name"`
	} `json:"plan"`
}

type deploy struct {
	ID        string
	Commit    string
	Timestamp time.Time
	Image     string
}
