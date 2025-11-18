#!/bin/bash

### Official Installation script for logget
# Author: Eneko Galan
# Date: 2025-11-10
# Description: This script downloads and installs logget from GitHub Releases
# Usage: ./scripts/install.sh [version|--uninstall|-u]
# Example: ./scripts/install.sh 1.0.0
# Example: ./scripts/install.sh latest
# Example: ./scripts/install.sh --uninstall
###

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
COLORS_SH="$SCRIPT_DIR/colors.sh"
if [ -f "$COLORS_SH" ]; then
    . "$COLORS_SH"
fi

BINARY_NAME="logget"
GITHUB_REPO="enegalan/logget"
TEMP_DIR=$(mktemp -d)

cleanup() {
    if [ -d "$TEMP_DIR" ]; then
        rm -rf "$TEMP_DIR"
    fi
}
trap cleanup EXIT

error_exit() {
    echo -e "${RED}Error: $1${NC}" >&2
    exit 1
}

detect_install_dir() {
    if [ -n "$LOGGET_INSTALL_DIR" ]; then
        echo "$LOGGET_INSTALL_DIR"
    elif [ -w "/usr/local/bin" ] 2>/dev/null; then
        echo "/usr/local/bin"
    elif [ "$(id -u)" = "0" ]; then
        echo "/usr/local/bin"
    elif [ -d "$HOME/bin" ] && [ -w "$HOME/bin" ]; then
        echo "$HOME/bin"
    elif [ -d "$HOME/.local/bin" ] && [ -w "$HOME/.local/bin" ]; then
        echo "$HOME/.local/bin"
    else
        mkdir -p "$HOME/bin" 2>/dev/null || error_exit "Cannot create install directory. Please set LOGGET_INSTALL_DIR environment variable."
        echo "$HOME/bin"
    fi
}

INSTALL_DIR=$(detect_install_dir)
USE_SUDO=false
if [ "$INSTALL_DIR" = "/usr/local/bin" ] && ! [ -w "/usr/local/bin" ] 2>/dev/null && [ "$(id -u)" != "0" ] && command -v sudo >/dev/null 2>&1; then
    USE_SUDO=true
fi

uninstall_logget() {
    local uninstall_dir=$(detect_install_dir)
    echo -e "${YELLOW}Uninstalling $BINARY_NAME from $uninstall_dir...${NC}"
    if [ -f "$uninstall_dir/$BINARY_NAME" ]; then
        local use_sudo_uninstall=false
        if [ "$uninstall_dir" = "/usr/local/bin" ] && ! [ -w "$uninstall_dir" ] 2>/dev/null && [ "$(id -u)" != "0" ] && command -v sudo >/dev/null 2>&1; then
            use_sudo_uninstall=true
        fi
        if [ "$use_sudo_uninstall" = true ]; then
            if sudo rm -f "$uninstall_dir/$BINARY_NAME"; then
                echo -e "${GREEN}Uninstallation complete!${NC}"
                echo -e "${BLUE}Removed: ${GREEN}$uninstall_dir/$BINARY_NAME${NC}"
            else
                error_exit "Failed to remove binary"
            fi
        else
            if rm -f "$uninstall_dir/$BINARY_NAME"; then
                echo -e "${GREEN}Uninstallation complete!${NC}"
                echo -e "${BLUE}Removed: ${GREEN}$uninstall_dir/$BINARY_NAME${NC}"
            else
                error_exit "Failed to remove binary"
            fi
        fi
    else
        echo -e "${YELLOW}$BINARY_NAME is not installed in $uninstall_dir${NC}"
        exit 0
    fi
}

if [ "$1" = "--uninstall" ] || [ "$1" = "-u" ]; then
    uninstall_logget
    exit 0
fi

detect_os() {
    local os=$(uname -s)
    case "$os" in
        Linux)
            echo "linux"
            ;;
        Darwin)
            echo "darwin"
            ;;
        *)
            error_exit "Unsupported operating system: $os. This script supports Linux and macOS only"
            ;;
    esac
}

