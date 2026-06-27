// Package rnnoise provides the RNNoise denoiser.
package rnnoise

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// New returns an RNNoise background-noise suppression filter.
func New() *zrt.Denoise {
	return zrt.NewDenoise(zrt.DenoiseOptions{Provider: ""})
}
