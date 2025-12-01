# NRZ-AI - Real-time Speech-to-Text with AI Conversation

🎙️ **Intelligent real-time speech transcription** using Whisper.cpp with advanced Voice Activity Detection (VAD) and optional AI conversation capabilities via Ollama.

## ✨ Features

- **🎯 Smart VAD**: RMS-based Voice Activity Detection with adaptive noise floor calibration
- **🔍 Wake Word Detection**: Optional privacy mode - activate listening only with "Jack" (configurable)
- **⚡ Real-time Processing**: Phrase-based transcription triggered by natural speech pauses
- **🤖 AI Conversation**: Optional integration with Ollama for intelligent responses to voice input
- **🧪 Testable Architecture**: Modular design with interfaces for easy unit testing and mocking
- **💬 Professional CLI**: Cobra-based command line interface with comprehensive options
- **📊 GPU Support**: ROCm/HIP acceleration for AMD graphics cards (CPU-only build available)
- **🎚️ Adaptive Thresholds**: Automatic noise floor detection and threshold adjustment
- **🛠️ Utility Commands**: Built-in tools for testing audio and listing AI models

## 🏗️ Architecture

```
cmd/nrz-ai/main.go          # Main application with Cobra CLI and SpeechProcessor
├── internal/audio/         # Audio capture and processing
│   ├── interfaces.go       # AudioCapture, AudioStream, AudioProcessor interfaces
│   ├── ffmpeg.go          # FFmpeg-based audio capture implementation
│   ├── processor.go       # Audio processing (bytes → float32, RMS calculation)
│   └── mock.go            # Mock implementations for testing
├── internal/vad/           # Voice Activity Detection
│   ├── interfaces.go       # VoiceActivityDetector interface
│   ├── rms.go             # RMS-based VAD with adaptive noise floor
│   └── mock.go            # Mock VAD for testing
├── internal/whisper/       # Speech-to-text transcription
│   ├── interfaces.go       # WhisperService interface
│   ├── service.go         # Whisper.cpp integration
│   └── mock.go            # Mock transcription for testing
└── internal/ai/            # AI conversation service
    ├── interfaces.go       # AIService, ConversationManager interfaces
    ├── ollama.go          # Ollama HTTP client implementation
    ├── conversation.go    # Thread-safe conversation management
    └── mock.go            # Mock AI service for testing
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

### Optional: Ollama for AI Conversation
```bash
# Install Ollama (for AI features)
curl -fsSL https://ollama.ai/install.sh | sh

# Start Ollama service
ollama serve

# Pull recommended model
ollama pull llama3.2:3b
```

### Optional: GPU Support (AMD)
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

### 2. Basic Usage
```bash
# Simple speech-to-text (French)
./dist/nrz-ai

# English transcription with medium model  
./dist/nrz-ai --language en --model ./models/ggml-medium.bin

# Enable AI conversation with default settings
./dist/nrz-ai --ai

# Custom AI setup
./dist/nrz-ai --ai --ollama-model llama3.2:1b --language en
```

### 3. Utility Commands
```bash
# Test your microphone
./dist/nrz-ai test-audio

# List available Ollama models
./dist/nrz-ai list-models --ollama-url http://localhost:11434
```

## ⚙️ Configuration

### Command Line Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--model` | `-m` | `./models/ggml-large-v3.bin` | Path to Whisper model file |
| `--language` | `-l` | `fr` | Language code (fr, en, es, etc.) |
| `--audio-source` | `-a` | `default` | PulseAudio source name |
| `--wake-word` | `-w` | `false` | Enable wake word detection |
| `--wake-word-text` | | `Jack` | Custom wake word to activate listening |
| `--ai` | | `false` | Enable AI conversation |
| `--ollama-url` | | `http://localhost:11434` | Ollama server URL |
| `--ollama-model` | | `llama3.2:3b` | Ollama model to use |
| `--system-prompt` | | French assistant prompt | AI system prompt |
| `--max-history` | | `10` | Max conversation messages to keep |
| `--verbose` | `-v` | `false` | Enable verbose logging |

