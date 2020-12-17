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
	"bytes"
	"errors"
	"fmt"
	"log"
	"rezolvr/model"
	"text/template"
)

// RESOLVED - a component's / resource's needs have been resolved
const RESOLVED = 1

// UNRESOLVED - a component's / resource's needs have not yet been resolved
const UNRESOLVED = 2

func markParamsResolvedStatus(params map[string]*model.Param, newRezolvrStatus int) {
	for curParamID, curParam := range params {
		curParam.RezolvrStatus = newRezolvrStatus
		params[curParamID] = curParam
	}
}

func markNeedsElementsResolvedStatus(elems map[string]*model.Resource, newRezolvrStatus int) {
	for curNeedID, curNeed := range elems {
		curNeed.RezolvrStatus = newRezolvrStatus
		markParamsResolvedStatus(curNeed.Params, newRezolvrStatus)
		elems[curNeedID] = curNeed
	}
}

func markUsesElementsResolvedStatus(elems map[string]*model.Resource, newRezolvrStatus int) {
	for curUseID, curUse := range elems {
		curUse.RezolvrStatus = newRezolvrStatus
		markParamsResolvedStatus(curUse.Params, newRezolvrStatus)
		elems[curUseID] = curUse
	}
}

func markProvidesElementsResolvedStatus(elems map[string]*model.Resource, newRezolvrStatus int) {
	for curProvideID, curProvide := range elems {
		curProvide.RezolvrStatus = newRezolvrStatus
		markParamsResolvedStatus(curProvide.Params, newRezolvrStatus)
		elems[curProvideID] = curProvide
	}
}

// MarkComponentResolvedStatus - convenience method to mark the resolved status of a component and it's resources
func MarkComponentResolvedStatus(curComponent *model.Component, newRezolvrStatus int) {
	curComponent.RezolvrStatus = newRezolvrStatus
	markNeedsElementsResolvedStatus(curComponent.Needs, newRezolvrStatus)
	markProvidesElementsResolvedStatus(curComponent.Provides, newRezolvrStatus)
	curComponent.NeedsRezolvrStatus = newRezolvrStatus
	curComponent.ProvidesRezolvrStatus = newRezolvrStatus

}

func markParamsWithValuesAsResolved(components map[string]*model.Component) {
	for _, curComponent := range components {
		allProvidesRezolvrStatus := RESOLVED
		for _, curProvides := range curComponent.Provides {
			providesRezolvrStatus := RESOLVED
			for _, curParam := range curProvides.Params {
				if len(curParam.Value) > 0 {
					curParam.RezolvrStatus = RESOLVED
				} else {
					providesRezolvrStatus = UNRESOLVED
					allProvidesRezolvrStatus = UNRESOLVED
				}
			}
			curProvides.RezolvrStatus = providesRezolvrStatus
		}
		curComponent.ProvidesRezolvrStatus = allProvidesRezolvrStatus
	}
}

// ResolveAllComponents - attempts to link "needs" to a component's "uses" and "provides" resources.
func ResolveAllComponents(state *model.State, componentsToResolve map[string]*model.Component) (map[string]*model.Component, error) {

	// Some parameters are already resolved, because they have a Value. Mark these appropriately.
	markParamsWithValuesAsResolved(componentsToResolve)
	markParamsWithValuesAsResolved(state.Components)

	initialUnresolvedComponents := make(map[string]*model.Component)
	for k, v := range componentsToResolve {
		initialUnresolvedComponents[k] = v
	}

	fullyResolvedComponents := make(map[string]*model.Component)

	unresolvedComponentCount := len(componentsToResolve)
	attempts := 0
	for unresolvedComponentCount > 0 {
		attempts++
		for curComponentID, curComponent := range componentsToResolve {
			delete(componentsToResolve, curComponentID)
			needsRezolvrStatus := resolveComponentNeeds(state, curComponent, initialUnresolvedComponents)
			if needsRezolvrStatus == RESOLVED {
				// Resolve both 'uses' and 'provides' sections
				resolveComponentUses(state, curComponent)

				providesRezolvrStatus := RESOLVED
				resolveComponentProvides(state, curComponent)
				curComponent.RezolvrStatus = providesRezolvrStatus
				fullyResolvedComponents[curComponentID] = curComponent
			} else {
				componentsToResolve[curComponentID] = curComponent
			}
		}
		// TODO: Catch circular & unresolved dependencies
		unresolvedComponentCount = len(componentsToResolve)
		if unresolvedComponentCount > 0 && attempts > model.RetryCount {
			err := errors.New("Dependency infinite loop encountered. This is usually due to a missing resource. Please fix")
			log.Println("Dependency infinite loop encountered... please fix")
			return nil, err
		}
	}
	return fullyResolvedComponents, nil
}

