package zrt

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
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
	// RtcBaseURL is the registry base URL (defaults to $ZRT_SIGNALING_URL or
	// api.videosdk.live).
	RtcBaseURL string
	// LogLevel is the worker log level (default "INFO").
	LogLevel string
	// Options, when set, is used as the base worker configuration and tuned for
	// registered (serve) mode. The fields above win over the matching fields here.
	Options *WorkerOptions
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

// Serve registers an agent with the ZRT registry over a WebSocket connection and
// listens for incoming sessions. It does not start a session on its own — call
// Invoke to start one. Serve blocks until the worker shuts down (Ctrl+C / SIGTERM).
func Serve(agent Agent, opts ServeOptions) error {
	if agent == nil {
		return fmt.Errorf("zrt.Serve: agent is required")
	}
	pipeline := agent.base().pipeline
	if pipeline == nil {
		return fmt.Errorf("zrt.Serve requires the agent to carry a pipeline: " +
			"use zrt.NewAgent(zrt.AgentOptions{..., Pipeline: pipeline}) " +
			"or set AgentOptions.Pipeline when building the embedded BaseAgent")
	}

	// The registration handle is the agent's id (set via AgentOptions.AgentID),
	// overridable per call via opts.AgentID.
	registeredID := cmp.Or(opts.AgentID, agent.base().id)
	if registeredID == "" {
		return fmt.Errorf("zrt.Serve requires the agent to have an id: " +
			"set AgentOptions.AgentID (the handle the registry routes to)")
	}
	displayName := cmp.Or(agent.base().name, registeredID)

	entrypoint := func(ctx context.Context, jobCtx *JobContext) error {
		session := NewAgentSession(agent, pipeline, AgentSessionOptions{})
		return session.Start(ctx, jobCtx, StartOptions{
			WaitForParticipant: true,
			RunUntilShutdown:   true,
		})
	}

	// Base worker options, forced into registered (serve) mode.
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
		return NewJobContext(&RoomOptions{Name: displayName}, nil)
	}

	return NewWorkerJob(entrypoint, jobctxFactory, workerOpts).Start()
}
