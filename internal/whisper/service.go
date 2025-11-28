package whisper

import (
	"errors"
	"log"
	"runtime"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

// Common errors
var (
	ErrModelNotLoaded = errors.New("whisper model not loaded")
)

// Service implements WhisperService interface
type Service struct {
	model    whisper.Model
	config   ModelConfig
	isLoaded bool
}

// NewService creates a new Whisper service
func NewService() *Service {
	return &Service{
		isLoaded: false,
	}
}

// LoadModel loads a Whisper model from the specified path
func (s *Service) LoadModel(modelPath string) error {
	model, err := whisper.New(modelPath)
	if err != nil {
		return err
	}
	
	s.model = model
	s.config.ModelPath = modelPath
	s.config.Threads = runtime.NumCPU()
	s.isLoaded = true
	
	log.Printf("📦 Whisper model loaded: %s", modelPath)
	return nil
}

// Transcribe transcribes audio samples to text
func (s *Service) Transcribe(audio []float32, language string) (TranscriptionResult, error) {
	if !s.isLoaded {
		return TranscriptionResult{}, ErrModelNotLoaded
	}
	
	// Create a fresh context for each transcription
	context, err := s.model.NewContext()
	if err != nil {
		return TranscriptionResult{}, err
	}

	context.SetLanguage(language)
	context.SetTranslate(s.config.Translate)
	context.SetThreads(uint(s.config.Threads))

	// Process the audio
	if err := context.Process(audio, nil, nil, nil); err != nil {
		return TranscriptionResult{}, err
	}

	// Extract all segments
	var text string
	var segments []Segment
	
	for {
		segment, err := context.NextSegment()
		if err != nil {
			break
		}
		text += segment.Text
		
		segments = append(segments, Segment{
			Text:     segment.Text,
			Start:    float64(segment.Start) / 1000.0, // Convert ms to seconds
			End:      float64(segment.End) / 1000.0,
			NoSpeech: segment.Text == "",
		})
	}

	return TranscriptionResult{
		Text:     text,
		Segments: segments,
		Language: language,
		Duration: float64(len(audio)) / 16000.0, // Assuming 16kHz sample rate
	}, nil
}

// SetLanguage sets the transcription language
func (s *Service) SetLanguage(language string) {
	s.config.Language = language
}

// Close closes the Whisper service and releases resources
func (s *Service) Close() error {
	if s.isLoaded && s.model != nil {
		s.model.Close()
		s.isLoaded = false
	}
	return nil
}