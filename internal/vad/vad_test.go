package vad

import "testing"

func TestNewMockVAD(t *testing.T) {
	mock := NewMockVAD()
	if mock == nil {
		t.Fatal("Expected mock VAD to be created")
	}

	if !mock.IsCalibrated() {
		t.Error("Expected mock VAD to start calibrated")
	}
}

func TestMockVAD_Initialize(t *testing.T) {
	mock := NewMockVAD()

	config := VADConfig{
		SampleRate:          16000,
		SilenceThreshold:    0.01,
		SilenceDurationMs:   800,
		MinSpeechDurationMs: 500,
		RMSWindowSize:       160,
		NoiseFloorSamples:   32000,
	}

	err := mock.Initialize(config)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !mock.IsInitialized() {
		t.Error("Expected VAD to be initialized")
	}
}

func TestMockVAD_ProcessSample(t *testing.T) {
	mock := NewMockVAD()

	// Set speech pattern: speech for 3 samples, then silence for 2 samples
	pattern := []bool{true, true, true, false, false}
	mock.SetSpeechPattern(pattern)

	// Test processing samples according to pattern
	results := []bool{}
	for i := 0; i < 7; i++ { // Process more than pattern length to test looping
		result := mock.ProcessSample(0.1)
		results = append(results, result)
	}

	// First 3 should be speech, next 2 silence, then pattern repeats
	expected := []bool{true, true, true, false, false, true, true}

	for i, result := range results {
		if result != expected[i] {
			t.Errorf("Sample %d: expected %v, got %v", i, expected[i], result)
		}
	}
}

func TestMockVAD_Reset(t *testing.T) {
	mock := NewMockVAD()

	pattern := []bool{true, false, false}
	mock.SetSpeechPattern(pattern)

	// Process some samples
	mock.ProcessSample(0.1) // Speech
	mock.ProcessSample(0.1) // Silence
	mock.ProcessSample(0.1) // Silence

	// Reset should clear state
	mock.Reset()

	if mock.IsSpeaking() {
		t.Error("Expected not speaking after reset")
	}

	if mock.GetSilenceDuration() != 0 {
		t.Errorf("Expected silence duration 0 after reset, got %d", mock.GetSilenceDuration())
	}
}
