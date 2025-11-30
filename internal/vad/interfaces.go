package vad

// VoiceActivityDetector handles voice activity detection
type VoiceActivityDetector interface {
	// Initialize initializes the VAD with configuration
	Initialize(config VADConfig) error

	// ProcessSample processes a single audio sample
	// Returns true if speech is detected, false for silence
	ProcessSample(sample float32) bool

	// IsSpeaking returns current speech state
	IsSpeaking() bool

	// GetSilenceDuration returns current silence duration in samples
	GetSilenceDuration() int

	// Reset resets the VAD state for next phrase
	Reset()

	// IsCalibrated returns true if noise floor calibration is complete
	IsCalibrated() bool
}

// VADConfig holds Voice Activity Detection configuration
type VADConfig struct {
	SampleRate          int
	SilenceThreshold    float32
	SilenceDurationMs   int
	MinSpeechDurationMs int
	RMSWindowSize       int
	NoiseFloorSamples   int
}

// VADState represents the current state of voice activity detection
type VADState struct {
	IsSpeaking        bool
	SilenceSamples    int
	SpeechSamples     int
	AdaptiveThreshold float32
	IsCalibrated      bool
}
