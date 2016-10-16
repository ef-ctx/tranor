// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"

	"github.com/cezarsa/form"
	"github.com/gorilla/mux"
	"github.com/tsuru/tsuru/api"
)

// fakeTsuruServer provides a non-thread-safe, partial implementation of the
// tsuru API.
type fakeTsuruServer struct {
	apps    []app
	envVars map[string][]envVar
	deploys map[string][]deploy
	server  *httptest.Server
	router  *mux.Router
}

func newFakeTsuruServer() *fakeTsuruServer {
	var s fakeTsuruServer
	s.buildRouter()
	s.server = httptest.NewServer(s.router)
	s.reset()
	return &s
}

func (s *fakeTsuruServer) buildRouter() {
	s.router = mux.NewRouter()
	r := s.router.PathPrefix("/1.0").Subrouter()
	r.HandleFunc("/apps", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			s.createApp(w, r)
		case "GET":
			s.listApps(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	r.HandleFunc("/apps/{appname}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PUT":
			s.updateApp(w, r)
		case "GET":
			s.getApp(w, r)
		case "DELETE":
			s.deleteApp(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	r.HandleFunc("/apps/{appname}/env", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			s.setEnvs(w, r)
		case "GET":
			s.getEnvs(w, r)
		case "DELETE":
			s.unsetEnvs(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	r.HandleFunc("/deploys", s.listDeploys)
	r.HandleFunc("/apps/{appname}/cname", s.addCName)
	r.HandleFunc("/apps/{appname}/quota", s.getAppQuota)
	r.HandleFunc("/services/instances", s.serviceInstances)
}

func (s *fakeTsuruServer) createApp(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var opts createAppOptions
	form.DecodeValues(&opts, r.Form)
	if opts.Name == "" || opts.Platform == "" {
		http.Error(w, "invalid params", http.StatusBadRequest)
		return
	}
	_, index := s.findApp(opts.Name)
	if index > -1 {
		http.Error(w, "app already exists", http.StatusConflict)
		return
	}
	if opts.Plan == "" {
		opts.Plan = "autogenerated"
	}
	repositoryURL := fmt.Sprintf("git@gandalf.example.com:%s.git", opts.Name)
	s.apps = append(s.apps, app{
		Name:        opts.Name,
		Platform:    opts.Platform,
		Description: opts.Description,
		TeamOwner:   opts.Team,
		Teams:       []string{opts.Team},
		Plan: struct {
			Name string `json:"name"`
		}{Name: opts.Plan},
		Pool:          opts.Pool,
		Owner:         "user@example.com",
		RepositoryURL: repositoryURL,
		Addr:          opts.Name + ".tsuru.example.com",
	})
	s.envVars[opts.Name] = []envVar{
		{Name: "TSURU_APPDIR", Value: "something"},
		{Name: "TSURU_APP_TOKEN", Value: "sometoken"},
		{Name: "TSURU_APPNAME", Value: opts.Name},
	}
	s.deploys[opts.Name] = nil
	s.writeJSON(w, map[string]string{"repository_url": repositoryURL})
}

func (s *fakeTsuruServer) listApps(w http.ResponseWriter, r *http.Request) {
	nameRegexp, err := regexp.Compile(r.URL.Query().Get("name"))
	if err != nil {
		http.Error(w, "invalid name regexp", http.StatusBadRequest)
		return
	}
	var apps []app
	for _, a := range s.apps {
		if nameRegexp.MatchString(a.Name) {
			apps = append(apps, a)
		}
	}
	if len(apps) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	s.writeJSON(w, apps)
}

func (s *fakeTsuruServer) updateApp(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var opts createAppOptions
	form.DecodeValues(&opts, r.Form)
	a, index := s.findApp(mux.Vars(r)["appname"])
	if index < 0 {
		http.Error(w, "app not found", http.StatusNotFound)
		return
	}
	if opts.Team != "" {
		a.TeamOwner = opts.Team
	}
	if opts.Description != "" {
		a.Description = opts.Description
	}
	if opts.Plan != "" {
		a.Plan.Name = opts.Plan
	}
	if opts.Pool != "" {
		a.Pool = opts.Pool
	}
	s.apps[index] = a
}

func (s *fakeTsuruServer) getApp(w http.ResponseWriter, r *http.Request) {
	a, index := s.findApp(mux.Vars(r)["appname"])
	if index < 0 {
		http.Error(w, "app not found", http.StatusNotFound)
		return
	}
	s.writeJSON(w, a)
}

func (s *fakeTsuruServer) deleteApp(w http.ResponseWriter, r *http.Request) {
	_, index := s.findApp(mux.Vars(r)["appname"])
	if index < 0 {
		http.Error(w, "app not found", http.StatusNotFound)
		return
	}
	s.apps[index] = s.apps[len(s.apps)-1]
	s.apps = s.apps[:len(s.apps)-1]
	w.WriteHeader(http.StatusOK)
}

func (s *fakeTsuruServer) setEnvs(w http.ResponseWriter, r *http.Request) {
	appName := mux.Vars(r)["appname"]
	envs, ok := s.envVars[appName]
	if !ok {
		http.Error(w, "app not found", http.StatusNotFound)
		return
	}
	var evars api.Envs
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	form.DecodeValues(&evars, r.Form)
	for _, e := range evars.Envs {
		envs = append(envs, envVar{
			Name:   e.Name,
			Value:  e.Value,
			Public: !evars.Private,
		})
	}
	s.envVars[appName] = envs
	w.WriteHeader(http.StatusOK)
}

func (s *fakeTsuruServer) getEnvs(w http.ResponseWriter, r *http.Request) {
	appName := mux.Vars(r)["appname"]
	envs, ok := s.envVars[appName]
	if !ok {
		http.Error(w, "app not found", http.StatusNotFound)
		return
	}
	s.writeJSON(w, envs)
}

func (s *fakeTsuruServer) unsetEnvs(w http.ResponseWriter, r *http.Request) {
	appName := mux.Vars(r)["appname"]
	envs, ok := s.envVars[appName]
	if !ok {
		http.Error(w, "app not found", http.StatusNotFound)
		return
	}
	envNames := r.URL.Query()["env"]
	var newEnvs []envVar
	for _, e := range envs {
		var exclude bool
		for _, envName := range envNames {
			if e.Name == envName && e.Public {
				exclude = true
				break
			}
		}
		if !exclude {
			newEnvs = append(newEnvs, e)
		}
	}
	s.envVars[appName] = newEnvs
	w.WriteHeader(http.StatusOK)
}

func (s *fakeTsuruServer) listDeploys(w http.ResponseWriter, r *http.Request) {
	appName := r.URL.Query().Get("app")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if appName == "" {
		http.Error(w, "missing app name in querystring", http.StatusBadRequest)
		return
	}
	deployList, ok := s.deploys[appName]
	if !ok {
		http.Error(w, "app not found", http.StatusNotFound)
		return
	}
	if len(deployList) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if limit == 0 || limit > len(deployList) {
		limit = len(deployList)
	}
	var deploys []deploy
	for i := len(deployList) - 1; i >= len(deployList)-limit; i-- {
		deploys = append(deploys, deployList[i])
	}
	s.writeJSON(w, deploys)
}

func (s *fakeTsuruServer) addCName(w http.ResponseWriter, r *http.Request) {
	cName := r.FormValue("cname")
	if cName == "" {
		http.Error(w, "missing param", http.StatusBadRequest)
		return
	}
	a, index := s.findApp(mux.Vars(r)["appname"])
	if index < 0 {
		http.Error(w, "app not found", http.StatusNotFound)
		return
	}
	for _, name := range a.CName {
		if name == cName {
			http.Error(w, "duplicate cname", http.StatusConflict)
			return
		}
	}
	a.CName = append(a.CName, cName)
	s.apps[index] = a
}

func (s *fakeTsuruServer) getAppQuota(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, map[string]interface{}{"Limit": -1})
}

func (s *fakeTsuruServer) serviceInstances(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, map[string]interface{}{})
}

func (s *fakeTsuruServer) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func (s *fakeTsuruServer) findApp(name string) (a app, index int) {
	index = -1
	for i := range s.apps {
		if s.apps[i].Name == name {
			a = s.apps[i]
			index = i
			break
		}
	}
	return a, index
}

func (s *fakeTsuruServer) url() string {
	return s.server.URL
}

func (s *fakeTsuruServer) token() string {
	return "whatever"
}

func (s *fakeTsuruServer) reset() {
	s.apps = nil
	s.envVars = make(map[string][]envVar)
	s.deploys = make(map[string][]deploy)
}
