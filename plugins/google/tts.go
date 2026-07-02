package google

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the Google text-to-speech provider.
type TTS struct {
	zrt.BaseTTS
	// Voice is the Google voice name.
	Voice string
	// LanguageCode is the BCP-47 language code.
	LanguageCode string
	// Model is the Google TTS model.
	Model string
	// SpeakingRate scales the speaking rate. nil uses the provider default.
	SpeakingRate *float64
	// Pitch shifts the voice pitch. nil uses the provider default.
	Pitch *float64
	// CredentialsJSON holds service-account credentials as a raw JSON string.
	CredentialsJSON string
	// ServiceAccountPath is the path to a service-account JSON key file.
	ServiceAccountPath string
	// ServiceAccountJSON holds the resolved service-account credentials as a JSON string.
	ServiceAccountJSON string
	// Stream synthesizes audio over a low-latency streaming connection when true.
	Stream bool
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
	// CredentialsJSON holds service-account credentials as a raw JSON string.
	CredentialsJSON string
	// ServiceAccountPath is the path to a service-account JSON key file.
	ServiceAccountPath string
	// Stream synthesizes audio over a low-latency streaming connection. nil defaults to true.
	Stream *bool
}

// NewTTS returns a Google TTS configured from opts.
func NewTTS(opts TTSOptions) *TTS {
	saJSON := resolveServiceAccountJSON(opts.CredentialsJSON, opts.ServiceAccountPath, opts.APIKey)
	t := &TTS{
		Voice:              zrt.StrOr(opts.Voice, "en-US-Neural2-F"),
		LanguageCode:       zrt.StrOr(opts.LanguageCode, "en-US"),
		Model:              opts.Model,
		SpeakingRate:       opts.SpeakingRate,
		Pitch:              opts.Pitch,
		CredentialsJSON:    opts.CredentialsJSON,
		ServiceAccountPath: opts.ServiceAccountPath,
		ServiceAccountJSON: saJSON,
		Stream:             zrt.BoolOr(opts.Stream, true),
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
	k := map[string]any{"language": t.LanguageCode, "voice": t.Voice, "tts_stream": t.Stream}
	if t.SpeakingRate != nil {
		k["speaking_rate"] = *t.SpeakingRate
	}
	if t.Pitch != nil {
		k["pitch"] = *t.Pitch
	}
	if t.ServiceAccountJSON != "" {
		k["tts_service_account_json"] = t.ServiceAccountJSON
	}
	return k
}
