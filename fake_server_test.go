// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

var versionPathRegexp = regexp.MustCompile(`^/(\d\.?)+`)

type fakeServer struct {
	server   *httptest.Server
	reqs     []http.Request
	payloads [][]byte
	resps    []preparedResponse
	t        *testing.T
}

func newFakeServer(t *testing.T) *fakeServer {
	s := fakeServer{t: t}
	server := httptest.NewServer(&s)
	s.server = server
	return &s
}

func (s *fakeServer) stop() {
	s.server.Close()
}

func (s *fakeServer) url() string {
	return s.server.URL
}

func (s *fakeServer) prepareResponse(r preparedResponse) {
	s.resps = append(s.resps, r)
}

func (s *fakeServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	payload, _ := ioutil.ReadAll(r.Body)
	s.reqs = append(s.reqs, *r)
	s.payloads = append(s.payloads, payload)
	var resp *preparedResponse
	requestURI := versionPathRegexp.ReplaceAllString(r.URL.RequestURI(), "")
	path := versionPathRegexp.ReplaceAllString(r.URL.Path, "")
	s.t.Logf("cleaned request: %s %s", r.Method, requestURI)
	for _, rs := range s.resps {
		pathToCompare := requestURI
		if rs.ignoreQS {
			pathToCompare = path
		}
		if rs.method == r.Method && rs.path == pathToCompare {
			resp = &rs
			break
		}
	}
	if resp == nil {
		s.t.Logf("not found: %s %s", r.Method, r.URL.RequestURI())
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(resp.code)
	w.Write(resp.payload)
}

type preparedResponse struct {
	method   string
	path     string
	code     int
	payload  []byte
	ignoreQS bool
}
