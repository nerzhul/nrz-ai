package ai

import "testing"

func TestMockAIService_Chat(t *testing.T) {
	mock := NewMockAIService()

	request := ChatRequest{
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		Model: "test-model",
	}

	response, err := mock.Chat(request)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if response.Model != "mock-model" {
		t.Errorf("Expected model 'mock-model', got '%s'", response.Model)
	}

	if response.Message.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", response.Message.Role)
	}
}

func TestMockAIService_SetResponses(t *testing.T) {
	mock := NewMockAIService()

	expectedResponses := []ChatResponse{
		{
			Model:   "test-model",
			Message: Message{Role: "assistant", Content: "Custom response 1"},
			Done:    true,
		},
		{
			Model:   "test-model",
			Message: Message{Role: "assistant", Content: "Custom response 2"},
			Done:    true,
		},
	}

	mock.SetResponses(expectedResponses)

	request := ChatRequest{
		Messages: []Message{{Role: "user", Content: "Test"}},
	}

	// First call should return first response
	response1, err := mock.Chat(request)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if response1.Message.Content != "Custom response 1" {
		t.Errorf("Expected 'Custom response 1', got '%s'", response1.Message.Content)
	}

	// Second call should return second response
	response2, err := mock.Chat(request)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if response2.Message.Content != "Custom response 2" {
		t.Errorf("Expected 'Custom response 2', got '%s'", response2.Message.Content)
	}
}

func TestMockAIService_ListModels(t *testing.T) {
	mock := NewMockAIService()

	models, err := mock.ListModels()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(models) != 2 {
		t.Errorf("Expected 2 default models, got %d", len(models))
	}

	if models[0] != "mock-model-1" {
		t.Errorf("Expected 'mock-model-1', got '%s'", models[0])
	}
}