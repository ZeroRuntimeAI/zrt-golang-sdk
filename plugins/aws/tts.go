// Package aws provides the AWS Polly text-to-speech provider.
package aws

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the AWS Polly text-to-speech provider.
type TTS struct {
	zrt.BaseTTS
	// Voice is the resolved Polly voice id.
	Voice string
	// Region is the resolved AWS region for Polly.
	Region string
	// Engine is the resolved Polly synthesis engine.
	Engine string
	// AWSSecretAccessKey is the resolved AWS secret access key.
	AWSSecretAccessKey string
	// AWSSessionToken is the resolved AWS session token.
	AWSSessionToken string
	// Speed is the resolved speech-rate multiplier.
	Speed float64
	// Pitch is the resolved pitch adjustment.
	Pitch float64
}

// TTSOptions configures an AWS Polly TTS instance.
type TTSOptions struct {
	// AWSAccessKeyID overrides the AWS_ACCESS_KEY_ID environment variable.
	AWSAccessKeyID string
	// AWSSecretAccessKey overrides the AWS_SECRET_ACCESS_KEY environment variable.
	AWSSecretAccessKey string
	// AWSSessionToken overrides the AWS_SESSION_TOKEN environment variable.
	AWSSessionToken string
	// Region is the AWS region. Defaults to "us-east-1".
	Region string
	// Voice is the Polly voice. Defaults to "Joanna".
	Voice string
	// Engine is the Polly synthesis engine. Defaults to "neural".
	Engine string
	// Speed is the speech-rate multiplier. nil uses the default (1.0).
	Speed *float64
	// Pitch is the pitch adjustment. nil uses the default (0.0).
	Pitch *float64
}

// NewTTS returns an AWS Polly TTS configured from opts.
func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{
		Voice:              zrt.StrOr(opts.Voice, "Joanna"),
		Region:             zrt.StrOr(opts.Region, "us-east-1"),
		Engine:             zrt.StrOr(opts.Engine, "neural"),
		AWSSecretAccessKey: opts.AWSSecretAccessKey,
		AWSSessionToken:    opts.AWSSessionToken,
		Speed:              zrt.FloatOr(opts.Speed, 1.0),
		Pitch:              zrt.FloatOr(opts.Pitch, 0.0),
	}
	t.InitTTS("aws", zrt.APIKeyOr(opts.AWSAccessKeyID, "AWS_ACCESS_KEY_ID"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "aws", Voice: t.Voice}
}
