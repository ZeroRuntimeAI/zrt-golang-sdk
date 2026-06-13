package zrt

import "sync"

// VoiceMailDetectorDefaultPrompt is the default classifier prompt.
const VoiceMailDetectorDefaultPrompt = "You are a voicemail detection classifier for an OUTBOUND calling system. A bot has called a phone number and you need to determine if a human answered or if the call went to voicemail based on the provided text. Answer in one word yes or no."

// VoiceMailDetector configures runtime voicemail detection and receives events.
type VoiceMailDetector struct {
	Callback           func(map[string]any)
	Duration           float64 // seconds
	CustomPrompt       string
	AutoHangup         bool
	DetectionThreshold float64

	mu        sync.Mutex
	detected  bool
	lastEvent map[string]any
}

// VoiceMailDetectorOptions configures a VoiceMailDetector.
type VoiceMailDetectorOptions struct {
	Callback            func(map[string]any)
	Duration            *float64 // default 2.0
	CustomPrompt        string
	AutoHangup          bool
	DetectionThreshold  *float64 // default 1.0
	MaxDetectionSeconds *int     // overrides Duration if set
}

// NewVoiceMailDetector builds a VoiceMailDetector from the given options.
func NewVoiceMailDetector(opts VoiceMailDetectorOptions) *VoiceMailDetector {
	duration := 2.0
	if opts.Duration != nil {
		duration = *opts.Duration
	}
	if opts.MaxDetectionSeconds != nil {
		duration = float64(*opts.MaxDetectionSeconds)
	}
	threshold := 1.0
	if opts.DetectionThreshold != nil {
		threshold = *opts.DetectionThreshold
	}
	return &VoiceMailDetector{
		Callback:           opts.Callback,
		Duration:           duration,
		CustomPrompt:       opts.CustomPrompt,
		AutoHangup:         opts.AutoHangup,
		DetectionThreshold: threshold,
	}
}

// IsDetected reports whether voicemail was detected.
func (v *VoiceMailDetector) IsDetected() bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.detected
}

// LastEvent returns the last detection event payload.
func (v *VoiceMailDetector) LastEvent() map[string]any {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.lastEvent
}

func (v *VoiceMailDetector) onRuntimeEvent(payload map[string]any) {
	v.mu.Lock()
	v.detected = true
	v.lastEvent = payload
	cb := v.Callback
	v.mu.Unlock()
	if cb == nil {
		return
	}
	func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("VoiceMailDetector callback panicked: %v", r)
			}
		}()
		cb(payload)
	}()
}
