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
	// AgentID attributes this message to the agent that produced it. Used for
	// multi-agent audit/filtering; it is not sent to the LLM.
	AgentID string
}

// AgentHandoff records a transfer of control between agents. It is a structural,
// audit-only item: it is kept in the ChatContext for inspection but is never sent
// to the LLM/runtime as a message (mirrors videosdk-agents).
type AgentHandoff struct {
	ID        string
	FromAgent string
	ToAgent   string
	Reason    string
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

type TruncateOptions struct {
	MaxItems  int
	MaxTokens int
}


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
