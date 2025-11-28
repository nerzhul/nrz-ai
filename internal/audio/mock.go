package audio

import (
	"errors"
	"io"
)

// MockAudioStream implements AudioStream for testing
type MockAudioStream struct {
	data      []byte
	position  int
	readError error
	closed    bool
}

// NewMockAudioStream creates a mock audio stream with test data
func NewMockAudioStream(data []byte) *MockAudioStream {
	return &MockAudioStream{
		data:     data,
		position: 0,
	}
}

// SetReadError sets an error to return on Read calls
func (m *MockAudioStream) SetReadError(err error) {
	m.readError = err
}

// Read reads data from the mock stream
func (m *MockAudioStream) Read(p []byte) (int, error) {
	if m.closed {
		return 0, errors.New("stream closed")
	}
	
	if m.readError != nil {
		return 0, m.readError
	}
	
	if m.position >= len(m.data) {
		return 0, io.EOF
	}
	
	n := copy(p, m.data[m.position:])
	m.position += n
	return n, nil
}

// Close closes the mock stream
func (m *MockAudioStream) Close() error {
	m.closed = true
	return nil
}

// MockAudioCapture implements AudioCapture for testing
type MockAudioCapture struct {
	audioStream AudioStream
	startError  error
	stopError   error
}

// NewMockAudioCapture creates a mock audio capture
func NewMockAudioCapture(stream AudioStream) *MockAudioCapture {
	return &MockAudioCapture{
		audioStream: stream,
	}
}

// SetStartError sets an error to return on StartCapture calls
func (m *MockAudioCapture) SetStartError(err error) {
	m.startError = err
}

// SetStopError sets an error to return on Stop calls
func (m *MockAudioCapture) SetStopError(err error) {
	m.stopError = err
}

// StartCapture returns the configured mock stream
func (m *MockAudioCapture) StartCapture(audioSource string) (AudioStream, error) {
	if m.startError != nil {
		return nil, m.startError
	}
	return m.audioStream, nil
}

// Stop simulates stopping capture
func (m *MockAudioCapture) Stop() error {
	return m.stopError
}