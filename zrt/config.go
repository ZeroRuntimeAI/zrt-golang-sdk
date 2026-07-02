package zrt

// EOUConfig configures end-of-utterance timing.
type EOUConfig struct {
	// Mode is "ADAPTIVE" or "DEFAULT".
	Mode string
	// MinMaxSpeechWaitTimeout is [min, max] seconds to wait for more speech.
	MinMaxSpeechWaitTimeout []float64
}

// DefaultEOUConfig returns an EOUConfig with default values.
func DefaultEOUConfig() *EOUConfig {
	return &EOUConfig{Mode: "DEFAULT", MinMaxSpeechWaitTimeout: []float64{0.5, 0.8}}
}

// InterruptConfig configures barge-in / interruption behavior.
type InterruptConfig struct {
	// Mode is "VAD_ONLY", "STT_ONLY" or "HYBRID".
	Mode string
	// InterruptMinDuration is the minimum speech duration, in seconds, required to trigger an interrupt. Defaults to 0.5.
	InterruptMinDuration float64
	// InterruptMinWords is the minimum number of recognized words required to trigger an interrupt. Defaults to 2.
	InterruptMinWords int
	// InterruptMinConfidence is the minimum STT confidence required to trigger an interrupt. Defaults to 0.0.
	InterruptMinConfidence float64
	// FalseInterruptPauseDuration is the legacy pause duration, in seconds, applied on a false interrupt. Defaults to 2.0.
	FalseInterruptPauseDuration float64
	// ResumeOnFalseInterrupt resumes playback after a false interrupt. Defaults to true.
	ResumeOnFalseInterrupt bool
	// FalseInterruptPauseDurationMS is the pause duration, in milliseconds, applied on a false interrupt. Defaults to 2000.
	FalseInterruptPauseDurationMS int
	// InterruptFadeDuration is the audio fade-out duration, in seconds, when interrupted. Defaults to 0.0.
	InterruptFadeDuration float64
	// InterruptFadeDurationMS is the audio fade-out duration, in milliseconds, when interrupted. Defaults to 400.
	InterruptFadeDurationMS int
}

// DefaultInterruptConfig returns an InterruptConfig with default values.
func DefaultInterruptConfig() *InterruptConfig {
	return &InterruptConfig{
		Mode:                          "HYBRID",
		InterruptMinDuration:          0.5,
		InterruptMinWords:             2,
		InterruptMinConfidence:        0.0,
		FalseInterruptPauseDuration:   2.0,
		ResumeOnFalseInterrupt:        true,
		FalseInterruptPauseDurationMS: 2000,
		InterruptFadeDuration:         0.0,
		InterruptFadeDurationMS:       400,
	}
}

// normalize derives the millisecond fade duration from the seconds value.
func (ic *InterruptConfig) normalize() {
	if ic.InterruptFadeDuration > 0 && ic.InterruptFadeDurationMS == 0 {
		ic.InterruptFadeDurationMS = int(ic.InterruptFadeDuration * 1000)
	}
}

// RealtimeConfig configures realtime (speech-to-speech) pipeline mode.
type RealtimeConfig struct {
	// Mode is "full_s2s", "hybrid_stt", "hybrid_tts" or "llm_only" ("" = auto).
	Mode string
	// ResponseModalities lists the output modalities, e.g. "TEXT" and/or "AUDIO". Empty = provider/pipeline default.
	ResponseModalities []string
}

// ContextWindow configures automatic context window management.
type ContextWindow struct {
	// MaxTokens caps the total context size in tokens. 0 = unset.
	MaxTokens int
	// MaxContextItems caps the number of retained context items. 0 = unset.
	MaxContextItems int
	// KeepRecentTurns is the number of most recent turns always kept. Defaults to 3.
	KeepRecentTurns int
	// MaxToolCallsPerTurn caps tool calls allowed per turn. Defaults to 10.
	MaxToolCallsPerTurn int
	// SummaryLLM is the LLM used to summarize older context when trimming. nil = disabled.
	SummaryLLM LLM
}

// DefaultContextWindow returns a ContextWindow with default values.
func DefaultContextWindow() *ContextWindow {
	return &ContextWindow{KeepRecentTurns: 3, MaxToolCallsPerTurn: 10}
}

