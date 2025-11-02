.PHONY: build clean test install uninstall release deps help generate-colors

BINARY_NAME=logget
VERSION=$(shell cat VERSION 2>/dev/null || echo "dev")
BUILD_DIR=build
GO_FILES=$(shell find . -name "*.go" -type f)

all: generate-colors build

build:
	@echo "Building $(BINARY_NAME)..."
	@go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME) .
	@echo "Build complete: ./$(BINARY_NAME)"

build-all:
	@echo "Building $(BINARY_NAME) for all platforms..."
	@cd scripts && ./build.sh

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

release:
	@echo "Creating release packages..."
	@cd scripts && ./release.sh

generate-colors:
	@echo "Generating colors.sh from helpers/colors.go..."
	@./scripts/generate_colors.sh

help:
	@echo "Available targets:"
	@echo "  build           - Build for current platform"
	@echo "  build-all       - Build for all platforms (Linux, Windows, macOS)"
	@echo "  clean           - Clean build artifacts"
	@echo "  test            - Run tests"
	@echo "  deps            - Install dependencies"
	@echo "  install         - Install binary to system"
	@echo "  uninstall       - Remove binary from system"
	@echo "  release         - Create release packages (ZIP and TAR.GZ)"
	@echo "  generate-colors - Generate scripts/colors.sh from helpers/colors.go"
	@echo "  help            - Show this help"
