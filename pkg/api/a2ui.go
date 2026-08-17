// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"fmt"
	"os"
	"strings"
)

// BoundValue represents a bound value in A2UI (either literal or data path)
type BoundValue struct {
	LiteralString  *string  `json:"literalString,omitempty"`
	LiteralBoolean *bool    `json:"literalBoolean,omitempty"`
	LiteralNumber  *float64 `json:"literalNumber,omitempty"`
	Path           *string  `json:"path,omitempty"`
}

// Helper initializers
func LiteralString(s string) *BoundValue { return &BoundValue{LiteralString: &s} }
func LiteralBoolean(b bool) *BoundValue  { return &BoundValue{LiteralBoolean: &b} }
func LiteralNumber(n float64) *BoundValue { return &BoundValue{LiteralNumber: &n} }
func Path(p string) *BoundValue          { return &BoundValue{Path: &p} }

func pointerToString(s string) *string { return &s }

// ComponentWrapper is the polymorphic wrapper for A2UI components
type ComponentWrapper struct {
	Card           *CardProps           `json:"Card,omitempty"`
	Column         *ColumnProps         `json:"Column,omitempty"`
	Row            *RowProps            `json:"Row,omitempty"`
	Text           *TextProps           `json:"Text,omitempty"`
	Button         *ButtonProps         `json:"Button,omitempty"`
	TextInput      *TextInputProps      `json:"TextInput,omitempty"`
	CheckBox       *CheckBoxProps       `json:"CheckBox,omitempty"`
	Image          *ImageProps          `json:"Image,omitempty"`
	MultipleChoice *MultipleChoiceProps `json:"MultipleChoice,omitempty"`
	WebFrameSrcdoc *WebFrameSrcdocProps `json:"WebFrameSrcdoc,omitempty"`
	Divider        *struct{}            `json:"Divider,omitempty"`
}

type A2UIComponent struct {
	ID        string           `json:"id"`
	Component ComponentWrapper `json:"component"`
}

type ChildrenProps struct {
	ExplicitList []string `json:"explicitList,omitempty"`
}

type CardProps struct {
	Title string  `json:"title,omitempty"`
	Child string  `json:"child,omitempty"`
	Style *string `json:"style,omitempty"`
}

type ColumnProps struct {
	Alignment          string        `json:"alignment,omitempty"`
	Distribution       string        `json:"distribution,omitempty"`
	CrossAxisAlignment string        `json:"crossAxisAlignment,omitempty"`
	MainAxisAlignment  string        `json:"mainAxisAlignment,omitempty"`
	Gap                int           `json:"gap,omitempty"`
	Children     ChildrenProps `json:"children"`
}

type RowProps struct {
	Alignment          string        `json:"alignment,omitempty"`
	Distribution       string        `json:"distribution,omitempty"`
	CrossAxisAlignment string        `json:"crossAxisAlignment,omitempty"`
	MainAxisAlignment  string        `json:"mainAxisAlignment,omitempty"`
	Gap                int           `json:"gap,omitempty"`
	Children     ChildrenProps `json:"children"`
}

type TextProps struct {
	Text      *BoundValue `json:"text"`
	UsageHint string      `json:"usageHint,omitempty"`
	Style     string      `json:"style,omitempty"`
}

type ButtonAction struct {
	Type    string          `json:"type"`
	Name    string          `json:"name"`
	Context []ButtonContext `json:"context,omitempty"`
}

type ButtonContext struct {
	Key   string      `json:"key"`
	Value *BoundValue `json:"value"`
}

type ButtonProps struct {
	Child   string        `json:"child"`
	Primary bool          `json:"primary,omitempty"`
	Label   string        `json:"label,omitempty"`
	Action  *ButtonAction `json:"action,omitempty"`
}

type TextInputProps struct {
	Label           string `json:"label"`
	Required        bool   `json:"required,omitempty"`
	DataBindingPath string `json:"dataBindingPath"`
}

type CheckBoxProps struct {
	Label *BoundValue `json:"label"`
	Value *BoundValue `json:"value"`
}

type ImageProps struct {
	URL *BoundValue `json:"url"`
}

type WebFrameSrcdocProps struct {
	HtmlContent *BoundValue `json:"htmlContent"`
	Height      int         `json:"height"`
}

type MultipleChoiceOption struct {
	Label *BoundValue `json:"label"`
	Value string      `json:"value"`
}

