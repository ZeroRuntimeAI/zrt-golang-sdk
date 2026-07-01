package aws

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// DefaultBedrockModel is the model used when none is supplied.
const DefaultBedrockModel = "amazon.nova-lite-v1:0"

// LLM is the AWS Bedrock LLM provider (Converse API).
//
// It streams text generation and function/tool calling from any Bedrock-hosted
// model (Amazon Nova, Anthropic Claude, Meta Llama, Mistral, Google Gemma).
// Credentials resolve as explicit option -> environment variable -> the default
// AWS credential chain (IAM role / shared profile) when left empty.
type LLM struct {
	zrt.BaseLLM
	Model           string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Temperature     float64
	MaxOutputTokens int
	TopP            *float64
	TopK            *int
	StopSequences   []string
	ToolChoice      string
	CacheSystem     *bool
	CacheTools      *bool
	StripThinking   *bool
	TextToolCalls   *bool

	AdditionalRequestFields map[string]any
}

// LLMOptions configures an AWS Bedrock LLM.
type LLMOptions struct {
	// Model is a Bedrock model id or inference-profile ARN. Defaults to
	// "amazon.nova-lite-v1:0"; falls back to BEDROCK_INFERENCE_PROFILE_ARN.
	Model string
	// Region is the AWS region for Bedrock Runtime. Falls back to
	// AWS_DEFAULT_REGION / AWS_REGION, then "us-east-1".
	Region string
	// AWSAccessKeyID falls back to AWS_ACCESS_KEY_ID.
	AWSAccessKeyID string
	// AWSSecretAccessKey falls back to AWS_SECRET_ACCESS_KEY.
	AWSSecretAccessKey string
	// AWSSessionToken falls back to AWS_SESSION_TOKEN. When no credentials are
	// resolved, the runtime uses the default AWS credential chain.
	AWSSessionToken string
	// Temperature is the sampling temperature. nil uses the default (0.7).
	Temperature *float64
	// MaxOutputTokens caps tokens generated per response. Defaults to 1024.
	MaxOutputTokens int
	// TopP is the nucleus-sampling probability mass.
	TopP *float64
	// TopK limits sampling to the K most likely tokens (sent via additional
	// request fields; model support varies).
	TopK *int
	// StopSequences are sequences that stop generation.
	StopSequences []string
	// ToolChoice is "auto", "required", "none", or a tool name. Defaults to "auto".
	ToolChoice string
	// CacheSystem adds a prompt-cache checkpoint after the system prompt.
	CacheSystem *bool
	// CacheTools adds a prompt-cache checkpoint after the tool definitions.
	CacheTools *bool
	// StripThinking removes <thinking>...</thinking> spans from the streamed
	// text. Defaults to enabled in the runtime.
	StripThinking *bool
	// TextToolCalls parses function calls a model prints as plain text instead
	// of native Converse tool use. Auto-enabled for models lacking native tool
	// use (e.g. Gemma) when left unset.
	TextToolCalls *bool
	// AdditionalRequestFields are extra additionalModelRequestFields merged into
	// the Converse request for model-specific parameters.
	AdditionalRequestFields map[string]any
}

// NewLLM creates an AWS Bedrock LLM from opts, applying defaults for unset fields.
func NewLLM(opts LLMOptions) *LLM {
	model := zrt.StrOr(opts.Model, zrt.EnvOr("BEDROCK_INFERENCE_PROFILE_ARN", DefaultBedrockModel))
	region := zrt.StrOr(opts.Region, zrt.EnvOr("AWS_DEFAULT_REGION", zrt.EnvOr("AWS_REGION", "us-east-1")))
	l := &LLM{
		Model:                   model,
		Region:                  region,
		AccessKeyID:             zrt.APIKeyOr(opts.AWSAccessKeyID, "AWS_ACCESS_KEY_ID"),
		SecretAccessKey:         zrt.APIKeyOr(opts.AWSSecretAccessKey, "AWS_SECRET_ACCESS_KEY"),
		SessionToken:            zrt.APIKeyOr(opts.AWSSessionToken, "AWS_SESSION_TOKEN"),
		Temperature:             zrt.FloatOr(opts.Temperature, 0.7),
		MaxOutputTokens:         zrt.IntZeroOr(opts.MaxOutputTokens, 1024),
		TopP:                    opts.TopP,
		TopK:                    opts.TopK,
		StopSequences:           opts.StopSequences,
		ToolChoice:              zrt.StrOr(opts.ToolChoice, "auto"),
		CacheSystem:             opts.CacheSystem,
		CacheTools:              opts.CacheTools,
		StripThinking:           opts.StripThinking,
		TextToolCalls:           opts.TextToolCalls,
		AdditionalRequestFields: opts.AdditionalRequestFields,
	}
	l.Init("bedrock", l.AccessKeyID)
	return l
}

// LLMConfig implements zrt.LLM and reports the model configuration.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "bedrock", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}

// Knobs returns the provider-specific tuning options that are set.
func (l *LLM) Knobs() map[string]any {
	k := map[string]any{}
	if l.TopP != nil {
		k["top_p"] = *l.TopP
	}
	if l.TopK != nil {
		k["top_k"] = *l.TopK
	}
	if len(l.StopSequences) > 0 {
		k["stop_sequences"] = l.StopSequences
	}
	if l.ToolChoice != "" {
		k["tool_choice"] = l.ToolChoice
	}
	if l.CacheSystem != nil {
		k["cache_system"] = *l.CacheSystem
	}
	if l.CacheTools != nil {
		k["cache_tools"] = *l.CacheTools
	}
	if l.StripThinking != nil {
		k["strip_thinking"] = *l.StripThinking
	}
	if l.TextToolCalls != nil {
		k["text_tool_calls"] = *l.TextToolCalls
	}
	if len(l.AdditionalRequestFields) > 0 {
		k["additional_request_fields"] = l.AdditionalRequestFields
	}
	return k
}

// AWSRegion reports the resolved Bedrock region for the runtime credentials.
func (l *LLM) AWSRegion() string { return l.Region }

// AWSAccessKeyID reports the resolved AWS access key id ("" for the default chain).
func (l *LLM) AWSAccessKeyID() string { return l.AccessKeyID }

// AWSSecretAccessKey reports the resolved AWS secret access key.
func (l *LLM) AWSSecretAccessKey() string { return l.SecretAccessKey }

// AWSSessionToken reports the resolved AWS session token.
func (l *LLM) AWSSessionToken() string { return l.SessionToken }
