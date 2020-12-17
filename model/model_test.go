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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func getSampleState() string {
	return `environmentVars:
  appEnvProps:
  - name: app_message
    value: Hello from Rezolvr!
components:
- name: welcome
  type: component.web.app
  provides:
  - name: welcomeappservice
    type: service.web.app
    params:
    - name: port
      value: "3000"
  uses:
  needs:`
}

func getSampleComponent() string {
	return `name: welcome
type: component.web.app
provides:
  - type: service.web.app
    name: welcomeappservice
    params:
      - name: port
        value: 3000
uses:
needs:`
}

func Test_LoadState(t *testing.T) {
	var content []byte

	// Create an empty state when no content exists
	state, err := LoadState(content)
	assert.Nil(t, err)
	assert.NotNil(t, state)

	// Test the basic parts of state
	sampleState := getSampleState()
	content = []byte(sampleState)
	state, err = LoadState(content)
	assert.Nil(t, err)
	assert.NotNil(t, state)

	// Ensure the environment variable from the state exists
	envProp := state.Components["environment.properties"].Provides["environment.properties:appEnvProps"].Params["app_message"].Value
	assert.NotNil(t, envProp)
	assert.Equal(t, "Hello from Rezolvr!", envProp)

	// Ensure the welcome resource from the state exists
	welcome := state.Components["component.web.app:welcome"].Provides["service.web.app:welcomeappservice"].Params["port"].Value
	assert.NotNil(t, welcome)
	assert.Equal(t, "3000", welcome)
}

func Test_PrepStateForPersistence(t *testing.T) {
	sampleState := getSampleState()
	originalContent := []byte(sampleState)
	state, err := LoadState(originalContent)
	content, err := PrepStateForPersistence(state)
	assert.Nil(t, err)
	assert.NotNil(t, content)
}

func Test_paramTransformation(t *testing.T) {
	paramArray := make([]Param, 2)
	paramArray[0] = Param{Name: "One", Value: "OneValue"}
	paramArray[1] = Param{Name: "Two", Value: "TwoValue"}

	var xform map[string]*Param
	xform = transformPersistentParams(paramArray)

	parmOne := xform["One"]
	parmTwo := xform["Two"]

	assert.Equal(t, "TwoValue", parmTwo.Value, "Value should equal 2")
	assert.Equal(t, "OneValue", parmOne.Value)
}

func Test_transformState(t *testing.T) {
	pState := persistedState{}

	paramArrayA := make([]NvParam, 2)
	paramArrayA[0] = NvParam{Name: "One", Value: "OneValue"}
	paramArrayA[1] = NvParam{Name: "Two", Value: "TwoValue"}

	paramArrayB := make([]NvParam, 2)
	paramArrayB[0] = NvParam{Name: "Three", Value: "ThreeValue"}
	paramArrayB[1] = NvParam{Name: "Four", Value: "FourValue"}

	paramMap := map[string][]NvParam{}
	paramMap["alpha"] = paramArrayA
	paramMap["beta"] = paramArrayB
	pState.EnvironmentVars = paramMap

	components := make([]persistedComponent, 1)
	components[0] = getPersistedComponent()
	pState.Components = components
	state, err := transformState(&pState)
	assert.Nil(t, err)

	// Environment variables end up as provided resources within a component
	xformedComponent, ok := state.Components["environment.properties"]
	assert.Equal(t, ok, true)

	providedEnvVars := xformedComponent.Provides["environment.properties:alpha"]
	alphaParams := providedEnvVars.Params
	twoVal, ok := alphaParams["Two"]
	assert.Equal(t, ok, true)

	assert.Equal(t, "TwoValue", twoVal.Value)

	// Make sure the components exists
	xformedComponent2, ok := state.Components["sometype:SomeComponent"]
	assert.Equal(t, ok, true, "Component should exist")

	assert.Equal(t, "SomeComponent", xformedComponent2.Name)

	// Remaining component-specific tests are handled by Test_transformPersistentComponent()
}

func getPersistedComponent() persistedComponent {
	pComp := persistedComponent{}
	pComp.Name = "SomeComponent"
	pComp.Type = "sometype"

	providedArray := make([]persistedResource, 2)
	providedArray[0] = persistedResource{Name: "Provided One", Type: "sometype"}
	paramArray := make([]Param, 2)
	paramArray[0] = Param{Name: "One", Value: "OneValue"}
	paramArray[1] = Param{Name: "Two", Value: "TwoValue"}
	providedArray[0].Params = paramArray

	providedArray[1] = persistedResource{Name: "Provided Two", Type: "sometype2"}
	pComp.Provides = providedArray

	neededArray := make([]persistedResource, 2)
	neededArray[0] = persistedResource{Name: "Needed One", Type: "sometype"}
	neededArray[1] = persistedResource{Name: "Needed Two", Type: "sometype2"}
	pComp.Needs = neededArray
	return pComp
}

func Test_transformPersistentComponent(t *testing.T) {
	pRes := getPersistedComponent()
	xform, err := transformPersistentComponent(&pRes)
	assert.Nil(t, err)

	foundProvided, ok := xform.Provides["sometype"+IDSeparator+"Provided One"]
	assert.Equal(t, ok, true)
	assert.Equal(t, "Provided One", foundProvided.Name)

	foundNeeded, ok := xform.Needs["sometype2"+IDSeparator+"Needed Two"]
	assert.Equal(t, ok, true)
	assert.Equal(t, "Needed Two", foundNeeded.Name)

	foundParam, ok := foundProvided.Params["One"]
	assert.Equal(t, ok, true)
	assert.Equal(t, "OneValue", foundParam.Value)
}

func Test_LoadComponent(t *testing.T) {
	blankComponent, err := LoadComponent(nil)
	assert.Nil(t, err)
	assert.NotNil(t, blankComponent)

	samplePersistedComponent := getSampleComponent()
	content := []byte(samplePersistedComponent)
	loadedComponent, err := LoadComponent(content)
	assert.Nil(t, err)
	assert.NotNil(t, loadedComponent)

	myParm := loadedComponent.Provides["service.web.app:welcomeappservice"].Params["port"].Value
	assert.Equal(t, "3000", myParm)
}
