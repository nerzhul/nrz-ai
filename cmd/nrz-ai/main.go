package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/nerzhul/nrz-ai/internal/ai"
	"github.com/nerzhul/nrz-ai/internal/audio"
	"github.com/nerzhul/nrz-ai/internal/vad"
	"github.com/nerzhul/nrz-ai/internal/whisper"
	"github.com/spf13/cobra"
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

// Config holds all configuration options
type Config struct {
	// Audio & Speech
	WhisperModel string
	Language     string
	AudioSource  string

	// Wake Word
	WakeWordEnabled bool
	WakeWord        string
	WakeWordSound   string

	// AI Configuration
	AIEnabled    bool
	OllamaURL    string
	OllamaModel  string
	SystemPrompt string

	// Advanced
	Verbose    bool
	MaxHistory int
}

// SpeechProcessor handles the main speech-to-text processing
type SpeechProcessor struct {
	audioCapture   audio.AudioCapture
	audioProcessor audio.AudioProcessor
	vadDetector    vad.VoiceActivityDetector
	whisperService whisper.WhisperService
	aiService      ai.AIService
	conversation   ai.ConversationManager

	audioBuffer   []float32
	language      string
	maxBufferSize int
	aiEnabled     bool

	// Wake word detection
	wakeWordEnabled bool
	wakeWord        string
	wakeWordSound   string
	wakeWordBuffer  []float32
	listeningActive bool
} // NewSpeechProcessor creates a new speech processor
func NewSpeechProcessor(
	capture audio.AudioCapture,
	processor audio.AudioProcessor,
	detector vad.VoiceActivityDetector,
	service whisper.WhisperService,
	aiSvc ai.AIService,
	conv ai.ConversationManager,
	wakeWordEnabled bool,
	wakeWord string,
	wakeWordSound string,
) *SpeechProcessor {
	return &SpeechProcessor{
		audioCapture:    capture,
		audioProcessor:  processor,
		vadDetector:     detector,
		whisperService:  service,
		aiService:       aiSvc,
		conversation:    conv,
		audioBuffer:     make([]float32, 0, sampleRate*maxBufferDurationS),
		language:        "fr",
		maxBufferSize:   sampleRate * maxBufferDurationS,
		aiEnabled:       aiSvc != nil,
		wakeWordEnabled: wakeWordEnabled,
		wakeWord:        wakeWord,
		wakeWordSound:   wakeWordSound,
		wakeWordBuffer:  make([]float32, 0, sampleRate*2), // 2 seconds for wake word detection
		listeningActive: !wakeWordEnabled,                 // If wake word disabled, always listen
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

// detectWakeWord checks if the wake word is present in the audio buffer
func (sp *SpeechProcessor) detectWakeWord() bool {
	if !sp.wakeWordEnabled || len(sp.wakeWordBuffer) < sampleRate/2 {
		return false
	}

	// Use Whisper to transcribe the wake word buffer
	result, err := sp.whisperService.Transcribe(sp.wakeWordBuffer, sp.language)
	if err != nil {
		return false
	}

	// Check if wake word is present (case-insensitive)
	text := strings.ToLower(strings.TrimSpace(result.Text))
	wakeWord := strings.ToLower(sp.wakeWord)

	return strings.Contains(text, wakeWord)
}

// resetWakeWordBuffer clears the wake word buffer
func (sp *SpeechProcessor) resetWakeWordBuffer() {
	sp.wakeWordBuffer = sp.wakeWordBuffer[:0]
}

// startListeningTimeout deactivates listening after 30 seconds of inactivity
func (sp *SpeechProcessor) startListeningTimeout() {
	time.Sleep(30 * time.Second)
	if sp.wakeWordEnabled {
		sp.listeningActive = false
		fmt.Printf("🔍 Listening timeout. Waiting for wake word '%s' again...\n", sp.wakeWord)
	}
}

// playWakeWordSound plays the wake word detection sound asynchronously
func (sp *SpeechProcessor) playWakeWordSound() {
	if sp.wakeWordSound == "" {
		return
	}

	// Play sound using ffplay in background (suppress output)
	go func() {
		cmd := exec.Command("ffplay", "-nodisp", "-autoexit", "-v", "quiet", sp.wakeWordSound)
		err := cmd.Run()
		if err != nil {
			log.Printf("🔊 Failed to play wake word sound: %v", err)
		}
	}()
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

	if sp.wakeWordEnabled {
		fmt.Printf("🔍 Listening for wake word '%s'...\n", sp.wakeWord)
	} else {
		fmt.Println("🔴 Processing audio stream...")
	}

	for {
		n, err := stream.Read(chunk)
		if err != nil {
			log.Printf("Error reading audio stream: %v", err)
			break
		}

		// Convert bytes to float32 samples
		samples := sp.audioProcessor.ProcessBytes(chunk[:n])

		for _, sample := range samples {
			// Handle wake word detection
			if sp.wakeWordEnabled {
				sp.wakeWordBuffer = append(sp.wakeWordBuffer, sample)

				// Keep wake word buffer to reasonable size (2 seconds)
				if len(sp.wakeWordBuffer) > sampleRate*2 {
					// Remove oldest samples
					copy(sp.wakeWordBuffer, sp.wakeWordBuffer[sampleRate/4:])
					sp.wakeWordBuffer = sp.wakeWordBuffer[:len(sp.wakeWordBuffer)-sampleRate/4]
				}

				// Check for wake word every 500ms
				if len(sp.wakeWordBuffer)%(sampleRate/2) == 0 {
					if sp.detectWakeWord() {
						fmt.Printf("🎯 Wake word '%s' detected! Activating listening...\n", sp.wakeWord)
						// Play wake word sound
						sp.playWakeWordSound()
						sp.listeningActive = true
						sp.resetWakeWordBuffer()
						// Start a timer to deactivate listening after 30 seconds of inactivity
						go sp.startListeningTimeout()
					}
				}

				// If not actively listening, skip regular processing
				if !sp.listeningActive {
					continue
				}
			}

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

		// Clean up the text
		cleanText := strings.TrimSpace(result.Text)

		fmt.Printf("[%s] 🎤 %s\n", timestamp, cleanText)

		// Send to AI if enabled and text is meaningful
		if sp.aiEnabled && len(cleanText) > 3 {
			sp.processWithAI(cleanText)
		}
	}
}

// processWithAI sends the transcribed text to the AI service
func (sp *SpeechProcessor) processWithAI(text string) {
	// Add user message to conversation
	userMsg := ai.Message{
		Role:    "user",
		Content: text,
	}
	sp.conversation.AddMessage(userMsg)

	// Prepare chat request
	request := ai.ChatRequest{
		Messages: sp.conversation.GetMessages(),
		Model:    "", // Will be set by the service
	}

	// Send to AI
	response, err := sp.aiService.Chat(request)
	if err != nil {
		log.Printf("❌ AI Error: %v", err)
		return
	}

	if response.Error != "" {
		log.Printf("❌ AI Response Error: %s", response.Error)
		return
	}

	// Validate response content
	if response.Message.Content == "" {
		log.Printf("⚠️  Warning: AI returned empty response")
		return
	}

	// Add AI response to conversation
	sp.conversation.AddMessage(response.Message)

	// Display AI response
	timestamp := time.Now().Format("15:04:05")
	cleanContent := strings.TrimSpace(response.Message.Content)

	fmt.Printf("[%s] 🤖 %s\n", timestamp, cleanContent)
} // resetForNextPhrase resets state for next phrase
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
	var config Config

	var rootCmd = &cobra.Command{
		Use:   "nrz-ai",
		Short: "Real-time Speech-to-Text with AI conversation",
		Long: `NRZ-AI is a real-time speech-to-text application with intelligent Voice Activity Detection,
optional wake word detection, and AI conversation capabilities using Ollama.

Features:
  • Smart VAD with adaptive noise floor calibration
  • Wake word detection for privacy (optional)
  • Real-time French/multilingual speech transcription  
  • Optional AI conversation with Ollama integration
  • Configurable models and audio sources`,
		Run: func(cmd *cobra.Command, args []string) {
			runApp(config)
		},
	}

	// Audio & Speech flags
	rootCmd.PersistentFlags().StringVarP(&config.WhisperModel, "model", "m",
		"./models/ggml-large-v3.bin", "Path to Whisper model file")
	rootCmd.PersistentFlags().StringVarP(&config.Language, "language", "l",
		"fr", "Language code (fr, en, es, etc.)")
	rootCmd.PersistentFlags().StringVarP(&config.AudioSource, "audio-source", "a",
		"default", "Audio source (PulseAudio device name)")

	// Wake Word flags
	rootCmd.PersistentFlags().BoolVarP(&config.WakeWordEnabled, "wake-word", "w", false,
		"Enable wake word detection (requires saying wake word before listening)")
	rootCmd.PersistentFlags().StringVar(&config.WakeWord, "wake-word-text",
		"Jack", "Wake word to activate listening")
	rootCmd.PersistentFlags().StringVar(&config.WakeWordSound, "wake-word-sound",
		"./sounds/pop-cartoon-328167.mp3", "Sound file to play when wake word is detected")

	// AI flags
	rootCmd.PersistentFlags().BoolVar(&config.AIEnabled, "ai", false,
		"Enable AI conversation with Ollama")
	rootCmd.PersistentFlags().StringVar(&config.OllamaURL, "ollama-url",
		"http://localhost:11434", "Ollama server URL")
	rootCmd.PersistentFlags().StringVar(&config.OllamaModel, "ollama-model",
		"llama3.2:3b", "Ollama model to use")
	rootCmd.PersistentFlags().StringVar(&config.SystemPrompt, "system-prompt",
		"Tu es un assistant vocal français intelligent et concis. Réponds brièvement et naturellement.",
		"AI system prompt")
	rootCmd.PersistentFlags().IntVar(&config.MaxHistory, "max-history", 10,
		"Maximum conversation history to keep")

	// Advanced flags
	rootCmd.PersistentFlags().BoolVarP(&config.Verbose, "verbose", "v", false,
		"Enable verbose logging")

	// Add subcommands
	rootCmd.AddCommand(createListModelsCmd())
	rootCmd.AddCommand(createTestAudioCmd())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runApp(config Config) {
	if config.Verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	fmt.Printf("🎙️  NRZ-AI - Real-time Speech-to-Text\n")
	fmt.Printf("📦 Whisper model: %s\n", config.WhisperModel)
	fmt.Printf("🎤 Audio source: %s\n", config.AudioSource)
	fmt.Printf("🗣️  Language: %s\n", config.Language)

	if config.WakeWordEnabled {
		fmt.Printf("🔍 Wake word: %s (listening mode)\n", config.WakeWord)
	}

	if config.AIEnabled {
		fmt.Printf("🤖 AI Service: Ollama (%s)\n", config.OllamaURL)
		fmt.Printf("🧠 Model: %s\n", config.OllamaModel)
	}

	// Create components using our architecture
	audioCapture := audio.NewFFmpegCapture()
	audioProcessor := audio.NewProcessor()
	vadDetector := vad.NewRMSDetector()
	whisperService := whisper.NewService()

	// Create AI components if enabled
	var aiService ai.AIService
	var conversation ai.ConversationManager

	if config.AIEnabled {
		aiService = ai.NewOllamaService(config.OllamaURL, config.OllamaModel)
		conversation = ai.NewConversation(config.MaxHistory)

		// Check if Ollama is available
		if !aiService.IsAvailable() {
			log.Printf("⚠️  Warning: Ollama service not available at %s", config.OllamaURL)
			log.Printf("   Make sure Ollama is running: ollama serve")
			log.Printf("   And the model is available: ollama pull %s", config.OllamaModel)
			config.AIEnabled = false
			aiService = nil
			conversation = nil
		} else {
			conversation.SetSystemPrompt(config.SystemPrompt)
			fmt.Printf("✅ AI service connected successfully\n")
		}
	}

	// Create speech processor
	processor := NewSpeechProcessor(audioCapture, audioProcessor, vadDetector, whisperService, aiService, conversation, config.WakeWordEnabled, config.WakeWord, config.WakeWordSound)

	// Initialize
	if err := processor.Initialize(config.WhisperModel, config.AudioSource, config.Language); err != nil {
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

	if config.AIEnabled {
		fmt.Println("💡 Tip: Speak naturally, AI will respond to your voice!")
	}

	if config.WakeWordEnabled {
		fmt.Printf("🎯 Say '%s' to activate listening, then speak normally\n", config.WakeWord)
	}

	fmt.Println("─────────────────────────────────────────────")

	// Start processing
	if err := processor.ProcessStream(config.AudioSource); err != nil {
		log.Fatalf("Failed to process stream: %v", err)
	}
}

func createListModelsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-models",
		Short: "List available Ollama models",
		Run: func(cmd *cobra.Command, args []string) {
			ollamaURL, _ := cmd.Flags().GetString("ollama-url")
			if ollamaURL == "" {
				ollamaURL = "http://localhost:11434"
			}

			service := ai.NewOllamaService(ollamaURL, "")
			if !service.IsAvailable() {
				log.Fatalf("❌ Ollama not available at %s", ollamaURL)
			}

			models, err := service.ListModels()
			if err != nil {
				log.Fatalf("❌ Failed to list models: %v", err)
			}

			fmt.Println("📋 Available Ollama models:")
			for _, model := range models {
				fmt.Printf("  • %s\n", model)
			}
		},
	}
}

func createTestAudioCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test-audio",
		Short: "Test audio input",
		Run: func(cmd *cobra.Command, args []string) {
			audioSource, _ := cmd.Flags().GetString("audio-source")
			if audioSource == "" {
				audioSource = "default"
			}

			fmt.Printf("🎤 Testing audio source: %s\n", audioSource)
			fmt.Println("This will capture 3 seconds of audio...")

			capture := audio.NewFFmpegCapture()
			stream, err := capture.StartCapture(audioSource)
			if err != nil {
				log.Fatalf("❌ Failed to start audio capture: %v", err)
			}
			defer stream.Close()

			// Read for 3 seconds
			buffer := make([]byte, 4096)
			totalBytes := 0
			timeout := time.After(3 * time.Second)

			for {
				select {
				case <-timeout:
					fmt.Printf("✅ Audio test complete. Captured %d bytes\n", totalBytes)
					if totalBytes > 0 {
						fmt.Println("✅ Audio input is working!")
					} else {
						fmt.Println("❌ No audio data received. Check your microphone.")
					}
					return
				default:
					n, err := stream.Read(buffer)
					if err != nil {
						log.Printf("❌ Audio read error: %v", err)
						return
					}
					totalBytes += n
				}
			}
		},
	}
}
