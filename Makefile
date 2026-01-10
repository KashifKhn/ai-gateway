# AI Gateway

.PHONY: build run clean test

# Default target
all: build

# Build the gateway
build:
	go build -o ai-gateway ./cmd/server

# Run the gateway (assumes OpenCode is already running)
run: build
	CONFIG_PATH=config/config.yaml ./ai-gateway

# Run with OpenCode (starts both)
run-full:
	./scripts/start.sh

# Clean build artifacts
clean:
	rm -f ai-gateway

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Generate API key
generate-key:
	./scripts/generate-key.sh

# Build for Linux (for deployment)
build-linux:
	GOOS=linux GOARCH=amd64 go build -o ai-gateway-linux ./cmd/server

# Install dependencies
deps:
	go mod tidy
	go mod download

# Help
help:
	@echo "Available targets:"
	@echo "  build        - Build the gateway binary"
	@echo "  run          - Build and run the gateway"
	@echo "  run-full     - Start OpenCode and gateway together"
	@echo "  clean        - Remove build artifacts"
	@echo "  test         - Run tests"
	@echo "  fmt          - Format code"
	@echo "  generate-key - Generate a new API key"
	@echo "  build-linux  - Build for Linux deployment"
	@echo "  deps         - Install/update dependencies"