type MultipleChoiceProps struct {
	Options              []MultipleChoiceOption `json:"options"`
	Selections           *BoundValue            `json:"selections"`
	MaxAllowedSelections int                    `json:"maxAllowedSelections,omitempty"`
}

type A2UITransaction struct {
	SurfaceUpdate   *SurfaceUpdatePayload   `json:"surfaceUpdate"`
	DataModelUpdate *DataModelUpdatePayload `json:"dataModelUpdate,omitempty"`
	BeginRendering  *BeginRenderingPayload  `json:"beginRendering"`
}

type SurfaceUpdatePayload struct {
	SurfaceID  string          `json:"surfaceId"`
	Components []A2UIComponent `json:"components"`
}

type DataModelUpdatePayload struct {
	SurfaceID string                 `json:"surfaceId"`
	Data      map[string]interface{} `json:"data"`
}

type BeginRenderingPayload struct {
	RootComponentID string            `json:"root"`
	SurfaceID       string            `json:"surfaceId"`
	Styles          map[string]string `json:"styles"`
}

// GetAgentStyles returns the complete set of premium dark theme variables from index.css
func GetAgentStyles() map[string]string {
	return map[string]string{
		"--font-sans":            "'Plus Jakarta Sans', system-ui, -apple-system, sans-serif",
		"--font-display":         "'Outfit', var(--font-sans)",
		"--font-mono":            "'JetBrains Mono', monospace",
		"--bg-main":              "#041f41",
		"--bg-card":              "rgba(9, 46, 92, 0.45)",
		"--bg-card-hover":        "rgba(15, 61, 122, 0.6)",
		"--bg-input":             "rgba(3, 23, 48, 0.6)",
		"--border-glow":          "rgba(0, 113, 206, 0.15)",
		"--border-muted":         "rgba(255, 255, 255, 0.05)",
		"--color-primary":        "#0071ce",
		"--color-primary-glow":   "rgba(0, 113, 206, 0.35)",
		"--color-secondary":      "#eb148d",
		"--color-secondary-glow": "rgba(235, 20, 141, 0.25)",
		"--color-accent":         "#ffc220",
		"--color-accent-glow":    "rgba(255, 194, 32, 0.25)",
		"--color-warning":        "#f59e0b",
		"--color-critical":       "#ef4444",
		"--color-critical-glow":  "rgba(239, 68, 68, 0.3)",
		"--text-primary":         "#f3f4f6",
		"--text-secondary":       "#9ca3af",
		"--text-muted":           "#6b7280",
	}
}

// NewA2UITransaction initializes a complete rendering envelope with injected styling custom properties
func NewA2UITransaction(surfaceID string, rootID string, components []A2UIComponent, initialData map[string]interface{}) *A2UITransaction {
	var dataModel *DataModelUpdatePayload
	if len(initialData) > 0 {
		dataModel = &DataModelUpdatePayload{
			SurfaceID: surfaceID,
			Data:      initialData,
		}
	}

	return &A2UITransaction{
		SurfaceUpdate: &SurfaceUpdatePayload{
			SurfaceID:  surfaceID,
			Components: components,
		},
		DataModelUpdate: dataModel,
		BeginRendering: &BeginRenderingPayload{
			RootComponentID: rootID,
			SurfaceID:       surfaceID,
			Styles:          GetAgentStyles(),
		},
	}
}

// Safe Map Access Helpers
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		}
	}
	return 0
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func getSlice(m map[string]interface{}, key string) []interface{} {
	if v, ok := m[key]; ok {
		if s, ok := v.([]interface{}); ok {
			return s
		}
	}
	return nil
}

func getMap(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key]; ok {
		if r, ok := v.(map[string]interface{}); ok {
			return r
		}
	}
	return nil
}

