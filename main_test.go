// Copyright 2016 EF CTX. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"reflect"
	"testing"

	"github.com/tsuru/tsuru-client/tsuru/admin"
	"github.com/tsuru/tsuru-client/tsuru/client"
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

func TestPlatformListIsRegistered(t *testing.T) {
	manager := buildManager("tranor")
	gotCommand, ok := manager.Commands["platform-list"]
	if !ok {
		t.Error("command platform-list not found")
	}
	if _, ok := gotCommand.(admin.PlatformList); !ok {
		t.Errorf("command %#v is not of type PlatformList{}", gotCommand)
	}
}

func TestTeamListIsRegistered(t *testing.T) {
	manager := buildManager("tranor")
	gotCommand, ok := manager.Commands["team-list"]
	if !ok {
		t.Error("command team-list not found")
	}
	if _, ok := gotCommand.(*client.TeamList); !ok {
		t.Errorf("command %#v is not of type TeamList{}", gotCommand)
	}
}

func TestTeamCreateIsRegistered(t *testing.T) {
	manager := buildManager("tranor")
	gotCommand, ok := manager.Commands["team-create"]
	if !ok {
		t.Error("command team-create not found")
	}
	if _, ok := gotCommand.(*client.TeamCreate); !ok {
		t.Errorf("command %#v is not of type TeamCreate{}", gotCommand)
	}
}

func TestTeamRemoveIsRegistered(t *testing.T) {
	manager := buildManager("tranor")
	gotCommand, ok := manager.Commands["team-remove"]
	if !ok {
		t.Error("command team-remove not found")
	}
	if _, ok := gotCommand.(*client.TeamRemove); !ok {
		t.Errorf("command %#v is not of type TeamRemove{}", gotCommand)
	}
}

func TestPlanListIsRegistered(t *testing.T) {
	manager := buildManager("tranor")
	gotCommand, ok := manager.Commands["plan-list"]
	if !ok {
		t.Error("command plan-list not found")
	}
	if _, ok := gotCommand.(*client.PlanList); !ok {
		t.Errorf("command %#v is not of type PlanList{}", gotCommand)
	}
}

func TestProjectCreateIsRegistered(t *testing.T) {
	manager := buildManager("tranor")
	gotCommand, ok := manager.Commands["project-create"]
	if !ok {
		t.Error("command project-create not found")
	}
	if _, ok := gotCommand.(*projectCreate); !ok {
		t.Errorf("command %#v is not of type projectCreate{}", gotCommand)
	}
}

func TestProjectRemoveIsRegistered(t *testing.T) {
	manager := buildManager("tranor")
	gotCommand, ok := manager.Commands["project-remove"]
	if !ok {
		t.Error("command project-remove not found")
	}
	if _, ok := gotCommand.(*projectRemove); !ok {
		t.Errorf("command %#v is not of type projectRemove{}", gotCommand)
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
