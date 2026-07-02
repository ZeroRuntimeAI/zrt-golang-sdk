package zrt

// BackgroundAudioHandlerConfig configures a background-audio loop.
type BackgroundAudioHandlerConfig struct {
	// File is the path to the audio file to play.
	File string
	// Volume is the playback gain. Defaults to 1.0.
	Volume float64 // default 1.0
	// Looping replays the file continuously when true.
	Looping bool
	// Enabled turns background audio on. Defaults to true.
	Enabled bool // default true
	// Mode is "playback" or "mixing". Defaults to "mixing".
	Mode string // "playback" or "mixing" (default "mixing")
	// ChunkSize is the audio chunk size in samples. Defaults to 320.
	ChunkSize int // default 320
}

// NewBackgroundAudioHandlerConfig returns a config with default values.
func NewBackgroundAudioHandlerConfig(file string) *BackgroundAudioHandlerConfig {
	return &BackgroundAudioHandlerConfig{File: file, Volume: 1.0, Enabled: true, Mode: "mixing", ChunkSize: 320}
}
