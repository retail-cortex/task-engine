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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestA2UINormalizeAllComponents(t *testing.T) {
	cardMap := map[string]interface{}{
		"type":  "card",
		"title": "Test Card",
		"style": "primary",
		"children": []interface{}{
			map[string]interface{}{
				"type":       "column",
				"alignment":  "center",
				"dist":       "spaceBetween",
				"gap":        10,
				"children": []interface{}{
					map[string]interface{}{
						"type":      "text",
						"text":      "Hello text",
						"style":     "h1",
						"usageHint": "h1",
					},
					map[string]interface{}{
						"type":  "row",
						"align": "end",
						"children": []interface{}{
							map[string]interface{}{
								"type":  "button",
								"label": "Click me",
								"style": "primary",
								"action": "CLICK",
								"actionData": map[string]interface{}{
									"boolVal": true,
									"numVal":  123.45,
									"intVal":  42,
									"strVal":  "hello",
								},
							},
						},
					},
					map[string]interface{}{
						"type": "table",
						"rows": []interface{}{
							map[string]interface{}{"label": "Key 1", "value": "Val 1"},
						},
					},
					map[string]interface{}{
						"type": "divider",
					},
					map[string]interface{}{
						"type":  "checkbox",
						"label": "Check boolean",
						"value": true,
					},
					map[string]interface{}{
						"type":  "checkbox",
						"label": "Check path",
						"value": "/state/checked",
					},
					map[string]interface{}{
						"type":  "checkbox",
						"label": "Check literal string",
						"value": "yes",
					},
					map[string]interface{}{
						"type":   "canvas",
						"layout": "racetrack",
						"beacon": map[string]interface{}{"x": 100, "y": 200},
					},
					map[string]interface{}{
						"type":        "webframe",
						"htmlContent": "<h1>iframe</h1>",
						"height":      400,
					},
					map[string]interface{}{
						"type": "select",
						"options": []interface{}{
							map[string]interface{}{"label": "Opt A", "value": "a"},
							map[string]interface{}{"label": "Opt B", "value": "b"},
						},
					},
					map[string]interface{}{
						"type": "switcher",
						"buttons": []interface{}{
							map[string]interface{}{"label": "Store 1", "action": "SWITCH", "actionData": map[string]interface{}{"id": "1"}},
						},
					},
				},
			},
		},
	}

	tx := NormalizeCardToA2UITransaction(cardMap, "test_surface")
	assert.NotNil(t, tx)
	assert.Equal(t, "test_surface", tx.SurfaceUpdate.SurfaceID)
	assert.NotEmpty(t, tx.SurfaceUpdate.Components)

	// Helper tests
	s := LiteralString("str")
	assert.NotNil(t, s.LiteralString)
	b := LiteralBoolean(false)
	assert.NotNil(t, b.LiteralBoolean)
	n := LiteralNumber(9.9)
	assert.NotNil(t, n.LiteralNumber)
	p := Path("/foo")
	assert.NotNil(t, p.Path)
	pStr := pointerToString("pt")
	assert.Equal(t, "pt", *pStr)

	m := map[string]interface{}{
		"b": true,
		"s": "not-bool",
		"i": 42,
		"f": 3.14,
		"l": []interface{}{1, 2},
	}
	assert.True(t, getBool(m, "b"))
	assert.False(t, getBool(m, "s"))
	assert.False(t, getBool(m, "missing"))
	assert.Equal(t, 42, getInt(m, "i"))
	assert.Equal(t, 3, getInt(m, "f"))
	assert.Equal(t, 0, getInt(m, "s"))
	assert.Len(t, getSlice(m, "l"), 2)
	assert.Nil(t, getSlice(m, "s"))
}
