#!/bin/bash

# Build script for logget - cross-platform compilation
# This script builds logget for Windows, Linux, and macOS

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building logget for multiple platforms...${NC}"

# Create build directory
mkdir -p build

# Get version from VERSION file or use default
VERSION=$(cat VERSION 2>/dev/null || echo "dev")

# Build for different platforms
echo -e "${YELLOW}Building for Linux (amd64)...${NC}"
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o build/logget-linux-amd64 .

echo -e "${YELLOW}Building for Windows (amd64)...${NC}"
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o build/logget-windows-amd64.exe .

echo -e "${YELLOW}Building for macOS (amd64)...${NC}"
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o build/logget-darwin-amd64 .

echo -e "${YELLOW}Building for macOS (arm64)...${NC}"
GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$VERSION" -o build/logget-darwin-arm64 .

echo -e "${YELLOW}Building for Linux (arm64)...${NC}"
GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$VERSION" -o build/logget-linux-arm64 .

echo -e "${GREEN}Build completed! Binaries are in the build/ directory:${NC}"
ls -la build/

echo -e "${GREEN}To install locally:${NC}"
echo "sudo cp build/logget-linux-amd64 /usr/local/bin/logget"
echo "sudo chmod +x /usr/local/bin/logget"
