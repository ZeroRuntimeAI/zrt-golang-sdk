package zrt

import (
	"slices"
	"strings"

	"github.com/google/uuid"

	pb "github.com/ZeroRuntimeAI/zrt-golang-sdk/internal/pb"
)

// ChatContent is a text content block.
type ChatContent struct {
	// Type is the content block type; defaults to "text".
	Type string // default "text"
	// Text is the block's text content.
	Text string
}

// ImageContent is an image content block.
type ImageContent struct {
	// Type is the content block type; defaults to "image".
	Type string // default "image"
	// URL is the image URL or data URI.
	URL string
	// Detail is the image detail level; defaults to "auto".
	Detail string // default "auto"
}

// FunctionCall represents a tool/function call.
type FunctionCall struct {
	// Name is the called function's name.
	Name string
	// Arguments is the call arguments as a JSON string.
	Arguments string
	// CallID uniquely identifies this call, linking it to its output.
	CallID string
}

// FunctionCallOutput represents the output of a tool/function call.
type FunctionCallOutput struct {
	// Name is the called function's name.
	Name string
	// Output is the function's result.
	Output string
	// CallID is the FunctionCall.CallID this output answers.
	CallID string
	// IsError reports whether the call failed.
	IsError bool
}

// ChatMessage is a single message in a chat context.
type ChatMessage struct {
	// Role is the message author's role.
	Role ChatRole
	// Content is the message text.
	Content string
	// MessageID uniquely identifies the message.
	MessageID string
	// ToolCalls the assistant requested on this message.
	ToolCalls []FunctionCall
	// ToolCallID, for a tool-result message, is the FunctionCall.CallID it answers.
	ToolCallID string
	// Images attached to this message.
	Images []ImageContent
	// AgentID attributes this message to the agent that produced it. Used for
	// multi-agent audit/filtering; it is not sent to the LLM.
	AgentID string
}

// AgentHandoff records a transfer of control between agents. It is an audit-only
// item: kept in the ChatContext for inspection but never sent to the LLM as a
// message.
type AgentHandoff struct {
	// ID uniquely identifies the handoff marker.
	ID string
	// FromAgent is the id of the agent transferring control.
	FromAgent string
	// ToAgent is the id of the agent receiving control.
	ToAgent string
	// Reason describes why the handoff occurred.
	Reason string
}

// ChatContext holds an ordered list of chat messages plus structural items
// (agent handoffs) that are tracked for attribution but not sent to the LLM.
type ChatContext struct {
	messages []ChatMessage
	handoffs []AgentHandoff
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

// AddAttributedMessage appends a message attributed to agentID and returns it.
// When messageID is empty a fresh UUID is generated.
func (c *ChatContext) AddAttributedMessage(role ChatRole, content, messageID, agentID string, images ...ImageContent) ChatMessage {
	msg := c.AddMessage(role, content, messageID, images...)
	c.messages[len(c.messages)-1].AgentID = agentID
	msg.AgentID = agentID
	return msg
}

// AddHandoff records a transfer of control between agents and returns the marker.
// The handoff is structural (audit-only) and is excluded from LLM/runtime messages.
func (c *ChatContext) AddHandoff(fromAgent, toAgent, reason string) AgentHandoff {
	h := AgentHandoff{ID: "handoff_" + uuid.NewString(), FromAgent: fromAgent, ToAgent: toAgent, Reason: reason}
	c.handoffs = append(c.handoffs, h)
	return h
}

// Handoffs returns a copy of the recorded agent handoffs (most recent last).
func (c *ChatContext) Handoffs() []AgentHandoff { return slices.Clone(c.handoffs) }

// LastHandoff returns the most recent handoff, or nil if none was recorded.
func (c *ChatContext) LastHandoff() *AgentHandoff {
	if n := len(c.handoffs); n > 0 {
		h := c.handoffs[n-1]
		return &h
	}
	return nil
}

// Copy returns a shallow copy of the context (messages and handoff markers).
func (c *ChatContext) Copy() *ChatContext {
	return &ChatContext{messages: slices.Clone(c.messages), handoffs: slices.Clone(c.handoffs)}
}

// TruncateOptions bounds a ChatContext by item count and/or token estimate.
type TruncateOptions struct {
	// MaxItems caps the number of messages kept (most recent). 0 means no cap.
	MaxItems int
	// MaxTokens caps the estimated token count, dropping oldest messages. 0 means no cap.
	MaxTokens int
}

// Truncate returns a copy of the context bounded by opts, keeping the most
// recent messages within the item and token limits (always at least one).
func (c *ChatContext) Truncate(opts TruncateOptions) *ChatContext {
	out := c.Copy()
	if opts.MaxItems > 0 && len(out.messages) > opts.MaxItems {
		out.messages = out.messages[len(out.messages)-opts.MaxItems:]
	}
	if opts.MaxTokens > 0 {
		tokensOf := func(m ChatMessage) int { return len(strings.Fields(m.Content)) * 2 }
		total := 0
		for _, m := range out.messages {
			total += tokensOf(m)
		}
		// Drop oldest messages until under the token budget (keep at least one).
		for len(out.messages) > 1 && total > opts.MaxTokens {
			total -= tokensOf(out.messages[0])
			out.messages = out.messages[1:]
		}
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

// ToContextMessages returns the messages as runtime context-message protos,
// including images and tool calls.
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

// ChatContextFromContextMessages builds a ChatContext from context messages.
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
