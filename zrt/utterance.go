package zrt

import (
	"sync"

	"github.com/google/uuid"
)

// UtteranceHandle tracks an in-flight spoken utterance. Use it to wait for the
// utterance to finish, interrupt it, or run a callback when its audio completes.
type UtteranceHandle struct {
	id            string
	interruptible bool

	mu          sync.Mutex
	doneCh      chan struct{}
	doneClosed  bool
	interrupted bool
	lastByteCBs []func()
}

// NewUtteranceHandle creates a handle. An empty id is auto-generated.
func NewUtteranceHandle(id string, interruptible bool) *UtteranceHandle {
	if id == "" {
		id = uuid.NewString()
	}
	return &UtteranceHandle{id: id, interruptible: interruptible, doneCh: make(chan struct{})}
}

// ID returns the utterance id.
func (u *UtteranceHandle) ID() string { return u.id }

// IsInterruptible reports whether the utterance can be interrupted.
func (u *UtteranceHandle) IsInterruptible() bool { return u.interruptible }

// Done reports whether the utterance has finished.
func (u *UtteranceHandle) Done() bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.doneClosed
}

// Interrupted reports whether the utterance was interrupted.
func (u *UtteranceHandle) Interrupted() bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.interrupted
}

// Wait blocks until the utterance finishes or is interrupted.
func (u *UtteranceHandle) Wait() {
	<-u.doneCh
}

// C returns a channel closed when the utterance finishes (for select).
func (u *UtteranceHandle) C() <-chan struct{} { return u.doneCh }

// Interrupt marks the utterance interrupted. With force, ignores interruptible.
func (u *UtteranceHandle) Interrupt(force bool) {
	u.mu.Lock()
	if !(u.interruptible || force) {
		u.mu.Unlock()
		return
	}
	u.interrupted = true
	cbs := u.closeLocked()
	u.mu.Unlock()
	fireCallbacks(cbs)
}

// OnLastAudioByte registers a callback invoked when the last audio byte plays
// (or immediately if already done).
func (u *UtteranceHandle) OnLastAudioByte(cb func()) {
	u.mu.Lock()
	if u.doneClosed {
		u.mu.Unlock()
		fireCallbacks([]func(){cb})
		return
	}
	u.lastByteCBs = append(u.lastByteCBs, cb)
	u.mu.Unlock()
}

func (u *UtteranceHandle) markDone() {
	u.mu.Lock()
	if u.doneClosed {
		u.mu.Unlock()
		return
	}
	cbs := u.closeLocked()
	u.mu.Unlock()
	fireCallbacks(cbs)
}

// closeLocked closes the done channel (idempotent) and returns pending
// last-audio-byte callbacks. Caller must hold u.mu.
func (u *UtteranceHandle) closeLocked() []func() {
	if !u.doneClosed {
		u.doneClosed = true
		close(u.doneCh)
	}
	cbs := u.lastByteCBs
	u.lastByteCBs = nil
	return cbs
}

func fireCallbacks(cbs []func()) {
	for _, cb := range cbs {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Errorf("on_last_audio_byte callback panicked: %v", r)
				}
			}()
			cb()
		}()
	}
}
