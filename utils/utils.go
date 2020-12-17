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
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"plugin"
	"rezolvr/model"
)

// CmdLineArgs is a simplified structure for managing the command line arguments
type CmdLineArgs struct {
	Command            string
	Subcommand         string
	EnvFile            string
	StateFile          string
	ExportFile         string
	OutputDir          string
	ComponentsToAdd    []string
	ComponentsToDelete []string
}

// ParseArgs - parse command line arguments
func ParseArgs(args []string) (*CmdLineArgs, error) {
	idx := 1
	if len(args) < 3 {
		return nil, errors.New("command line arguments missing")
	}
	cla := CmdLineArgs{}
	cla.ComponentsToAdd = make([]string, 0)
	cla.Command = args[idx]
	// The default output location is './out/'
	cla.OutputDir = "./out/"

	idx++
	// The keyword 'whatif' is reserved for future use
	if cla.Command == "whatif" {
		cla.Subcommand = args[idx]
		idx++
	}
	for idx < len(args) {
		flag := args[idx]
		idx++
		// Make sure each flag has a target value
		if idx >= len(args) {
			return nil, errors.New("Unmatching command line args")
		}
		target := args[idx]
		idx++
		if flag == "-a" || flag == "--add-component" {
			cla.ComponentsToAdd = append(cla.ComponentsToAdd, target)
		} else if flag == "-d" || flag == "--delete-component" {
			cla.ComponentsToDelete = append(cla.ComponentsToDelete, target)
		} else if flag == "-e" || flag == "--environment" {
			cla.EnvFile = target
		} else if flag == "-s" || flag == "--source" {
			cla.StateFile = target
		} else if flag == "-x" || flag == "--export" {
			cla.ExportFile = target
		} else if flag == "-o" || flag == "--output-dir" {
			cla.OutputDir = target
		} else {
			log.Printf("Warning: Command line argument unknown: %s. Ignoring.", flag)
		}
	}

	return &cla, nil
}

// LoadPlugin - Attempt to dynamically load a plugin
func LoadPlugin(pluginPathAndName string) (model.RezolvrDriver, error) {
	curPlugin, err := plugin.Open(pluginPathAndName)
	if err != nil {
		msg := fmt.Sprintf("Error loading plugin: %v", err)
		log.Println(msg)
		return nil, err
	}

	foundPlugin, err := curPlugin.Lookup("RezolvrDriver")
	if err != nil {
		msg := fmt.Sprintf("Error loading the Rezolvr plugin: %v . Error: %v", pluginPathAndName, err)
		log.Println(msg)
		return nil, err
	}

	rDriver, ok := foundPlugin.(model.RezolvrDriver)
	if !ok {
		log.Println("Plugin methods not found")
		return nil, err
	}
	return rDriver, nil
}

// LoadFile - Load a file from disk
func LoadFile(filename string, mustExist bool) ([]byte, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		if mustExist == false && os.IsNotExist(err) {
			emptyContent := make([]byte, 1)
			return emptyContent, nil
		}
		return nil, err
	}
	return content, nil
}

// SaveFile - save the file to disk
func SaveFile(filename string, content []byte) error {
	os.Rename(filename, filename+".backup")
	err := ioutil.WriteFile(filename, content, 0644)
	return err
}
