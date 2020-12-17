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
)

// Represents a link from one object within a diagram to another object
// These attributes appear to be much different than a typical cell's style
type linkCellStyle struct {
	HTML                 int
	LabelBackgroundColor string
	EndArrow             string
	EndFill              int
	EndSize              int
	JettySize            string
	OrthogonalLoop       int
	StrokeWidth          int
	FontSize             int
	EntryX               float64
	EntryY               float64
	EntryDx              int
	EntryDy              int
	ExitX                float64
	ExitY                float64
	ExitDx               int
	ExitDy               int
}

// Convert a linked cell style into a string. This is used as the style attrib on an mxCell
func (lcs *linkCellStyle) String() string {
	return fmt.Sprintf("html=%d;labelBackgroundColor=%s;endArrow=%s;endFill=%d;endSize=%d;jettySize=%s;orthogonalLoop=%d;strokeWidth=%d;fontSize=%d;entryX=%f;entryY=%f;entryDx=%d;entryDy=%d;exitX=%f;exitY=%f;exitDx=%d;exitDy=%d;", lcs.HTML, lcs.LabelBackgroundColor, lcs.EndArrow, lcs.EndFill, lcs.EndSize, lcs.JettySize, lcs.OrthogonalLoop, lcs.StrokeWidth, lcs.FontSize, lcs.EntryX, lcs.EntryY, lcs.EntryDx, lcs.EntryDy, lcs.ExitX, lcs.ExitY, lcs.ExitDx, lcs.ExitDy)
}

// Create a new linked cell style with default values
func getNewLinkCellStyle() *linkCellStyle {
	newCell := linkCellStyle{HTML: 1, LabelBackgroundColor: "#ffffff", EndArrow: "classic", EndFill: 1, EndSize: 6, JettySize: "auto", OrthogonalLoop: 1, StrokeWidth: 1, FontSize: 14, EntryX: 0.5, EntryY: 0, EntryDx: 0, EntryDy: 0, ExitX: 0, ExitY: 0.75, ExitDx: 0, ExitDy: 0}
	return &newCell
}

// cellStyle represents the styling of an object within a diagram
type cellStyle struct {
	HTML                  int
	FillColor             string
	StrokeColor           string
	VerticalAlign         string
	LabelPosition         string
	VerticalLabelPosition string
	Align                 string
	SpacingTop            int
	FontSize              int
	FontStyle             int
	Image                 string
}

// Convert a cell style into a string. This is used as the style attrib on an mxCell
func (cs *cellStyle) String() string {
	return fmt.Sprintf("html=%d;fillColor=%s;strokeColor=%s;verticalAlign=%s;labelPosition=%s;verticalLabelPosition=%s;align=%s;spacingTop=%d;fontSize=%d;fontStyle=%d;image;image=%s;", cs.HTML, cs.FillColor, cs.StrokeColor, cs.VerticalAlign, cs.LabelPosition, cs.VerticalLabelPosition, cs.Align, cs.SpacingTop, cs.FontSize, cs.FontStyle, cs.Image)
}

// Create a new linked cell style with default values
func getNewCellStyle(newImage string) *cellStyle {
	newCell := cellStyle{HTML: 1, FillColor: "#5184F3", StrokeColor: "none", VerticalAlign: "top", LabelPosition: "center", VerticalLabelPosition: "bottom", Align: "center", SpacingTop: -6, FontSize: 12, FontStyle: 0, Image: newImage}
	return &newCell
}

// mxGraphModel is the top-level node in a diagram
type mxGraphModel struct {
	XMLName    xml.Name `xml:"mxGraphModel"`
	Dx         int      `xml:"dx,attr"`
	Dy         int      `xml:"dy,attr"`
	Grid       int      `xml:"grid,attr"`
	GridSize   int      `xml:"gridSize,attr"`
	Guides     int      `xml:"guides,attr"`
	Tooltips   int      `xml:"tooltips,attr"`
	Connect    int      `xml:"connect,attr"`
	Arrows     int      `xml:"arrows,attr"`
	Fold       int      `xml:"fold,attr"`
	Page       int      `xml:"page,attr"`
	PageScale  int      `xml:"pageScale,attr"`
	PageWidth  int      `xml:"pageWidth,attr"`
	PageHeight int      `xml:"pageHeight,attr"`
	Math       int      `xml:"math,attr"`
	Shadow     int      `xml:"shadow,attr"`
	Root       *root    `xml:"root"`
}

// Despite the name, this is the first (and only) child of an mxGraphModel
type root struct {
	XMLName xml.Name  `xml:"root"`
	MxCells []*mxCell `xml:"mxCell"`
}

// A cell represents a single item / object within a diagram
type mxCell struct {
	XMLName xml.Name    `xml:"mxCell"`
	ID      string      `xml:"id,attr"`
	Parent  string      `xml:"parent,attr"`
	Value   string      `xml:"value,attr"`
	Style   string      `xml:"style,attr"`
	Vertex  int         `xml:"vertex,attr"`
	Edge    string      `xml:"edge,attr,omitempty"`
	Source  string      `xml:"source,attr"`
	Target  string      `xml:"target,attr"`
	Geom    *mxGeometry `xml:"mxGeometry"`
}

// mxGeometry represents the location of an object. It can also contain child points
type mxGeometry struct {
	XMLName xml.Name   `xml:"mxGeometry"`
	X       int        `xml:"x,attr"`
	Y       int        `xml:"y,attr"`
	Width   int        `xml:"width,attr"`
	Height  int        `xml:"height,attr"`
	As      string     `xml:"as,attr"`
	Points  []*mxPoint `xml:"mxPoint"`
}

// mxPoint is a specific point within a diagram
type mxPoint struct {
	XMLName xml.Name `xml:"mxPoint"`
	X       int      `xml:"x,attr"`
	Y       int      `xml:"y,attr"`
	As      string   `xml:"as,attr"`
}

// A mapping between resources and their diagram image equivalents
// In the future, this should be externalized in a config file
var mappings = map[string]string{
	"service.web.app":        "img/lib/ibm/applications/application_logic.svg",
	"service.db.postgres":    "img/lib/clip_art/computers/Database_128x128.png",
	"environment.properties": "img/lib/ibm/applications/runtime_services.svg",
	"default":                "img/lib/ibm/applications/runtime_services.svg",
}

func getIconForResourceType(resourceType string) string {
	iconName, ok := mappings[resourceType]
	if !ok {
		// Default to the generic 'runtime services' icon
		iconName = mappings["default"]
	}
	return iconName
}
