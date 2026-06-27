// Package aws provides the AWS Polly text-to-speech provider.
package aws

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the AWS Polly text-to-speech provider.
type TTS struct {
	zrt.BaseTTS
	Voice  string
	Region string
	Engine string
}

// TTSOptions configures an AWS Polly TTS instance.
type TTSOptions struct {
	// AWSAccessKeyID overrides the AWS_ACCESS_KEY_ID environment variable.
	AWSAccessKeyID string
	// Region is the AWS region. Defaults to "us-east-1".
	Region string
	// Voice is the Polly voice. Defaults to "Joanna".
	Voice string
	// Engine is the Polly synthesis engine. Defaults to "neural".
	Engine string
}

// NewTTS returns an AWS Polly TTS configured from opts.
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
