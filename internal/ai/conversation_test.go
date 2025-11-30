package ai

import (
	"testing"
)

func TestNewConversation(t *testing.T) {
	conv := NewConversation(5)
	if conv == nil {
		t.Fatal("Expected conversation to be created")
	}

	if conv.GetMessageCount() != 0 {
		t.Errorf("Expected 0 messages, got %d", conv.GetMessageCount())
	}
}

func TestConversation_AddMessage(t *testing.T) {
	conv := NewConversation(3)

	msg1 := Message{Role: "user", Content: "Hello"}
	msg2 := Message{Role: "assistant", Content: "Hi there!"}

	conv.AddMessage(msg1)
	conv.AddMessage(msg2)

	messages := conv.GetMessages()
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	if messages[0].Content != "Hello" {
		t.Errorf("Expected first message 'Hello', got '%s'", messages[0].Content)
	}
}

func TestConversation_MaxHistory(t *testing.T) {
	conv := NewConversation(2)

	conv.AddMessage(Message{Role: "user", Content: "Message 1"})
	conv.AddMessage(Message{Role: "assistant", Content: "Response 1"})
	conv.AddMessage(Message{Role: "user", Content: "Message 2"})

	messages := conv.GetMessages()

	// Should keep the last 2 messages
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages (max history), got %d", len(messages))
	}

	if messages[0].Content != "Response 1" {
		t.Errorf("Expected 'Response 1', got '%s'", messages[0].Content)
	}
}

func TestConversation_SystemPrompt(t *testing.T) {
	conv := NewConversation(5)

	conv.SetSystemPrompt("You are a helpful assistant")
	conv.AddMessage(Message{Role: "user", Content: "Hello"})

	messages := conv.GetMessages()
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages (system + user), got %d", len(messages))
	}

	if messages[0].Role != "system" {
		t.Errorf("Expected system message first, got role '%s'", messages[0].Role)
	}

	if messages[0].Content != "You are a helpful assistant" {
		t.Errorf("Expected system prompt, got '%s'", messages[0].Content)
	}
}

func TestConversation_ClearHistory(t *testing.T) {
	conv := NewConversation(5)

	conv.SetSystemPrompt("You are a helpful assistant")
	conv.AddMessage(Message{Role: "user", Content: "Hello"})
	conv.AddMessage(Message{Role: "assistant", Content: "Hi!"})

	conv.ClearHistory()

	messages := conv.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message (system only) after clear, got %d", len(messages))
	}

	if messages[0].Role != "system" {
		t.Errorf("Expected system message to remain after clear, got role '%s'", messages[0].Role)
	}
}
