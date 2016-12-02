// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	tsuruServer = newFakeTsuruServer()
	testHost := os.Getenv("TSURU_TEST_HOST")
	testToken := os.Getenv("TSURU_TEST_TOKEN")
	if testHost != "" && testToken != "" {
		tsuruServer = &actualTsuruServer{
			tsuruHost:  testHost,
			tsuruToken: testToken,
		}
	}
	var exitCode int
	defer func() {
		tsuruServer.reset()
		os.Exit(exitCode)
	}()
	exitCode = m.Run()
}
