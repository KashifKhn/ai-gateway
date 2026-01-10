.PHONY: build run clean test docker-build docker-up docker-down deploy

all: build

build:
	go build -o ai-gateway ./cmd/server

run: build
	CONFIG_PATH=config/config.yaml ./ai-gateway

run-full:
	./scripts/start.sh

clean:
	rm -f ai-gateway ai-gateway-linux

test:
	go test -v ./...

fmt:
	go fmt ./...

lint:
	golangci-lint run

generate-key:
	./scripts/generate-key.sh

build-linux:
	GOOS=linux GOARCH=amd64 go build -o ai-gateway-linux ./cmd/server

deps:
	go mod tidy
	go mod download

docker-build:
	docker build -t ai-gateway .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

deploy:
	./scripts/deploy.sh

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
	@echo "  docker-build - Build Docker image"
	@echo "  docker-up    - Start with Docker Compose"
	@echo "  docker-down  - Stop Docker Compose"
	@echo "  docker-logs  - View Docker logs"
	@echo "  deploy       - Full deployment script"
