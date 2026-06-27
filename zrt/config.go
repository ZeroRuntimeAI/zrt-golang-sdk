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
	Mode                          string
	InterruptMinDuration          float64
	InterruptMinWords             int
	InterruptMinConfidence        float64
	FalseInterruptPauseDuration   float64
	ResumeOnFalseInterrupt        bool
	FalseInterruptPauseDurationMS int
	InterruptFadeDuration         float64
	InterruptFadeDurationMS       int
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
	Mode               string
	ResponseModalities []string
}

// ContextWindow configures automatic context window management.
type ContextWindow struct {
	MaxTokens           int
	MaxContextItems     int
	KeepRecentTurns     int
	MaxToolCallsPerTurn int
	SummaryLLM          LLM
}

// DefaultContextWindow returns a ContextWindow with default values.
func DefaultContextWindow() *ContextWindow {
	return &ContextWindow{KeepRecentTurns: 3, MaxToolCallsPerTurn: 10}
}

// S3StorageConfig configures S3 (or S3-compatible) recording storage.
type S3StorageConfig struct {
	Bucket               string
	Region               string
	Prefix               string
	AccessKeyID          string
	SecretAccessKey      string
	SessionToken         string
	EndpointURL          string
	StorageClass         string
	ServerSideEncryption string
	KMSKeyID             string
	ACL                  string
	MultipartUpload      bool
	MultipartPartSizeMB  int
	UploadTimeoutSeconds int
	MaxRetryAttempts     int
	Tags                 map[string]string
	UserMetadata         map[string]string
	ContentTypeOverride  string
}

// NewS3StorageConfig returns an S3StorageConfig with default values
// (multipart_upload=true, multipart_part_size_mb=8, max_retry_attempts=3).
func NewS3StorageConfig() *S3StorageConfig {
	return &S3StorageConfig{MultipartUpload: true, MultipartPartSizeMB: 8, MaxRetryAttempts: 3}
}

// RecordingTranscriptConfig configures transcript output for a recording.
type RecordingTranscriptConfig struct {
	Enabled               bool
	Format                string // "json", "srt", "vtt"
	IncludeWordTimestamps bool
	IncludeConfidence     bool
	SpeakerLabels         bool
	Language              string
}

// RecordingConfig configures session recording.
type RecordingConfig struct {
	Enabled            bool
	AutoStart          bool
	Format             string // "wav", "ogg_opus", "mp3", "flac"
	ChannelMode        string // "mixed", "dual_channel"
	SampleRate         int
	BitrateKbps        int
	Storage            *S3StorageConfig
	Transcript         *RecordingTranscriptConfig
	MaxDurationSeconds int
	MaxFileSizeMB      int
	RecordingBeep      bool
	RedactDTMF         bool
	CustomMetadata     map[string]string
	RecordingName      string
	RecordingGroup     string
	ApplyDenoise       bool
	NormalizeAudio     bool
	TrimSilence        bool
	RecordVideo        bool
	RecordScreenShare  bool
	RecordScreenAudio  bool
}

// NewRecordingConfig returns a RecordingConfig with default values
// (auto_start=true, format="ogg_opus", channel_mode="dual_channel").
func NewRecordingConfig() *RecordingConfig {
	return &RecordingConfig{AutoStart: true, Format: "ogg_opus", ChannelMode: "dual_channel"}
}
