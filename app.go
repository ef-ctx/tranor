// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

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
