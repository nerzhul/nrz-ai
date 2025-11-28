package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

const (
	sampleRate          = 16000
	readChunkSize       = 4096  // Size of FFmpeg read buffer
	rmsWindowSize       = 160   // RMS calculation window (10ms at 16kHz)
	silenceThreshold    = 0.01  // RMS threshold for speech detection
	silenceDurationMs   = 800   // Silence duration in ms to trigger transcription
	minSpeechDurationMs = 500   // Minimum speech duration to process
	maxBufferDurationS  = 30    // Max buffer duration in seconds
	noiseFloorSamples   = 32000 // Samples to calculate initial noise floor (2 seconds)
)

func main() {
	modelPath := os.Getenv("WHISPER_MODEL")
	if modelPath == "" {
		modelPath = "./models/ggml-large-v3.bin"
	}
	language := os.Getenv("WHISPER_LANGUAGE")
	if language == "" {
		language = "fr"
	}
	audioSource := os.Getenv("AUDIO_SOURCE")
	if audioSource == "" {
		audioSource = "default" // Can also use specific device like "alsa_input.usb-RODE..."
	}

	fmt.Printf("🎙️  Real-time Speech-to-Text (Streaming Mode)\n")
	fmt.Printf("📦 Loading Whisper model: %s\n", modelPath)
	fmt.Printf("🎤 Audio source: %s\n", audioSource)

	model, err := whisper.New(modelPath)
	if err != nil {
		log.Fatalf("Failed to load Whisper model: %v", err)
	}
	defer model.Close()

	fmt.Printf("🖥️  Using %d CPU threads\n", runtime.NumCPU())
	fmt.Println("🔴 Streaming audio with FFmpeg...")
	fmt.Println("─────────────────────────────────────────────")

	// Create audio pipe - capture from default microphone input
	cmd := exec.Command("ffmpeg",
		"-f", "pulse",
		"-i", audioSource,
		"-ar", "16000",
		"-ac", "1",
		"-f", "f32le",
		"-loglevel", "quiet", // Suppress FFmpeg logs
		"-")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to create pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start FFmpeg: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n\n✅ Stopping recording")
		cmd.Process.Kill()
		os.Exit(0)
	}()

	// Streaming buffer and VAD state
	audioBuffer := make([]float32, 0, sampleRate*maxBufferDurationS)
	chunk := make([]byte, readChunkSize)

	// RMS-based VAD state
	rmsBuffer := make([]float32, 0, rmsWindowSize)
	silenceSamples := 0
	silenceThresholdSamples := (silenceDurationMs * sampleRate) / 1000
	minSpeechSamples := (minSpeechDurationMs * sampleRate) / 1000
	isSpeaking := false

	// Adaptive noise floor
	noiseFloorSamplesCount := 0
	noiseFloorSum := 0.0
	adaptiveThreshold := float32(silenceThreshold)
	calibrating := true

	if calibrating {
		log.Printf("🎚️  Calibrating noise floor for %.1f seconds...", float64(noiseFloorSamples)/float64(sampleRate))
	}

	for {
		n, err := stdout.Read(chunk)
		if err != nil {
			log.Printf("Error reading audio stream: %v", err)
			break
		}

		// Convert bytes to float32
		for i := 0; i < n; i += 4 {
			if i+4 <= n {
				sample := float32FromBytes(chunk[i : i+4])
				audioBuffer = append(audioBuffer, sample)

				// Add to RMS calculation buffer
				rmsBuffer = append(rmsBuffer, sample*sample) // Square for RMS
				if len(rmsBuffer) > rmsWindowSize {
					rmsBuffer = rmsBuffer[1:] // Keep sliding window
				}

				// Calculate RMS level
				var rmsSum float32
				for _, val := range rmsBuffer {
					rmsSum += val
				}
				rmsLevel := float32(0.0)
				if len(rmsBuffer) > 0 {
					meanSquare := rmsSum / float32(len(rmsBuffer))
					rmsLevel = float32(math.Sqrt(float64(meanSquare)))
				}

				// Adaptive noise floor calibration
				if calibrating && noiseFloorSamplesCount < noiseFloorSamples {
					noiseFloorSum += float64(rmsLevel)
					noiseFloorSamplesCount++
					if noiseFloorSamplesCount >= noiseFloorSamples {
						noiseFloor := noiseFloorSum / float64(noiseFloorSamples)
						adaptiveThreshold = float32(noiseFloor * 3.0) // 3x noise floor
						if adaptiveThreshold < float32(silenceThreshold) {
							adaptiveThreshold = float32(silenceThreshold)
						}
						calibrating = false
						log.Printf("🎚️  Noise floor calibrated: %.6f, adaptive threshold: %.6f", noiseFloor, adaptiveThreshold)
					}
					continue // Skip VAD during calibration
				}

				// Voice Activity Detection using RMS
				if rmsLevel > adaptiveThreshold {
					// Speech detected
					if !isSpeaking {
						log.Printf("🎤 Speech started (RMS: %.6f > %.6f)", rmsLevel, adaptiveThreshold)
						isSpeaking = true
					}
					silenceSamples = 0
				} else if isSpeaking {
					// Increment silence counter
					silenceSamples++
					// if silenceSamples%1000 == 0 {
					// 	log.Printf("🤫 Silence accumulating: %d samples (%.2fs), RMS: %.6f", silenceSamples, float64(silenceSamples)/float64(sampleRate), rmsLevel)
					// }
				}
			}
		}

		// Check if we should process (silence detected after speech)
		if isSpeaking && silenceSamples >= silenceThresholdSamples {
			log.Printf("🔔 Trigger condition met: isSpeaking=%v, silenceSamples=%d, threshold=%d", isSpeaking, silenceSamples, silenceThresholdSamples)
			if len(audioBuffer) >= minSpeechSamples {
				log.Printf("📊 Processing %d samples (%.2f seconds)", len(audioBuffer), float64(len(audioBuffer))/float64(sampleRate))
				currentText := processAudioStream(model, language, audioBuffer)

				// Always print if we have text
				if currentText != "" {
					timestamp := time.Now().Format("15:04:05")
					fmt.Printf("[%s] 💬 %s\n", timestamp, currentText)
				}
			} else {
				log.Printf("⚠️  Not enough speech samples: %d < %d", len(audioBuffer), minSpeechSamples)
			}

			// Reset for next phrase
			audioBuffer = audioBuffer[:0]
			silenceSamples = 0
			isSpeaking = false
			log.Println("⏸️  Silence detected, ready for next phrase")
		}

		// Debug info every second
		if len(audioBuffer)%(sampleRate) == 0 && isSpeaking {
			log.Printf("🔍 Status: buffer=%d samples (%.1fs), silence=%d samples (%.1fs), speaking=%v",
				len(audioBuffer), float64(len(audioBuffer))/float64(sampleRate),
				silenceSamples, float64(silenceSamples)/float64(sampleRate),
				isSpeaking)
		}

		// Prevent buffer overflow
		if len(audioBuffer) >= sampleRate*maxBufferDurationS {
			log.Println("⚠️  Max buffer reached, processing...")
			currentText := processAudioStream(model, language, audioBuffer)
			if currentText != "" {
				timestamp := time.Now().Format("15:04:05")
				fmt.Printf("[%s] 💬 %s\n", timestamp, currentText)
			}
			audioBuffer = audioBuffer[:0]
			silenceSamples = 0
			isSpeaking = false
		}
	}
}

func float32FromBytes(b []byte) float32 {
	bits := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	return *(*float32)(unsafe.Pointer(&bits))
}

func processAudioStream(model whisper.Model, language string, audio []float32) string {
	// Create a fresh context for each chunk
	context, err := model.NewContext()
	if err != nil {
		log.Printf("Failed to create context: %v", err)
		return ""
	}

	context.SetLanguage(language)
	context.SetTranslate(false)
	context.SetThreads(uint(runtime.NumCPU()))

	// Process the audio
	if err := context.Process(audio, nil, nil, nil); err != nil {
		log.Printf("Failed to transcribe: %v", err)
		return ""
	}

	// Extract all segments
	var text string
	for {
		segment, err := context.NextSegment()
		if err != nil {
			break
		}
		text += segment.Text
	}

	return text
}
