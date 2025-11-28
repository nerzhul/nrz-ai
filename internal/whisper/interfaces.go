package whisper

// TranscriptionResult represents the result of a transcription
type TranscriptionResult struct {
	Text      string
	Segments  []Segment
	Language  string
	Duration  float64
}

// Segment represents a segment of transcribed text
type Segment struct {
	Text     string
	Start    float64
	End      float64
	NoSpeech bool
}

// WhisperService handles speech-to-text transcription
type WhisperService interface {
	// LoadModel loads a Whisper model from the specified path
	LoadModel(modelPath string) error
	
	// Transcribe transcribes audio samples to text
	Transcribe(audio []float32, language string) (TranscriptionResult, error)
	
	// SetLanguage sets the transcription language
	SetLanguage(language string)
	
	// Close closes the Whisper service and releases resources
	Close() error
}

// ModelConfig holds configuration for Whisper model
type ModelConfig struct {
	ModelPath    string
	Language     string
	Threads      int
	Translate    bool
}