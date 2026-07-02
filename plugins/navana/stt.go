package navana

import (
	"os"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

// STT is the Navana speech-to-text provider.
type STT struct {
	zrt.BaseSTT
	// CustomerID is the Navana customer identifier sent with requests.
	CustomerID string
	// Model is the transcription model id.
	Model string
	// SampleRate is the input audio sample rate in Hz.
	SampleRate int
}

// STTOptions configures NewSTT.
type STTOptions struct {
	// APIKey is the Navana API key. If empty, the NAVANA_API_KEY environment variable is used.
	APIKey string
	// CustomerID is the Navana customer identifier. If empty, the NAVANA_CUSTOMER_ID environment variable is used.
	CustomerID string
	// Model is the transcription model id. Defaults to "hi-general-v2-8khz".
	Model string
	// SampleRate is the input audio sample rate in Hz. Defaults to 8000.
	SampleRate int
}

// NewSTT builds an STT from opts, applying defaults for unset fields.
func NewSTT(opts STTOptions) *STT {
	s := &STT{
		Model:      zrt.StrOr(opts.Model, "hi-general-v2-8khz"),
		SampleRate: zrt.IntZeroOr(opts.SampleRate, 8000),
		CustomerID: zrt.StrOr(opts.CustomerID, os.Getenv("NAVANA_CUSTOMER_ID")),
	}
	s.Init("navana", zrt.APIKeyOr(opts.APIKey, "NAVANA_API_KEY"))
	return s
}

// STTConfig implements zrt.STT.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "navana", Model: s.Model}
}

// Knobs returns provider-specific runtime settings, including the customer_id when set.
func (s *STT) Knobs() map[string]any {
	if s.CustomerID != "" {
		return map[string]any{"customer_id": s.CustomerID}
	}
	return map[string]any{}
}
