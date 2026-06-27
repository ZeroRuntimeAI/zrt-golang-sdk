package zrt

import (
	"context"
	"slices"
)

// ActiveAgent returns the session's current active agent (the most recent handoff
// target, or the agent the session started with).
func (s *AgentSession) ActiveAgent() Agent {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.agent
}

// seedHandoffRegistry builds the agent-id -> agent map and the set of alternate ids
// (which the runtime applies persona for) from the root agent's handoff agents and alternates.
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
// agent-switch marker, registering the agent so the switch can run its lifecycle.
// Any other result passes through unchanged.
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
// agent, or "" if there has been none.
func (s *AgentSession) LastHandoffReason() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if n := len(s.handoffs); n > 0 {
		return s.handoffs[n-1].Reason
	}
	return ""
}

// applyAgentSwitch performs the SDK side of a handoff after an AgentSwitched event:
// swaps the active agent, applies persona for non-alternate agents, settles in-flight
// turn state, records the handoff, and runs the new agent's OnEnter. The outgoing
// agent's OnExit is NOT called — OnExit is reserved for session end.
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

	// Settle in-flight turn state so it does not leak into the incoming agent: clear
	// the utterance handle (a stuck one blocks waiters) and stop thinking/background
	// audio. Context caches are preserved so inherited history stays consistent.
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

	// A2A topics are keyed on the active agent's id; re-point the subscription to the
	// new agent's id.
	s.mu.Lock()
	resubA2A := s.a2aSubscribed
	s.mu.Unlock()
	if resubA2A {
		if err := s.SubscribeA2A(context.Background()); err != nil {
			logger.Errorf("[handoff] failed to re-subscribe A2A for %q: %v", to, err)
		}
	}

	// Non-alternate agents were not pre-loaded by the runtime, so apply their
	// instructions/tools here. Alternates were already applied runtime-side; skip them.
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

	// Run OnEnter off the event loop. Bind the session into the context so a shared
	// agent resolves this session.
	go func() {
		if err := newAgent.OnEnter(s.bindBus(context.Background())); err != nil {
			logger.Errorf("[handoff] on_enter error: %v", err)
		}
	}()
}
