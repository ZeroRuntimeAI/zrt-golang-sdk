package zrt

import (
	"cmp"
	"os"
)

// Denoise is a noise-reduction descriptor.
type Denoise struct {
	providerName string
	// ModelID is the provider model identifier.
	ModelID string
	// ModelSampleRate is the sample rate (Hz) the model expects.
	ModelSampleRate int
	// ChunkMS is the audio chunk size in milliseconds.
	ChunkMS int
	// GatewayToken is the auth token for the denoise gateway.
	GatewayToken string
	// BaseURL overrides the denoise gateway base URL.
	BaseURL string

	hasModelSampleRate bool
	hasChunkMS         bool
}

// DenoiseOptions configures a Denoise descriptor.
type DenoiseOptions struct {
	// Provider is the denoise provider name (e.g. "rnnoise", "sanas", "aicoustics").
	Provider string
	// ModelID is the provider model identifier.
	ModelID string
	// ModelSampleRate is the sample rate (Hz) the model expects; nil leaves it unset.
	ModelSampleRate *int
	// ChunkMS is the audio chunk size in milliseconds; nil leaves it unset.
	ChunkMS *int
	// GatewayToken is the auth token for the denoise gateway. Defaults to the
	// ZRT_AUTH_TOKEN environment variable.
	GatewayToken string
	// BaseURL overrides the denoise gateway base URL.
	BaseURL string
}

// NewDenoise builds a Denoise descriptor.
func NewDenoise(opts DenoiseOptions) *Denoise {
	gw := cmp.Or(opts.GatewayToken, os.Getenv("ZRT_AUTH_TOKEN"))
	d := &Denoise{
		providerName: opts.Provider,
		ModelID:      opts.ModelID,
		GatewayToken: gw,
		BaseURL:      opts.BaseURL,
	}
	if opts.ModelSampleRate != nil {
		d.ModelSampleRate = *opts.ModelSampleRate
		d.hasModelSampleRate = true
	}
	if opts.ChunkMS != nil {
		d.ChunkMS = *opts.ChunkMS
		d.hasChunkMS = true
	}
	return d
}

// ProviderName returns the denoise provider name.
func (d *Denoise) ProviderName() string { return d.providerName }

// DenoiseRNNoise builds an RNNoise denoiser (provider="rnnoise").
func DenoiseRNNoise() *Denoise { return NewDenoise(DenoiseOptions{Provider: "rnnoise"}) }

// DenoiseSanas builds a Sanas denoiser.
func DenoiseSanas(modelID string, sampleRate, chunkMS int, gatewayToken, baseURL string) *Denoise {
	modelID = cmp.Or(modelID, "VI_G_NC3.0")
	sampleRate = cmp.Or(sampleRate, 16000)
	chunkMS = cmp.Or(chunkMS, 20)
	return NewDenoise(DenoiseOptions{Provider: "sanas", ModelID: modelID, ModelSampleRate: &sampleRate, ChunkMS: &chunkMS, GatewayToken: gatewayToken, BaseURL: baseURL})
}

// DenoiseAicoustics builds an ai-coustics denoiser.
func DenoiseAicoustics(modelID string, sampleRate, chunkMS int, gatewayToken, baseURL string) *Denoise {
	if modelID == "" {
		modelID = "rook-l-48khz"
	}
	if sampleRate == 0 {
		sampleRate = 48000
	}
	if chunkMS == 0 {
		chunkMS = 10
	}
	return NewDenoise(DenoiseOptions{Provider: "aicoustics", ModelID: modelID, ModelSampleRate: &sampleRate, ChunkMS: &chunkMS, GatewayToken: gatewayToken, BaseURL: baseURL})
}