// NormalizeCardToA2UITransaction compiles a legacy nested card map into a flat A2UI v0.8.0 Transaction
func NormalizeCardToA2UITransaction(legacyCard map[string]interface{}, surfaceID string) *A2UITransaction {
	components := []A2UIComponent{}
	compCounter := 0
	initialData := map[string]interface{}{}

	nextID := func(prefix string) string {
		compCounter++
		return fmt.Sprintf("%s_%d", prefix, compCounter)
	}

	var processNode func(node map[string]interface{}) string
	var processChildren func(childrenList []interface{}) []string

	processChildren = func(childrenList []interface{}) []string {
		processedIDs := []string{}
		switcherButtons := []map[string]interface{}{}

		for _, c := range childrenList {
			if m, ok := c.(map[string]interface{}); ok {
				if strings.ToLower(getString(m, "type")) == "button" && getString(m, "action") == "SET_STORE" {
					switcherButtons = append(switcherButtons, m)
				}
			}
		}

		if len(switcherButtons) >= 2 {
			choiceID := nextID("multiplechoice")
			optionsList := []MultipleChoiceOption{}
			for _, btn := range switcherButtons {
				label := getString(btn, "label")
				actionData := getMap(btn, "actionData")
				siteID := getString(actionData, "siteID")
				if siteID == "" {
					siteID = getString(actionData, "siteId")
				}
				optionsList = append(optionsList, MultipleChoiceOption{
					Label: LiteralString(label),
					Value: siteID,
				})
			}

			selectionsPath := "/store_switcher/selected"
			initialData[selectionsPath] = ""

			components = append(components, A2UIComponent{
				ID: choiceID,
				Component: ComponentWrapper{
					MultipleChoice: &MultipleChoiceProps{
						Options:              optionsList,
						Selections:           Path(selectionsPath),
						MaxAllowedSelections: 1,
					},
				},
			})

			btnID := nextID("button")
			btnLabelID := nextID("text")
			components = append(components, A2UIComponent{
				ID: btnLabelID,
				Component: ComponentWrapper{
					Text: &TextProps{
						Text: LiteralString("Switch Store"),
					},
				},
			})
			components = append(components, A2UIComponent{
				ID: btnID,
				Component: ComponentWrapper{
					Button: &ButtonProps{
						Child: btnLabelID,
						Action: &ButtonAction{
							Type: "submit",
							Name: "SET_STORE",
							Context: []ButtonContext{
								{
									Key:   "siteID",
									Value: Path(selectionsPath),
								},
							},
						},
					},
				},
			})

			switcherSet := map[interface{}]bool{}
			for _, b := range switcherButtons {
				// Address-based identification hack in Go
				switcherSet[fmt.Sprintf("%v", b)] = true
			}

			for _, child := range childrenList {
				if m, ok := child.(map[string]interface{}); ok {
					if switcherSet[fmt.Sprintf("%v", m)] {
						continue
					}
					childID := processNode(m)
					if childID != "" {
						processedIDs = append(processedIDs, childID)
					}
				}
			}

			processedIDs = append(processedIDs, choiceID, btnID)
		} else {
			for _, child := range childrenList {
				if m, ok := child.(map[string]interface{}); ok {
					childID := processNode(m)
					if childID != "" {
						processedIDs = append(processedIDs, childID)
					}
				}
			}
		}

		return processedIDs
	}

	processNode = func(node map[string]interface{}) string {
		nodeType := strings.ToLower(getString(node, "type"))

		switch nodeType {
		case "card":
			compID := nextID("card")
			properties := CardProps{}
			childrenIDs := []string{}

			title := getString(node, "title")
			if title != "" {
				properties.Title = title
				titleID := nextID("text")
				components = append(components, A2UIComponent{
					ID: titleID,
					Component: ComponentWrapper{
						Text: &TextProps{
							Text:      LiteralString(title),
							UsageHint: "h2",
						},
					},
				})
				childrenIDs = append(childrenIDs, titleID)
			}

			style := getString(node, "style")
			if style != "" {
				properties.Style = &style
			}

			childrenIDs = append(childrenIDs, processChildren(getSlice(node, "children"))...)

			if len(childrenIDs) == 1 {
				properties.Child = childrenIDs[0]
			} else if len(childrenIDs) > 1 {
				wrapperColID := nextID("column")
				components = append(components, A2UIComponent{
					ID: wrapperColID,
					Component: ComponentWrapper{
						Column: &ColumnProps{
							Alignment:    "stretch",
							Distribution: "start",
							Children: ChildrenProps{
								ExplicitList: childrenIDs,
							},
						},
					},
				})
				properties.Child = wrapperColID
			}

			components = append(components, A2UIComponent{
				ID: compID,
				Component: ComponentWrapper{
					Card: &properties,
				},
			})
			return compID

		case "column":
			compID := nextID("column")
			align := getString(node, "alignment")
			if align == "" {
				align = getString(node, "align")
			}
			if align == "" {
				align = "stretch"
			}
			dist := getString(node, "distribution")
			if dist == "" {
				dist = getString(node, "dist")
			}
			if dist == "" {
				dist = "start"
			}
			properties := ColumnProps{
				Alignment:          align,
				Distribution:       dist,
				CrossAxisAlignment: mapAlignment(align),
				MainAxisAlignment:  mapAlignment(dist),
				Gap:                getInt(node, "gap"),
			}
			properties.Children = ChildrenProps{
				ExplicitList: processChildren(getSlice(node, "children")),
			}
			components = append(components, A2UIComponent{
				ID: compID,
				Component: ComponentWrapper{
					Column: &properties,
				},
			})
			return compID

		case "row":
			compID := nextID("row")
			align := getString(node, "alignment")
			if align == "" {
				align = getString(node, "align")
			}
			if align == "" {
				align = "stretch"
			}
			dist := getString(node, "distribution")
			if dist == "" {
				dist = getString(node, "dist")
			}
			if dist == "" {
				dist = "start"
			}
			properties := RowProps{
				Alignment:          align,
				Distribution:       dist,
				CrossAxisAlignment: mapAlignment(align),
				MainAxisAlignment:  mapAlignment(dist),
				Gap:                getInt(node, "gap"),
			}
			properties.Children = ChildrenProps{
				ExplicitList: processChildren(getSlice(node, "children")),
			}
			components = append(components, A2UIComponent{
				ID: compID,
				Component: ComponentWrapper{
					Row: &properties,
				},
			})
			return compID

		case "button":
			compID := nextID("button")

			actionName := getString(node, "action")
			actionData := getMap(node, "actionData")
			contextList := []ButtonContext{}
			for k, v := range actionData {
				valObj := BoundValue{}
				switch val := v.(type) {
				case bool:
					valObj.LiteralBoolean = &val
				case float64:
					valObj.LiteralNumber = &val
				case int:
					valF := float64(val)
					valObj.LiteralNumber = &valF
				default:
					valS := fmt.Sprintf("%v", v)
					valObj.LiteralString = &valS
				}
				contextList = append(contextList, ButtonContext{
					Key:   k,
					Value: &valObj,
				})
			}

			actionProperties := &ButtonAction{
				Type:    "submit",
				Name:    actionName,
				Context: contextList,
			}

			labelText := getString(node, "label")
			labelID := nextID("text")
			components = append(components, A2UIComponent{
				ID: labelID,
				Component: ComponentWrapper{
					Text: &TextProps{
						Text: LiteralString(labelText),
					},
				},
			})

			components = append(components, A2UIComponent{
				ID: compID,
				Component: ComponentWrapper{
					Button: &ButtonProps{
						Action:  actionProperties,
						Child:   labelID,
						Label:   labelText,
						Primary: getString(node, "style") == "primary",
					},
				},
			})
			return compID

		case "text":
			compID := nextID("text")
			textVal := getString(node, "text")
			if textVal == "" {
				textVal = getString(node, "content")
			}
			properties := TextProps{
				Text: LiteralString(textVal),
			}
			style := getString(node, "style")
			usageHint := getString(node, "usageHint")
			if style != "" || usageHint != "" {
				hint := usageHint
				if hint == "" {
					hint = style
				}
				properties.UsageHint = normalizeUsageHint(hint)
				properties.Style = normalizeUsageHint(hint)
			}
			components = append(components, A2UIComponent{
				ID: compID,
				Component: ComponentWrapper{
					Text: &properties,
				},
			})
			return compID

		case "table":
			compID := nextID("column")
			rowIDs := []string{}
			rowsList := getSlice(node, "rows")

			for _, r := range rowsList {
				if row, ok := r.(map[string]interface{}); ok {
					rowID := nextID("row")
					lbl := getString(row, "label")
					val := getString(row, "value")

					lblID := nextID("text")
					components = append(components, A2UIComponent{
						ID: lblID,
						Component: ComponentWrapper{
							Text: &TextProps{
								Text: LiteralString(lbl),
							},
						},
					})

					valID := nextID("text")
					components = append(components, A2UIComponent{
						ID: valID,
						Component: ComponentWrapper{
							Text: &TextProps{
								Text:      LiteralString(val),
								UsageHint: "caption",
							},
						},
					})

					components = append(components, A2UIComponent{
						ID: rowID,
						Component: ComponentWrapper{
							Row: &RowProps{
								Distribution: "spaceBetween",
								Children: ChildrenProps{
									ExplicitList: []string{lblID, valID},
								},
							},
						},
					})
					rowIDs = append(rowIDs, rowID)
				}
			}

			components = append(components, A2UIComponent{
				ID: compID,
				Component: ComponentWrapper{
					Column: &ColumnProps{
						Gap: 8,
						Children: ChildrenProps{
							ExplicitList: rowIDs,
						},
					},
				},
			})
			return compID

		case "divider":
			compID := nextID("divider")
			components = append(components, A2UIComponent{
				ID: compID,
				Component: ComponentWrapper{
					Divider: &struct{}{},
				},
			})
			return compID

		case "checkbox":
			compID := nextID("checkbox")
			labelText := getString(node, "label")
			val := node["value"]

			labelBound := LiteralString(labelText)
			valueBound := BoundValue{}

			if b, ok := val.(bool); ok {
				valueBound.LiteralBoolean = &b
			} else if s, ok := val.(string); ok {
				if strings.HasPrefix(s, "/") {
					valueBound.Path = &s
				} else {
					valueBound.LiteralString = &s
				}
			} else {
				fallbackB := getBool(node, "value")
				valueBound.LiteralBoolean = &fallbackB
			}

			components = append(components, A2UIComponent{
				ID: compID,
				Component: ComponentWrapper{
					CheckBox: &CheckBoxProps{
						Label: labelBound,
						Value: &valueBound,
					},
				},
			})
			return compID

		case "canvas":
			compID := nextID("image")
			layout := getString(node, "layout")
			if layout == "" {
				layout = "linear"
			}
			beacon := getMap(node, "beacon")

			agentHost := os.Getenv("AGENT_HOST_URL")
			if agentHost == "" {
				agentHost = "https://gemini-task-agent-dev-10781708810.us-central1.run.app"
			}

			imgURL := fmt.Sprintf("%s/api/v1/blueprint?layout=%s", agentHost, layout)
			if beacon != nil {
				x := beacon["x"]
				y := beacon["y"]
				if x != nil && y != nil {
					imgURL += fmt.Sprintf("&x=%v&y=%v", x, y)
				}
			}

			components = append(components, A2UIComponent{
				ID: compID,
				Component: ComponentWrapper{
					Image: &ImageProps{
						URL: LiteralString(imgURL),
					},
				},
			})
			return compID

		case "webframe", "webframesrcdoc":
			compID := nextID("webframe")
			htmlVal := getString(node, "htmlContent")
			if htmlVal == "" {
				htmlVal = getString(node, "html_content")
			}
			heightVal := getInt(node, "height")
			if heightVal == 0 {
				heightVal = 300
			}

			components = append(components, A2UIComponent{
				ID: compID,
				Component: ComponentWrapper{
					WebFrameSrcdoc: &WebFrameSrcdocProps{
						HtmlContent: LiteralString(htmlVal),
						Height:      heightVal,
					},
				},
			})
			return compID

		case "select":
			compID := nextID("select")
			optionsList := []MultipleChoiceOption{}
			optsSlice := getSlice(node, "options")
			for _, o := range optsSlice {
				if opt, ok := o.(map[string]interface{}); ok {
					optionsList = append(optionsList, MultipleChoiceOption{
						Label: LiteralString(getString(opt, "label")),
						Value: getString(opt, "value"),
					})
				}
			}

			components = append(components, A2UIComponent{
				ID: compID,
				Component: ComponentWrapper{
					MultipleChoice: &MultipleChoiceProps{
						Options:              optionsList,
						Selections:           Path(getString(node, "name")),
						MaxAllowedSelections: 1,
					},
				},
			})
			return compID

		case "input":
			compID := nextID("textinput")
			components = append(components, A2UIComponent{
				ID: compID,
				Component: ComponentWrapper{
					TextInput: &TextInputProps{
						Label:           getString(node, "label"),
						Required:        getBool(node, "required"),
						DataBindingPath: getString(node, "name"),
					},
				},
			})
			return compID
		}

		return ""
	}

	rootID := processNode(legacyCard)
	if rootID == "" {
		return nil
	}

	return NewA2UITransaction(surfaceID, rootID, components, initialData)
}

func normalizeUsageHint(hint string) string {
	hint = strings.ToLower(hint)
	switch hint {
	case "h1", "h2", "h3", "body", "caption":
		return hint
	case "primary":
		return "body"
	case "secondary":
		return "caption"
	default:
		return "body"
	}
}

func mapAlignment(val string) string {
	val = strings.ToLower(val)
	switch val {
	case "start":
		return "Start"
	case "center", "middle":
		return "Center"
	case "end":
		return "End"
	case "stretch":
		return "Stretch"
	case "spacebetween", "space-between":
		return "SpaceBetween"
	case "spacearound", "space-around":
		return "SpaceAround"
	case "spaceevenly", "space-evenly":
		return "SpaceEvenly"
	default:
		return "Start"
	}
}
