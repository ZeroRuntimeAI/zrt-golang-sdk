package zrt

import (
	"cmp"
	"context"
	"fmt"
	"maps"
	"os"
	"runtime"
	"strconv"
	"time"
)

// ServeOptions configures Serve.
type ServeOptions struct {
	// AgentID overrides the handle the registry registers this agent under and
	// that Invoke targets. When empty it defaults to agent.ID().
	AgentID string
	// AuthToken is the ZRT auth token. When empty it is resolved from the
	// environment (ZRT_AUTH_TOKEN, or ZRT_API_KEY + ZRT_SECRET_KEY).
	AuthToken string
	// OnReady, when set, is called once registration is confirmed. It may block /
	// call zrt.Invoke; it runs on its own goroutine and Serve keeps running.
	OnReady func()
	// Capacity is the concurrent session capacity. When 0 a CPU-based default is
	// used. ZRT_MAX_CONCURRENT_SESSIONS, when set, overrides this value.
	Capacity int
	// LoadThreshold is the load fraction above which new sessions are rejected
	// (default 0.75).
	LoadThreshold float64
	// InitializeTimeout bounds how long the worker waits for the agent/models to
	// initialize before it is considered failed. Raise it when model load is
	// slow. Zero uses the default (10s).
	InitializeTimeout time.Duration
	// RtcBaseURL is the registry base URL (defaults to $ZRT_SIGNALING_URL or
	// api.videosdk.live).
	RtcBaseURL string
	// LogLevel is the worker log level (default "INFO").
	LogLevel string
	// SessionOptions configures each served AgentSession — e.g. a DTMFHandler, a
	// VoiceMailDetector, a background-audio bed, or wake-up. The same options are
	// shared by every concurrent session this worker serves.
	SessionOptions AgentSessionOptions
	// RoomOptions configures the room each served session joins — e.g. Vision (to
	// subscribe the caller's video) or BackgroundAudio. It is cloned per session;
	// Name defaults to the agent's display name when empty.
	RoomOptions *RoomOptions
	// Options, when set, is used as the base worker configuration and tuned for
	// registered (serve) mode. The fields above win over the matching fields here.
	Options *WorkerOptions
	// AgentFactory, when set, builds a fresh Agent (with its own Pipeline) for
	// each incoming call, giving full per-call isolation for stateful agents. The
	// agent argument to Serve may then be nil. Without it a single shared agent
	// serves every call.
	AgentFactory func() Agent
}

// maxConcurrentSessions resolves the worker capacity. ZRT_MAX_CONCURRENT_SESSIONS
// wins; then an explicit capacity; otherwise a CPU-based default.
func maxConcurrentSessions(capacity int) int {
	if raw := os.Getenv("ZRT_MAX_CONCURRENT_SESSIONS"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 1 {
			return n
		}
	}
	if capacity > 0 {
		return capacity
	}
	return max(4, runtime.NumCPU()*8)
}

// Serve registers an agent with the ZRT registry and listens for incoming sessions.
// It does not start a session on its own — call Invoke to start one. Serve blocks
// until the worker shuts down (Ctrl+C / SIGTERM).
//
// Without opts.AgentFactory a single agent instance is shared across all
// concurrent sessions. Set opts.AgentFactory to build a fresh Agent + Pipeline
// per call for full per-call isolation. Capacity defaults to a CPU-based value;
// tune it with opts.Capacity or ZRT_MAX_CONCURRENT_SESSIONS.
func Serve(agent Agent, opts ServeOptions) error {
	factory := opts.AgentFactory
	template := agent
	if template == nil {
		if factory == nil {
			return fmt.Errorf("zrt.Serve: agent is required (pass an Agent, or set ServeOptions.AgentFactory)")
		}
		template = factory()
		if template == nil {
			return fmt.Errorf("zrt.Serve: ServeOptions.AgentFactory returned nil; it must return an Agent")
		}
	}

	if template.base().pipeline == nil {
		return fmt.Errorf("zrt.Serve requires the agent to carry a pipeline: " +
			"use zrt.NewAgent(zrt.AgentOptions{...}) " +
			"or set AgentOptions.Pipeline when building the embedded BaseAgent")
	}

	registeredID := cmp.Or(opts.AgentID, template.base().id)
	if registeredID == "" {
		return fmt.Errorf("zrt.Serve requires the agent to have an id: " +
			"set AgentOptions.AgentID (the handle the registry routes to)")
	}
	displayName := cmp.Or(template.base().name, registeredID)

	entrypoint := func(ctx context.Context, jobCtx *JobContext) error {
		callAgent := template
		if factory != nil {
			callAgent = factory()
			if callAgent == nil {
				return fmt.Errorf("zrt.Serve: ServeOptions.AgentFactory returned nil")
			}
		}
		callPipeline := callAgent.base().pipeline
		md := maps.Clone(jobCtx.Metadata)
		if md == nil {
			md = map[string]any{}
		}
		callAgent.base().metadata = md
		if jobCtx.RoomOptions != nil {
			callAgent.base().roomID = jobCtx.RoomOptions.RoomID
		}
		session := NewAgentSession(callAgent, callPipeline, opts.SessionOptions)
		return session.Start(ctx, jobCtx, StartOptions{
			WaitForParticipant: true,
			RunUntilShutdown:   true,
		})
	}

	workerOpts := opts.Options
	usingDefaults := workerOpts == nil
	if workerOpts == nil {
		workerOpts = NewWorkerOptions()
	}
	workerOpts.Register = true
	workerOpts.AgentID = registeredID

	if usingDefaults || opts.Capacity > 0 {
		workerOpts.MaxProcesses = maxConcurrentSessions(opts.Capacity)
	}
	if opts.AuthToken != "" {
		workerOpts.AuthToken = opts.AuthToken
	}
	if opts.LoadThreshold > 0 {
		workerOpts.LoadThreshold = opts.LoadThreshold
	}
	if opts.InitializeTimeout > 0 {
		workerOpts.InitializeTimeout = opts.InitializeTimeout
	}
	if opts.RtcBaseURL != "" {
		workerOpts.SignalingBaseURL = opts.RtcBaseURL
	}
	if opts.LogLevel != "" {
		workerOpts.LogLevel = opts.LogLevel
	}
	if opts.OnReady != nil {
		workerOpts.OnReady = opts.OnReady
	}

	jobctxFactory := func() *JobContext {
		ro := &RoomOptions{Name: displayName}
		if opts.RoomOptions != nil {
			clone := *opts.RoomOptions
			ro = &clone
			if ro.Name == "" {
				ro.Name = displayName
			}
		}
		return NewJobContext(ro, nil)
	}

	return NewWorkerJob(entrypoint, jobctxFactory, workerOpts).Start()
}
