package whisper

import "testing"

func TestNewMockWhisperService(t *testing.T) {
	mock := NewMockWhisperService()
	if mock == nil {
		t.Fatal("Expected mock service to be created")
	}

	if mock.IsLoaded() {
		t.Error("Expected service to start unloaded")
	}
}

func TestMockWhisperService_LoadModel(t *testing.T) {
	mock := NewMockWhisperService()

	err := mock.LoadModel("test-model.bin")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !mock.IsLoaded() {
		t.Error("Expected model to be loaded")
	}
}

func TestMockWhisperService_Transcribe(t *testing.T) {
	mock := NewMockWhisperService()

	// Test transcribe without loaded model
	_, err := mock.Transcribe([]float32{0.1, 0.2}, "fr")
	if err == nil {
		t.Error("Expected error when model not loaded")
	}

	// Load model and test successful transcribe
	mock.LoadModel("test-model.bin")

	expectedResult := TranscriptionResult{
		Text:     "Bonjour le monde",
		Language: "fr",
		Duration: 2.0,
	}
	mock.SetTranscribeResult(expectedResult)

	result, err := mock.Transcribe([]float32{0.1, 0.2}, "fr")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.Text != expectedResult.Text {
		t.Errorf("Expected text '%s', got '%s'", expectedResult.Text, result.Text)
	}
}
