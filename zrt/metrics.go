package zrt

import "sync"

// MetricsCollector buffers per-turn metrics snapshots emitted by the runtime.
type MetricsCollector struct {
	mu    sync.Mutex
	turns []map[string]any
}

func (m *MetricsCollector) append(payload map[string]any) {
	m.mu.Lock()
	m.turns = append(m.turns, payload)
	m.mu.Unlock()
}

// Turns returns a copy of the buffered metric snapshots.
func (m *MetricsCollector) Turns() []map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]map[string]any, len(m.turns))
	copy(out, m.turns)
	return out
}

// Clear empties the buffer.
func (m *MetricsCollector) Clear() {
	m.mu.Lock()
	m.turns = nil
	m.mu.Unlock()
}

// metricsCollector is the process-wide metrics buffer.
var metricsCollector = &MetricsCollector{}

// Metrics returns the process-wide metrics collector.
func Metrics() *MetricsCollector { return metricsCollector }
