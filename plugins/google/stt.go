package google

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

// STT is the Google Cloud Speech-to-Text provider.
type STT struct {
	zrt.BaseSTT
	// Model is the recognition model ID, e.g. "latest_long".
	Model string
	// Language is the BCP-47 recognition language, e.g. "en-US". Multiple codes are comma-separated.
	Language string
	// ServiceAccountJSON holds the resolved service-account credentials as a JSON string.
	ServiceAccountJSON string
	// ProjectID is the Google Cloud project ID.
	ProjectID string
	// Stream transcribes audio over a low-latency streaming connection when true.
	Stream bool
	// Location is the Cloud Speech-to-Text regional endpoint, e.g. "us" or "eu".
	Location string
	// AudioChannelCount is the number of audio channels in the input.
	AudioChannelCount int
	// InterimResults emits partial transcripts before an utterance is finalized when true.
	InterimResults bool
	// Punctuate adds automatic punctuation to transcripts when true.
	Punctuate bool
	// ProfanityFilter replaces profane words with placeholder characters when true.
	ProfanityFilter bool
	// EnableSpokenPunctuation transcribes spoken punctuation cues as punctuation marks when true.
	EnableSpokenPunctuation bool
	// EnableSpokenEmojis transcribes spoken emoji names as emoji characters when true.
	EnableSpokenEmojis bool
	// EnableWordTimeOffsets returns start and end timestamps for each recognized word when true.
	EnableWordTimeOffsets bool
	// EnableWordConfidence includes per-word confidence scores in results when true.
	EnableWordConfidence bool
	// MaxAlternatives is the maximum number of alternative transcriptions to return.
	MaxAlternatives int
	// EnableVoiceActivityEvents emits speech-start and speech-end events independently of transcripts when true.
	EnableVoiceActivityEvents bool
	// SpeechStartTimeout is the seconds to wait for speech to begin before a voice-activity timeout. nil uses the API default.
	SpeechStartTimeout *float64
	// SpeechEndTimeout is the seconds of silence after speech before an end-of-speech event fires. nil uses the API default.
	SpeechEndTimeout *float64
	// MinSpeakerCount is the minimum expected number of speakers for diarization. nil disables diarization.
	MinSpeakerCount *int
	// MaxSpeakerCount is the maximum expected number of speakers for diarization. nil disables diarization.
	MaxSpeakerCount *int
	// MinConfidenceThreshold is a client-side filter dropping transcripts below this confidence.
	MinConfidenceThreshold float64
	// SampleRate is the input audio sample rate in Hz.
	SampleRate int
}

// STTOptions configures a Google Speech-to-Text instance. Credentials are
// resolved in priority order: CredentialsJSON, ServiceAccountPath, APIKey, the
// GOOGLE_APPLICATION_CREDENTIALS env var, then a credentials.json file in the
// working directory.
type STTOptions struct {
	// APIKey is the Google API key, used as a fallback credential source. If empty, the GOOGLE_API_KEY environment variable is used.
	APIKey string
	// CredentialsJSON holds service-account credentials as a raw JSON string.
	CredentialsJSON string
	// ServiceAccountPath is the path to a service-account JSON key file.
	ServiceAccountPath string
	// ProjectID is the Google Cloud project ID. Extracted from the service-account JSON when unset.
	ProjectID string
	// Languages is one or more BCP-47 recognition language codes; multiple codes enable multi-language detection.
	Languages []string
	// Language is a single BCP-47 recognition language code. Defaults to "en-US".
	Language string
	// Model is the recognition model ID. Defaults to "latest_long".
	Model string
	// Stream transcribes over a low-latency streaming connection. nil defaults to true.
	Stream *bool
	// Location is the Cloud Speech-to-Text regional endpoint. Defaults to "us".
	Location string
	// AudioChannelCount is the number of audio channels in the input. Defaults to 1.
	AudioChannelCount int
	// InterimResults emits partial transcripts before finalization. nil defaults to true.
	InterimResults *bool
	// Punctuate adds automatic punctuation to transcripts. nil defaults to true.
	Punctuate *bool
	// ProfanityFilter replaces profane words with placeholder characters. nil defaults to false.
	ProfanityFilter *bool
	// EnableSpokenPunctuation transcribes spoken punctuation cues as marks. nil defaults to false.
	EnableSpokenPunctuation *bool
	// EnableSpokenEmojis transcribes spoken emoji names as emoji. nil defaults to false.
	EnableSpokenEmojis *bool
	// EnableWordTimeOffsets returns per-word timestamps. nil defaults to false.
	EnableWordTimeOffsets *bool
	// EnableWordConfidence includes per-word confidence scores. nil defaults to false.
	EnableWordConfidence *bool
	// MaxAlternatives is the maximum number of alternative transcriptions. Defaults to 1.
	MaxAlternatives int
	// EnableVoiceActivityEvents emits speech-start/end events. nil defaults to false.
	EnableVoiceActivityEvents *bool
	// SpeechStartTimeout is the seconds to wait for speech to begin. nil uses the API default.
	SpeechStartTimeout *float64
	// SpeechEndTimeout is the seconds of silence after speech before end-of-speech. nil uses the API default.
	SpeechEndTimeout *float64
	// MinSpeakerCount is the minimum expected number of speakers for diarization. nil disables diarization.
	MinSpeakerCount *int
	// MaxSpeakerCount is the maximum expected number of speakers for diarization. nil disables diarization.
	MaxSpeakerCount *int
	// MinConfidenceThreshold is a client-side filter dropping transcripts below this confidence. nil defaults to 0.0.
	MinConfidenceThreshold *float64
	// SampleRate is the input audio sample rate in Hz. Defaults to 48000.
	SampleRate int
}

