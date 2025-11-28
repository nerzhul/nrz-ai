package audio

import (
	"io"
	"os/exec"
)

// FFmpegStream implements AudioStream using FFmpeg
type FFmpegStream struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
}

// Read reads audio data from FFmpeg stdout
func (f *FFmpegStream) Read(data []byte) (int, error) {
	return f.stdout.Read(data)
}

// Close closes the FFmpeg stream
func (f *FFmpegStream) Close() error {
	if f.cmd != nil && f.cmd.Process != nil {
		f.cmd.Process.Kill()
	}
	if f.stdout != nil {
		return f.stdout.Close()
	}
	return nil
}

// FFmpegCapture implements AudioCapture using FFmpeg
type FFmpegCapture struct{}

// NewFFmpegCapture creates a new FFmpeg audio capture
func NewFFmpegCapture() *FFmpegCapture {
	return &FFmpegCapture{}
}

// StartCapture starts capturing audio from the specified source
func (f *FFmpegCapture) StartCapture(audioSource string) (AudioStream, error) {
	cmd := exec.Command("ffmpeg",
		"-f", "pulse",
		"-i", audioSource,
		"-ar", "16000",
		"-ac", "1",
		"-f", "f32le",
		"-loglevel", "quiet",
		"-")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &FFmpegStream{
		cmd:    cmd,
		stdout: stdout,
	}, nil
}

// Stop stops the audio capture (not used in streaming mode)
func (f *FFmpegCapture) Stop() error {
	return nil
}