// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"reflect"
	"testing"

	"github.com/tsuru/tsuru/cmd"
)

func TestBaseCommandsAreRegistered(t *testing.T) {
	baseManager := cmd.BuildBaseManager("tranor", "", "", nil)
	manager := buildManager("tranor")
	for name, expectedCommand := range baseManager.Commands {
		gotCommand, ok := manager.Commands[name]
		if !ok {
			t.Errorf("Command %q not found", name)
		}
		if reflect.TypeOf(gotCommand) != reflect.TypeOf(expectedCommand) {
			t.Errorf("Command %q: want %#v. Got %#v", name, expectedCommand, gotCommand)
		}
	}
}

func TestEnvListIsRegistered(t *testing.T) {
	manager := buildManager("tranor")
	gotCommand, ok := manager.Commands["env-list"]
	if !ok {
		t.Error("command env-list not found")
	}
	if _, ok := gotCommand.(envList); !ok {
		t.Errorf("command %#v is not of type envList{}", gotCommand)
	}
}