func resolveNeedParams(needParams map[string]*model.Param, providedParams map[string]*model.Param, res *model.Component) int {
	rezolvrStatus := RESOLVED
	for _, curNeedParam := range needParams {
		if curNeedParam.RezolvrStatus == UNRESOLVED {
			// Find a value for the parameter
			providedParam, ok := providedParams[curNeedParam.Name]
			if ok {
				if providedParam.RezolvrStatus == UNRESOLVED {
					rezolvrStatus = UNRESOLVED
				} else {
					curNeedParam.Value = providedParam.Value
					curNeedParam.RezolvrStatus = RESOLVED
				}
			} else {
				// Check to see if there's a default value
				if len(curNeedParam.DefaultValue) > 0 {
					curNeedParam.Value = curNeedParam.DefaultValue
					curNeedParam.RezolvrStatus = RESOLVED
				} else if curNeedParam.Required {
					// No value has been found for a required parameter. Throw an error
					msg := fmt.Sprintf("Missing required parameter for %s - %s: %s", res.Name, res.Type, curNeedParam.Name)
					log.Println(msg)
					rezolvrStatus = UNRESOLVED
				}
			}
		}
	}
	return rezolvrStatus
}

func locateProviderForNeed(needType string, needName string, providesMap map[string]*model.Resource) (*model.Resource, bool) {
	key := needType + model.IDSeparator + needName
	found, ok := providesMap[key]
	return found, ok
}

func getProviderFromState(pType string, providedName string, state *model.State) (*model.Resource, bool) {
	var provider *model.Resource
	if state == nil {
		return provider, false
	}
	for _, curComponent := range state.Components {
		key := pType + model.IDSeparator + providedName
		provider, ok := curComponent.Provides[key]
		if ok {
			return provider, true
		}
	}
	return provider, false
}

func locateProviderFromModifiedComponents(pType string, providedName string, components map[string]*model.Component) (*model.Resource, bool) {
	var provider *model.Resource
	for _, curComponent := range components {
		key := pType + model.IDSeparator + providedName
		provider, ok := curComponent.Provides[key]
		if ok {
			return provider, true
		}
	}
	return provider, false
}

