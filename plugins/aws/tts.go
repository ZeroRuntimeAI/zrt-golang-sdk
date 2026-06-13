// Package aws provides the AWS Polly text-to-speech provider.
package aws

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the AWS Polly text-to-speech descriptor.
type TTS struct {
	zrt.BaseTTS
	Voice  string
	Region string
	Engine string
}

// TTSOptions configures TTS.
type TTSOptions struct {
	AWSAccessKeyID string
	Region         string // default "us-east-1"
	Voice          string // default "Joanna"
	Engine         string // default "neural"
}

// NewTTS builds a TTS.
func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{
		Voice:  zrt.StrOr(opts.Voice, "Joanna"),
		Region: zrt.StrOr(opts.Region, "us-east-1"),
		Engine: zrt.StrOr(opts.Engine, "neural"),
	}
	t.InitTTS("aws", zrt.APIKeyOr(opts.AWSAccessKeyID, "AWS_ACCESS_KEY_ID"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "aws", Voice: t.Voice}
}
