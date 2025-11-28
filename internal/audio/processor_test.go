package audio

import (
	"testing"
	"unsafe"
)

func TestNewProcessor(t *testing.T) {
	processor := NewProcessor()
	if processor == nil {
		t.Fatal("Expected processor to be created")
	}
}

func TestProcessor_ProcessBytes(t *testing.T) {
	processor := NewProcessor()
	
	// Create test data: 2 float32 samples (8 bytes)
	sample1 := float32(0.5)
	sample2 := float32(-0.3)
	
	data := make([]byte, 8)
	// Convert float32 to bytes (little-endian)
	bits1 := *(*uint32)(unsafe.Pointer(&sample1))
	bits2 := *(*uint32)(unsafe.Pointer(&sample2))
	
	data[0] = byte(bits1)
	data[1] = byte(bits1 >> 8)
	data[2] = byte(bits1 >> 16)
	data[3] = byte(bits1 >> 24)
	
	data[4] = byte(bits2)
	data[5] = byte(bits2 >> 8)
	data[6] = byte(bits2 >> 16)
	data[7] = byte(bits2 >> 24)
	
	samples := processor.ProcessBytes(data)
	
	if len(samples) != 2 {
		t.Errorf("Expected 2 samples, got %d", len(samples))
	}
	
	if abs(samples[0]-sample1) > 0.0001 {
		t.Errorf("Expected sample1 %.6f, got %.6f", sample1, samples[0])
	}
	
	if abs(samples[1]-sample2) > 0.0001 {
		t.Errorf("Expected sample2 %.6f, got %.6f", sample2, samples[1])
	}
}

func TestProcessor_CalculateRMS_EmptySlice(t *testing.T) {
	processor := NewProcessor()
	
	samples := []float32{}
	rms := processor.CalculateRMS(samples, 10)
	
	if rms != 0.0 {
		t.Errorf("Expected RMS 0.0 for empty slice, got %.6f", rms)
	}
}

func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}