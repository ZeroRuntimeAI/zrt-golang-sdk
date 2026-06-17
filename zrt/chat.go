package zrt

import (
	"slices"
	"strings"

	"github.com/google/uuid"

	pb "github.com/ZeroRuntimeAI/zrt-golang-sdk/internal/pb"
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
	// ToolCalls the assistant requested on this message.
	ToolCalls []FunctionCall
	// ToolCallID, for a tool-result message, is the FunctionCall.CallID it answers.
	ToolCallID string
	// Images attached to this message (sent to the runtime as ImageContent values).
	Images []ImageContent
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

// AddMessage appends a message and returns it. Optional images are attached to
// the message. When messageID is empty a fresh UUID is generated.
func (c *ChatContext) AddMessage(role ChatRole, content string, messageID string, images ...ImageContent) ChatMessage {
	if messageID == "" {
		messageID = uuid.NewString()
	}
	msg := ChatMessage{Role: role, Content: content, MessageID: messageID}
	if len(images) > 0 {
		msg.Images = slices.Clone(images)
	}
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

func (c *ChatContext) ToContextMessages() []*pb.ContextMessageProto {
	out := make([]*pb.ContextMessageProto, 0, len(c.messages))
	for _, m := range c.messages {
		pm := &pb.ContextMessageProto{
			Role:       string(m.Role),
			Content:    m.Content,
			MessageId:  m.MessageID,
			ToolCallId: m.ToolCallID,
		}
		for _, img := range m.Images {
			pm.Images = append(pm.Images, &pb.ImageContentProto{Url: img.URL, Detail: img.Detail})
		}
		for _, tc := range m.ToolCalls {
			pm.ToolCalls = append(pm.ToolCalls, &pb.ToolCallProto{
				CallId:        tc.CallID,
				Name:          tc.Name,
				ArgumentsJson: tc.Arguments,
			})
		}
		out = append(out, pm)
	}
	return out
}

// ChatContextFromContextMessages builds a ChatContext from runtime wire messages
func ChatContextFromContextMessages(messages []*pb.ContextMessageProto) *ChatContext {
	c := &ChatContext{}
	for _, m := range messages {
		msg := ChatMessage{
			Role:       ChatRole(m.GetRole()),
			Content:    m.GetContent(),
			MessageID:  m.GetMessageId(),
			ToolCallID: m.GetToolCallId(),
		}
		for _, img := range m.GetImages() {
			msg.Images = append(msg.Images, ImageContent{URL: img.GetUrl(), Detail: img.GetDetail()})
		}
		for _, tc := range m.GetToolCalls() {
			msg.ToolCalls = append(msg.ToolCalls, FunctionCall{
				CallID:    tc.GetCallId(),
				Name:      tc.GetName(),
				Arguments: tc.GetArgumentsJson(),
			})
		}
		c.messages = append(c.messages, msg)
	}
	return c
}
