package zrt

// AgentCard describes an agent for agent-to-agent (A2A) discovery.
type AgentCard struct {
	AgentID      string
	Name         string
	Description  string
	Capabilities []string
	Metadata     map[string]any
}

// A2AMessage is an agent-to-agent message.
type A2AMessage struct {
	SourceAgentID string
	TargetAgentID string
	MessageType   string // default "text"
	Payload       map[string]any
}
