package zrt

// BackgroundAudioHandlerConfig configures a background-audio loop.
type BackgroundAudioHandlerConfig struct {
	File      string
	Volume    float64 // default 1.0
	Looping   bool
	Enabled   bool   // default true
	Mode      string // "playback" or "mixing" (default "mixing")
	ChunkSize int    // default 320
}

// NewBackgroundAudioHandlerConfig returns a config with default values.
func NewBackgroundAudioHandlerConfig(file string) *BackgroundAudioHandlerConfig {
	return &BackgroundAudioHandlerConfig{File: file, Volume: 1.0, Enabled: true, Mode: "mixing", ChunkSize: 320}
}
