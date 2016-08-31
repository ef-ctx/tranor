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
		var skip bool
		for _, c := range baseCommandsToRemove {
			if name == c {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		gotCommand, ok := manager.Commands[name]
		if !ok {
			t.Errorf("Command %q not found", name)
		}
		if reflect.TypeOf(gotCommand) != reflect.TypeOf(expectedCommand) {
			t.Errorf("Command %q: want %#v. Got %#v", name, expectedCommand, gotCommand)
		}
	}
}

func TestDefaultTargetCommandsArentRegistered(t *testing.T) {
	manager := buildManager("tranor")
	cmds := []string{"target-add", "target-list", "target-remove"}
	for _, cmd := range cmds {
		if _, ok := manager.Commands[cmd]; ok {
			t.Errorf("command %q should not be registered", cmd)
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

func TestBuiltinTargetSetIsOverwritten(t *testing.T) {
	manager := buildManager("tranor")
	gotCommand, ok := manager.Commands["target-set"]
	if !ok {
		t.Error("command target-set not found")
	}
	if _, ok := gotCommand.(targetSet); !ok {
		t.Errorf("command %#v is not of type targetSet{}", gotCommand)
	}
}
