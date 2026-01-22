.PHONY: build install clean test fmt lint run help

BINARY_NAME=silo
OUTPUT_DIR=bin
VERSION?=dev

help:
	@echo "Silo CLI - Build & Development"
	@echo ""
	@echo "Usage:"
	@echo "  make build        Build the binary"
	@echo "  make install      Install the binary to /usr/local/bin"
	@echo "  make clean        Remove build artifacts"
	@echo "  make test         Run tests"
	@echo "  make fmt          Format code"
	@echo "  make lint         Run linter"
	@echo "  make run          Build and run"
	@echo ""

build:
	@echo "Building $(BINARY_NAME)..."
	@VERSION=$(VERSION) OUTPUT=$(OUTPUT_DIR)/$(BINARY_NAME) bash scripts/build.sh

install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(OUTPUT_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "✓ Installed successfully"

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(OUTPUT_DIR)
	@echo "✓ Clean complete"

test:
	@echo "Running tests..."
	@go test -v ./...

fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Format complete"

lint:
	@echo "Running linter..."
	@golangci-lint run ./...

run: build
	@$(OUTPUT_DIR)/$(BINARY_NAME) $(ARGS)

dev:
	@go run cmd/silo/main.go $(ARGS)
