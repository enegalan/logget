.PHONY: build clean test install uninstall deps help

BINARY_NAME=logget
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR=build
GO_FILES=$(shell find . -name "*.go" -type f)

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME) .
	@echo "Build complete: ./$(BINARY_NAME)"

build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@echo "Building for Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	@echo "Building for Windows (amd64)..."
	@GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Building for macOS (amd64)..."
	@GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	@echo "Building for macOS (arm64)..."
	@GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	@echo "Building for Linux (arm64)..."
	@GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	@echo "Build complete! Binaries are in $(BUILD_DIR)/"

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

test:
	@echo "Running tests..."
	@go test -v ./...

deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BINARY_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "Installation complete"

uninstall:
	@echo "Uninstalling $(BINARY_NAME) from /usr/local/bin..."
	@if [ -f /usr/local/bin/$(BINARY_NAME) ]; then \
		sudo rm -f /usr/local/bin/$(BINARY_NAME); \
		echo "Uninstallation complete"; \
	else \
		echo "$(BINARY_NAME) is not installed in /usr/local/bin"; \
	fi

help:
	@echo "Available targets:"
	@echo "  build        - Build for current platform"
	@echo "  build-all    - Build for all platforms (Linux, Windows, macOS)"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  deps         - Install dependencies"
	@echo "  install      - Install binary to system"
	@echo "  uninstall    - Remove binary from system"
	@echo "  dev          - Build and run example"
	@echo "  help         - Show this help"
