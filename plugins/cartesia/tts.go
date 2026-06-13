// Package cartesia provides the Cartesia text-to-speech provider.
package cartesia

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

const defaultVoice = "f8f5f1b2-f02d-4d8e-a40d-fd850a487b3d"

// TTS is the Cartesia text-to-speech descriptor.
type TTS struct {
	zrt.BaseTTS
	Voice                string
	voiceEmbedding       []float64
	Model                string
	Language             string
	Speed                *float64
	Volume               *float64
	Emotion              string
	PronunciationDictID  string
	MaxBufferDelayMS     *int
	EnableWordTimestamps bool
}

// TTSOptions configures TTS. Provide either Voice (a voice id)
// or VoiceEmbedding (a raw embedding); when an embedding is given, Voice is ignored.
type TTSOptions struct {
	// APIKey overrides the CARTESIA_API_KEY environment variable.
	APIKey               string
	Voice                string // default voice id
	VoiceEmbedding       []float64
	Model                string // default "sonic-2"
	Language             string // default "en"
	Speed                *float64
	Volume               *float64
	Emotion              string
	PronunciationDictID  string
	MaxBufferDelayMS     *int
	EnableWordTimestamps bool
}

// NewTTS builds a TTS.
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

// VoiceEmbedding returns the configured raw voice embedding (or nil).
func (t *TTS) VoiceEmbedding() []float64 { return t.voiceEmbedding }

// Knobs implements the credential knob source.
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
