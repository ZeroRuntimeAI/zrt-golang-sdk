package zrt

import (
	"context"
	"slices"
)

// ActiveAgent returns the session's current active agent: the most recent handoff
// target, or the agent the session started with. Tool dispatch and lifecycle use
// this so a handoff retargets them to the new agent.
func (s *AgentSession) ActiveAgent() Agent {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.agent
}

// seedHandoffRegistry builds the agent-id -> agent map and the set of ids the
// runtime already knows as alternates (and therefore applies persona for itself),
// from the root agent's configured handoff agents and named alternates.
func (s *AgentSession) seedHandoffRegistry(root Agent) {
	s.agentsByID = map[string]Agent{}
	s.alternateIDs = map[string]bool{}
	if root == nil {
		return
	}
	rb := root.base()
	if rb.id != "" {
		s.agentsByID[rb.id] = root
	}
	for _, ag := range rb.handoffAgents {
		if ag == nil {
			continue
		}
		id := ag.base().id
		if id == "" {
			continue
		}
		ag.base().session = s
		s.agentsByID[id] = ag
		s.alternateIDs[id] = true
	}
	for _, alt := range rb.alternates {
		if alt == nil || alt.AgentID == "" {
			continue
		}
		s.alternateIDs[alt.AgentID] = true
	}
}

// registerHandoffAgent adds an agent to the session registry. Used when a tool
// returns a fresh Agent object that was not pre-registered at session start.
// Returns the agent id.
func (s *AgentSession) registerHandoffAgent(ag Agent) string {
	if ag == nil {
		return ""
	}
	id := ag.base().id
	s.mu.Lock()
	if s.agentsByID == nil {
		s.agentsByID = map[string]Agent{}
	}
	s.agentsByID[id] = ag
	s.mu.Unlock()
	return id
}

// interceptToolResult converts a tool result that is itself an Agent into an
// agent-switch marker (videosdk-style "return an Agent" handoff), registering the
// agent so the switch can run its lifecycle. Any other result passes through
// unchanged. It is invoked on the tool-execution goroutine before serialization.
func (s *AgentSession) interceptToolResult(result any) any {
	ag, ok := result.(Agent)
	if !ok {
		return result
	}
	id := s.registerHandoffAgent(ag)
	inherit := ag.base().inheritContext
	from := ""
	if cur := s.ActiveAgent(); cur != nil {
		from = cur.base().id
	}
	return AgentSwitch(id, AgentSwitchOptions{From: from, Reason: "agent_returned_agent", InheritContext: &inherit})
}

// Handoffs returns a copy of the recorded agent handoffs (most recent last).
func (s *AgentSession) Handoffs() []AgentHandoff {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.handoffs)
}

// LastHandoffReason returns the reason of the most recent handoff into the active
// agent, or "" if there has been none. Useful inside a new agent's OnEnter to
// personalize its greeting.
func (s *AgentSession) LastHandoffReason() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if n := len(s.handoffs); n > 0 {
		return s.handoffs[n-1].Reason
	}
	return ""
}

// applyAgentSwitch performs the SDK side of a handoff once the runtime emits an
// AgentSwitched event: it swaps the active agent, applies persona for SDK-managed
// (non-alternate) agents, settles in-flight turn state so nothing leaks from the
// outgoing agent, records an attribution marker, and runs the new agent's OnEnter.
// The outgoing agent's OnExit is intentionally NOT called — OnExit is reserved for
// session end, mirroring videosdk-agents.
func (s *AgentSession) applyAgentSwitch(from, to, reason string) {
	s.mu.Lock()
	newAgent := s.agentsByID[to]
	old := s.agent
	isAlternate := s.alternateIDs[to]
	s.mu.Unlock()

	if newAgent == nil {
		logger.Warnf("[handoff] target agent_id %q is not registered in the SDK; the active agent will not be swapped, so the new agent's tools and OnEnter will not run. Pre-register it via AgentOptions.Agents.", to)
		return
	}
	if old != nil && old.base().id == to {
		return // already active
	}

	// Settle in-flight turn state from the outgoing agent so it does not leak into
	// the incoming one: a stuck utterance handle would block waiters, and stale
	// thinking/background audio would keep playing. Context caches
	// (transcriptMirror / chatHistoryCache) are deliberately preserved so the
	// runtime's inherited history stays consistent.
	if u := s.CurrentUtterance(); u != nil {
		u.markDone()
	}
	s.mu.Lock()
	s.currentUtterance = nil
	s.agent = newAgent
	s.handoffs = append(s.handoffs, AgentHandoff{ID: "handoff_" + to, FromAgent: from, ToAgent: to, Reason: reason})
	s.mu.Unlock()
	s.autoStopThinkingAudio()

	newAgent.base().session = s
	s.pipeline.setAgent(newAgent)

	// For agents the runtime did not pre-load as alternates it warned and left the
	// old persona in place, so apply the new instructions/tools ourselves. For
	// pre-registered alternates the runtime already applied them — skip to avoid a
	// redundant round-trip.
	if !isAlternate {
		nb := newAgent.base()
		if t := s.transportRef(); t != nil {
			if err := t.sendUpdateInstructions(nb.instructions); err != nil {
				logger.Errorf("[handoff] update instructions failed: %v", err)
			}
			if err := t.sendUpdateTools(nb.tools); err != nil {
				logger.Errorf("[handoff] update tools failed: %v", err)
			}
		}
	}

	// Run the incoming agent's OnEnter off the event loop (mirrors session start).
	go func() {
		if err := newAgent.OnEnter(context.Background()); err != nil {
			logger.Errorf("[handoff] on_enter error: %v", err)
		}
	}()
}
