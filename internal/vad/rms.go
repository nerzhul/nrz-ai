package vad

import (
	"log"
)

// RMSDetector implements VoiceActivityDetector using RMS-based detection
type RMSDetector struct {
	config         VADConfig
	rmsBuffer      []float32
	silenceSamples int
	speechSamples  int
	isSpeaking     bool

	// Adaptive noise floor
	noiseFloorSamplesCount int
	noiseFloorSum          float64
	adaptiveThreshold      float32
	calibrating            bool
}

// NewRMSDetector creates a new RMS-based voice activity detector
func NewRMSDetector() *RMSDetector {
	return &RMSDetector{
		rmsBuffer:   make([]float32, 0),
		calibrating: true,
	}
}

// Initialize initializes the VAD with configuration
func (r *RMSDetector) Initialize(config VADConfig) error {
	r.config = config
	r.rmsBuffer = make([]float32, 0, config.RMSWindowSize)
	r.adaptiveThreshold = config.SilenceThreshold
	r.calibrating = true

	log.Printf("🎯 VAD Initialized - RMS window: %d, silence threshold: %.3f, duration: %dms",
		config.RMSWindowSize, config.SilenceThreshold, config.SilenceDurationMs)

	if r.calibrating {
		log.Printf("🎚️  Calibrating noise floor for %.1f seconds...",
			float64(config.NoiseFloorSamples)/float64(config.SampleRate))
	}

	return nil
}

// ProcessSample processes a single audio sample
func (r *RMSDetector) ProcessSample(sample float32) bool {
	// Add to RMS calculation buffer
	r.rmsBuffer = append(r.rmsBuffer, sample*sample)
	if len(r.rmsBuffer) > r.config.RMSWindowSize {
		r.rmsBuffer = r.rmsBuffer[1:] // Keep sliding window
	}

	// Calculate RMS level
	rmsLevel := r.calculateRMS()

	// Adaptive noise floor calibration
	if r.calibrating && r.noiseFloorSamplesCount < r.config.NoiseFloorSamples {
		r.noiseFloorSum += float64(rmsLevel)
		r.noiseFloorSamplesCount++
		if r.noiseFloorSamplesCount >= r.config.NoiseFloorSamples {
			noiseFloor := r.noiseFloorSum / float64(r.config.NoiseFloorSamples)
			r.adaptiveThreshold = float32(noiseFloor * 3.0) // 3x noise floor
			if r.adaptiveThreshold < r.config.SilenceThreshold {
				r.adaptiveThreshold = r.config.SilenceThreshold
			}
			r.calibrating = false
			log.Printf("🎚️  Noise floor calibrated: %.6f, adaptive threshold: %.6f",
				noiseFloor, r.adaptiveThreshold)
		}
		return false // Skip VAD during calibration
	}

	// Voice Activity Detection using RMS
	if rmsLevel > r.adaptiveThreshold {
		// Speech detected
		if !r.isSpeaking {
			log.Printf("🎤 Speech started (RMS: %.6f > %.6f)", rmsLevel, r.adaptiveThreshold)
			r.isSpeaking = true
		}
		r.silenceSamples = 0
		r.speechSamples++
	} else if r.isSpeaking {
		// Increment silence counter
		r.silenceSamples++
	}

	return r.isSpeaking
}

// IsSpeaking returns current speech state
func (r *RMSDetector) IsSpeaking() bool {
	return r.isSpeaking
}

// GetSilenceDuration returns current silence duration in samples
func (r *RMSDetector) GetSilenceDuration() int {
	return r.silenceSamples
}

// Reset resets the VAD state for next phrase
func (r *RMSDetector) Reset() {
	r.silenceSamples = 0
	r.speechSamples = 0
	r.isSpeaking = false
	log.Println("⏸️  VAD reset, ready for next phrase")
}

// IsCalibrated returns true if noise floor calibration is complete
func (r *RMSDetector) IsCalibrated() bool {
	return !r.calibrating
}

// calculateRMS calculates RMS level from current buffer
func (r *RMSDetector) calculateRMS() float32 {
	if len(r.rmsBuffer) == 0 {
		return 0.0
	}

	var sum float32
	for _, val := range r.rmsBuffer {
		sum += val
	}

	meanSquare := sum / float32(len(r.rmsBuffer))
	// Simple approximation of square root
	if meanSquare <= 0 {
		return 0.0
	}

	// Newton's method for square root
	x := meanSquare
	for i := 0; i < 5; i++ {
		x = (x + meanSquare/x) / 2
	}

	return x
}
