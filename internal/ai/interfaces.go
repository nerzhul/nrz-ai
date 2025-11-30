package ai

// Message represents a single message in a conversation
type Message struct {
	Role    string `json:"role"`    // "user", "assistant", "system"
	Content string `json:"content"` // Message content
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model"`
	Stream      bool      `json:"stream,omitempty"`
	Temperature float32   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	Model     string    `json:"model"`
	Message   Message   `json:"message"`
	Done      bool      `json:"done"`
	Error     string    `json:"error,omitempty"`
	CreatedAt string    `json:"created_at,omitempty"`
	Context   []int     `json:"context,omitempty"`
}

// AIService interface for AI backend services
type AIService interface {
	// Chat sends a message to the AI and returns the response
	Chat(request ChatRequest) (ChatResponse, error)
	
	// ChatStream sends a message and returns a streaming response
	ChatStream(request ChatRequest) (<-chan ChatResponse, error)
	
	// ListModels returns available models
	ListModels() ([]string, error)
	
	// IsAvailable checks if the service is available
	IsAvailable() bool
	
	// Close closes any connections
	Close() error
}

// ConversationManager handles conversation context
type ConversationManager interface {
	// AddMessage adds a message to the conversation
	AddMessage(message Message)
	
	// GetMessages returns all messages in the conversation
	GetMessages() []Message
	
	// ClearHistory clears the conversation history
	ClearHistory()
	
	// SetSystemPrompt sets the system prompt
	SetSystemPrompt(prompt string)
}