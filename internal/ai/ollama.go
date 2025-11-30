package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaService implements AIService for Ollama backend
type OllamaService struct {
	baseURL    string
	httpClient *http.Client
	model      string
	timeout    time.Duration
}

// NewOllamaService creates a new Ollama service
func NewOllamaService(baseURL, model string) *OllamaService {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3.2:3b"
	}

	return &OllamaService{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		model:   model,
		timeout: 30 * time.Second,
	}
}

// Chat sends a message to Ollama and returns the response
func (o *OllamaService) Chat(request ChatRequest) (ChatResponse, error) {
	request.Model = o.model
	request.Stream = false

	reqBody, err := json.Marshal(request)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := o.httpClient.Post(
		fmt.Sprintf("%s/api/chat", o.baseURL),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return ChatResponse{}, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var response ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return ChatResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return response, nil
}

// ChatStream sends a message and returns a streaming response
func (o *OllamaService) ChatStream(request ChatRequest) (<-chan ChatResponse, error) {
	request.Model = o.model
	request.Stream = true

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := o.httpClient.Post(
		fmt.Sprintf("%s/api/chat", o.baseURL),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	responseChan := make(chan ChatResponse)

	go func() {
		defer close(responseChan)
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		for {
			var response ChatResponse
			if err := decoder.Decode(&response); err != nil {
				if err != io.EOF {
					response.Error = fmt.Sprintf("decode error: %v", err)
					responseChan <- response
				}
				return
			}

			responseChan <- response

			if response.Done {
				return
			}
		}
	}()

	return responseChan, nil
}

// ListModels returns available models from Ollama
func (o *OllamaService) ListModels() ([]string, error) {
	resp, err := o.httpClient.Get(fmt.Sprintf("%s/api/tags", o.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to get models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d", resp.StatusCode)
	}

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]string, len(result.Models))
	for i, model := range result.Models {
		models[i] = model.Name
	}

	return models, nil
}

// IsAvailable checks if Ollama is running and accessible
func (o *OllamaService) IsAvailable() bool {
	resp, err := o.httpClient.Get(fmt.Sprintf("%s/api/tags", o.baseURL))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// Close closes the HTTP client (no-op for this implementation)
func (o *OllamaService) Close() error {
	return nil
}

// SetModel changes the model used for requests
func (o *OllamaService) SetModel(model string) {
	o.model = model
}

// GetModel returns the current model
func (o *OllamaService) GetModel() string {
	return o.model
}

// SetTimeout sets the request timeout
func (o *OllamaService) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
	o.httpClient.Timeout = timeout
}