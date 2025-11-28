# NRZ-AI - Real-time Speech-to-Text

Go application for real-time audio transcription using Whisper.cpp by ggerganov.

## 📋 Prerequisites

### System
```bash
# Ubuntu/Debian
sudo apt-get install -y ffmpeg pulseaudio-utils build-essential git

# Arch Linux  
sudo pacman -S ffmpeg pulseaudio base-devel git
```

### Installing Whisper.cpp

```bash
# Clone and compile whisper.cpp
git clone https://github.com/ggerganov/whisper.cpp.git ~/whisper.cpp
cd ~/whisper.cpp
make

# Set CGO variables
export CGO_LDFLAGS="-L$HOME/whisper.cpp"
export CGO_CFLAGS="-I$HOME/whisper.cpp"
```

### Download a Whisper Model

```bash
# Create models directory
mkdir -p models

# Download base model (142 MB)
bash ~/whisper.cpp/models/download-ggml-model.sh base
cp ~/whisper.cpp/models/ggml-base.bin models/

# Or small (466 MB) for better accuracy
bash ~/whisper.cpp/models/download-ggml-model.sh small
cp ~/whisper.cpp/models/ggml-small.bin models/
```

## 🚀 Building

```bash
# Required CGO variables
export CGO_LDFLAGS="-L$HOME/whisper.cpp"
export CGO_CFLAGS="-I$HOME/whisper.cpp"

# Install Go dependencies
go mod tidy

# Build
go build -o nrz-ai ./cmd/nrz-ai
```

## ⚙️ Configuration

Environment variables:

```bash
# Whisper model to use (default: ./models/ggml-base.bin)
export WHISPER_MODEL="./models/ggml-base.bin"

# Transcription language (default: fr)
export WHISPER_LANGUAGE="fr"
```

## 🎤 Usage

```bash
# Start real-time transcription
./nrz-ai

# Or with custom parameters
WHISPER_MODEL="./models/ggml-small.bin" WHISPER_LANGUAGE="en" ./nrz-ai
```

Transcription is displayed in real-time every 3 seconds. Press **Ctrl+C** to stop.

## 🎯 Available Models

| Model | Size | Quality | Recommended Use |
|-------|------|---------|----------------|
| `tiny` | 75 MB | Basic | Quick tests |
| `base` | 142 MB | Good | General use |
| `small` | 466 MB | Better | Conversations |
| `medium` | 1.5 GB | Very good | Professional transcription |
| `large` | 2.9 GB | Excellent | Maximum accuracy |

## 🔧 Troubleshooting

### Error "whisper.h: No such file or directory"

```bash
export CGO_LDFLAGS="-L$HOME/whisper.cpp"
export CGO_CFLAGS="-I$HOME/whisper.cpp"
go build ./cmd/nrz-ai
```

### FFmpeg Error

Test your microphone:
```bash
ffmpeg -f pulse -i default -t 5 test.wav
```

## 📝 License

MIT
