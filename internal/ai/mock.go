package ai

// MockAIService implements AIService for testing
type MockAIService struct {
	responses     []ChatResponse
	currentIndex  int
	models        []string
	isAvailable   bool
	chatError     error
	streamError   error
	modelsError   error
}

// NewMockAIService creates a new mock AI service
func NewMockAIService() *MockAIService {
	return &MockAIService{
		responses:   make([]ChatResponse, 0),
		models:      []string{"mock-model-1", "mock-model-2"},
		isAvailable: true,
	}
}

// SetResponses sets the responses to return for Chat calls
func (m *MockAIService) SetResponses(responses []ChatResponse) {
	m.responses = responses
	m.currentIndex = 0
}

// SetModels sets the models to return for ListModels calls
func (m *MockAIService) SetModels(models []string) {
	m.models = models
}

// SetAvailable sets whether the service appears available
func (m *MockAIService) SetAvailable(available bool) {
	m.isAvailable = available
}

// SetChatError sets an error to return for Chat calls
func (m *MockAIService) SetChatError(err error) {
	m.chatError = err
}

// SetStreamError sets an error to return for ChatStream calls
func (m *MockAIService) SetStreamError(err error) {
	m.streamError = err
}

// SetModelsError sets an error to return for ListModels calls
func (m *MockAIService) SetModelsError(err error) {
	m.modelsError = err
}

// Chat returns the next mock response or error
func (m *MockAIService) Chat(request ChatRequest) (ChatResponse, error) {
	if m.chatError != nil {
		return ChatResponse{}, m.chatError
	}

	if len(m.responses) == 0 {
		return ChatResponse{
			Model: "mock-model",
			Message: Message{
				Role:    "assistant",
				Content: "Mock response for: " + request.Messages[len(request.Messages)-1].Content,
			},
			Done: true,
		}, nil
	}

	if m.currentIndex >= len(m.responses) {
		m.currentIndex = 0 // Loop back to start
	}

	response := m.responses[m.currentIndex]
	m.currentIndex++
	return response, nil
}

// ChatStream returns a mock streaming response
func (m *MockAIService) ChatStream(request ChatRequest) (<-chan ChatResponse, error) {
	if m.streamError != nil {
		return nil, m.streamError
	}

	responseChan := make(chan ChatResponse, 1)

	go func() {
		defer close(responseChan)

		if len(m.responses) == 0 {
			// Default mock streaming response
			responseChan <- ChatResponse{
				Model: "mock-model",
				Message: Message{
					Role:    "assistant",
					Content: "Mock streaming response",
				},
				Done: true,
			}
			return
		}

		// Send all configured responses
		for _, response := range m.responses {
			responseChan <- response
		}
	}()

	return responseChan, nil
}

// ListModels returns the configured mock models
func (m *MockAIService) ListModels() ([]string, error) {
	if m.modelsError != nil {
		return nil, m.modelsError
	}
	return m.models, nil
}

// IsAvailable returns the configured availability
func (m *MockAIService) IsAvailable() bool {
	return m.isAvailable
}

// Close simulates closing the service
func (m *MockAIService) Close() error {
	return nil
}

// MockConversationManager implements ConversationManager for testing
type MockConversationManager struct {
	messages     []Message
	systemPrompt string
}

// NewMockConversationManager creates a new mock conversation manager
func NewMockConversationManager() *MockConversationManager {
	return &MockConversationManager{
		messages: make([]Message, 0),
	}
}

// AddMessage adds a message to the mock conversation
func (m *MockConversationManager) AddMessage(message Message) {
	m.messages = append(m.messages, message)
}

// GetMessages returns all messages in the mock conversation
func (m *MockConversationManager) GetMessages() []Message {
	return m.messages
}

// ClearHistory clears the mock conversation history
func (m *MockConversationManager) ClearHistory() {
	m.messages = make([]Message, 0)
}

// SetSystemPrompt sets the mock system prompt
func (m *MockConversationManager) SetSystemPrompt(prompt string) {
	m.systemPrompt = prompt
}