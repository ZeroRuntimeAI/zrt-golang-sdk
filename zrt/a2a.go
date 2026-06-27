package zrt

import (
	"context"
	"encoding/json"
)

// a2aTopicPrefix is the pub/sub topic prefix for A2A messages: a message to agent
// X is published to topic "a2a/X".
const a2aTopicPrefix = "a2a/"

// SendA2A sends an agent-to-agent (A2A) message to targetAgentID.
//
// A2A is transported over pub/sub (topic "a2a/<targetAgentID>"), NOT a runtime
// RPC: the message is published as a JSON envelope
// {source_agent_id, message_json, correlation_id}. The receiving agent must have
// called SubscribeA2A to receive it (surfaced as an "a2a_message" event).
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

// SubscribeA2A starts receiving A2A messages addressed to this agent (topic
// "a2a/<agent id>"). Each inbound envelope is re-emitted as an "a2a_message" event
// carrying {source_agent_id, message_json, correlation_id}.
func (s *AgentSession) SubscribeA2A(ctx context.Context) error {
	own := ""
	if s.agent != nil {
		own = s.agent.base().id
	}
	if own == "" {
		logger.Warnf("AgentSession.SubscribeA2A: agent has no id — ignored")
		return nil
	}
	return s.SubscribePubSub(ctx, a2aTopicPrefix+own, func(m PubSubMessage) {
		if m.Message == "" {
			return
		}
		var env map[string]any
		if err := json.Unmarshal([]byte(m.Message), &env); err != nil {
			logger.Errorf("AgentSession.SubscribeA2A: failed to decode message: %v", err)
			return
		}
		str := func(k string) string { v, _ := env[k].(string); return v }
		s.Emit("a2a_message", map[string]any{
			"source_agent_id": str("source_agent_id"),
			"message_json":    str("message_json"),
			"correlation_id":  str("correlation_id"),
		})
	})
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
