package whisper

import "errors"

// MockWhisperService implements WhisperService for testing
type MockWhisperService struct {
	isLoaded           bool
	loadError          error
	transcribeError    error
	transcribeResult   TranscriptionResult
	language           string
	closeError         error
}

// NewMockWhisperService creates a mock Whisper service
func NewMockWhisperService() *MockWhisperService {
	return &MockWhisperService{
		isLoaded: false,
		language: "fr",
	}
}

// SetLoadError sets an error to return on LoadModel calls
func (m *MockWhisperService) SetLoadError(err error) {
	m.loadError = err
}

// SetTranscribeError sets an error to return on Transcribe calls
func (m *MockWhisperService) SetTranscribeError(err error) {
	m.transcribeError = err
}

// SetTranscribeResult sets the result to return on Transcribe calls
func (m *MockWhisperService) SetTranscribeResult(result TranscriptionResult) {
	m.transcribeResult = result
}

// SetCloseError sets an error to return on Close calls
func (m *MockWhisperService) SetCloseError(err error) {
	m.closeError = err
}

// LoadModel simulates loading a model
func (m *MockWhisperService) LoadModel(modelPath string) error {
	if m.loadError != nil {
		return m.loadError
	}
	m.isLoaded = true
	return nil
}

// Transcribe simulates transcribing audio
func (m *MockWhisperService) Transcribe(audio []float32, language string) (TranscriptionResult, error) {
	if !m.isLoaded {
		return TranscriptionResult{}, errors.New("model not loaded")
	}
	
	if m.transcribeError != nil {
		return TranscriptionResult{}, m.transcribeError
	}
	
	return m.transcribeResult, nil
}

// SetLanguage sets the transcription language
func (m *MockWhisperService) SetLanguage(language string) {
	m.language = language
}

// GetLanguage returns the current language (for testing)
func (m *MockWhisperService) GetLanguage() string {
	return m.language
}

// Close simulates closing the service
func (m *MockWhisperService) Close() error {
	if m.closeError != nil {
		return m.closeError
	}
	m.isLoaded = false
	return nil
}

// IsLoaded returns whether the model is loaded (for testing)
func (m *MockWhisperService) IsLoaded() bool {
	return m.isLoaded
}