detect_arch() {
    local arch=$(uname -m)
    case "$arch" in
        x86_64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        *)
            error_exit "Unsupported architecture: $arch"
            ;;
    esac
}

OS=$(detect_os)
ARCH=$(detect_arch)
echo -e "${BLUE}Detected OS: ${GREEN}$OS${NC}"
echo -e "${BLUE}Detected architecture: ${GREEN}$ARCH${NC}"

check_dependencies() {
    local missing_tools=""
    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        missing_tools="${missing_tools}curl or wget "
    fi
    if ! command -v tar >/dev/null 2>&1; then
        missing_tools="${missing_tools}tar "
    fi
    if [ -n "$missing_tools" ]; then
        error_exit "Missing required tools: $missing_tools"
    fi
    if command -v curl >/dev/null 2>&1; then
        DOWNLOAD_CMD="curl"
    else
        DOWNLOAD_CMD="wget"
    fi
}

check_dependencies

download_file() {
    local url=$1
    local output=$2
    echo -e "${YELLOW}Downloading from: $url${NC}"
    if [ "$DOWNLOAD_CMD" = "curl" ]; then
        if ! curl -fsSL -o "$output" "$url"; then
            error_exit "Failed to download from $url"
        fi
    else
        if ! wget -q -O "$output" "$url"; then
            error_exit "Failed to download from $url"
        fi
    fi
}

get_latest_version() {
    local api_url="https://api.github.com/repos/$GITHUB_REPO/releases/latest"
    local version
    local response
    local response_file="$TEMP_DIR/github_api_response.json"
    local http_code
    echo -e "${YELLOW}Fetching latest version...${NC}" >&2
    if [ "$DOWNLOAD_CMD" = "curl" ]; then
        http_code=$(curl -sSL -o "$response_file" -w "%{http_code}" "$api_url" 2>&1)
        if [ $? -ne 0 ] || [ "$http_code" != "200" ]; then
            if [ -f "$response_file" ]; then
                local error_msg=$(grep -o '"message": "[^"]*' "$response_file" | sed 's/"message": "//' | head -1)
                if [ -n "$error_msg" ]; then
                    error_exit "GitHub API error: $error_msg"
                fi
            fi
            error_exit "Failed to fetch latest version from GitHub API (HTTP $http_code). Check your internet connection and GitHub availability."
        fi
        response=$(cat "$response_file")
        version=$(echo "$response" | grep -o '"tag_name": "[^"]*' | sed 's/"tag_name": "//' | sed 's/^v//' | head -1)
    else
        response=$(wget -q -O - "$api_url" 2>&1)
        if [ $? -ne 0 ] || [ -z "$response" ] || echo "$response" | grep -q '"message"'; then
            local error_msg=$(echo "$response" | grep -o '"message": "[^"]*' | sed 's/"message": "//' | head -1)
            if [ -n "$error_msg" ]; then
                error_exit "GitHub API error: $error_msg"
            fi
            error_exit "Failed to fetch latest version from GitHub API. Check your internet connection and GitHub availability."
        fi
        version=$(echo "$response" | grep -o '"tag_name": "[^"]*' | sed 's/"tag_name": "//' | sed 's/^v//' | head -1)
    fi
    if [ -z "$version" ]; then
        error_exit "Failed to parse version from GitHub API response. The repository might not have any releases yet."
    fi
    echo "$version"
}

validate_version() {
    local version=$1
    local url="https://github.com/$GITHUB_REPO/releases/download/v$version/logget-$version-$OS-$ARCH.tar.gz"
    local status_code
    if [ "$DOWNLOAD_CMD" = "curl" ]; then
        status_code=$(curl -s -o /dev/null -w "%{http_code}" "$url" || echo "000")
    else
        if wget --spider -q "$url" 2>/dev/null; then
            status_code="200"
        else
            status_code="404"
        fi
    fi
    if [ "$status_code" != "200" ]; then
        error_exit "Version $version not found or not available for $OS-$ARCH"
    fi
}

