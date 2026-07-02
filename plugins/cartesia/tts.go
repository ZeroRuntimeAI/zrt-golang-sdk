// Package cartesia provides the Cartesia text-to-speech provider.
package cartesia

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

const defaultVoice = "f8f5f1b2-f02d-4d8e-a40d-fd850a487b3d"

// TTS is the Cartesia text-to-speech provider.
type TTS struct {
	zrt.BaseTTS
	// Voice is the resolved Cartesia voice id ("" when a voice embedding is used).
	Voice          string
	voiceEmbedding []float64
	// Model is the resolved Cartesia model.
	Model string
	// Language is the resolved spoken language.
	Language string
	// Speed scales the speaking rate; nil leaves it unset.
	Speed *float64
	// Volume scales the output volume; nil leaves it unset.
	Volume *float64
	// Emotion sets the emotional style.
	Emotion string
	// PronunciationDictID selects a custom pronunciation dictionary.
	PronunciationDictID string
	// MaxBufferDelayMS caps the synthesis buffering delay, in milliseconds; nil leaves it unset.
	MaxBufferDelayMS *int
	// EnableWordTimestamps requests per-word timing information.
	EnableWordTimestamps bool
}

// TTSOptions configures a Cartesia TTS instance. Provide either Voice (a voice
// id) or VoiceEmbedding (a raw embedding); when an embedding is given, Voice is
// ignored.
type TTSOptions struct {
	// APIKey overrides the CARTESIA_API_KEY environment variable.
	APIKey string
	// Voice is the Cartesia voice id. Defaults to a built-in voice.
	Voice string
	// VoiceEmbedding is a raw voice embedding used in place of Voice.
	VoiceEmbedding []float64
	// Model is the Cartesia model. Defaults to "sonic-2".
	Model string
	// Language is the spoken language. Defaults to "en".
	Language string
	// Speed scales the speaking rate.
	Speed *float64
	// Volume scales the output volume.
	Volume *float64
	// Emotion sets the emotional style.
	Emotion string
	// PronunciationDictID selects a custom pronunciation dictionary.
	PronunciationDictID string
	// MaxBufferDelayMS caps the synthesis buffering delay, in milliseconds.
	MaxBufferDelayMS *int
	// EnableWordTimestamps requests per-word timing information.
	EnableWordTimestamps bool
}

// NewTTS returns a Cartesia TTS configured from opts.
func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{
		Model:                zrt.StrOr(opts.Model, "sonic-2"),
		Language:             zrt.StrOr(opts.Language, "en"),
		Speed:                opts.Speed,
		Volume:               opts.Volume,
		Emotion:              opts.Emotion,
		PronunciationDictID:  opts.PronunciationDictID,
		MaxBufferDelayMS:     opts.MaxBufferDelayMS,
		EnableWordTimestamps: opts.EnableWordTimestamps,
	}
	if len(opts.VoiceEmbedding) > 0 {
		t.voiceEmbedding = append([]float64(nil), opts.VoiceEmbedding...)
		t.Voice = ""
	} else {
		t.Voice = zrt.StrOr(opts.Voice, defaultVoice)
	}
	t.InitTTS("cartesia", zrt.APIKeyOr(opts.APIKey, "CARTESIA_API_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "cartesia", Voice: t.Voice}
}

// VoiceEmbedding returns the configured raw voice embedding, or nil if none was set.
func (t *TTS) VoiceEmbedding() []float64 { return t.voiceEmbedding }

// Knobs returns the provider-specific TTS settings.
func (t *TTS) Knobs() map[string]any {
	k := map[string]any{
		"model":                  t.Model,
		"language":               t.Language,
		"enable_word_timestamps": t.EnableWordTimestamps,
	}
	if t.Speed != nil {
		k["speed"] = *t.Speed
	}
	if t.Volume != nil {
		k["volume"] = *t.Volume
	}
	if t.Emotion != "" {
		k["emotion"] = t.Emotion
	}
	if t.MaxBufferDelayMS != nil {
		k["max_buffer_delay_ms"] = *t.MaxBufferDelayMS
	}
	if t.PronunciationDictID != "" {
		k["pronunciation_dict_id"] = t.PronunciationDictID
	}
	return k
}
