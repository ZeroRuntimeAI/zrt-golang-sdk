package zrt

import "maps"

// AgentSwitchKey is the key that signals an in-call agent handoff when present
// in a tool's return value.
const AgentSwitchKey = "__agent_switch__"

// AgentSwitchOptions configures an agent_switch payload.
type AgentSwitchOptions struct {
	// From is the name of the agent handing off, included in the payload when non-empty.
	From string
	// Reason describes why the handoff occurs, included in the payload when non-empty.
	Reason string
	// InheritContext defaults to true when nil.
	InheritContext *bool
	// Extra holds additional keys merged into the returned tool result.
	Extra map[string]any
}

// AgentSwitch builds the value a tool returns to hand off the call to another
// agent named to.
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
	maps.Copy(result, opts.Extra)
	return result
}
