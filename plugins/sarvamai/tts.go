package sarvamai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is a Sarvam AI text-to-speech engine.
type TTS struct {
	zrt.BaseTTS
	// Voice selects the speaker. Defaults to "shubh".
	Voice string
	// Model selects the model. Defaults to "bulbul:v3".
	Model string
	// Language is the language code. Defaults to "en-IN".
	Language string
	// Streaming enables streaming synthesis. Defaults to true.
	Streaming bool
	// Pitch adjusts the voice pitch. Defaults to 0.0.
	Pitch float64
	// Pace adjusts the speaking rate. Defaults to 1.0.
	Pace float64
	// Loudness adjusts the output loudness. Defaults to 1.0.
	Loudness float64
	// Temperature controls synthesis variability. Defaults to 0.6.
	Temperature float64
	// Preprocessing enables text preprocessing. Defaults to false.
	Preprocessing bool
	// Bitrate is the output bitrate. Defaults to "128k".
	Bitrate string
	// OutputAudioCodec is the output audio codec. Defaults to "linear16".
	OutputAudioCodec string
	// MinBufferSize is the minimum buffer size before synthesis begins. Defaults to 50.
	MinBufferSize int
	// MaxChunkLength is the maximum text chunk length per synthesis request. Defaults to 150.
	MaxChunkLength int
}

// TTSOptions configures a Sarvam AI TTS engine. Pointer fields left nil
// fall back to their default values.
type TTSOptions struct {
	// APIKey is the Sarvam AI API key. If empty, the SARVAM_API_KEY environment variable is used.
	APIKey string
	// Model selects the model. Defaults to "bulbul:v3".
	Model string
	// Language is the language code. Defaults to "en-IN".
	Language string
	// Speaker selects the voice. Defaults to "shubh".
	Speaker string
	// Streaming enables streaming synthesis. Defaults to true.
	Streaming *bool
	// Pitch adjusts the voice pitch. Defaults to 0.0.
	Pitch *float64
	// Pace adjusts the speaking rate. Defaults to 1.0.
	Pace *float64
	// Loudness adjusts the output loudness. Defaults to 1.0.
	Loudness *float64
	// Temperature controls synthesis variability. Defaults to 0.6.
	Temperature *float64
	// Preprocessing enables text preprocessing. Defaults to false.
	Preprocessing *bool
	// Bitrate is the output bitrate. Defaults to "128k".
	Bitrate string
	// OutputAudioCodec is the output audio codec. Defaults to "linear16".
	OutputAudioCodec string
	// MinBufferSize is the minimum buffer size before synthesis begins. Defaults to 50.
	MinBufferSize int
	// MaxChunkLength is the maximum text chunk length per synthesis request. Defaults to 150.
	MaxChunkLength int
}

// NewTTS creates a Sarvam AI TTS engine from the given options.
func NewTTS(opts TTSOptions) *TTS {
	minBuf := opts.MinBufferSize
	if minBuf == 0 {
		minBuf = 50
	}
	maxChunk := opts.MaxChunkLength
	if maxChunk == 0 {
		maxChunk = 150
	}
	t := &TTS{
		Voice:            zrt.StrOr(opts.Speaker, "shubh"),
		Model:            zrt.StrOr(opts.Model, "bulbul:v3"),
		Language:         zrt.StrOr(opts.Language, "en-IN"),
		Streaming:        zrt.BoolOr(opts.Streaming, true),
		Pitch:            zrt.FloatOr(opts.Pitch, 0.0),
		Pace:             zrt.FloatOr(opts.Pace, 1.0),
		Loudness:         zrt.FloatOr(opts.Loudness, 1.0),
		Temperature:      zrt.FloatOr(opts.Temperature, 0.6),
		Preprocessing:    zrt.BoolOr(opts.Preprocessing, false),
		Bitrate:          zrt.StrOr(opts.Bitrate, "128k"),
		OutputAudioCodec: zrt.StrOr(opts.OutputAudioCodec, "linear16"),
		MinBufferSize:    minBuf,
		MaxChunkLength:   maxChunk,
	}
	t.InitTTS("sarvamai", zrt.APIKeyOr(opts.APIKey, "SARVAM_API_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "sarvamai", Model: t.Model, Language: t.Language, Voice: t.Voice}
}

// Knobs implements the credential knob source.
func (t *TTS) Knobs() map[string]any {
	return map[string]any{
		"model":         t.Model,
		"language":      t.Language,
		"streaming":     t.Streaming,
		"pitch":         t.Pitch,
		"pace":          t.Pace,
		"loudness":      t.Loudness,
		"temperature":   t.Temperature,
		"preprocessing": t.Preprocessing,
		"bitrate":       t.Bitrate,
	}
}