### Subcommands

| Command | Description |
|---------|-------------|
| `list-models` | List available Ollama models |
| `test-audio` | Test microphone input for 3 seconds |

### Available Models

| Model | Size | VRAM | Accuracy | Use Case |
|-------|------|------|----------|----------|
| `tiny` | 75 MB | ~390 MB | Basic | Quick testing |
| `base` | 142 MB | ~500 MB | Good | General use |
| `small` | 466 MB | ~852 MB | Better | Clear speech |
| `medium` | 1.5 GB | ~2.1 GB | Very good | Professional |
| **`large-v3`** | **3.1 GB** | **~4.2 GB** | **Excellent** | **Production (recommended)** |

## 🎤 Usage Examples

### Speech-to-Text Only
```bash
# French transcription (default)
./dist/nrz-ai

# English with medium model
./dist/nrz-ai --language en --model ./models/ggml-medium.bin

# Specific microphone device
./dist/nrz-ai --audio-source alsa_input.usb-RODE_RODE_AI-Micro-00.analog-stereo
```

### Wake Word Mode (Privacy)
```bash
# Enable wake word detection with default "Jack"
./dist/nrz-ai --wake-word

# Custom wake word
./dist/nrz-ai --wake-word --wake-word-text "Ordinateur"

# Wake word with AI conversation
./dist/nrz-ai --wake-word --ai

# English wake word mode
./dist/nrz-ai --wake-word --wake-word-text "Computer" --language en
```
```bash
# Enable AI with defaults (French assistant)
./dist/nrz-ai --ai

# English AI conversation
./dist/nrz-ai --ai --language en --system-prompt "You are a helpful English assistant."

# Custom Ollama setup
./dist/nrz-ai --ai --ollama-url http://192.168.1.100:11434 --ollama-model llama3.2:1b
```

### Utility Commands
```bash
# Test microphone for 3 seconds
./dist/nrz-ai test-audio --audio-source default

# List available AI models
./dist/nrz-ai list-models

# Get help for any command
./dist/nrz-ai --help
./dist/nrz-ai list-models --help
```

### Finding Audio Sources
```bash
# List available PulseAudio sources
pactl list short sources

# Test microphone manually
ffmpeg -f pulse -i default -ar 16000 -ac 1 -t 5 test.wav

# Or use built-in test
./dist/nrz-ai test-audio
```
### Finding Audio Sources
```bash
# List available PulseAudio sources
pactl list short sources

# Test microphone manually
ffmpeg -f pulse -i default -ar 16000 -ac 1 -t 5 test.wav

# Or use built-in test
./dist/nrz-ai test-audio
```

### Example Wake Word + AI Output
```
🎙️  NRZ-AI - Real-time Speech-to-Text
📦 Whisper model: ./models/ggml-large-v3.bin
🎤 Audio source: default
🗣️  Language: fr
🔍 Wake word: Jack (listening mode)
🤖 AI Service: Ollama (http://localhost:11434)
🧠 Model: llama3.2:3b
✅ AI service connected successfully
🎯 Say 'Jack' to activate listening, then speak normally
─────────────────────────────────────────────
🔍 Listening for wake word 'Jack'...
🎯 Wake word 'Jack' detected! Activating listening...
[15:04:12] 🎤 Bonjour, comment ça va ?
[15:04:13] 🤖 Bonjour ! Je vais très bien, merci. Comment puis-je vous aider ?
🔍 Listening timeout. Waiting for wake word 'Jack' again...
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
# Test with built-in command
./dist/nrz-ai test-audio

# Manual test
ffmpeg -f pulse -i default -ar 16000 -ac 1 -f f32le -t 2 - | wc -c
# Should output > 0

# Check PulseAudio
pulseaudio --check -v
```

**VAD not triggering:**
- Adjust `silenceThreshold` in `internal/vad/rms.go`
- Check noise floor calibration logs with `--verbose`
- Verify microphone input levels

### AI Issues

