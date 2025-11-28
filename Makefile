.PHONY: whispercpp build clean help model test coverage test-integration test-all

WHISPER_DIR := deps/whisper.cpp
WHISPER_REPO := https://github.com/ggerganov/whisper.cpp.git
WHISPER_VERSION := v1.8.2
MODEL_DIR := models

help:
	@echo "Available targets:"
	@echo "  make whispercpp     - Clone and build whisper.cpp"
	@echo "  make model          - Download Whisper large-v3 model"
	@echo "  make build          - Build nrz-ai binary"
	@echo "  make test           - Run unit tests"
	@echo "  make test-integration - Run integration tests"
	@echo "  make test-all       - Run all tests"
	@echo "  make coverage       - Run tests with coverage report"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make cleanall       - Remove everything including whisper.cpp"

whispercpp:
	@if [ ! -d "$(WHISPER_DIR)" ]; then \
		echo "🔽 Cloning whisper.cpp $(WHISPER_VERSION)..."; \
		mkdir -p deps; \
		git clone --branch $(WHISPER_VERSION) --depth 1 $(WHISPER_REPO) $(WHISPER_DIR); \
	else \
		echo "✅ whisper.cpp already exists"; \
	fi
	@echo "🔨 Building whisper.cpp with GPU support (ROCm/HIP)..."
	@cd $(WHISPER_DIR) && cmake -B build \
		-DCMAKE_BUILD_TYPE=Release \
		-DBUILD_SHARED_LIBS=ON \
		-DGGML_HIP=ON \
		-DCMAKE_C_COMPILER=/opt/rocm/llvm/bin/clang \
		-DCMAKE_CXX_COMPILER=/opt/rocm/llvm/bin/clang++ \
		-DGPU_TARGETS=gfx1100 && \
	cmake --build build --config Release
	@echo "✅ whisper.cpp built successfully with GPU support"

model:
	@mkdir -p $(MODEL_DIR)
	@if [ ! -f "$(MODEL_DIR)/ggml-large-v3.bin" ]; then \
		echo "📥 Downloading Whisper large-v3 model (~3.1 GB)..."; \
		cd $(WHISPER_DIR) && bash models/download-ggml-model.sh large-v3 && \
		cp models/ggml-large-v3.bin ../../$(MODEL_DIR)/; \
		echo "✅ Model downloaded to $(MODEL_DIR)/ggml-large-v3.bin"; \
	else \
		echo "✅ Model already exists"; \
	fi

build: whispercpp
	@echo "🔨 Building nrz-ai..."
	@export CGO_LDFLAGS="-L$(PWD)/$(WHISPER_DIR)/build/src -L$(PWD)/$(WHISPER_DIR)/build/ggml/src -lwhisper -lggml -Wl,-rpath,$(PWD)/$(WHISPER_DIR)/build/src -Wl,-rpath,$(PWD)/$(WHISPER_DIR)/build/ggml/src -Wl,-rpath,/opt/rocm/lib" && \
	 export CGO_CFLAGS="-I$(PWD)/$(WHISPER_DIR)/include -I$(PWD)/$(WHISPER_DIR)/ggml/include -I/opt/rocm/include" && \
	 mkdir -p dist && \
	 go build -o dist/nrz-ai ./cmd/nrz-ai
	@echo "✅ nrz-ai built successfully"

# Run unit tests
test:
	@echo "🧪 Running unit tests..."
	@export CGO_LDFLAGS="-L$(PWD)/$(WHISPER_DIR)/build/src -L$(PWD)/$(WHISPER_DIR)/build/ggml/src -lwhisper -lggml -Wl,-rpath,$(PWD)/$(WHISPER_DIR)/build/src -Wl,-rpath,$(PWD)/$(WHISPER_DIR)/build/ggml/src -Wl,-rpath,/opt/rocm/lib" && \
	 export CGO_CFLAGS="-I$(PWD)/$(WHISPER_DIR)/include -I$(PWD)/$(WHISPER_DIR)/ggml/include -I/opt/rocm/include" && \
	 go test -v ./...
	@echo "✅ Unit tests completed"

# Run tests with coverage
coverage:
	@echo "🧪 Running tests with coverage..."
	@mkdir -p coverage
	@export CGO_LDFLAGS="-L$(PWD)/$(WHISPER_DIR)/build/src -L$(PWD)/$(WHISPER_DIR)/build/ggml/src -lwhisper -lggml -Wl,-rpath,$(PWD)/$(WHISPER_DIR)/build/src -Wl,-rpath,$(PWD)/$(WHISPER_DIR)/build/ggml/src -Wl,-rpath,/opt/rocm/lib" && \
	 export CGO_CFLAGS="-I$(PWD)/$(WHISPER_DIR)/include -I$(PWD)/$(WHISPER_DIR)/ggml/include -I/opt/rocm/include" && \
	 go test -v -coverprofile=coverage/coverage.out ./...
	@go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "✅ Coverage report generated: coverage/coverage.html"

# Run integration tests  
test-integration:
	@echo "🧪 Running integration tests..."
	@export CGO_LDFLAGS="-L$(PWD)/$(WHISPER_DIR)/build/src -L$(PWD)/$(WHISPER_DIR)/build/ggml/src -lwhisper -lggml -Wl,-rpath,$(PWD)/$(WHISPER_DIR)/build/src -Wl,-rpath,$(PWD)/$(WHISPER_DIR)/build/ggml/src -Wl,-rpath,/opt/rocm/lib" && \
	 export CGO_CFLAGS="-I$(PWD)/$(WHISPER_DIR)/include -I$(PWD)/$(WHISPER_DIR)/ggml/include -I/opt/rocm/include" && \
	 go test -v ./cmd/nrz-ai/...
	@echo "✅ Integration tests completed"

# Run all tests
test-all: test test-integration

clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -f dist/nrz-ai
	@echo "✅ Clean complete"

cleanall: clean
	@echo "🧹 Removing whisper.cpp..."
	@rm -rf $(WHISPER_DIR)
	@echo "✅ Clean all complete"
