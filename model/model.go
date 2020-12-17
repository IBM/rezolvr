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
	"log"

	"gopkg.in/yaml.v2"
)

//NvParam represents a simpler name/value parameter
type NvParam struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

// persistedResource is a flattened representation of a resource. It represents a resource stored in YAML
type persistedResource struct {
	Name   string
	Type   string  `yaml:"type"`
	Params []Param `yaml:"params"`
}

// PersistedComponent represents "something in the system" (how it's stored in YAML)
type persistedComponent struct {
	Name        string
	Type        string `yaml:"type"`
	Driver      string `yaml:"driver"`
	Description string
	Provides    []persistedResource
	Uses        []persistedResource
	Needs       []persistedResource
}

type persistedState struct {
	EnvironmentVars map[string][]NvParam `yaml:"environmentVars"`
	Components      []persistedComponent `yaml:"components"`
}

// LoadState - given the contents of a YAML file, convert it into a collection of components
func LoadState(content []byte) (*State, error) {
	//content, err := loadFile(fileName, false)
	if len(content) < 2 {
		// No existing state was found; create a new empty state
		log.Println("The state is empty. Creating a brand new state model...")
		s := State{}
		components := make(map[string]*Component)
		c := Component{Name: "environment.properties"}
		c.Needs = make(map[string]*Resource)
		c.Uses = make(map[string]*Resource)
		c.Provides = make(map[string]*Resource)
		components["environment.properties"] = &c
		s.Components = components
		return &s, nil
	}

	persistedState := &persistedState{}
	err := yaml.Unmarshal(content, persistedState)
	if err != nil {
		return nil, err
	}

	state, err := transformState(persistedState)
	return state, err
}

// PrepStateForPersistence save the state back to the filesystem
func PrepStateForPersistence(state *State) ([]byte, error) {
	pState, err := flattenState(state)
	if err != nil {
		return nil, err
	}
	content, err := yaml.Marshal(pState)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func transformState(persistedState *persistedState) (*State, error) {
	state := &State{}
	state.Components = make(map[string]*Component)

	// Transform environment variables into their internal representation
	envProvides := make(map[string]*Resource)
	for curEnvVarCategoryName, curEnvVarCategoryNv := range persistedState.EnvironmentVars {
		curEnvVarCategory := transformNvParamsToParams(curEnvVarCategoryNv)
		localProps := transformPersistentParams(curEnvVarCategory)
		resID := "environment.properties:" + curEnvVarCategoryName

		envProvidedResource := Resource{}
		envProvidedResource.Name = curEnvVarCategoryName
		envProvidedResource.Type = "environment.properties"
		envProvidedResource.Params = localProps
		envProvides[resID] = &envProvidedResource
	}

	// The environment properties should appear to be another component
	propComponent := Component{}
	propComponent.Name = "environment.properties"
	propComponent.Type = "environment.properties"
	propComponent.Provides = envProvides
	state.Components[propComponent.Name] = &propComponent

	// Transform components into their internal representation
	for _, curPersistedComponent := range persistedState.Components {
		curComponent, err := transformPersistentComponent(&curPersistedComponent)
		if err == nil {
			key := curComponent.Type + IDSeparator + curComponent.Name
			state.Components[key] = curComponent
		}
	}

	return state, nil
}

func flattenParams(resources map[string]*Resource) *[]persistedResource {

	transformedResources := make([]persistedResource, len(resources))
	resCount := 0
	for _, curResource := range resources {

		// Flatten params
		parmCount := 0
		transformedParams := make([]Param, len(curResource.Params))
		for _, curParam := range curResource.Params {
			curParam.RezolvrStatus = 0
			transformedParams[parmCount] = *curParam
			parmCount++
		}
		transformedResources[resCount] = persistedResource{Name: curResource.Name, Type: curResource.Type, Params: transformedParams}
		resCount++
	}
	return &transformedResources
}

// flattenState converts state information into a format that's consistent with the YAML files
func flattenState(state *State) (*persistedState, error) {

	// Flattend envrionment variables
	ps := persistedState{}
	ps.EnvironmentVars = make(map[string][]NvParam)
	// Convert environment properties back into a map
	allEnvProps := state.Components["environment.properties"].Provides
	for _, curEnvProp := range allEnvProps {
		envVarCategoryName := curEnvProp.Name
		envVarArray := make([]NvParam, len(curEnvProp.Params))
		count := 0
		for _, curProp := range curEnvProp.Params {
			nvParam := NvParam{Name: curProp.Name, Value: curProp.Value}
			envVarArray[count] = nvParam
			count++
		}
		ps.EnvironmentVars[envVarCategoryName] = envVarArray
	}

	// Flatten components
	componentCount := 0
	totalComponents := len(state.Components) - 1 // Don't include environment.properties
	ps.Components = make([]persistedComponent, totalComponents)
	for compName, curComp := range state.Components {
		if compName != "environment.properties" {
			pc := persistedComponent{}
			pc.Name = curComp.Name
			pc.Type = curComp.Type
			pc.Driver = curComp.Driver
			pc.Description = curComp.Description

			pc.Needs = *flattenParams(curComp.Needs)
			pc.Uses = *flattenParams(curComp.Uses)
			pc.Provides = *flattenParams(curComp.Provides)

			ps.Components[componentCount] = pc
			componentCount++
		}
	}
	return &ps, nil
}

// LoadComponent - given the contents of a YAML file, load the file and convert it into a component
func LoadComponent(content []byte) (*Component, error) {
	if content == nil {
		return &Component{}, nil
	}

	persistedComponent := &persistedComponent{}
	err := yaml.Unmarshal(content, persistedComponent)
	if err != nil {
		return nil, err
	}

	component, err := transformPersistentComponent(persistedComponent)
	if err != nil {
		return nil, err
	}

	return component, nil
}

func transformNvParamsToParams(nvParams []NvParam) []Param {
	params := make([]Param, len(nvParams))
	for idx, nvParam := range nvParams {
		p := Param{Name: nvParam.Name, Value: nvParam.Value}
		params[idx] = p
	}
	return params
}

func transformPersistentParams(params []Param) map[string]*Param {
	paramMap := make(map[string]*Param)
	for idx := range params {
		val := &params[idx]
		paramMap[val.Name] = val
	}
	return paramMap
}

func transformPersistentResource(pResource *[]persistedResource) map[string]*Resource {
	resources := make(map[string]*Resource)
	for _, val := range *pResource {
		newResource := Resource{}
		newResource.Name = val.Name
		newResource.Type = val.Type
		newResource.Params = transformPersistentParams(val.Params)
		// If the name is blank, then don't include it as the resourceID
		if len(newResource.Name) < 1 {
			resources[newResource.Type] = &newResource
		} else {
			resourceID := newResource.Type + IDSeparator + newResource.Name
			resources[resourceID] = &newResource
		}
	}
	return resources
}

func transformPersistentComponent(pComponent *persistedComponent) (*Component, error) {
	component := &Component{}
	component.Name = pComponent.Name
	component.Type = pComponent.Type
	component.Driver = pComponent.Driver
	component.Description = pComponent.Description

	component.Provides = transformPersistentResource(&pComponent.Provides)
	component.Uses = transformPersistentResource(&pComponent.Uses)
	component.Needs = transformPersistentResource(&pComponent.Needs)

	return component, nil
}
