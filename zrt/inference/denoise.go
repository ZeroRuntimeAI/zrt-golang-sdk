package inference

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// DenoiseAicousticsOptions configures an ai-coustics denoiser.
type DenoiseAicousticsOptions struct {
	// ModelID is the ai-coustics model id. Defaults to "rook-l-48khz".
	ModelID string
	// SampleRate is the model sample rate in Hz. Defaults to 48000.
	SampleRate int
	// ChunkMS is the audio chunk size in milliseconds. Defaults to 10.
	ChunkMS int
	// GatewayToken authenticates against the inference gateway.
	GatewayToken string
	// BaseURL overrides the inference gateway endpoint.
	BaseURL string
}

// AICousticsDenoise builds an ai-coustics denoiser from the given options.
func AICousticsDenoise(o DenoiseAicousticsOptions) *zrt.Denoise {
	return zrt.DenoiseAicoustics(o.ModelID, o.SampleRate, o.ChunkMS, o.GatewayToken, o.BaseURL)
}

// DenoiseSanasOptions configures a Sanas denoiser.
type DenoiseSanasOptions struct {
	// ModelID is the Sanas model id. Defaults to "VI_G_NC3.0".
	ModelID string
	// SampleRate is the model sample rate in Hz. Defaults to 16000.
	SampleRate int
	// ChunkMS is the audio chunk size in milliseconds. Defaults to 20.
	ChunkMS int
	// GatewayToken authenticates against the inference gateway.
	GatewayToken string
	// BaseURL overrides the inference gateway endpoint.
	BaseURL string
}

// SanasDenoise builds a Sanas denoiser from the given options.
func SanasDenoise(o DenoiseSanasOptions) *zrt.Denoise {
	return zrt.DenoiseSanas(o.ModelID, o.SampleRate, o.ChunkMS, o.GatewayToken, o.BaseURL)
}
