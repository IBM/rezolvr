package main

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
// limitations under the License.package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"rezolvr/model"
	"rezolvr/utils"

	xmlexport "rezolvr/exports/xmlexport"
	"rezolvr/validation"
)

// Package-level variables
var state *model.State
var allNewComponents []*model.Component
var pluginDir string
var driverName string
var rezolvrPlugin model.RezolvrDriver
var platformSettings map[string]*model.Platform

func loadRezolvrFiles(cliArgs *utils.CmdLineArgs) error {
	// Load the component file(s) and the environment file
	allNewComponents = make([]*model.Component, len(cliArgs.ComponentsToAdd))
	for idx, curComponentFile := range cliArgs.ComponentsToAdd {
		log.Printf("Attempting to load file: %s\n", curComponentFile)
		content, err := utils.LoadFile(curComponentFile, true)
		if err != nil {
			log.Printf("Error loading file: %v\n", err)
			return err
		}
		curComponent, err := model.LoadComponent(content)
		if err != nil {
			log.Printf("Error loading component: %v\n", err)
			return err
		}
		allNewComponents[idx] = curComponent
	}

	environmentFile := cliArgs.EnvFile
	content, err := utils.LoadFile(environmentFile, false)
	initialEnv, err := model.LoadComponent(content)
	if err != nil {
		log.Printf("Error loading environment details: %v", err)
		return err
	}
	driverName = initialEnv.Driver

	content, err = utils.LoadFile(cliArgs.StateFile, false)
	if err != nil {
		return err
	}
	state, err = model.LoadState(content)
	if err != nil {
		log.Printf("Error loading state: %v", err)
		return err
	}

	// Combine the environment properties with the existing state environment properties
	// New environment properties take precedent over existing state envrionment properties
	stateProps := state.Components["environment.properties"].Provides
	for envKey, envVal := range initialEnv.Provides {
		stateProps[envKey] = envVal
	}

	// Attempt to load a plugin to handle the transformation
	driverFullName := pluginDir + driverName + "/plugin" + driverName + ".so"
	log.Printf("Attempting to load plugin: %s\n", driverFullName)
	rezolvrPlugin, err = utils.LoadPlugin(driverFullName)
	if err != nil || rezolvrPlugin == nil {
		msg := fmt.Sprintf("Unsuitable driver found: *%v*\n", driverName)
		err = errors.New(msg)
		return err
	}

	// Load platform-specific settings
	platformSettings = make(map[string]*model.Platform)
	for _, envVal := range initialEnv.Uses {
		if envVal.Type == "platform.settings" {
			curPlatform := model.Platform{Params: envVal.Params}
			platformSettings[envVal.Name] = &curPlatform
		}
	}
	return nil
}

func applyUpdatedComponents(cliArgs *utils.CmdLineArgs) error {

	// Ensure the state of previously-defined components / resources is valid
	err := validation.ValidateState(state)
	if err != nil {
		return err
	}

	// If there are components to remove, remove them from the state
	recalculateAllComponents := false
	if len(cliArgs.ComponentsToDelete) > 0 {
		validation.RemoveComponentsFromState(state, cliArgs.ComponentsToDelete)
		recalculateAllComponents = true
	}

	var componentsToResolve map[string]*model.Component
	if recalculateAllComponents == true {
		// There was a delete, so every component - including new components - should be "re-resolved"
		componentsToResolve = map[string]*model.Component{}
		for curExistingComponentID, curExistingComponent := range state.Components {
			if !(curExistingComponentID == "environment.properties") {
				componentsToResolve[curExistingComponentID] = curExistingComponent
			}
		}
		// Include newly added components as well
		for _, v := range allNewComponents {
			componentID := v.Type + model.IDSeparator + v.Name
			componentsToResolve[componentID] = v
		}

	} else {
		// Check to see if any existing components are dependent upon the new component
		// This needs to be recursive to get a complete list of all components which must be re-resolved
		componentsNeedingUpdate := map[string]*model.Component{}
		for _, v := range allNewComponents {
			componentID := v.Type + model.IDSeparator + v.Name
			componentsNeedingUpdate[componentID] = v
		}

		log.Println("Locating impacted components which must be resolved...")
		componentsToResolve = validation.GetImpactedComponents(state, componentsNeedingUpdate)
	}

	log.Println("Resolving components...")
	allUpdatedComponents, err := utils.ResolveAllComponents(state, componentsToResolve)
	if err != nil {
		return err
	}
	// Transform the components into output files
	log.Println("Transforming components...")
	rezolvrPlugin.TransformComponents(allUpdatedComponents, state, pluginDir+driverName+"/", cliArgs.OutputDir, platformSettings)

	// Add the updated components to the state
	log.Println("Adding updated components to the state of the system...")
	for k, v := range allUpdatedComponents {
		state.Components[k] = v
	}

	// Persist the updated state
	log.Println("Saving the system state...")
	content, err := model.PrepStateForPersistence(state)
	err = utils.SaveFile(cliArgs.StateFile, content)
	return nil
}

func main() {
	log.Println("rezolvr version: 0.0.1")

	cliArgs, err := utils.ParseArgs(os.Args)
	if err != nil {
		log.Println(err)
		log.Fatal("Usage: rezolvr apply -a/-r <component file(s)> -e <environment file> -s <state file>")
	}
	var ok bool
	pluginDir, ok = os.LookupEnv("REZOLVR_PLUGINDIR")
	if !ok {
		homeDir := os.Getenv("HOME")
		pluginDir = homeDir + "/.rezolvr/plugins/"
	}
	log.Printf("Plugin directory: %s\n", pluginDir)

	// Only 'apply' and 'export' are supported for now
	if cliArgs.Command == "export" {
		content, err := utils.LoadFile(cliArgs.StateFile, false)
		if err != nil {
			log.Fatal("Error loading state file")
		}
		state, err := model.LoadState(content)
		if err != nil {
			msg := fmt.Sprintf("Error loading state: %v", err)
			log.Fatal(msg)
		}
		err = xmlexport.ExportState(state, cliArgs.ExportFile)
		if err != nil {
			fmt.Printf("Error exporting state: %v\n", err)
		}
	} else if cliArgs.Command == "apply" {
		err = loadRezolvrFiles(cliArgs)
		if err == nil {
			err = applyUpdatedComponents(cliArgs)
		}
		if err != nil {
			log.Fatalf("Error encountered: %v\n", err)
		}
	} else {
		log.Fatal("Usage: only the 'apply' and 'export' commands are supported")
	}
	log.Println("Rezolvr completed")
}
