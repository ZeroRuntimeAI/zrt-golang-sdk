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
	// AgentID is the handle the runtime registers this agent under and that
	// DispatchAgent targets. When empty it falls back to the agent's name/id.
	AgentID string
	// AuthToken is the ZRT auth token. When empty it is resolved from the
	// environment (ZRT_AUTH_TOKEN, or ZRT_API_KEY + ZRT_SECRET_KEY).
	AuthToken string
	// Options, when set, is used as the base worker configuration and tuned for
	// registered (serve) mode. AgentID / AuthToken above win over the matching
	// fields here. When nil, sensible defaults are used and worker capacity
	// defaults to a CPU-based value (override via ZRT_MAX_CONCURRENT_SESSIONS).
	Options *WorkerOptions
}

// maxConcurrentSessions is the internal default worker capacity (max concurrent
// sessions), tunable via the ZRT_MAX_CONCURRENT_SESSIONS env var.
func maxConcurrentSessions() int {
	if raw := os.Getenv("ZRT_MAX_CONCURRENT_SESSIONS"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 1 {
			return n
		}
	}
	return max(4, runtime.NumCPU()*8)
}
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

	displayName := cmp.Or(agent.base().name, agent.base().id, "Agent")

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

	switch {
	case opts.AgentID != "":
		workerOpts.AgentID = opts.AgentID
	case workerOpts.AgentID == "" || workerOpts.AgentID == "ZeroRuntimeAgent":
		workerOpts.AgentID = displayName
	}
	if usingDefaults {
		workerOpts.MaxProcesses = maxConcurrentSessions()
	}
	if opts.AuthToken != "" && workerOpts.AuthToken == "" {
		workerOpts.AuthToken = opts.AuthToken
	}

	jobctxFactory := func() *JobContext {
		return NewJobContext(&RoomOptions{Name: displayName}, nil)
	}

	return NewWorkerJob(entrypoint, jobctxFactory, workerOpts).Start()
}