func resolveComponentNeeds(state *model.State, comp *model.Component, componentsToResolve map[string]*model.Component) int {
	// Given the existing state of the system, and a target environment,
	// attempt to resolve a component's needs
	log.Printf("The type of component to be resolved: %s\n", comp.Type)
	env := state.Components["environment.properties"]
	needsRezolvrStatus := RESOLVED

	for _, curNeed := range comp.Needs {
		// Combine all of the found parameters into a single map
		// The precedence from highest to lowest: environment, modified component, state
		combinedProvidedParams := make(map[string]*model.Param)

		// Determine if the resource exists in the current state of the system
		stateProvider, stateProviderOk := getProviderFromState(curNeed.Type, curNeed.Name, state)
		if stateProviderOk {
			for k, v := range stateProvider.Params {
				combinedProvidedParams[k] = v
			}
		}

		// Look for resources not yet fully resolved
		modResourceProvider, modResourceProviderOk := locateProviderFromModifiedComponents(curNeed.Type, curNeed.Name, componentsToResolve)
		if modResourceProviderOk {
			for k, v := range modResourceProvider.Params {
				combinedProvidedParams[k] = v
			}
		}

		// First attempt to find this need as a provider in the env
		// (This takes precedence over what's stored in the current state)
		envProvider, envProviderOk := locateProviderForNeed(curNeed.Type, curNeed.Name, env.Provides)
		if envProviderOk {
			for k, v := range envProvider.Params {
				combinedProvidedParams[k] = v
			}
		}

		if !stateProviderOk && !envProviderOk && !modResourceProviderOk {
			msg := fmt.Sprintf("Need missing for %s - %s: %s %s", comp.Name, comp.Type, curNeed.Type, curNeed.Name)
			log.Println(msg)
			return UNRESOLVED
		}

		paramsRezolvrStatus := resolveNeedParams(curNeed.Params, combinedProvidedParams, comp)
		curNeed.RezolvrStatus = paramsRezolvrStatus
		if paramsRezolvrStatus == UNRESOLVED {
			needsRezolvrStatus = UNRESOLVED
		}
	}
	comp.NeedsRezolvrStatus = needsRezolvrStatus
	log.Printf("Resulting status for current needs: %v", comp.NeedsRezolvrStatus)
	return comp.NeedsRezolvrStatus
}

func resolveComponentUses(state *model.State, component *model.Component) int {

	// Determine if any formulas exist in the 'uses' section
	log.Printf("Resolving any 'uses' formulas for: %s \n", component.Type)

	// The following two variables are made available to the eval() method
	data := map[string]interface{}{
		"Needs":     component.Needs,
		"Component": component,
	}

	for _, curUse := range component.Uses {

		for _, curUseParam := range curUse.Params {

			if len(curUseParam.Formula) > 0 {
				log.Printf("Current formula to resolve: %v\n", curUseParam.Formula)
				t := template.Must(template.New("").Parse(curUseParam.Formula))
				buf := &bytes.Buffer{}
				err := t.Execute(buf, data)
				if err != nil {
					msg := fmt.Sprintf("Error resolving a 'Uses' formula: %s", curUseParam.Formula)
					log.Printf(msg)
				} else {
					stringVal := buf.String()
					curUseParam.Value = stringVal
				}
			}
		}
	}
	markUsesElementsResolvedStatus(component.Uses, RESOLVED)
	component.UsesRezolvrStatus = RESOLVED

	log.Println("All uses forumlas successfully executed")
	return component.UsesRezolvrStatus
}

func resolveComponentProvides(state *model.State, component *model.Component) int {
	// This should only be called after all of the resource needs have been resolved
	// Determine if any formulas exist in the 'provides' section
	log.Printf("Resolving any 'provides' formulas for: %s \n", component.Type)

	// The following two variables are made available to the eval() method
	data := map[string]interface{}{
		"Needs":     component.Needs,
		"Component": component,
	}

	for _, curProvide := range component.Provides {

		for _, curProvideParam := range curProvide.Params {

			if len(curProvideParam.Formula) > 0 {
				log.Printf("Current formula to resolve: %v\n", curProvideParam.Formula)
				t := template.Must(template.New("").Parse(curProvideParam.Formula))
				buf := &bytes.Buffer{}
				err := t.Execute(buf, data)
				if err != nil {
					msg := fmt.Sprintf("Error resolving a 'provides' formula: %s", curProvideParam.Formula)
					log.Printf(msg)
				} else {
					stringVal := buf.String()
					curProvideParam.Value = stringVal
				}
			}
		}
	}
	markProvidesElementsResolvedStatus(component.Provides, RESOLVED)
	component.ProvidesRezolvrStatus = RESOLVED

	log.Println("All provides forumlas successfully executed")
	return component.ProvidesRezolvrStatus
}
