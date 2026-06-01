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

package agents

// AgentDefinition defines the profile, behavior, and tool-access scopes of a conversational model session.
type AgentDefinition struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	SystemInstruction string   `json:"system_instruction"`
	AllowedTools      []string `json:"allowed_tools"`
}

var registry = map[string]*AgentDefinition{
	"shift_agent": {
		ID:          "shift_agent",
		Name:        "Shift Assistant (Hanna)",
		Description: "AI shift assistant designed to assist associates with shift initialization, task execution, asset overrides, and peer-to-peer trades.",
		SystemInstruction: `You are Hanna, a highly capable, professional, and direct retail shift assistant.
You assist the active user in executing their assigned task queues, query compliance SOPs, and log asset overrides or trade proposals.
Maintain a professional, direct, and precise baseline. Keep your answers strictly factual and actionable.`,
		AllowedTools: []string{
			"get_tasks",
			"override_asset",
			"propose_trade",
			"accept_trade",
			"reject_trade",
			"query_sop",
		},
	},
}

// Get returns the AgentDefinition by its ID.
func Get(id string) (*AgentDefinition, bool) {
	agent, ok := registry[id]
	return agent, ok
}

// List returns all registered agent definitions.
func List() []*AgentDefinition {
	list := make([]*AgentDefinition, 0, len(registry))
	for _, agent := range registry {
		list = append(list, agent)
	}
	return list
}
