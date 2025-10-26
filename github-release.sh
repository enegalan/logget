#!/bin/bash

# GitHub Release automation script
# Requires: gh CLI tool (https://cli.github.com/)

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Configuration
VERSION=$(cat VERSION 2>/dev/null || echo "dev")
REPO="your-username/logget"  # Change this to your repo
RELEASES_DIR="releases"

echo -e "${YELLOW}Creating GitHub release for version $VERSION...${NC}"

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo "Error: GitHub CLI (gh) is not installed"
    echo "Install it from: https://cli.github.com/"
    exit 1
fi

# Check if user is authenticated
if ! gh auth status &> /dev/null; then
    echo "Error: Not authenticated with GitHub CLI"
    echo "Run: gh auth login"
    exit 1
fi

# Create release with all files
gh release create "v$VERSION" \
    --title "logget v$VERSION" \
    --notes "## What's New in v$VERSION

- Extract browser logs and network data from web pages
- Cross-platform support (Linux, Windows, macOS)
- Multiple output formats (JSON, human-readable)
- Chrome DevTools Protocol integration

## Downloads

- **All platforms**: \`logget-$VERSION.zip\` or \`logget-$VERSION.tar.gz\`
- **Individual platforms**: See platform-specific files below

## Verification

Verify file integrity using checksums:
\`\`\`bash
sha256sum -c checksums.txt
\`\`\`

## Installation

1. Download the appropriate file for your platform
2. Extract the binary
3. Move to your PATH: \`sudo mv logget /usr/local/bin/\`
4. Make executable: \`sudo chmod +x /usr/local/bin/logget\`

## Usage

\`\`\`bash
logget https://example.com --logs --network
\`\`\`" \
    "$RELEASES_DIR"/*

echo -e "${GREEN}Release created successfully!${NC}"
echo -e "${YELLOW}View at: https://github.com/$REPO/releases/tag/v$VERSION${NC}"
