package zrt

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// WakeUp returns the configured wake-up timeout (seconds; 0 = disabled).
func (s *AgentSession) WakeUp() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.wakeUp
}

// SetWakeUp sets the wake-up timeout (seconds).
func (s *AgentSession) SetWakeUp(v int) {
	s.mu.Lock()
	s.wakeUp = v
	s.mu.Unlock()
}

// IsBackgroundAudioEnabled reports whether background audio is enabled for the
// bound room.
func (s *AgentSession) IsBackgroundAudioEnabled() bool {
	if jc := s.boundJobCtx(); jc != nil && jc.RoomOptions != nil {
		return jc.RoomOptions.BackgroundAudio
	}
	return false
}

func (s *AgentSession) boundJobCtx() *JobContext { return s.jobCtx }

// UpdateInterruptConfigOptions configures UpdateInterruptConfig. nil fields are
// left unchanged.
type UpdateInterruptConfigOptions struct {
	// Mode is the interruption mode: "VAD_ONLY", "STT_ONLY", or "HYBRID" (case-insensitive).
	Mode *string
	// InterruptMinDuration is the minimum speech duration, in seconds, before an interruption is honored.
	InterruptMinDuration *float64
	// InterruptMinWords is the minimum number of words before an interruption is honored.
	InterruptMinWords *int
	// CooldownMS is the cooldown between interruptions, in milliseconds.
	CooldownMS *int
	// FalseInterruptPauseDurationMS is the pause length treated as a false interrupt, in milliseconds.
	FalseInterruptPauseDurationMS *int
	// ResumeOnFalseInterrupt resumes the agent's turn after a false interrupt.
	ResumeOnFalseInterrupt *bool
}

// UpdateInterruptConfig updates interruption behavior at runtime.
func (s *AgentSession) UpdateInterruptConfig(ctx context.Context, opts UpdateInterruptConfigOptions) error {
	params := map[string]string{}
	if opts.Mode != nil {
		m := strings.ToLower(strings.TrimSpace(*opts.Mode))
		valid := map[string]bool{"vad_only": true, "stt_only": true, "hybrid": true}
		if !valid[m] {
			return fmt.Errorf("update_interrupt_config: mode=%q invalid; expected VAD_ONLY / STT_ONLY / HYBRID", *opts.Mode)
		}
		params["mode"] = m
	}
	if opts.InterruptMinDuration != nil {
		params["min_duration_ms"] = strconv.Itoa(int(*opts.InterruptMinDuration*1000 + 0.5))
	}
	if opts.InterruptMinWords != nil {
		params["min_words"] = strconv.Itoa(*opts.InterruptMinWords)
	}
	if opts.CooldownMS != nil {
		params["cooldown_ms"] = strconv.Itoa(*opts.CooldownMS)
	}
	if opts.FalseInterruptPauseDurationMS != nil {
		params["false_interrupt_pause_duration_ms"] = strconv.Itoa(*opts.FalseInterruptPauseDurationMS)
	}
	if opts.ResumeOnFalseInterrupt != nil {
		params["resume_on_false_interrupt"] = strconv.FormatBool(*opts.ResumeOnFalseInterrupt)
	}
	if len(params) == 0 {
		return nil
	}
	return s.UpdateProvider(ctx, "interrupt", "interrupt", params)
}

// WarmTransfer is not supported; use AgentSwitch for multi-agent handoff.
func (s *AgentSession) WarmTransfer(ctx context.Context) error {
	return fmt.Errorf("%w: AgentSession.WarmTransfer; use AgentSwitch(...) for multi-agent handoff", ErrNotImplemented)
}

// ContextHistory returns the locally-cached conversation history, filtered by
// the given options. Use FetchContextHistory to pull the authoritative history.
func (s *AgentSession) ContextHistory(lastN int, includeFunctionCalls, includeSystemMessages bool) []map[string]any {
	s.mu.Lock()
	source := s.chatHistoryCache
	if len(source) == 0 && len(s.transcriptMirror) > 0 {
		source = s.transcriptMirror
	}
	snap := slices.Clone(source)
	s.mu.Unlock()
	return filterHistory(snap, lastN, includeFunctionCalls, includeSystemMessages)
}

// ---- RecordingManager ----

var recordingActiveStates = map[string]bool{"recording": true, "finalizing": true, "uploading": true}

