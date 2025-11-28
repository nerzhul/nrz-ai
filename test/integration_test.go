package tests

import (
	"testing"

	"github.com/nerzhul/nrz-ai/internal/audio"
	"github.com/nerzhul/nrz-ai/internal/vad"
	"github.com/nerzhul/nrz-ai/internal/whisper"
)

func TestAudioProcessorIntegration(t *testing.T) {
	// Test audio processor avec vraie implémentation
	processor := audio.NewProcessor()
	
	// Test avec des données audio simulées
	testData := []byte{0x00, 0x00, 0x00, 0x3F, 0x00, 0x00, 0x80, 0xBF} // ~1.0 et ~-1.0
	samples := processor.ProcessBytes(testData)
	
	if len(samples) != 2 {
		t.Errorf("Expected 2 samples, got %d", len(samples))
	}
	
	rms := processor.CalculateRMS(samples, 10)
	if rms < 0.5 || rms > 1.5 {
		t.Errorf("Expected RMS around 0.7-1.0, got %.3f", rms)
	}
}

func TestMockIntegrationWorkflow(t *testing.T) {
	// Test du workflow complet avec mocks
	
	// 1. Audio capture mock
	testAudioData := []byte{0x00, 0x00, 0x00, 0x3F, 0x00, 0x00, 0x80, 0xBF}
	stream := audio.NewMockAudioStream(testAudioData)
	capture := audio.NewMockAudioCapture(stream)
	
	// 2. Audio processor réel
	processor := audio.NewProcessor()
	
	// 3. VAD mock
	vadDetector := vad.NewMockVAD()
	vadDetector.SetSpeechPattern([]bool{true, true, false, false}) // Speech puis silence
	
	// 4. Whisper mock
	whisperService := whisper.NewMockWhisperService()
	expectedResult := whisper.TranscriptionResult{
		Text:     "Test transcription",
		Language: "fr",
		Duration: 1.0,
	}
	whisperService.SetTranscribeResult(expectedResult)
	whisperService.LoadModel("test-model.bin")
	
	// Test du workflow
	audioStream, err := capture.StartCapture("test-source")
	if err != nil {
		t.Fatalf("Failed to start capture: %v", err)
	}
	defer audioStream.Close()
	
	// Lit des données audio
	buffer := make([]byte, len(testAudioData))
	n, err := audioStream.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read audio: %v", err)
	}
	
	// Traite les données audio
	samples := processor.ProcessBytes(buffer[:n])
	if len(samples) == 0 {
		t.Error("Expected audio samples to be processed")
	}
	
	// VAD processing
	config := vad.VADConfig{SampleRate: 16000}
	vadDetector.Initialize(config)
	
	for _, sample := range samples {
		vadDetector.ProcessSample(sample)
	}
	
	// Test transcription
	result, err := whisperService.Transcribe(samples, "fr")
	if err != nil {
		t.Fatalf("Failed to transcribe: %v", err)
	}
	
	if result.Text != expectedResult.Text {
		t.Errorf("Expected text '%s', got '%s'", expectedResult.Text, result.Text)
	}
}