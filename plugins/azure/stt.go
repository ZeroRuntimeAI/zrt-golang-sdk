// Package azure provides the Azure STT, TTS and Voice Live providers.
package azure

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// STT is the Azure speech-to-text descriptor.
type STT struct {
	zrt.BaseSTT
	Model        string
	Language     string
	SpeechRegion string
}

// STTOptions configures STT.
type STTOptions struct {
	SpeechKey    string
	SpeechRegion string // default from AZURE_REGION, else "eastus"
	Language     string // default "en-US"
}

// NewSTT builds an STT.
func NewSTT(opts STTOptions) *STT {
	s := &STT{
		Model:        "",
		Language:     zrt.StrOr(opts.Language, "en-US"),
		SpeechRegion: zrt.StrOr(opts.SpeechRegion, zrt.EnvOr("AZURE_REGION", "eastus")),
	}
	s.Init("azure", zrt.APIKeyOr(opts.SpeechKey, "AZURE_SPEECH_KEY"))
	return s
}

// STTConfig implements zrt.STT.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "azure", Model: s.Model, Language: s.Language}
}