// S3StorageConfig configures S3 (or S3-compatible) recording storage.
type S3StorageConfig struct {
	// Bucket is the destination S3 bucket name.
	Bucket string
	// Region is the S3 region.
	Region string
	// Prefix is the key prefix prepended to uploaded objects.
	Prefix string
	// AccessKeyID is the S3 access key ID.
	AccessKeyID string
	// SecretAccessKey is the S3 secret access key.
	SecretAccessKey string
	// SessionToken is the optional temporary session token.
	SessionToken string
	// EndpointURL overrides the S3 endpoint for S3-compatible storage.
	EndpointURL string
	// StorageClass is the S3 storage class, e.g. "STANDARD".
	StorageClass string
	// ServerSideEncryption is the server-side encryption mode, e.g. "AES256" or "aws:kms".
	ServerSideEncryption string
	// KMSKeyID is the KMS key ID used when ServerSideEncryption is KMS-based.
	KMSKeyID string
	// ACL is the canned object ACL, e.g. "private".
	ACL string
	// MultipartUpload enables multipart uploads. Defaults to true.
	MultipartUpload bool
	// MultipartPartSizeMB is the multipart part size in megabytes. Defaults to 8.
	MultipartPartSizeMB int
	// UploadTimeoutSeconds is the upload timeout in seconds. 0 = unset.
	UploadTimeoutSeconds int
	// MaxRetryAttempts is the number of upload retry attempts. Defaults to 3.
	MaxRetryAttempts int
	// Tags are S3 object tags applied to uploaded objects.
	Tags map[string]string
	// UserMetadata is custom user metadata attached to uploaded objects.
	UserMetadata map[string]string
	// ContentTypeOverride overrides the detected object Content-Type.
	ContentTypeOverride string
}

// NewS3StorageConfig returns an S3StorageConfig with default values
// (multipart_upload=true, multipart_part_size_mb=8, max_retry_attempts=3).
func NewS3StorageConfig() *S3StorageConfig {
	return &S3StorageConfig{MultipartUpload: true, MultipartPartSizeMB: 8, MaxRetryAttempts: 3}
}

// RecordingTranscriptConfig configures transcript output for a recording.
type RecordingTranscriptConfig struct {
	// Enabled turns on transcript generation for the recording.
	Enabled bool
	// Format is the transcript output format.
	Format string // "json", "srt", "vtt"
	// IncludeWordTimestamps includes per-word timestamps in the transcript.
	IncludeWordTimestamps bool
	// IncludeConfidence includes recognition confidence scores in the transcript.
	IncludeConfidence bool
	// SpeakerLabels adds speaker labels to the transcript. Defaults to true.
	SpeakerLabels bool
	// Language is the transcript language hint. Empty = auto/detect.
	Language string
}

// RecordingConfig configures session recording.
type RecordingConfig struct {
	// Enabled turns on session recording.
	Enabled bool
	// AutoStart begins recording automatically when the session starts. Defaults to true.
	AutoStart bool
	// Format is the recording audio format.
	Format string // "wav", "ogg_opus", "mp3", "flac"
	// ChannelMode is the channel layout of the recording.
	ChannelMode string // "mixed", "dual_channel"
	// SampleRate is the recording sample rate in Hz. 0 = default.
	SampleRate int
	// BitrateKbps is the encoder bitrate in kbps for lossy formats. 0 = default.
	BitrateKbps int
	// Storage configures where the recording is uploaded. nil = default storage.
	Storage *S3StorageConfig
	// Transcript configures transcript output for the recording. nil = no transcript.
	Transcript *RecordingTranscriptConfig
	// MaxDurationSeconds caps recording length in seconds. 0 = unlimited.
	MaxDurationSeconds int
	// MaxFileSizeMB caps recording file size in megabytes. 0 = unlimited.
	MaxFileSizeMB int
	// RecordingBeep plays a beep to indicate recording is active.
	RecordingBeep bool
	// RedactDTMF removes DTMF tones from the recording.
	RedactDTMF bool
	// CustomMetadata is arbitrary metadata attached to the recording.
	CustomMetadata map[string]string
	// RecordingName is the name assigned to the recording.
	RecordingName string
	// RecordingGroup groups related recordings under a common name.
	RecordingGroup string
	// ApplyDenoise applies noise reduction to the recorded audio.
	ApplyDenoise bool
	// NormalizeAudio normalizes the level of the recorded audio.
	NormalizeAudio bool
	// TrimSilence removes silence from the recorded audio.
	TrimSilence bool
	// RecordVideo also records the video track.
	RecordVideo bool
	// RecordScreenShare also records the screen-share track.
	RecordScreenShare bool
	// RecordScreenAudio also records the screen-share audio track.
	RecordScreenAudio bool
}

// NewRecordingConfig returns a RecordingConfig with default values
// (auto_start=true, format="ogg_opus", channel_mode="dual_channel").
func NewRecordingConfig() *RecordingConfig {
	return &RecordingConfig{AutoStart: true, Format: "ogg_opus", ChannelMode: "dual_channel"}
}
