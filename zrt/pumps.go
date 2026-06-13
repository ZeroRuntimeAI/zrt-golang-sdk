package zrt

import "sync"

// customSTTPump feeds runtime audio chunks into a CustomSTTHook and forwards the
// hook's transcripts back over a buffered channel.
type customSTTPump struct {
	in     chan CustomSTTAudioChunk
	once   sync.Once
	closed bool
	mu     sync.Mutex
}

func newCustomSTTPump(hook CustomSTTHook, send func(CustomSTTResult)) *customSTTPump {
	p := &customSTTPump{in: make(chan CustomSTTAudioChunk, 256)}
	out := hook(p.in)
	go func() {
		for r := range out {
			send(r)
		}
	}()
	return p
}

func (p *customSTTPump) push(chunk CustomSTTAudioChunk) {
	p.mu.Lock()
	closed := p.closed
	p.mu.Unlock()
	if closed {
		return
	}
	select {
	case p.in <- chunk:
	default:
		logger.Debugf("customSTTPump audio queue full — dropping chunk (utterance_id=%s)", chunk.UtteranceID)
	}
}

func (p *customSTTPump) close() {
	p.mu.Lock()
	p.closed = true
	p.mu.Unlock()
	p.once.Do(func() { close(p.in) })
}

// customTTSPump feeds runtime synthesis requests into a CustomTTSHook and
// forwards the hook's audio back over a buffered channel.
type customTTSPump struct {
	in     chan CustomTTSSynthesize
	once   sync.Once
	closed bool
	mu     sync.Mutex
}

func newCustomTTSPump(hook CustomTTSHook, send func(CustomTTSAudioChunk)) *customTTSPump {
	p := &customTTSPump{in: make(chan CustomTTSSynthesize, 64)}
	out := hook(p.in)
	go func() {
		for c := range out {
			send(c)
		}
	}()
	return p
}

func (p *customTTSPump) push(req CustomTTSSynthesize) {
	p.mu.Lock()
	closed := p.closed
	p.mu.Unlock()
	if closed {
		return
	}
	select {
	case p.in <- req:
	default:
		logger.Warnf("customTTSPump synthesize queue full — dropping request (utterance_id=%s)", req.UtteranceID)
	}
}

func (p *customTTSPump) close() {
	p.mu.Lock()
	p.closed = true
	p.mu.Unlock()
	p.once.Do(func() { close(p.in) })
}
