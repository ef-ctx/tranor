// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
	"net/http"

	"github.com/tsuru/tsuru/cmd"
)

var tsuruServer interface {
	url() string
	token() string
	reset()
}

type actualTsuruServer struct {
	tsuruHost  string
	tsuruToken string
}

func (s *actualTsuruServer) reset() {
	client := cmd.NewClient(http.DefaultClient, &cmd.Context{}, &cmd.Manager{})
	cleanup, err := setupFakeConfig(s.url(), s.token())
	if err != nil {
		panic(err)
	}
	defer cleanup()
	appPrefixes := []string{"myproj", "superproj", "proj1"}
	for _, prefix := range appPrefixes {
		filters := map[string]string{"name": "^" + prefix}
		apps, err := listApps(client, filters)
		if err != nil {
			panic(err)
		}
		_, err = deleteApps(apps, client, ioutil.Discard)
		if err != nil {
			panic(err)
		}
	}
}

func (s *actualTsuruServer) token() string {
	return s.tsuruToken
}

func (s *actualTsuruServer) url() string {
	return s.tsuruHost
}
