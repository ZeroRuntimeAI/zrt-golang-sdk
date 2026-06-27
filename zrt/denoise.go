package zrt

import (
	"cmp"
	"os"
)

// Denoise is a noise-reduction descriptor.
type Denoise struct {
	providerName    string
	ModelID         string
	ModelSampleRate int
	ChunkMS         int
	GatewayToken    string
	BaseURL         string

	hasModelSampleRate bool
	hasChunkMS         bool
}

// DenoiseOptions configures a Denoise descriptor.
type DenoiseOptions struct {
	Provider        string
	ModelID         string
	ModelSampleRate *int
	ChunkMS         *int
	GatewayToken    string
	BaseURL         string
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
	modelID = cmp.Or(modelID, "sparrow-xxs-48khz")
	sampleRate = cmp.Or(sampleRate, 48000)
	chunkMS = cmp.Or(chunkMS, 10)
	return NewDenoise(DenoiseOptions{Provider: "aicoustics", ModelID: modelID, ModelSampleRate: &sampleRate, ChunkMS: &chunkMS, GatewayToken: gatewayToken, BaseURL: baseURL})
}
