// Package rnnoise provides the RNNoise denoiser.
package rnnoise

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// New builds an RNNoise denoiser.
func New() *zrt.Denoise {
	return zrt.NewDenoise(zrt.DenoiseOptions{Provider: ""})
}
