package ai

import (
	"sync"
)

// Conversation implements ConversationManager
type Conversation struct {
	messages     []Message
	systemPrompt string
	mutex        sync.RWMutex
	maxHistory   int
}

// NewConversation creates a new conversation manager
func NewConversation(maxHistory int) *Conversation {
	if maxHistory <= 0 {
		maxHistory = 50 // Default to last 50 messages
	}
	
	return &Conversation{
		messages:   make([]Message, 0),
		maxHistory: maxHistory,
	}
}

// AddMessage adds a message to the conversation
func (c *Conversation) AddMessage(message Message) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.messages = append(c.messages, message)

	// Keep only the last maxHistory messages (excluding system prompt)
	if len(c.messages) > c.maxHistory {
		// Find system message if it exists
		systemIndex := -1
		for i, msg := range c.messages {
			if msg.Role == "system" {
				systemIndex = i
				break
			}
		}

		if systemIndex >= 0 {
			// Keep system message + last maxHistory-1 messages
			systemMsg := c.messages[systemIndex]
			c.messages = append([]Message{systemMsg}, c.messages[len(c.messages)-c.maxHistory+1:]...)
		} else {
			// No system message, just keep last maxHistory messages
			c.messages = c.messages[len(c.messages)-c.maxHistory:]
		}
	}
}

// GetMessages returns all messages in the conversation
func (c *Conversation) GetMessages() []Message {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Create a copy to avoid race conditions
	messages := make([]Message, len(c.messages))
	copy(messages, c.messages)
	
	return messages
}

// ClearHistory clears the conversation history but keeps system prompt
func (c *Conversation) ClearHistory() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Find and keep only the system message
	for _, msg := range c.messages {
		if msg.Role == "system" {
			c.messages = []Message{msg}
			return
		}
	}
	
	// No system message found, clear everything
	c.messages = make([]Message, 0)
}

// SetSystemPrompt sets or updates the system prompt
func (c *Conversation) SetSystemPrompt(prompt string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.systemPrompt = prompt
	
	// Remove existing system message if any
	for i, msg := range c.messages {
		if msg.Role == "system" {
			c.messages = append(c.messages[:i], c.messages[i+1:]...)
			break
		}
	}

	// Add new system message at the beginning
	if prompt != "" {
		systemMsg := Message{
			Role:    "system",
			Content: prompt,
		}
		c.messages = append([]Message{systemMsg}, c.messages...)
	}
}

// GetSystemPrompt returns the current system prompt
func (c *Conversation) GetSystemPrompt() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.systemPrompt
}

// GetMessageCount returns the number of messages
func (c *Conversation) GetMessageCount() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.messages)
}