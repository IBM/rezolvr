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

// RezolvrDriver is the interface plugins use.
type RezolvrDriver interface {
	PrintMessage()
	TransformComponents(updatedComponents map[string]*Component, state *State,
		pluginDir string, outputDir string, platformSettings map[string]*Platform)
}

// IDSeparator is used to concatenate a resource's type and name
const IDSeparator = ":"

// RetryCount - Specify the number of recursive retries to use before marking something as unresolved
const RetryCount = 50

// Param represents a parameter associated with either a ProvidedResource or a Needed Resource
type Param struct {
	Name          string `yaml:"name"`
	Formula       string `yaml:"formula,omitempty"`
	Value         string `yaml:"value"`
	DefaultValue  string `yaml:"defaultValue,omitempty"`
	Required      bool   `yaml:"required,omitempty"`
	RezolvrStatus int    `yaml:",omitempty"`
}

// Resource represents something that a resource needs, uses, or provides
type Resource struct {
	Name          string
	Type          string
	Params        map[string]*Param
	RezolvrStatus int
}

// Platform settings are used for platform-specific names and values
type Platform struct {
	Params map[string]*Param
}

// Candidate terms:
// Larger "things": Package, Group, Pack, Bundle, Assortment, Bale, Component, Group, Parcel
// Smaller "things": Capability, Feature, Service, Component, Joule, Resource

// Component represents "something in the system"
type Component struct {
	Name                  string
	Type                  string
	Driver                string
	Description           string
	Provides              map[string]*Resource
	Uses                  map[string]*Resource
	Needs                 map[string]*Resource
	RezolvrStatus         int
	NeedsRezolvrStatus    int
	UsesRezolvrStatus     int
	ProvidesRezolvrStatus int
}

// State manages the overall state of the system
type State struct {
	Components map[string]*Component
}
