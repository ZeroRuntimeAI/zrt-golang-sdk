package zrt

// AgentSwitchKey is the magic key signalling an in-call agent handoff when
// returned from a tool.
const AgentSwitchKey = "__agent_switch__"

// AgentSwitchOptions configures an agent_switch payload.
type AgentSwitchOptions struct {
	From   string
	Reason string
	// InheritContext defaults to true when nil.
	InheritContext *bool
	Extra          map[string]any
}

// AgentSwitch builds the dict a tool returns to trigger a handoff to another
// agent.
func AgentSwitch(to string, opts AgentSwitchOptions) map[string]any {
	payload := map[string]any{"to": to}
	if opts.From != "" {
		payload["from"] = opts.From
	}
	if opts.Reason != "" {
		payload["reason"] = opts.Reason
	}
	inherit := true
	if opts.InheritContext != nil {
		inherit = *opts.InheritContext
	}
	payload["inherit_context"] = inherit
	result := map[string]any{AgentSwitchKey: payload}
	for k, v := range opts.Extra {
		result[k] = v
	}
	return result
}
