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

package validation

import (
	"errors"
	"log"
	"rezolvr/model"
	"rezolvr/utils"
)

// RemoveComponentsFromState - Remove components which have been marked for deletion
func RemoveComponentsFromState(state *model.State, componentsToDelete []string) {
	for _, res := range componentsToDelete {
		log.Printf("About to delete: %v\n", res)
		delete(state.Components, res)
	}
}

// GetImpactedComponents locates all new and existing components which must be re-resolved
func GetImpactedComponents(state *model.State, componentsNeedingUpdate map[string]*model.Component) map[string]*model.Component {
	componentAdded := false

	for _, curComponent := range componentsNeedingUpdate {
		utils.MarkComponentResolvedStatus(curComponent, utils.UNRESOLVED)
		for curExistingComponentID, curExistingComponent := range state.Components {
			if !(curExistingComponentID == "environment.properties") {
				for curNeedID := range curExistingComponent.Needs {
					_, ok := curComponent.Provides[curNeedID]
					if ok {
						// This may have been previously found. Make sure it's new
						_, okExists := componentsNeedingUpdate[curExistingComponentID]
						if !okExists {
							log.Printf("Component found which needs recalc: %s - %s\n", curExistingComponentID, curNeedID)
							componentsNeedingUpdate[curExistingComponentID] = curExistingComponent
							componentAdded = true
						}
						break
					}
				}
			}
		}
	}

	if componentAdded {
		log.Println("Dependencies found on existing components. Recursing.")
		GetImpactedComponents(state, componentsNeedingUpdate)
	} else {
		log.Println("No new dependencies found... returning.")
	}
	return componentsNeedingUpdate
}

// ValidateState ensures that all existing components have been resolved
func ValidateState(state *model.State) error {
	log.Println("Validating the integrity of the current state...")

	// Collect all of the 'provides' resources into a single map
	allProvides := make(map[string]*model.Resource)
	for _, curComponent := range state.Components {
		curProvides := curComponent.Provides
		for k, v := range curProvides {
			allProvides[k] = v
		}
	}

	// Transform environment variables into their internal representation
	for _, curComp := range state.Components {
		if curComp.Needs != nil {
			for curNeedID := range curComp.Needs {
				_, ok := allProvides[curNeedID]
				if !ok {
					return errors.New("An unmatched component was located within the state: " + curComp.Name + " - " + curNeedID)
				}
			}
		}
	}
	log.Println("State validation complete")

	return nil
}
