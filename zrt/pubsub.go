package zrt

import (
	"context"
	"encoding/json"
)

// PubSubMessage is a message received on a subscribed room pub/sub topic.
type PubSubMessage struct {
	Topic      string
	Message    string
	SenderID   string
	SenderName string
	Timestamp  string
	Payload    any
}

// PublishMessage publishes a message to a room pub/sub topic. The runtime forwards
// it to the room; persistence is requested via cfg.Options (e.g. {"persist": true}).
func (s *AgentSession) PublishMessage(ctx context.Context, cfg PubSubPublishConfig) error {
	if cfg.Topic == "" {
		logger.Warnf("PublishMessage: empty topic — ignored")
		return nil
	}
	optionsJSON := ""
	if len(cfg.Options) > 0 {
		if b, err := json.Marshal(cfg.Options); err == nil {
			optionsJSON = string(b)
		}
	}
	payloadJSON := ""
	if cfg.Payload != nil {
		if b, err := json.Marshal(cfg.Payload); err == nil {
			payloadJSON = string(b)
		}
	}
	if t := s.transportRef(); t != nil {
		return t.sendPublishMessage(cfg.Topic, cfg.Message, optionsJSON, payloadJSON)
	}
	return nil
}

// SubscribePubSub subscribes to a room pub/sub topic. The runtime joins the topic and
// forwards each received message to the SDK; messages are delivered via the
// "pubsub_message" session event and, if cb is non-nil, to cb (filtered to topic).
func (s *AgentSession) SubscribePubSub(ctx context.Context, topic string, cb func(PubSubMessage)) error {
	if topic == "" {
		logger.Warnf("SubscribePubSub: empty topic — ignored")
		return nil
	}
	if cb != nil {
		s.On("pubsub_message", func(payload any) {
			m, ok := payload.(map[string]any)
			if !ok {
				return
			}
			t, _ := m["topic"].(string)
			if t != topic {
				return
			}
			msg, _ := m["message"].(string)
			sid, _ := m["sender_id"].(string)
			sname, _ := m["sender_name"].(string)
			ts, _ := m["timestamp"].(string)
			cb(PubSubMessage{
				Topic:      t,
				Message:    msg,
				SenderID:   sid,
				SenderName: sname,
				Timestamp:  ts,
				Payload:    m["payload"],
			})
		})
	}
	if t := s.transportRef(); t != nil {
		return t.sendSubscribePubSub(topic)
	}
	return nil
}
