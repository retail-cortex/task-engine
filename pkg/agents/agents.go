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
