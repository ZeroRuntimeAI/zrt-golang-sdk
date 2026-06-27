// Package azure provides the Azure STT, TTS and Voice Live providers.
package azure

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// STT is an Azure speech-to-text engine.
type STT struct {
	zrt.BaseSTT
	Model        string
	Language     string
	SpeechRegion string
}

// STTOptions configures an Azure STT.
type STTOptions struct {
	// SpeechKey overrides the AZURE_SPEECH_KEY environment variable.
	SpeechKey string
	// SpeechRegion selects the service region. Defaults to the AZURE_REGION
	// environment variable, or "eastus".
	SpeechRegion string
	// Language is the recognition language. Defaults to "en-US".
	Language string
}

// NewSTT returns an Azure STT configured from opts.
func NewSTT(opts STTOptions) *STT {
	s := &STT{
		Model:        "",
		Language:     zrt.StrOr(opts.Language, "en-US"),
		SpeechRegion: zrt.StrOr(opts.SpeechRegion, zrt.EnvOr("AZURE_REGION", "eastus")),
	}
	s.Init("azure", zrt.APIKeyOr(opts.SpeechKey, "AZURE_SPEECH_KEY"))
	return s
}

// STTConfig returns the provider, model, and language for this engine.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "azure", Model: s.Model, Language: s.Language}
}
