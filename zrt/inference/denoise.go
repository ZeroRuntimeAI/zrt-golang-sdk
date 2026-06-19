package inference

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

type DenoiseAicousticsOptions struct {
	ModelID      string
	SampleRate   int
	ChunkMS      int
	GatewayToken string
	BaseURL      string
}

func AICousticsDenoise(o DenoiseAicousticsOptions) *zrt.Denoise {
	return zrt.DenoiseAicoustics(o.ModelID, o.SampleRate, o.ChunkMS, o.GatewayToken, o.BaseURL)
}

type DenoiseSanasOptions struct {
	ModelID      string
	SampleRate   int
	ChunkMS      int
	GatewayToken string
	BaseURL      string
}

func SanasDenoise(o DenoiseSanasOptions) *zrt.Denoise {
	return zrt.DenoiseSanas(o.ModelID, o.SampleRate, o.ChunkMS, o.GatewayToken, o.BaseURL)
}
