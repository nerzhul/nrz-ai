package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nerzhul/nrz-ai/internal/audio"
	"github.com/nerzhul/nrz-ai/internal/vad"
	"github.com/nerzhul/nrz-ai/internal/whisper"
)

const (
	sampleRate          = 16000
	readChunkSize       = 4096
	silenceThreshold    = 0.01
	silenceDurationMs   = 800
	minSpeechDurationMs = 500
	maxBufferDurationS  = 30
	rmsWindowSize       = 160
	noiseFloorSamples   = 32000
)

// SpeechProcessor handles the main speech-to-text processing
type SpeechProcessor struct {
	audioCapture   audio.AudioCapture
	audioProcessor audio.AudioProcessor
	vadDetector    vad.VoiceActivityDetector
	whisperService whisper.WhisperService

	audioBuffer   []float32
	language      string
	maxBufferSize int
}

// NewSpeechProcessor creates a new speech processor
func NewSpeechProcessor(
	capture audio.AudioCapture,
	processor audio.AudioProcessor,
	detector vad.VoiceActivityDetector,
	service whisper.WhisperService,
) *SpeechProcessor {
	return &SpeechProcessor{
		audioCapture:   capture,
		audioProcessor: processor,
		vadDetector:    detector,
		whisperService: service,
		audioBuffer:    make([]float32, 0, sampleRate*maxBufferDurationS),
		language:       "fr",
		maxBufferSize:  sampleRate * maxBufferDurationS,
	}
}

// Initialize initializes all components
func (sp *SpeechProcessor) Initialize(modelPath, audioSource, language string) error {
	// Load Whisper model
	if err := sp.whisperService.LoadModel(modelPath); err != nil {
		return fmt.Errorf("failed to load Whisper model: %w", err)
	}

	sp.whisperService.SetLanguage(language)
	sp.language = language

	// Initialize VAD
	vadConfig := vad.VADConfig{
		SampleRate:          sampleRate,
		SilenceThreshold:    silenceThreshold,
		SilenceDurationMs:   silenceDurationMs,
		MinSpeechDurationMs: minSpeechDurationMs,
		RMSWindowSize:       rmsWindowSize,
		NoiseFloorSamples:   noiseFloorSamples,
	}

	return sp.vadDetector.Initialize(vadConfig)
}

// ProcessStream processes the audio stream
func (sp *SpeechProcessor) ProcessStream(audioSource string) error {
	stream, err := sp.audioCapture.StartCapture(audioSource)
	if err != nil {
		return fmt.Errorf("failed to start audio capture: %w", err)
	}
	defer stream.Close()

	chunk := make([]byte, readChunkSize)
	silenceThresholdSamples := (silenceDurationMs * sampleRate) / 1000
	minSpeechSamples := (minSpeechDurationMs * sampleRate) / 1000

	fmt.Println("🔴 Processing audio stream...")

	for {
		n, err := stream.Read(chunk)
		if err != nil {
			log.Printf("Error reading audio stream: %v", err)
			break
		}

		// Convert bytes to float32 samples
		samples := sp.audioProcessor.ProcessBytes(chunk[:n])

		for _, sample := range samples {
			sp.audioBuffer = append(sp.audioBuffer, sample)

			// Process sample with VAD
			sp.vadDetector.ProcessSample(sample)

			// Check if we should transcribe (silence detected after speech)
			if sp.vadDetector.IsSpeaking() &&
				sp.vadDetector.GetSilenceDuration() >= silenceThresholdSamples {

				if len(sp.audioBuffer) >= minSpeechSamples {
					sp.transcribeAndOutput()
				}

				sp.resetForNextPhrase()
			}
		}

		// Prevent buffer overflow
		if len(sp.audioBuffer) >= sp.maxBufferSize {
			log.Println("⚠️  Max buffer reached, processing...")
			sp.transcribeAndOutput()
			sp.resetForNextPhrase()
		}
	}

	return nil
}

// transcribeAndOutput transcribes current buffer and outputs result
func (sp *SpeechProcessor) transcribeAndOutput() {
	log.Printf("📊 Processing %d samples (%.2f seconds)",
		len(sp.audioBuffer), float64(len(sp.audioBuffer))/float64(sampleRate))

	result, err := sp.whisperService.Transcribe(sp.audioBuffer, sp.language)
	if err != nil {
		log.Printf("Failed to transcribe: %v", err)
		return
	}

	if result.Text != "" {
		timestamp := time.Now().Format("15:04:05")
		fmt.Printf("[%s] 💬 %s\n", timestamp, result.Text)
	}
}

// resetForNextPhrase resets state for next phrase
func (sp *SpeechProcessor) resetForNextPhrase() {
	sp.audioBuffer = sp.audioBuffer[:0]
	sp.vadDetector.Reset()
}

// Close closes all resources
func (sp *SpeechProcessor) Close() error {
	if err := sp.audioCapture.Stop(); err != nil {
		log.Printf("Error stopping audio capture: %v", err)
	}
	return sp.whisperService.Close()
}

func main() {
	// Configuration from environment variables
	modelPath := getEnvOrDefault("WHISPER_MODEL", "./models/ggml-large-v3.bin")
	language := getEnvOrDefault("WHISPER_LANGUAGE", "fr")
	audioSource := getEnvOrDefault("AUDIO_SOURCE", "default")

	fmt.Printf("🎙️  Real-time Speech-to-Text (Streaming Mode)\n")
	fmt.Printf("📦 Loading Whisper model: %s\n", modelPath)
	fmt.Printf("🎤 Audio source: %s\n", audioSource)
	fmt.Printf("🗣️  Language: %s\n", language)

	// Create components using our new architecture
	audioCapture := audio.NewFFmpegCapture()
	audioProcessor := audio.NewProcessor()
	vadDetector := vad.NewRMSDetector()
	whisperService := whisper.NewService()

	// Create speech processor
	processor := NewSpeechProcessor(audioCapture, audioProcessor, vadDetector, whisperService)

	// Initialize
	if err := processor.Initialize(modelPath, audioSource, language); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	defer processor.Close()

	// Handle shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n\n✅ Stopping recording")
		processor.Close()
		os.Exit(0)
	}()

	fmt.Println("─────────────────────────────────────────────")

	// Start processing
	if err := processor.ProcessStream(audioSource); err != nil {
		log.Fatalf("Failed to process stream: %v", err)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
