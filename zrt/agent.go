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
	// base exposes the embedded BaseAgent to the SDK.
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
	Instructions             string
	Tools                    []*FunctionTool
	AgentID                  string
	MCPServers               []MCPServer
	InheritContext           bool
	KnowledgeBase            *KnowledgeBase
	Greeting                 string
	GreetingNonInterruptible bool
	// VoiceSuffix overrides the spoken-style suffix. nil means "not set", which
	// (with AppendVoiceSuffix) appends the default voice suffix.
	VoiceSuffix *string
	// AppendVoiceSuffix controls whether a voice suffix is appended. nil defaults true.
	AppendVoiceSuffix *bool
	// Alternates are additional named agents for in-call handoff.
	Alternates []*NamedAgent
	// ContextWindow configures server-side context management for this agent.
	ContextWindow *ContextWindow
}

// BaseAgent holds an agent's configuration and SDK wiring. Embed it.
type BaseAgent struct {
	instructions             string
	tools                    []*FunctionTool
	id                       string
	mcpServers               []MCPServer
	inheritContext           bool
	knowledgeBase            *KnowledgeBase
	greeting                 string
	greetingNonInterruptible bool
	voiceSuffix              *string
	appendVoiceSuffix        bool
	alternates               []*NamedAgent
	cw                       *ContextWindow

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
		instructions:             opts.Instructions,
		tools:                    slices.Clone(opts.Tools),
		id:                       cmp.Or(opts.AgentID, "Agent"),
		mcpServers:               opts.MCPServers,
		inheritContext:           opts.InheritContext,
		knowledgeBase:            opts.KnowledgeBase,
		greeting:                 opts.Greeting,
		greetingNonInterruptible: opts.GreetingNonInterruptible,
		voiceSuffix:              opts.VoiceSuffix,
		appendVoiceSuffix:        BoolOr(opts.AppendVoiceSuffix, true),
		alternates:               opts.Alternates,
		cw:                       opts.ContextWindow,
		llmStreamHookTimeoutMS:   100,
	}
}

//lint:ignore U1000 base is called on Agent values throughout the SDK and is satisfied by external types that embed BaseAgent, which staticcheck cannot see in-package.
func (a *BaseAgent) base() *BaseAgent { return a }

// contextWindow returns the agent's configured context window (nil-safe).
func (a *BaseAgent) contextWindow() *ContextWindow { return a.cw }

// Instructions returns the system instructions.
func (a *BaseAgent) Instructions() string { return a.instructions }

// SetInstructions replaces the system instructions (local copy only; use
// AgentSession.UpdateInstructions to push to the runtime).
func (a *BaseAgent) SetInstructions(v string) { a.instructions = v }

// ID returns the agent id.
func (a *BaseAgent) ID() string { return a.id }

// Tools returns the registered tools.
func (a *BaseAgent) Tools() []*FunctionTool { return a.tools }

// UpdateTools replaces the local tool set (use AgentSession.UpdateTools to push).
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
