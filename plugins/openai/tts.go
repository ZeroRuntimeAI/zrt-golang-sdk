package openai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the OpenAI text-to-speech provider.
type TTS struct {
	zrt.BaseTTS
	// Voice is the voice id. Defaults to "ash".
	Voice string
	// Model is the TTS model id. Defaults to "gpt-4o-mini-tts".
	Model string
	// Speed is the speech-rate multiplier; nil = provider default.
	Speed *float64
	// Stream enables streamed audio. Defaults to true.
	Stream bool
}

// TTSOptions configures NewTTS.
type TTSOptions struct {
	// APIKey is the OpenAI API key; falls back to OPENAI_API_KEY.
	APIKey string
	// Voice is the voice id. Defaults to "ash".
	Voice string
	// Model is the TTS model id. Defaults to "gpt-4o-mini-tts".
	Model string
	// SampleRate is the output audio sample rate in Hz. Defaults to 24000.
	SampleRate int
	// Speed is the speech-rate multiplier; nil = provider default.
	Speed *float64
	// Stream enables streamed audio; nil applies true.
	Stream *bool
}

// NewTTS builds a TTS from opts, applying defaults and resolving the API key.
func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{
		Voice:  zrt.StrOr(opts.Voice, "ash"),
		Model:  zrt.StrOr(opts.Model, "gpt-4o-mini-tts"),
		Speed:  opts.Speed,
		Stream: zrt.BoolOr(opts.Stream, true),
	}
	t.InitTTS("openai", zrt.APIKeyOr(opts.APIKey, "OPENAI_API_KEY"), zrt.IntZeroOr(opts.SampleRate, 24000))
	return t
}

// TTSConfig returns the runtime configuration for this provider.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "openai", Model: t.Model, Voice: t.Voice}
}

// Knobs returns the set of provider parameters to pass to the runtime.
func (t *TTS) Knobs() map[string]any {
	k := map[string]any{"tts_stream": t.Stream}
	if t.Speed != nil {
		k["speed"] = *t.Speed
	}
	return k
}