// RecordingManager exposes the current recording status.
type RecordingManager struct {
	session *AgentSession
}

// RecordingManager returns the session's recording manager.
func (s *AgentSession) RecordingManager() *RecordingManager {
	return &RecordingManager{session: s}
}

// LastStatus returns the latest recording status (may be nil).
func (m *RecordingManager) LastStatus() map[string]any { return m.session.RecordingState() }

// IsRecording reports whether a recording is active.
func (m *RecordingManager) IsRecording() bool {
	s := m.LastStatus()
	if s == nil {
		return false
	}
	state, _ := s["state"].(string)
	return recordingActiveStates[state]
}

// RecordingID returns the active recording id.
func (m *RecordingManager) RecordingID() string { return statusString(m.LastStatus(), "recording_id") }

// OutputURI returns the recording output URI.
func (m *RecordingManager) OutputURI() string { return statusString(m.LastStatus(), "output_uri") }

// TrackKind returns the recording track kind.
func (m *RecordingManager) TrackKind() string { return statusString(m.LastStatus(), "track_kind") }

// DurationSeconds returns the current recording duration.
func (m *RecordingManager) DurationSeconds() float64 {
	if s := m.LastStatus(); s != nil {
		if v, ok := s["duration_seconds"].(float32); ok {
			return float64(v)
		}
		if v, ok := s["duration_seconds"].(float64); ok {
			return v
		}
	}
	return 0
}

// FileSizeBytes returns the current recording size in bytes.
func (m *RecordingManager) FileSizeBytes() uint64 {
	if s := m.LastStatus(); s != nil {
		if v, ok := s["file_size_bytes"].(uint64); ok {
			return v
		}
	}
	return 0
}

func statusString(s map[string]any, key string) string {
	if s == nil {
		return ""
	}
	v, _ := s[key].(string)
	return v
}

// ---- AudioTrack ----

// AudioTrack is a handle to the agent's outgoing audio.
type AudioTrack struct {
	session         *AgentSession
	sampleRate      int
	lastAudioByteCB func(durationSeconds float64)
}

// AudioTrack returns the session's outgoing audio track.
func (s *AgentSession) AudioTrack() *AudioTrack {
	s.mu.Lock()
	if s.audioTrackCache != nil {
		t := s.audioTrackCache
		s.mu.Unlock()
		return t
	}
	sampleRate := 48000
	if tts, ok := s.pipeline.tts.(interface{ SampleRate() int }); ok && s.pipeline.tts != nil {
		sampleRate = tts.SampleRate()
	}
	t := &AudioTrack{session: s, sampleRate: sampleRate}
	s.audioTrackCache = t
	s.mu.Unlock()
	s.On("last_audio_byte", func(payload any) {
		if t.lastAudioByteCB == nil {
			return
		}
		dur := 0.0
		if m, ok := payload.(map[string]any); ok {
			switch v := m["duration_seconds"].(type) {
			case float32:
				dur = float64(v)
			case float64:
				dur = v
			}
		}
		t.lastAudioByteCB(dur)
	})
	return t
}

// SampleRate returns the track sample rate.
func (t *AudioTrack) SampleRate() int { return t.sampleRate }

// IsSpeaking reports whether the agent is currently speaking.
func (t *AudioTrack) IsSpeaking() bool { return t.session.AgentState() == AgentStateSpeaking }

// CanPause reports whether the TTS provider supports pause.
func (t *AudioTrack) CanPause() bool {
	caps := t.session.TTSCapabilities()
	if caps == nil {
		return false
	}
	v, _ := caps["can_pause"].(bool)
	return v
}

// AddNewBytes sends a raw PCM frame as the agent's outgoing audio.
func (t *AudioTrack) AddNewBytes(ctx context.Context, pcm []byte, sampleRate int) error {
	sampleRate = cmp.Or(sampleRate, t.sampleRate)
	return t.session.PushAudioFrame(ctx, pcm, sampleRate)
}

// Interrupt cancels the current generation.
func (t *AudioTrack) Interrupt(ctx context.Context) error {
	return t.session.CancelGeneration(ctx)
}

// OnLastAudioByte registers a callback fired when the last audio byte plays.
func (t *AudioTrack) OnLastAudioByte(cb func(durationSeconds float64)) { t.lastAudioByteCB = cb }
