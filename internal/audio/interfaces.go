package audio

// AudioStream represents an audio input stream
type AudioStream interface {
	// Read reads audio data from the stream
	// Returns bytes read and any error
	Read([]byte) (int, error)
	
	// Close closes the audio stream
	Close() error
}

// AudioCapture handles audio capture from various sources
type AudioCapture interface {
	// StartCapture starts capturing audio from the specified source
	// Returns an AudioStream for reading audio data
	StartCapture(audioSource string) (AudioStream, error)
	
	// Stop stops the audio capture
	Stop() error
}

// AudioProcessor processes raw audio bytes into float32 samples
type AudioProcessor interface {
	// ProcessBytes converts raw audio bytes to float32 samples
	ProcessBytes(data []byte) []float32
	
	// CalculateRMS calculates RMS level from audio samples
	CalculateRMS(samples []float32, windowSize int) float32
}