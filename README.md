# NRZ-AI - Real-time Speech-to-Text with Voice Activity Detection

🎙️ **Intelligent real-time French speech transcription** using Whisper.cpp with advanced Voice Activity Detection (VAD) for phrase-based processing.

## ✨ Features

- **🎯 Smart VAD**: RMS-based Voice Activity Detection with adaptive noise floor calibration
- **⚡ Real-time Processing**: Phrase-based transcription triggered by natural speech pauses
- **🧪 Testable Architecture**: Modular design with interfaces for easy unit testing and mocking
- **🔧 Configurable**: Environment-based configuration for models, languages, and audio sources
- **📊 GPU Support**: ROCm/HIP acceleration for AMD graphics cards
- **🎚️ Adaptive Thresholds**: Automatic noise floor detection and threshold adjustment

## 🏗️ Architecture

```
cmd/nrz-ai/main.go          # Main application with SpeechProcessor
├── internal/audio/         # Audio capture and processing
│   ├── interfaces.go       # AudioCapture, AudioStream, AudioProcessor interfaces
│   ├── ffmpeg.go          # FFmpeg-based audio capture implementation
│   ├── processor.go       # Audio processing (bytes → float32, RMS calculation)
│   └── mock.go            # Mock implementations for testing
├── internal/vad/           # Voice Activity Detection
│   ├── interfaces.go       # VoiceActivityDetector interface
│   ├── rms.go             # RMS-based VAD with adaptive noise floor
│   └── mock.go            # Mock VAD for testing
└── internal/whisper/       # Speech-to-text transcription
    ├── interfaces.go       # WhisperService interface
    ├── service.go         # Whisper.cpp integration
    └── mock.go            # Mock transcription for testing
```

## 📋 Prerequisites

### System Dependencies
```bash
# Ubuntu/Debian
sudo apt-get install -y ffmpeg pulseaudio-utils build-essential git cmake

# Arch/NixOS
sudo pacman -S ffmpeg pulseaudio base-devel git cmake
# or with nix: nix-shell -p ffmpeg pulseaudio cmake gcc
```

### GPU Support (Optional - AMD)
```bash
# For ROCm/HIP acceleration
sudo apt install rocm-dev hip-dev
# or equivalent for your distribution
```

## 🚀 Quick Start

### 1. Build Everything
```bash
# Download and build whisper.cpp, download model, build nrz-ai
make all
```

### 2. Run with Defaults
```bash
# Uses French, large-v3 model, default microphone
WHISPER_MODEL=./models/ggml-large-v3.bin ./dist/nrz-ai
```

## ⚙️ Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `WHISPER_MODEL` | `./models/ggml-large-v3.bin` | Path to Whisper model file |
| `WHISPER_LANGUAGE` | `fr` | Language code (fr, en, es, etc.) |
| `AUDIO_SOURCE` | `default` | PulseAudio source name |

### Available Models

| Model | Size | VRAM | Accuracy | Use Case |
|-------|------|------|----------|----------|
| `tiny` | 75 MB | ~390 MB | Basic | Quick testing |
| `base` | 142 MB | ~500 MB | Good | General use |
| `small` | 466 MB | ~852 MB | Better | Clear speech |
| `medium` | 1.5 GB | ~2.1 GB | Very good | Professional |
| **`large-v3`** | **3.1 GB** | **~4.2 GB** | **Excellent** | **Production (recommended)** |

## 🎤 Usage Examples

### Basic Usage
```bash
# Start with French transcription
./dist/nrz-ai

# English transcription with medium model
WHISPER_MODEL=./models/ggml-medium.bin WHISPER_LANGUAGE=en ./dist/nrz-ai

# Specific microphone device
AUDIO_SOURCE=alsa_input.usb-RODE_RODE_AI-Micro-00.analog-stereo ./dist/nrz-ai
```

### Finding Audio Sources
```bash
# List available PulseAudio sources
pactl list short sources

# Test microphone
ffmpeg -f pulse -i default -ar 16000 -ac 1 -t 5 test.wav
```

