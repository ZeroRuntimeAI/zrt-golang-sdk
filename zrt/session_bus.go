package zrt

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
)

// SessionBus lets a single shared agent resolve the correct per-call session.
// The current session's bus id is carried on context.Context, so one agent
// instance can serve many concurrent sessions without cross-wiring callers.

type sessionBusKeyT struct{}

// sessionBusKey is the context key under which the current session's bus id is stored.
var sessionBusKey = sessionBusKeyT{}

var (
	sessionBusMu  sync.RWMutex
	sessionBusReg = map[string]*AgentSession{}
)

// sessionBusNewID returns a fresh, unique bus id for a session.
func sessionBusNewID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Avoid panicking in a constructor if crypto/rand fails.
		return hex.EncodeToString(b[:])
	}
	return hex.EncodeToString(b[:])
}

func sessionBusRegister(id string, s *AgentSession) {
	if id == "" {
		return
	}
	sessionBusMu.Lock()
	sessionBusReg[id] = s
	sessionBusMu.Unlock()
}

func sessionBusUnregister(id string) {
	if id == "" {
		return
	}
	sessionBusMu.Lock()
	delete(sessionBusReg, id)
	sessionBusMu.Unlock()
}

func sessionBusLookup(id string) *AgentSession {
	sessionBusMu.RLock()
	s := sessionBusReg[id]
	sessionBusMu.RUnlock()
	return s
}

// bindSession returns a copy of ctx carrying id as the current session binding.
func bindSession(ctx context.Context, id string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, sessionBusKey, id)
}

// SessionFromContext returns the AgentSession bound in ctx, or nil if none is
// bound. Use it inside a function-tool handler (or any ctx-carrying callback) to
// reach the call's own session when the agent instance is shared:
//
//	result := func(ctx context.Context, args map[string]any) (any, error) {
//	    s := zrt.SessionFromContext(ctx)
//	    _, err := s.Say(ctx, "...")
//	    return nil, err
//	}
func SessionFromContext(ctx context.Context) *AgentSession {
	if ctx == nil {
		return nil
	}
	id, _ := ctx.Value(sessionBusKey).(string)
	if id == "" {
		return nil
	}
	return sessionBusLookup(id)
}

// bindBus returns ctx carrying this session's bus id.
func (s *AgentSession) bindBus(ctx context.Context) context.Context {
	return bindSession(ctx, s.busID)
}
