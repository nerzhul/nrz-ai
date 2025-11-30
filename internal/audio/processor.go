package audio

import (
	"math"
	"unsafe"
)

// Processor implements AudioProcessor interface
type Processor struct{}

// NewProcessor creates a new audio processor
func NewProcessor() *Processor {
	return &Processor{}
}

// ProcessBytes converts raw audio bytes to float32 samples
func (p *Processor) ProcessBytes(data []byte) []float32 {
	samples := make([]float32, 0, len(data)/4)

	for i := 0; i < len(data); i += 4 {
		if i+4 <= len(data) {
			sample := p.float32FromBytes(data[i : i+4])
			samples = append(samples, sample)
		}
	}

	return samples
}

// CalculateRMS calculates RMS level from audio samples
func (p *Processor) CalculateRMS(samples []float32, windowSize int) float32 {
	if len(samples) == 0 {
		return 0.0
	}

	// Use the last windowSize samples or all samples if less available
	start := 0
	if len(samples) > windowSize {
		start = len(samples) - windowSize
	}

	var sum float32
	count := 0
	for i := start; i < len(samples); i++ {
		sum += samples[i] * samples[i]
		count++
	}

	if count == 0 {
		return 0.0
	}

	meanSquare := sum / float32(count)
	return float32(math.Sqrt(float64(meanSquare)))
}

// float32FromBytes converts 4 bytes to float32 (little-endian)
func (p *Processor) float32FromBytes(b []byte) float32 {
	bits := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	return *(*float32)(unsafe.Pointer(&bits))
}
