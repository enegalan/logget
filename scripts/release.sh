#!/bin/bash

# Release script for logget
# Creates compressed releases for all platforms using build.sh

set -e

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
COLORS_SH="$SCRIPT_DIR/colors.sh"
source "$COLORS_SH"

# Configuration
BINARY_NAME="logget"
VERSION_FILE="$PROJECT_ROOT/VERSION"
BUILD_DIR="$PROJECT_ROOT/build"
RELEASES_DIR="$PROJECT_ROOT/releases"

# Read version from file
if [ ! -f "$VERSION_FILE" ]; then
    echo -e "${RED}Error: $VERSION_FILE not found${NC}"
    exit 1
fi

VERSION=$(cat "$VERSION_FILE" | tr -d '\n\r')
echo -e "${BLUE}Building releases for version: ${GREEN}$VERSION${NC}"
echo -e "${YELLOW}Project root: $PROJECT_ROOT${NC}"

# Create directories
mkdir -p "$RELEASES_DIR"

# Clean previous releases
echo -e "${YELLOW}Cleaning previous releases...${NC}"
rm -rf "$RELEASES_DIR"/*

# Use build.sh to create all binaries
echo -e "${YELLOW}Building binaries using build.sh...${NC}"
"$SCRIPT_DIR/build.sh"

# Create release packages
echo -e "${YELLOW}Creating release packages...${NC}"

# Create ZIP archive with all binaries
ZIP_NAME="${BINARY_NAME}-${VERSION}.zip"
echo "Creating $ZIP_NAME..."
cd "$BUILD_DIR"
zip -r "$RELEASES_DIR/$ZIP_NAME" .
cd - > /dev/null

# Create TAR.GZ archive with all binaries
TAR_NAME="${BINARY_NAME}-${VERSION}.tar.gz"
echo "Creating $TAR_NAME..."
cd "$BUILD_DIR"
tar -czf "$RELEASES_DIR/$TAR_NAME" .
cd - > /dev/null

# Create individual platform packages
echo -e "${YELLOW}Creating individual platform packages...${NC}"

# Linux packages
cd "$BUILD_DIR"
zip "$RELEASES_DIR/${BINARY_NAME}-${VERSION}-linux-amd64.zip" "${BINARY_NAME}-linux-amd64"
zip "$RELEASES_DIR/${BINARY_NAME}-${VERSION}-linux-arm64.zip" "${BINARY_NAME}-linux-arm64"
tar -czf "$RELEASES_DIR/${BINARY_NAME}-${VERSION}-linux-amd64.tar.gz" "${BINARY_NAME}-linux-amd64"
tar -czf "$RELEASES_DIR/${BINARY_NAME}-${VERSION}-linux-arm64.tar.gz" "${BINARY_NAME}-linux-arm64"

# Windows packages
zip "$RELEASES_DIR/${BINARY_NAME}-${VERSION}-windows-amd64.zip" "${BINARY_NAME}-windows-amd64.exe"
tar -czf "$RELEASES_DIR/${BINARY_NAME}-${VERSION}-windows-amd64.tar.gz" "${BINARY_NAME}-windows-amd64.exe"

# macOS packages
zip "$RELEASES_DIR/${BINARY_NAME}-${VERSION}-darwin-amd64.zip" "${BINARY_NAME}-darwin-amd64"
zip "$RELEASES_DIR/${BINARY_NAME}-${VERSION}-darwin-arm64.zip" "${BINARY_NAME}-darwin-arm64"
tar -czf "$RELEASES_DIR/${BINARY_NAME}-${VERSION}-darwin-amd64.tar.gz" "${BINARY_NAME}-darwin-amd64"
tar -czf "$RELEASES_DIR/${BINARY_NAME}-${VERSION}-darwin-arm64.tar.gz" "${BINARY_NAME}-darwin-arm64"
cd - > /dev/null

# Display results
echo -e "${GREEN}Release creation complete!${NC}"
echo -e "${BLUE}Files created in $RELEASES_DIR/:${NC}"
ls -la "$RELEASES_DIR"

echo -e "\n${GREEN}Release packages:${NC}"
echo -e "  ${YELLOW}All platforms:${NC}"
echo -e "    - $RELEASES_DIR/${BINARY_NAME}-${VERSION}.zip"
echo -e "    - $RELEASES_DIR/${BINARY_NAME}-${VERSION}.tar.gz"
echo -e "  ${YELLOW}Individual platforms:${NC}"
echo -e "    - Linux (amd64/arm64): .zip and .tar.gz"
echo -e "    - Windows (amd64): .zip and .tar.gz"
echo -e "    - macOS (amd64/arm64): .zip and .tar.gz"

echo -e "\n${BLUE}Version: $VERSION${NC}"
echo -e "${GREEN}Ready for distribution!${NC}"
