package deepgram

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is a Deepgram Aura-2 streaming text-to-speech engine.
type TTS struct {
	zrt.BaseTTS
	// Model is the Aura-2 voice id.
	Model string
	// Voice is the Aura voice id used for synthesis. Defaults to Model.
	Voice string
	// Stream reports whether audio is streamed as it is synthesized.
	Stream bool
	// Encoding is the output audio encoding.
	Encoding string
	// BaseURL is the Deepgram Speak WebSocket endpoint.
	BaseURL string
}

// TTSOptions configures a Deepgram TTS. Nil pointer fields fall back to their defaults.
type TTSOptions struct {
	// APIKey overrides the DEEPGRAM_API_KEY environment variable.
	APIKey string
	// Model is the Aura-2 voice id, e.g. "aura-2-thalia-en". Defaults to "aura-2-andromeda-en".
	Model string
	// Voice overrides Model as the Aura voice id. Defaults to Model.
	Voice string
	// Language is accepted for forward compatibility and is not currently applied.
	Language string
	// Stream streams audio as it is synthesized. Defaults to true.
	Stream *bool
	// Encoding is the output audio encoding. Defaults to "linear16".
	Encoding string
	// BaseURL is the Deepgram Speak WebSocket endpoint. Defaults to "wss://api.deepgram.com/v1/speak".
	BaseURL string
}

// NewTTS returns a Deepgram TTS configured from opts.
func NewTTS(opts TTSOptions) *TTS {
	model := zrt.StrOr(opts.Model, "aura-2-andromeda-en")
	t := &TTS{
		Model:    model,
		Voice:    zrt.StrOr(opts.Voice, model),
		Stream:   zrt.BoolOr(opts.Stream, true),
		Encoding: zrt.StrOr(opts.Encoding, "linear16"),
		BaseURL:  zrt.StrOr(opts.BaseURL, "wss://api.deepgram.com/v1/speak"),
	}
	t.InitTTS("deepgram", zrt.APIKeyOr(opts.APIKey, "DEEPGRAM_API_KEY"), 24000)
	return t
}

// TTSConfig returns the provider, model, and voice for this engine.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "deepgram", Model: t.Model, Voice: t.Voice}
}

// Knobs returns the Deepgram-specific options as a key/value map.
func (t *TTS) Knobs() map[string]any {
	return map[string]any{
		"tts_stream":   t.Stream,
		"tts_encoding": t.Encoding,
		"base_url":     t.BaseURL,
	}
}
