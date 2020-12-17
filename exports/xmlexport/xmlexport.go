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

package xmlexport

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"rezolvr/model"
)

// createLinkageCell creates an mxCell which serves as an arrow from the source cell to the target cell
func createLinkageCell(sourceCell *mxCell, targetCell *mxCell, ID string, parentID string) *mxCell {

	// Default style values are acceptable
	fullStyle := fmt.Sprintf("%v", getNewLinkCellStyle())

	// Use default geometric settings for this cell
	g := mxGeometry{Width: 110, Height: 110, As: "geometry"}

	// Create linkage between the source and target cells
	p0 := mxPoint{X: sourceCell.Geom.X, Y: sourceCell.Geom.Y, As: "sourcePoint"}
	p1 := mxPoint{X: targetCell.Geom.X, Y: targetCell.Geom.Y, As: "targetPoint"}

	g.Points = make([]*mxPoint, 2)
	g.Points[0] = &p0
	g.Points[1] = &p1

	// Create the cell with all of the previously defined components
	c := mxCell{
		ID:     ID,
		Parent: parentID,
		Value:  "",
		Style:  fullStyle,
		Vertex: 1,
		Geom:   &g,
		Source: sourceCell.ID,
		Target: targetCell.ID,
		Edge:   "1",
	}

	return &c
}

// createLinkageCellFromProvidedResource takes a "Provides" resource and creates an associated diagram object
func createCellFromProvidedResource(pr *model.Resource, parentID string) *mxCell {

	// Look up the image to use for this object, and create the associated style
	iconName := getIconForResourceType(pr.Type)
	fullStyle := fmt.Sprintf("%v", getNewCellStyle(iconName))

	// Use default geometric settings for this cell. They will be revised after all cells have been created
	g := mxGeometry{X: 100, Y: 200, Width: 110, Height: 110, As: "geometry"}

	// Create the cell with all of the previously defined components
	c := mxCell{
		ID:     pr.Name + model.IDSeparator + pr.Type,
		Parent: parentID,
		Value:  pr.Name + model.IDSeparator + pr.Type,
		Style:  fullStyle,
		Vertex: 1,
		Geom:   &g,
	}
	return &c
}

func positionCells(allCells map[string]*mxCell) {
	// Position all of the cells so they don't overlap
	curX := 100
	curY := 100
	for _, curCell := range allCells {
		if curCell.ID != "0" && curCell.ID != "1" {
			curCell.Geom.X = curX
			curCell.Geom.Y = curY
			curX = (curX + 500) % 1000
			if curX < 101 {
				curY = curY + 200
			}
		}
	}
}

// ExportState exports the contents of the state to a format that can be read by a diagramming tool
func ExportState(state *model.State, exportFilename string) error {
	log.Printf("Exporting state to file: %v\n", exportFilename)

	allCells := map[string]*mxCell{}

	// Create the base document structure with mxGraphModel and root objects
	r := root{}
	gm := mxGraphModel{Dx: 1168, Dy: 738, Grid: 1, GridSize: 10, Guides: 1, Tooltips: 1, Connect: 1, Arrows: 1, Fold: 1, Page: 1, PageScale: 1, PageWidth: 850, PageHeight: 1100, Math: 0, Shadow: 0, Root: &r}

	// Add the two basic (empty) cells
	parentID := "1"
	allCells["0"] = &mxCell{ID: "0"}
	allCells[parentID] = &mxCell{ID: parentID, Parent: "0"}

	// Iterate over all provided resources and create a cell for each "provides"
	for _, curRes := range state.Components {
		for _, curProvides := range curRes.Provides {
			myCell := createCellFromProvidedResource(curProvides, parentID)
			allCells[curProvides.Type+model.IDSeparator+curProvides.Name] = myCell
		}
	}

	positionCells(allCells)

	// Create links between the "needs" and the "provides"
	for _, curRes := range state.Components {
		// TODO: Assumption is that everything is tied to the first "provides". Consider changing this in the future
		var providesKey string
		for curProvidesKey := range curRes.Provides {
			providesKey = curProvidesKey
		}

		var linkCount int = 0
		for _, curNeeds := range curRes.Needs {
			needsKey := curNeeds.Type + model.IDSeparator + curNeeds.Name
			sourceCell, ok := allCells[providesKey]
			if !ok {
				log.Printf("Warning: Provided resource not found: %v\n", providesKey)
			} else {
				targetCell, ok := allCells[needsKey]
				if !ok {
					log.Printf("Warning: Needed resource not found: %v\n", needsKey)
				} else {
					// Create the link between the two cells
					linkedID := "diaglink" + fmt.Sprint(linkCount)
					linkedCell := createLinkageCell(sourceCell, targetCell, linkedID, parentID)
					allCells[linkedID] = linkedCell
					linkCount++
				}
			}
		}
	}

	// Convert the map to an array (I wish this was simpler)
	for _, curCell := range allCells {
		r.MxCells = append(r.MxCells, curCell)
	}

	// Marshall the XML into an array of bytes, and save the results to the file system
	output, err := xml.MarshalIndent(gm, " ", "  ")
	if err != nil {
		log.Printf("Unable to marshal the XML: %v\n", err)
		return err
	}
	err = saveFile(exportFilename, output)
	if err != nil {
		log.Printf("Error saving the file: %v\n", err)
		return err
	}
	return nil
}

func saveFile(filename string, content []byte) error {
	os.Rename(filename, filename+".backup")
	err := ioutil.WriteFile(filename, content, 0644)
	return err
}
