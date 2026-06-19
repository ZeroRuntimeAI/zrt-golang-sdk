package zrt

import (
	"slices"
	"sync"
)

// EventHandler receives a single event payload (usually a map[string]any).
type EventHandler func(payload any)

type eventReg struct {
	fn EventHandler
}

// EventEmitter is a minimal synchronous event emitter.
//
// Handlers are invoked synchronously in registration order; panics in a handler
// are recovered and logged so one bad handler cannot break the emit loop.
type EventEmitter struct {
	mu       sync.Mutex
	handlers map[string][]*eventReg
	closed   bool
}

func (e *EventEmitter) ensure() {
	if e.handlers == nil {
		e.handlers = make(map[string][]*eventReg)
	}
}

// On registers a handler for event and returns a function that unsubscribes it.
func (e *EventEmitter) On(event string, fn EventHandler) func() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.ensure()
	reg := &eventReg{fn: fn}
	e.handlers[event] = append(e.handlers[event], reg)
	return func() { e.removeReg(event, reg) }
}

func (e *EventEmitter) removeReg(event string, reg *eventReg) {
	e.mu.Lock()
	defer e.mu.Unlock()
	regs := e.handlers[event]
	out := regs[:0]
	for _, r := range regs {
		if r != reg {
			out = append(out, r)
		}
	}
	e.handlers[event] = out
}

// Emit invokes every handler registered for event with payload.
func (e *EventEmitter) Emit(event string, payload any) {
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return
	}
	regs := slices.Clone(e.handlers[event])
	e.mu.Unlock()
	for _, r := range regs {
		invokeHandler(r.fn, payload, event)
	}
}

func invokeHandler(fn EventHandler, payload any, event string) {
	defer func() {
		if rec := recover(); rec != nil {
			logger.Errorf("event handler for %q panicked: %v", event, rec)
		}
	}()
	fn(payload)
}

// safeHook runs a user-supplied pipeline hook, recovering and logging any panic
// so that one misbehaving hook cannot crash the session event loop.
func safeHook(name string, fn func()) {
	defer func() {
		if rec := recover(); rec != nil {
			logger.Errorf("pipeline hook %q panicked: %v", name, rec)
		}
	}()
	fn()
}

func (e *EventEmitter) closeEmitter() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.closed = true
	e.handlers = nil
}