// NewSTT creates a Google Speech-to-Text provider from opts, applying defaults
// for any unset fields and resolving service-account credentials and the project
// ID.
func NewSTT(opts STTOptions) *STT {
	lang := opts.Language
	if len(opts.Languages) > 0 {
		lang = strings.Join(opts.Languages, ",")
	}
	lang = zrt.StrOr(lang, "en-US")

	saJSON := resolveServiceAccountJSON(opts.CredentialsJSON, opts.ServiceAccountPath, opts.APIKey)
	projectID := zrt.StrOr(opts.ProjectID, projectIDFromJSON(saJSON))

	s := &STT{
		Model:                     zrt.StrOr(opts.Model, "latest_long"),
		Language:                  lang,
		ServiceAccountJSON:        saJSON,
		ProjectID:                 projectID,
		Stream:                    zrt.BoolOr(opts.Stream, true),
		Location:                  zrt.StrOr(opts.Location, "us"),
		AudioChannelCount:         zrt.IntZeroOr(opts.AudioChannelCount, 1),
		InterimResults:            zrt.BoolOr(opts.InterimResults, true),
		Punctuate:                 zrt.BoolOr(opts.Punctuate, true),
		ProfanityFilter:           zrt.BoolOr(opts.ProfanityFilter, false),
		EnableSpokenPunctuation:   zrt.BoolOr(opts.EnableSpokenPunctuation, false),
		EnableSpokenEmojis:        zrt.BoolOr(opts.EnableSpokenEmojis, false),
		EnableWordTimeOffsets:     zrt.BoolOr(opts.EnableWordTimeOffsets, false),
		EnableWordConfidence:      zrt.BoolOr(opts.EnableWordConfidence, false),
		MaxAlternatives:           zrt.IntZeroOr(opts.MaxAlternatives, 1),
		EnableVoiceActivityEvents: zrt.BoolOr(opts.EnableVoiceActivityEvents, false),
		SpeechStartTimeout:        opts.SpeechStartTimeout,
		SpeechEndTimeout:          opts.SpeechEndTimeout,
		MinSpeakerCount:           opts.MinSpeakerCount,
		MaxSpeakerCount:           opts.MaxSpeakerCount,
		MinConfidenceThreshold:    zrt.FloatOr(opts.MinConfidenceThreshold, 0.0),
		SampleRate:                zrt.IntZeroOr(opts.SampleRate, 48000),
	}
	s.Init("google_stt", zrt.APIKeyOr(opts.APIKey, "GOOGLE_API_KEY"))
	return s
}

// STTConfig implements zrt.STT and reports the model configuration.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "google_stt", Model: s.Model, Language: s.Language}
}

// Knobs returns the provider-specific STT settings that are set.
func (s *STT) Knobs() map[string]any {
	k := map[string]any{
		"location":                     s.Location,
		"stream":                       s.Stream,
		"audio_channel_count":          s.AudioChannelCount,
		"interim_results":              s.InterimResults,
		"punctuate":                    s.Punctuate,
		"profanity_filter":             s.ProfanityFilter,
		"enable_spoken_punctuation":    s.EnableSpokenPunctuation,
		"enable_spoken_emojis":         s.EnableSpokenEmojis,
		"enable_word_time_offsets":     s.EnableWordTimeOffsets,
		"enable_word_confidence":       s.EnableWordConfidence,
		"max_alternatives":             s.MaxAlternatives,
		"enable_voice_activity_events": s.EnableVoiceActivityEvents,
		"min_confidence_threshold":     s.MinConfidenceThreshold,
	}
	if s.ServiceAccountJSON != "" {
		k["service_account_json"] = s.ServiceAccountJSON
	}
	if s.ProjectID != "" {
		k["project_id"] = s.ProjectID
	}
	if s.SpeechStartTimeout != nil {
		k["speech_start_timeout"] = *s.SpeechStartTimeout
	}
	if s.SpeechEndTimeout != nil {
		k["speech_end_timeout"] = *s.SpeechEndTimeout
	}
	if s.MinSpeakerCount != nil {
		k["min_speaker_count"] = *s.MinSpeakerCount
	}
	if s.MaxSpeakerCount != nil {
		k["max_speaker_count"] = *s.MaxSpeakerCount
	}
	return k
}

func coerceToJSONContent(value string) string {
	if value == "" {
		return ""
	}
	if strings.HasPrefix(strings.TrimSpace(value), "{") {
		return value
	}
	b, err := os.ReadFile(value)
	if err != nil {
		return ""
	}
	return string(b)
}

func resolveServiceAccountJSON(credentialsJSON, serviceAccountPath, apiKey string) string {
	for _, candidate := range []string{credentialsJSON, serviceAccountPath, apiKey, os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"), "credentials.json"} {
		if content := coerceToJSONContent(candidate); content != "" {
			return content
		}
	}
	return ""
}

func projectIDFromJSON(saJSON string) string {
	if saJSON == "" {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(saJSON), &m); err != nil {
		return ""
	}
	if v, ok := m["project_id"].(string); ok {
		return v
	}
	return ""
}