## 🧪 Development & Testing

### Build Individual Components
```bash
make whispercpp          # Build whisper.cpp only
make model              # Download model only  
make build              # Build nrz-ai only
```

### Run Tests
```bash
make test               # Unit tests
make coverage           # Tests with coverage report
make test-integration   # Integration tests
make test-all           # All tests
```

### Example Test Output
```bash
$ make test
🧪 Running unit tests...
ok      github.com/nerzhul/nrz-ai/internal/audio       0.002s  coverage: 85.7% of statements
ok      github.com/nerzhul/nrz-ai/internal/vad         0.001s  coverage: 92.3% of statements
✅ Unit tests completed
```

## 🔧 Troubleshooting

### Build Issues

**CGO linking errors:**
```bash
# Verify whisper.cpp is built
ls deps/whisper.cpp/build/src/libwhisper.so

# Clean and rebuild
make clean && make build
```

**Model loading errors:**
```bash
# Check model exists and has correct permissions
ls -la models/ggml-large-v3.bin

# Re-download model
rm models/ggml-large-v3.bin && make model
```

### Audio Issues

**No audio input:**
```bash
# Test microphone with FFmpeg
ffmpeg -f pulse -i default -ar 16000 -ac 1 -f f32le -t 2 - | wc -c
# Should output > 0

# Check PulseAudio
pulseaudio --check -v
```

**VAD not triggering:**
- Adjust `silenceThreshold` in `internal/vad/rms.go`
- Check noise floor calibration logs
- Verify microphone input levels

### Performance Issues

**High CPU usage:**
- Use smaller model (`medium` instead of `large-v3`)
- Enable GPU acceleration (AMD: install ROCm)
- Reduce `rmsWindowSize` or `sampleRate`

## 🎯 Voice Activity Detection

The VAD system uses sophisticated RMS-based detection:

1. **🎚️ Noise Floor Calibration** (2s): Measures background noise
2. **📊 Adaptive Thresholds**: Sets detection threshold to 3× noise floor
3. **🔄 Phrase Detection**: Triggers transcription after 800ms of silence
4. **⏰ Smart Timing**: Minimum 500ms speech before processing

### VAD Configuration
```go
// In internal/vad/interfaces.go
type VADConfig struct {
    SilenceThreshold    float32  // Base RMS threshold (0.01)
    SilenceDurationMs   int      // Silence to trigger (800ms)
    MinSpeechDurationMs int      // Min speech duration (500ms)
    RMSWindowSize       int      // RMS calculation window (160 = 10ms)
}
```

## 📊 Performance

### Benchmarks (AMD Ryzen + RX 7900)
- **Latency**: ~200-500ms from speech end to transcription
- **Accuracy**: 95%+ for clear French speech (large-v3)
- **CPU Usage**: ~15-25% (large-v3), ~8-12% (medium)
- **Memory**: ~4-6GB RAM + 4GB VRAM (large-v3)

## 🔄 Recent Updates

- ✅ **v2.0**: Complete architecture refactor with interfaces
- ✅ **RMS-based VAD**: Replaced simple amplitude detection
- ✅ **Adaptive thresholds**: Automatic noise floor calibration
- ✅ **Unit testing**: Mock interfaces for all components
- ✅ **Modular design**: Separate audio, VAD, and Whisper modules

## 🤝 Contributing

1. **Architecture**: Follow the interface-based design
2. **Testing**: Add tests for new features using mocks
3. **Performance**: Profile before optimizing
4. **Documentation**: Update README for API changes

### Adding New Audio Sources
```go
// Implement audio.AudioCapture interface
type MyAudioCapture struct{}
func (m *MyAudioCapture) StartCapture(source string) (audio.AudioStream, error) {
    // Your implementation
}
```

## 📝 License

MIT License - see LICENSE file

---

**💡 Pro Tip**: For best results, use a good microphone, minimize background noise, and let the system calibrate for 2 seconds before speaking.
