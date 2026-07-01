package navana

import (
	"os"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

type STT struct {
	zrt.BaseSTT
	CustomerID string
	Model      string
	SampleRate int
}

type STTOptions struct {
	APIKey     string
	CustomerID string
	Model      string
	SampleRate int
}

func NewSTT(opts STTOptions) *STT {
	s := &STT{
		Model:      zrt.StrOr(opts.Model, "hi-general-v2-8khz"),
		SampleRate: zrt.IntZeroOr(opts.SampleRate, 8000),
		CustomerID: zrt.StrOr(opts.CustomerID, os.Getenv("NAVANA_CUSTOMER_ID")),
	}
	s.Init("navana", zrt.APIKeyOr(opts.APIKey, "NAVANA_API_KEY"))
	return s
}

func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "navana", Model: s.Model}
}

func (s *STT) Knobs() map[string]any {
	if s.CustomerID != "" {
		return map[string]any{"customer_id": s.CustomerID}
	}
	return map[string]any{}
}
