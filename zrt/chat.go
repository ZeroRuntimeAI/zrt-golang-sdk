package zrt

import (
	"slices"
	"strings"

	"github.com/google/uuid"
)

// ChatContent is a text content block.
type ChatContent struct {
	Type string // default "text"
	Text string
}

// ImageContent is an image content block.
type ImageContent struct {
	Type   string // default "image"
	URL    string
	Detail string // default "auto"
}

// FunctionCall represents a tool/function call.
type FunctionCall struct {
	Name      string
	Arguments string
	CallID    string
}

// FunctionCallOutput represents the output of a tool/function call.
type FunctionCallOutput struct {
	Name    string
	Output  string
	CallID  string
	IsError bool
}

// ChatMessage is a single message in a chat context.
type ChatMessage struct {
	Role      ChatRole
	Content   string
	MessageID string
	ToolCalls []FunctionCall
}

// ChatContext holds an ordered list of chat messages.
type ChatContext struct {
	messages []ChatMessage
}

// NewChatContext returns an empty chat context.
func NewChatContext() *ChatContext { return &ChatContext{} }

// EmptyChatContext returns an empty chat context.
func EmptyChatContext() *ChatContext { return &ChatContext{} }

// Items returns the underlying message slice.
func (c *ChatContext) Items() []ChatMessage { return c.messages }

// Messages returns a copy of the messages.
func (c *ChatContext) Messages() []ChatMessage {
	return slices.Clone(c.messages)
}

// TurnCount returns the number of user messages.
func (c *ChatContext) TurnCount() int {
	n := 0
	for _, m := range c.messages {
		if m.Role == ChatRoleUser {
			n++
		}
	}
	return n
}

// EstimatedTokens returns a rough token estimate (2 per word).
func (c *ChatContext) EstimatedTokens() int {
	total := 0
	for _, m := range c.messages {
		total += len(strings.Fields(m.Content)) * 2
	}
	return total
}

// AddMessage appends a message and returns it.
func (c *ChatContext) AddMessage(role ChatRole, content string, messageID string) ChatMessage {
	if messageID == "" {
		messageID = uuid.NewString()
	}
	msg := ChatMessage{Role: role, Content: content, MessageID: messageID}
	c.messages = append(c.messages, msg)
	return msg
}

// Copy returns a shallow copy of the context.
func (c *ChatContext) Copy() *ChatContext {
	return &ChatContext{messages: slices.Clone(c.messages)}
}

// Truncate returns a context keeping at most maxItems most-recent messages
// (maxItems <= 0 keeps all).
func (c *ChatContext) Truncate(maxItems int) *ChatContext {
	out := c.Copy()
	if maxItems > 0 && len(out.messages) > maxItems {
		out.messages = out.messages[len(out.messages)-maxItems:]
	}
	return out
}

// ToOpenAIMessages returns the messages as role/content maps.
func (c *ChatContext) ToOpenAIMessages() []map[string]string {
	out := make([]map[string]string, 0, len(c.messages))
	for _, m := range c.messages {
		out = append(out, map[string]string{"role": string(m.Role), "content": m.Content})
	}
	return out
}
