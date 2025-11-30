package vad

// MockVAD implements VoiceActivityDetector for testing
type MockVAD struct {
	initialized      bool
	isSpeaking       bool
	silenceDuration  int
	isCalibrated     bool
	speechDetections []bool
	currentSample    int
}

// NewMockVAD creates a mock voice activity detector
func NewMockVAD() *MockVAD {
	return &MockVAD{
		speechDetections: make([]bool, 0),
		isCalibrated:     true, // Start calibrated for testing
	}
}

// SetSpeechPattern sets a pattern of speech detections for testing
func (m *MockVAD) SetSpeechPattern(pattern []bool) {
	m.speechDetections = pattern
	m.currentSample = 0
}

// SetCalibrated sets the calibration state
func (m *MockVAD) SetCalibrated(calibrated bool) {
	m.isCalibrated = calibrated
}

// Initialize initializes the mock VAD
func (m *MockVAD) Initialize(config VADConfig) error {
	m.initialized = true
	return nil
}

// ProcessSample processes a sample and returns the next speech detection in the pattern
func (m *MockVAD) ProcessSample(sample float32) bool {
	if len(m.speechDetections) == 0 {
		return false
	}

	if m.currentSample >= len(m.speechDetections) {
		m.currentSample = 0 // Loop back to start
	}

	speaking := m.speechDetections[m.currentSample]
	m.currentSample++

	if speaking {
		m.isSpeaking = true
		m.silenceDuration = 0
	} else if m.isSpeaking {
		m.silenceDuration++
	}

	return speaking
}

// IsSpeaking returns current speech state
func (m *MockVAD) IsSpeaking() bool {
	return m.isSpeaking
}

// GetSilenceDuration returns current silence duration
func (m *MockVAD) GetSilenceDuration() int {
	return m.silenceDuration
}

// Reset resets the VAD state
func (m *MockVAD) Reset() {
	m.isSpeaking = false
	m.silenceDuration = 0
}

// IsCalibrated returns calibration state
func (m *MockVAD) IsCalibrated() bool {
	return m.isCalibrated
}

// IsInitialized returns initialization state (for testing)
func (m *MockVAD) IsInitialized() bool {
	return m.initialized
}
