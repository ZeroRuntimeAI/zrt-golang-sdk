package zrt

import "sync"

const dtmfSequenceBufferMax = 32

// DTMFHandler dispatches DTMF digit/sequence callbacks.
type DTMFHandler struct {
	mu             sync.Mutex
	digitHandlers  map[string]func(string)
	seqHandlers    map[string]func(string)
	sequenceBuffer string
}

// NewDTMFHandler creates a DTMF handler.
func NewDTMFHandler() *DTMFHandler {
	return &DTMFHandler{digitHandlers: map[string]func(string){}, seqHandlers: map[string]func(string){}}
}

// OnDigit registers a callback for a single digit.
func (h *DTMFHandler) OnDigit(digit string, cb func(string)) {
	if digit == "" {
		return
	}
	h.mu.Lock()
	h.digitHandlers[digit] = cb
	h.mu.Unlock()
}

// OnSequence registers a callback for a digit sequence.
func (h *DTMFHandler) OnSequence(sequence string, cb func(string)) {
	if sequence == "" {
		return
	}
	h.mu.Lock()
	h.seqHandlers[sequence] = cb
	h.mu.Unlock()
}

func (h *DTMFHandler) dispatch(digit string) {
	h.mu.Lock()
	if cb, ok := h.digitHandlers[digit]; ok {
		invokeDTMF(cb, digit)
	}
	h.sequenceBuffer += digit
	if len(h.sequenceBuffer) > dtmfSequenceBufferMax {
		h.sequenceBuffer = h.sequenceBuffer[len(h.sequenceBuffer)-dtmfSequenceBufferMax:]
	}
	for seq, cb := range h.seqHandlers {
		if seq != "" && len(h.sequenceBuffer) >= len(seq) && h.sequenceBuffer[len(h.sequenceBuffer)-len(seq):] == seq {
			invokeDTMF(cb, seq)
			h.sequenceBuffer = ""
			break
		}
	}
	h.mu.Unlock()
}

func invokeDTMF(cb func(string), arg string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("DTMFHandler callback panicked: %v", r)
		}
	}()
	cb(arg)
}
