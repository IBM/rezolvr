// © Copyright IBM Corporation 2020. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"rezolvr/model"
	"testing"
)

func Test_MarkResourceResolvedStatus(t *testing.T) {
	resource := model.Component{}
	resource.Provides = map[string]*model.Resource{}
	resource.Needs = map[string]*model.Resource{}

	pr := model.Resource{}
	pr.Params = map[string]*model.Param{}
	param1 := model.Param{Name: "Name1", Value: "Value1"}
	pr.Params["Name1"] = &param1
	resource.Provides["test1"] = &pr

	nr := model.Resource{}
	nr.Params = map[string]*model.Param{}
	param2 := model.Param{Name: "Name2", Value: "Value2"}
	nr.Params["Name2"] = &param2
	resource.Needs["test2"] = &nr

	// All entries start with a RezolvrStatus of zero. Ensure all get updated appropriately
	MarkComponentResolvedStatus(&resource, RESOLVED)
	if resource.NeedsRezolvrStatus != RESOLVED || resource.ProvidesRezolvrStatus != RESOLVED {
		t.Error("Resource resolved status not properly updated")
	}
	if pr.RezolvrStatus != RESOLVED || nr.RezolvrStatus != RESOLVED {
		t.Error("Needs and Provides resource status - not properly set")
	}
	if param1.RezolvrStatus != RESOLVED || param2.RezolvrStatus != RESOLVED {
		t.Error("Resource parameter - rezolvr status - not properly updated")
	}
}

func Test_getProviderFromState(t *testing.T) {

	// Test empty values
	_, ok := getProviderFromState("", "", &model.State{})
	if ok {
		t.Error("Resource found when none provided")
	}

	var pState *model.State
	_, ok = getProviderFromState("", "", pState)
	if ok {
		t.Error("Resource found when none provided")
	}
}
