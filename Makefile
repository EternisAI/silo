.PHONY: build install clean test fmt lint run help build-daemon install-daemon install-service

BINARY_NAME=silo
DAEMON_NAME=silod
OUTPUT_DIR=bin
VERSION?=dev

help:
	@echo "Silo CLI - Build & Development"
	@echo ""
	@echo "Usage:"
	@echo "  make build            Build the CLI binary"
	@echo "  make build-daemon     Build the daemon binary"
	@echo "  make build-all        Build both CLI and daemon"
	@echo "  make install          Install the CLI to /usr/local/bin"
	@echo "  make install-daemon   Install the daemon to /usr/local/bin"
	@echo "  make install-service  Install systemd service"
	@echo "  make install-all      Install CLI, daemon, and systemd service"
	@echo "  make clean            Remove build artifacts"
	@echo "  make test             Run tests"
	@echo "  make fmt              Format code"
	@echo "  make lint             Run linter"
	@echo "  make run              Build and run CLI"
	@echo "  make run-daemon       Build and run daemon"
	@echo ""

build:
	@echo "Building $(BINARY_NAME)..."
	@VERSION=$(VERSION) OUTPUT=$(OUTPUT_DIR)/$(BINARY_NAME) bash scripts/build.sh

build-daemon:
	@echo "Building $(DAEMON_NAME)..."
	@VERSION=$(VERSION) OUTPUT=$(OUTPUT_DIR)/$(DAEMON_NAME) MAIN=cmd/silod/main.go bash scripts/build.sh

build-all: build build-daemon

install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(OUTPUT_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "✓ Installed successfully"

install-daemon: build-daemon
	@echo "Installing $(DAEMON_NAME) to /usr/local/bin..."
	@sudo cp $(OUTPUT_DIR)/$(DAEMON_NAME) /usr/local/bin/$(DAEMON_NAME)
	@sudo chmod +x /usr/local/bin/$(DAEMON_NAME)
	@echo "✓ Daemon installed successfully"

install-service: install-daemon
	@echo "Installing systemd service..."
	@sudo cp scripts/silod.service /etc/systemd/system/silod.service
	@sudo systemctl daemon-reload
	@echo "✓ Service installed. Enable with: sudo systemctl enable silod"
	@echo "  Start with: sudo systemctl start silod"

install-all: install install-daemon install-service

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

run-daemon: build-daemon
	@$(OUTPUT_DIR)/$(DAEMON_NAME)

dev:
	@go run cmd/silo/main.go $(ARGS)

dev-daemon:
	@go run cmd/silod/main.go