# Determine version to install
VERSION=${1:-""}
if [ -z "$VERSION" ]; then
    VERSION=$(get_latest_version)
    echo -e "${BLUE}Installing latest version: ${GREEN}$VERSION${NC}"
else
    echo -e "${BLUE}Installing version: ${GREEN}$VERSION${NC}"
    validate_version "$VERSION"
fi

# Construct download URL
DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/v$VERSION/logget-$VERSION-$OS-$ARCH.tar.gz"
ARCHIVE_NAME="logget-$VERSION-$OS-$ARCH.tar.gz"
ARCHIVE_PATH="$TEMP_DIR/$ARCHIVE_NAME"

# Download archive
download_file "$DOWNLOAD_URL" "$ARCHIVE_PATH"

# Verify download
if [ ! -f "$ARCHIVE_PATH" ] || [ ! -s "$ARCHIVE_PATH" ]; then
    error_exit "Downloaded file is empty or missing"
fi

# Extract archive
echo -e "${YELLOW}Extracting archive...${NC}"
cd "$TEMP_DIR"
if ! tar -xzf "$ARCHIVE_PATH"; then
    error_exit "Failed to extract archive"
fi

# Find the binary
BINARY_PATH="$TEMP_DIR/logget-$OS-$ARCH"
if [ ! -f "$BINARY_PATH" ]; then
    error_exit "Binary not found in archive: logget-$OS-$ARCH"
fi

# Set executable permissions
chmod +x "$BINARY_PATH"

# Check if logget is already installed
if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    echo -e "${YELLOW}logget is already installed at $INSTALL_DIR/$BINARY_NAME${NC}"
    printf "Do you want to overwrite it? (y/N): "
    read REPLY
    case "$REPLY" in
        [Yy]|"yes"|"Yes"|"YES")
            ;;
        *)
            echo -e "${YELLOW}Installation cancelled${NC}"
            exit 0
            ;;
    esac
fi

# Install binary
echo -e "${YELLOW}Installing to $INSTALL_DIR/$BINARY_NAME...${NC}"
if [ "$USE_SUDO" = true ]; then
    if ! sudo cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"; then
        error_exit "Failed to install binary"
    fi
    if ! sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"; then
        error_exit "Failed to set executable permissions"
    fi
else
    if ! cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"; then
        error_exit "Failed to install binary"
    fi
    if ! chmod +x "$INSTALL_DIR/$BINARY_NAME"; then
        error_exit "Failed to set executable permissions"
    fi
fi

# Verify installation
if [ -f "$INSTALL_DIR/$BINARY_NAME" ] && [ -x "$INSTALL_DIR/$BINARY_NAME" ]; then
    INSTALLED_VERSION=$("$INSTALL_DIR/$BINARY_NAME" --version 2>/dev/null || echo "unknown")
    echo -e "${GREEN}Installation complete!${NC}"
    echo -e "${BLUE}Installed version: ${GREEN}$INSTALLED_VERSION${NC}"
    echo -e "${BLUE}Binary location: ${GREEN}$INSTALL_DIR/$BINARY_NAME${NC}"
    case "$INSTALL_DIR" in
        "$HOME"*)
            case ":$PATH:" in
                *":$INSTALL_DIR:"*)
                    ;;
                *)
                    echo -e "${YELLOW}Note: $INSTALL_DIR is not in your PATH.${NC}"
                    echo -e "${YELLOW}Add this line to your shell profile (~/.bashrc, ~/.zshrc, etc.):${NC}"
                    echo -e "${BLUE}export PATH=\"\$PATH:$INSTALL_DIR\"${NC}"
                    ;;
            esac
            ;;
    esac
else
    error_exit "Installation verification failed"
fi
