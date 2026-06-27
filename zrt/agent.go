package zrt

import (
	"cmp"
	"context"
	"slices"
)

// Agent is the behavior contract: instructions, tools, and lifecycle hooks.
//
// Embed BaseAgent in your own type and implement OnEnter/OnExit:
//
//	type Assistant struct{ zrt.BaseAgent }
//	func (a *Assistant) OnEnter(ctx context.Context) error { _, err := a.Session().Say(ctx, "Hi!"); return err }
//	func (a *Assistant) OnExit(ctx context.Context) error  { return nil }
type Agent interface {
	// OnEnter is called once the session is live (after any wait-for-participant).
	OnEnter(ctx context.Context) error
	// OnExit is called as the session closes.
	OnExit(ctx context.Context) error
	base() *BaseAgent
}

// NamedAgent is an alternate agent definition usable for in-call handoff.
type NamedAgent struct {
	AgentID      string
	Instructions string
	Tools        []*FunctionTool
	Greeting     string
}

// AgentOptions configures a BaseAgent.
type AgentOptions struct {
	Instructions              string
	Name                      string
	Pipeline                  *Pipeline
	MaxSessionDurationSeconds *int
	Tools                     []*FunctionTool
	AgentID                   string
	MCPServers                []MCPServer
	InheritContext            bool
	KnowledgeBase             *KnowledgeBase
	Greeting                  string
	GreetingNonInterruptible  bool
	// VoiceSuffix overrides the spoken-style suffix. nil means "not set", which
	// (with AppendVoiceSuffix) appends the default voice suffix.
	VoiceSuffix *string
	// AppendVoiceSuffix controls whether a voice suffix is appended. nil defaults true.
	AppendVoiceSuffix *bool
	// Alternates are additional named agents for in-call handoff.
	Alternates []*NamedAgent
	Agents     []Agent
	// ContextWindow configures context management for this agent.
	ContextWindow *ContextWindow
}

// BaseAgent holds an agent's configuration and SDK wiring. Embed it.
type BaseAgent struct {
	instructions              string
	name                      string
	pipeline                  *Pipeline
	maxSessionDurationSeconds *int
	tools                     []*FunctionTool
	id                        string
	mcpServers                []MCPServer
	inheritContext            bool
	knowledgeBase             *KnowledgeBase
	greeting                  string
	greetingNonInterruptible  bool
	voiceSuffix               *string
	appendVoiceSuffix         bool
	alternates                []*NamedAgent
	handoffAgents             []Agent
	cw                        *ContextWindow

	session                  *AgentSession
	thinkingBackgroundConfig map[string]any
	preloadBGAudio           [][2]any

	beforeLLMHook          bool
	llmStreamHookEnabled   bool
	llmStreamHookTimeoutMS int
}

// NewBaseAgent builds a BaseAgent from opts. Embed the result in your agent type.
func NewBaseAgent(opts AgentOptions) BaseAgent {
	return BaseAgent{
		instructions:              opts.Instructions,
		name:                      opts.Name,
		pipeline:                  opts.Pipeline,
		maxSessionDurationSeconds: opts.MaxSessionDurationSeconds,
		tools:                     slices.Clone(opts.Tools),
		id:                        cmp.Or(opts.AgentID, opts.Name, "Agent"),
		mcpServers:                opts.MCPServers,
		inheritContext:            opts.InheritContext,
		knowledgeBase:             opts.KnowledgeBase,
		greeting:                  opts.Greeting,
		greetingNonInterruptible:  opts.GreetingNonInterruptible,
		voiceSuffix:               opts.VoiceSuffix,
		appendVoiceSuffix:         BoolOr(opts.AppendVoiceSuffix, true),
		alternates:                opts.Alternates,
		handoffAgents:             slices.Clone(opts.Agents),
		cw:                        opts.ContextWindow,
		llmStreamHookTimeoutMS:    100,
	}
}

// *BaseAgent is itself a usable Agent (default no-op OnEnter/OnExit).
var _ Agent = (*BaseAgent)(nil)

func NewAgent(opts AgentOptions) *BaseAgent {
	a := NewBaseAgent(opts)
	return &a
}
func (a *BaseAgent) OnEnter(ctx context.Context) error { return nil }

func (a *BaseAgent) OnExit(ctx context.Context) error { return nil }

