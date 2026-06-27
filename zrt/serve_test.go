package zrt

import (
	"context"
	"sync"
	"testing"
)

// TestSessionBusResolvesPerContext verifies the F1 fix: one shared agent serving two
// sessions resolves the correct session per call via the SessionBus binding in context —
// no cross-wiring. Without the bus, agent.Session(ctx) would fall back to the single shared
// field and both calls would see the last-attached session.
func TestSessionBusResolvesPerContext(t *testing.T) {
	agent := NewAgent(AgentOptions{AgentID: "shared", Pipeline: NewPipeline(PipelineOptions{})})

	s1 := NewAgentSession(agent, agent.base().pipeline, AgentSessionOptions{})
	s2 := NewAgentSession(agent, agent.base().pipeline, AgentSessionOptions{})

	ctx1 := s1.bindBus(context.Background())
	ctx2 := s2.bindBus(context.Background())

	// The shared agent resolves each call's own session from its context.
	if got := agent.Session(ctx1); got != s1 {
		t.Fatalf("agent.Session(ctx1) = %p, want s1 %p", got, s1)
	}
	if got := agent.Session(ctx2); got != s2 {
		t.Fatalf("agent.Session(ctx2) = %p, want s2 %p", got, s2)
	}

	// SessionFromContext returns the same.
	if SessionFromContext(ctx1) != s1 || SessionFromContext(ctx2) != s2 {
		t.Fatal("SessionFromContext did not resolve the bound session")
	}

	// No binding -> falls back to the most recently attached session (capacity-1 behavior).
	if got := agent.Session(context.Background()); got != s2 {
		t.Fatalf("unbound agent.Session = %p, want fallback s2 %p", got, s2)
	}
}

// TestSessionBusConcurrent hammers two sessions on one shared agent from many goroutines and
// asserts each context always resolves to its own session (race detector exercises locking).
func TestSessionBusConcurrent(t *testing.T) {
	agent := NewAgent(AgentOptions{AgentID: "shared", Pipeline: NewPipeline(PipelineOptions{})})
	s1 := NewAgentSession(agent, agent.base().pipeline, AgentSessionOptions{})
	s2 := NewAgentSession(agent, agent.base().pipeline, AgentSessionOptions{})
	ctx1 := s1.bindBus(context.Background())
	ctx2 := s2.bindBus(context.Background())

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); if agent.Session(ctx1) != s1 { t.Error("ctx1 resolved wrong session") } }()
		go func() { defer wg.Done(); if agent.Session(ctx2) != s2 { t.Error("ctx2 resolved wrong session") } }()
	}
	wg.Wait()
}

// TestSessionBusUnregister verifies a closed session is removed from the registry so its
// context no longer resolves it.
func TestSessionBusUnregister(t *testing.T) {
	agent := NewAgent(AgentOptions{AgentID: "shared", Pipeline: NewPipeline(PipelineOptions{})})
	s := NewAgentSession(agent, agent.base().pipeline, AgentSessionOptions{})
	ctx := s.bindBus(context.Background())
	if SessionFromContext(ctx) != s {
		t.Fatal("session not registered")
	}
	sessionBusUnregister(s.busID)
	if SessionFromContext(ctx) != nil {
		t.Fatal("session should be gone from the bus after unregister")
	}
}
