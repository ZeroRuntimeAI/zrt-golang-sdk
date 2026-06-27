package zrt

import (
	"context"
	"encoding/json"
)

// a2aTopicPrefix is the pub/sub topic prefix for A2A messages.
const a2aTopicPrefix = "a2a/"

// SendA2A sends an agent-to-agent (A2A) message to targetAgentID.
//
// The receiving agent must have called SubscribeA2A; the message is delivered
// there as an "a2a_message" event.
func (s *AgentSession) SendA2A(ctx context.Context, targetAgentID, messageJSON, correlationID string) error {
	if targetAgentID == "" {
		logger.Warnf("AgentSession.SendA2A: targetAgentID is required — ignored")
		return nil
	}
	sourceID := ""
	if s.agent != nil {
		sourceID = s.agent.base().id
	}
	envelope, err := json.Marshal(map[string]any{
		"source_agent_id": sourceID,
		"message_json":    messageJSON,
		"correlation_id":  correlationID,
	})
	if err != nil {
		return err
	}
	return s.PublishMessage(ctx, PubSubPublishConfig{
		Topic:   a2aTopicPrefix + targetAgentID,
		Message: string(envelope),
	})
}

// SubscribeA2A starts receiving A2A messages addressed to this agent. Each
// inbound message is emitted as an "a2a_message" event carrying
// {source_agent_id, message_json, correlation_id}.
//
// Safe to call more than once: a prior subscription is replaced, so
// re-subscribing (for example after a handoff) does not stack duplicate
// listeners.
func (s *AgentSession) SubscribeA2A(ctx context.Context) error {
	own := ""
	if a := s.ActiveAgent(); a != nil {
		own = a.base().id
	}
	if own == "" {
		logger.Warnf("AgentSession.SubscribeA2A: agent has no id — ignored")
		return nil
	}
	newTopic := a2aTopicPrefix + own

	s.mu.Lock()
	// Already subscribed to this topic: nothing to do.
	if s.a2aSubscribed && s.a2aTopic == newTopic && s.a2aUnsub != nil {
		s.mu.Unlock()
		return nil
	}
	oldUnsub := s.a2aUnsub
	s.a2aUnsub = nil
	s.mu.Unlock()

	// Drop any prior A2A handler before re-subscribing to a new topic.
	if oldUnsub != nil {
		oldUnsub()
	}

	unsub, err := s.subscribePubSubH(ctx, newTopic, func(m PubSubMessage) {
		if m.Message == "" {
			return
		}
		var env map[string]any
		if err := json.Unmarshal([]byte(m.Message), &env); err != nil {
			logger.Errorf("AgentSession.SubscribeA2A: failed to decode message: %v", err)
			return
		}
		if env == nil {
			// Valid JSON but not an object (e.g. "null") — not a real envelope.
			return
		}
		str := func(k string) string { v, _ := env[k].(string); return v }
		s.Emit("a2a_message", map[string]any{
			"source_agent_id": str("source_agent_id"),
			"message_json":    str("message_json"),
			"correlation_id":  str("correlation_id"),
		})
	})

	s.mu.Lock()
	s.a2aSubscribed = true
	s.a2aTopic = newTopic
	s.a2aUnsub = unsub
	s.mu.Unlock()
	return err
}

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