//lint:ignore U1000 base is called on Agent values throughout the SDK and is satisfied by external types that embed BaseAgent, which staticcheck cannot see in-package.
func (a *BaseAgent) base() *BaseAgent { return a }

// contextWindow returns the agent's configured context window (nil-safe).
func (a *BaseAgent) contextWindow() *ContextWindow { return a.cw }

// Instructions returns the system instructions.
func (a *BaseAgent) Instructions() string { return a.instructions }

// SetInstructions replaces the system instructions locally. Use
// AgentSession.UpdateInstructions to apply them to a live session.
func (a *BaseAgent) SetInstructions(v string) { a.instructions = v }

// ID returns the agent id (the registration handle Serve registers under and
// Invoke targets). It is derived from AgentOptions.AgentID, then Name.
func (a *BaseAgent) ID() string { return a.id }

// AgentID is an alias for ID — the agent's registration handle.
func (a *BaseAgent) AgentID() string { return a.id }

// Name returns the agent's display name (falls back to the id when unset).
func (a *BaseAgent) Name() string { return cmp.Or(a.name, a.id) }

// Pipeline returns the voice stack this agent carries (nil if it has none).
func (a *BaseAgent) Pipeline() *Pipeline { return a.pipeline }

// MaxSessionDurationSeconds returns the per-agent session-duration cap (nil if unset).
func (a *BaseAgent) MaxSessionDurationSeconds() *int { return a.maxSessionDurationSeconds }

// Greeting returns the agent's opening line, if any.
func (a *BaseAgent) Greeting() string { return a.greeting }

// InheritContext reports whether this agent inherits the conversation context on
// a handoff into it.
func (a *BaseAgent) InheritContext() bool { return a.inheritContext }

// HandoffAgents returns the full agent objects reachable from this agent via
// handoff (configured through AgentOptions.Agents).
func (a *BaseAgent) HandoffAgents() []Agent { return a.handoffAgents }

// Tools returns the registered tools.
func (a *BaseAgent) Tools() []*FunctionTool { return a.tools }

// UpdateTools replaces the tool set locally. Use AgentSession.UpdateTools to
// apply them to a live session.
func (a *BaseAgent) UpdateTools(tools []*FunctionTool) {
	a.tools = slices.Clone(tools)
}

// Session returns the bound AgentSession (nil before start).
func (a *BaseAgent) Session() *AgentSession { return a.session }

// Hangup ends the call.
func (a *BaseAgent) Hangup(ctx context.Context) error {
	if a.session != nil {
		return a.session.Hangup(ctx, "manual_hangup")
	}
	return nil
}

// SetThinkingAudio configures the background audio played while the agent thinks.
// Pass an empty file to disable.
func (a *BaseAgent) SetThinkingAudio(file string, volume float64) {
	if file == "" {
		a.thinkingBackgroundConfig = nil
		return
	}
	a.thinkingBackgroundConfig = map[string]any{"file_url": file, "volume": volume, "looping": true}
}

// PlayBackgroundAudio plays a background audio file.
func (a *BaseAgent) PlayBackgroundAudio(ctx context.Context, file string, volume float64, looping bool) error {
	if a.session == nil || file == "" {
		return nil
	}
	return a.session.PlayBackgroundAudio(ctx, map[string]any{"file_url": file, "volume": volume, "looping": looping})
}

// StopBackgroundAudio stops background audio playback.
func (a *BaseAgent) StopBackgroundAudio(ctx context.Context) error {
	if a.session == nil {
		return nil
	}
	return a.session.StopBackgroundAudio(ctx)
}

// PreloadBackgroundAudio preloads a background audio file.
func (a *BaseAgent) PreloadBackgroundAudio(ctx context.Context, file string, volume float64) error {
	if a.session == nil || file == "" {
		return nil
	}
	return a.session.PreloadBackgroundAudio(ctx, map[string]any{"file_url": file, "volume": volume})
}

// CaptureFrames returns the latest buffered vision frames (most recent last).
func (a *BaseAgent) CaptureFrames(numFrames int) []map[string]any {
	if numFrames <= 0 {
		numFrames = 1
	}
	if a.session == nil || a.session.pipeline == nil {
		return nil
	}
	return a.session.pipeline.GetLatestFrames(numFrames)
}
