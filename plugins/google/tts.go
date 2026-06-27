package google

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the Google text-to-speech provider.
type TTS struct {
	zrt.BaseTTS
	Voice        string
	LanguageCode string
	Model        string
	SpeakingRate *float64
	Pitch        *float64
}

// TTSOptions configures a Google TTS instance.
type TTSOptions struct {
	// APIKey overrides the GOOGLE_API_KEY environment variable.
	APIKey string
	// Voice is the Google voice. Defaults to "en-US-Neural2-F".
	Voice string
	// LanguageCode is the BCP-47 language code. Defaults to "en-US".
	LanguageCode string
	// Model is the Google model.
	Model string
	// SpeakingRate scales the speaking rate.
	SpeakingRate *float64
	// Pitch shifts the voice pitch.
	Pitch *float64
}

// NewTTS returns a Google TTS configured from opts.
func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{
		Voice:        zrt.StrOr(opts.Voice, "en-US-Neural2-F"),
		LanguageCode: zrt.StrOr(opts.LanguageCode, "en-US"),
		Model:        opts.Model,
		SpeakingRate: opts.SpeakingRate,
		Pitch:        opts.Pitch,
	}
	t.InitTTS("google", zrt.APIKeyOr(opts.APIKey, "GOOGLE_API_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "google", Voice: t.Voice}
}

// Knobs returns the provider-specific TTS settings.
func (t *TTS) Knobs() map[string]any {
	k := map[string]any{"language": t.LanguageCode, "voice": t.Voice}
	if t.SpeakingRate != nil {
		k["speaking_rate"] = *t.SpeakingRate
	}
	if t.Pitch != nil {
		k["pitch"] = *t.Pitch
	}
	return k
}