**Ollama connection failed:**
```bash
# Check Ollama is running
curl -f http://localhost:11434/api/version

# Start Ollama
ollama serve

# Install required model
ollama pull llama3.2:3b

# Test with list-models command
./dist/nrz-ai list-models
```

**AI responses too slow:**
- Use smaller model (`llama3.2:1b` instead of `3b`)
- Check Ollama server resources
- Reduce `--max-history` parameter

### Performance Issues

**High CPU usage:**
- Use smaller model (`medium` instead of `large-v3`)
- Enable GPU acceleration (AMD: install ROCm)
- Reduce `rmsWindowSize` or `sampleRate` in constants

**AI conversation lag:**
- Use smaller Ollama model (`llama3.2:1b`)
- Reduce conversation history: `--max-history 5`
- Check Ollama server performance

## 🔐 Wake Word Detection

The wake word feature provides **privacy-first voice activation**. When enabled, the system only processes full transcription after detecting the configured wake word.

### How It Works

1. **🔍 Continuous Monitoring**: Listens for the wake word using small audio buffers (2 seconds)
2. **🎯 Wake Word Detection**: Uses Whisper to detect the configured word (default: "Jack")
3. **⚡ Activation**: Once detected, enables full speech processing for 30 seconds
4. **🔒 Timeout**: Returns to wake word mode after 30 seconds of inactivity

### Configuration Options

```bash
# Basic wake word (default: "Jack")
./dist/nrz-ai --wake-word

# Custom wake word in French
./dist/nrz-ai --wake-word --wake-word-text "Ordinateur"

# English wake word
./dist/nrz-ai --wake-word --wake-word-text "Computer" --language en

# Wake word + AI conversation
./dist/nrz-ai --wake-word --ai --wake-word-text "Assistant"
```

### Privacy Benefits

- **🔒 No always-on transcription**: Only processes speech after wake word
- **📱 Lower resource usage**: Minimal processing when waiting for wake word
- **🎛️ User control**: Explicit activation prevents accidental recordings
- **⏱️ Auto-timeout**: Automatically returns to private mode after inactivity

### Performance Notes

- Wake word detection uses short 0.5-2 second audio buffers
- Detection runs every 500ms to balance responsiveness and CPU usage
- Uses the same Whisper model as main transcription for accuracy

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

- ✅ **v3.1**: Wake word detection for privacy-first voice activation
- ✅ **v3.0**: Complete CLI rewrite with Cobra framework  
- ✅ **v2.5**: AI conversation integration with Ollama
- ✅ **v2.1**: Conversation management with history limits
- ✅ **v2.0**: Complete architecture refactor with interfaces
- ✅ **Utility commands**: Built-in audio testing and model listing
- ✅ **Professional CLI**: Comprehensive flags and help system
- ✅ **RMS-based VAD**: Replaced simple amplitude detection
- ✅ **Adaptive thresholds**: Automatic noise floor calibration
- ✅ **Unit testing**: Mock interfaces for all components

## 🤝 Contributing

1. **Architecture**: Follow the interface-based design
2. **Testing**: Add tests for new features using mocks
3. **Performance**: Profile before optimizing
4. **Documentation**: Update README for API changes

### Adding New AI Backends
```go
// Implement ai.AIService interface
type MyAIService struct{}
func (m *MyAIService) Chat(request ai.ChatRequest) (ai.ChatResponse, error) {
    // Your implementation
}
func (m *MyAIService) IsAvailable() bool { /* ... */ }
```

### Adding Conversation Features
```go
// Extend ai.ConversationManager interface for new features
type EnhancedConversation struct {
    ai.ConversationManager
    // Additional fields
}
```

## 📝 License

MIT License - see LICENSE file

---

**💡 Pro Tip**: Start with `./dist/nrz-ai --ai` for the full voice conversation experience. Use smaller models (`llama3.2:1b`) for faster responses on lower-end hardware.

**🎙️ Voice Tip**: Speak naturally and pause briefly between thoughts - the VAD will detect when you're finished and trigger both transcription and AI response automatically.
