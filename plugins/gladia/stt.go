// Package gladia provides the Gladia speech-to-text provider.
package gladia

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// STT is a Gladia speech-to-text engine.
type STT struct {
	zrt.BaseSTT
	// Model is the Gladia recognition model.
	Model string
	// Language is the recognition language hint.
	Language string
	// CodeSwitching re-detects the language on every utterance when true; otherwise the language is detected once and held for the session.
	CodeSwitching bool
	// InputSampleRate is the sample rate (Hz) of the incoming audio before any resampling.
	InputSampleRate int
	// OutputSampleRate is the sample rate (Hz) forwarded to the Gladia WebSocket session.
	OutputSampleRate int
	// Encoding is the PCM encoding format sent over the WebSocket (e.g. "wav/pcm", "wav/alaw", "wav/ulaw").
	Encoding string
	// BitDepth is the bit depth of the PCM samples (8, 16, 24, or 32).
	BitDepth int
	// Channels is the number of audio channels.
	Channels int
	// ReceivePartialTranscripts emits partial (non-final) transcripts before an utterance is complete when true.
	ReceivePartialTranscripts bool
}

// STTOptions configures a Gladia STT.
type STTOptions struct {
	// APIKey overrides the GLADIA_API_KEY environment variable.
	APIKey string
	// Model selects the recognition model. Defaults to "solaria-1".
	Model string
	// Languages lists the recognition languages; the first entry is used.
	// Defaults to "english".
	Languages []string
	// CodeSwitching re-detects the language on every utterance when true; otherwise the language is detected once and held. nil = provider default (true).
	CodeSwitching *bool
	// InputSampleRate is the sample rate (Hz) of the incoming audio before any resampling. Defaults to 48000.
	InputSampleRate int
	// OutputSampleRate is the sample rate (Hz) forwarded to the Gladia session. Defaults to 16000.
	OutputSampleRate int
	// Encoding is the PCM encoding format sent over the WebSocket. Defaults to "wav/pcm".
	Encoding string
	// BitDepth is the bit depth of the PCM samples. Defaults to 16.
	BitDepth int
	// Channels is the number of audio channels. Defaults to 1 (mono).
	Channels int
	// ReceivePartialTranscripts emits partial (non-final) transcripts before an utterance is complete. nil = provider default (false).
	ReceivePartialTranscripts *bool
}

// NewSTT returns a Gladia STT configured from opts.
func NewSTT(opts STTOptions) *STT {
	lang := "english"
	if len(opts.Languages) > 0 {
		lang = opts.Languages[0]
	}
	s := &STT{
		Model:                     zrt.StrOr(opts.Model, "solaria-1"),
		Language:                  lang,
		CodeSwitching:             zrt.BoolOr(opts.CodeSwitching, true),
		InputSampleRate:           zrt.IntZeroOr(opts.InputSampleRate, 48000),
		OutputSampleRate:          zrt.IntZeroOr(opts.OutputSampleRate, 16000),
		Encoding:                  zrt.StrOr(opts.Encoding, "wav/pcm"),
		BitDepth:                  zrt.IntZeroOr(opts.BitDepth, 16),
		Channels:                  zrt.IntZeroOr(opts.Channels, 1),
		ReceivePartialTranscripts: zrt.BoolOr(opts.ReceivePartialTranscripts, false),
	}
	s.Init("gladia", zrt.APIKeyOr(opts.APIKey, "GLADIA_API_KEY"))
	return s
}

// STTConfig returns the provider, model, and language for this engine.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "gladia", Model: s.Model, Language: s.Language}
}
