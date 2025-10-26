# logget

A command-line tool similar to `curl` that extracts browser logs and network data from web pages using an embedded Chromium browser.

## Features

- **Console Log Collection**: Capture all console.log, console.error, console.warn, and console.info messages
- **Network Monitoring**: Track all HTTP requests (fetch, XMLHttpRequest) with headers, status codes, and timing
- **Cross-Platform**: Works on Windows, Linux, and macOS
- **No Chrome Required**: Uses embedded Chromium via `go-rod`
- **JSON Output**: Structured data output for easy parsing
- **Custom Headers**: Add custom HTTP headers like curl
- **Configurable Timeout**: Set custom timeout values

## Installation

### Pre-built Binaries

Download the appropriate binary for your platform from the releases page.

### Build from Source

```bash
git clone https://github.com/yourusername/logget.git
cd logget
go mod tidy
go build -o logget .
```

### Cross-platform Build

```bash
chmod +x build.sh
./build.sh
```

## Usage

```bash
logget <url> [flags]
```

### Basic Examples

```bash
# Show console logs
logget https://example.com --show-logs

# Show network requests
logget https://example.com --show-network

# Show both logs and network data
logget https://example.com --show-logs --show-network

# Output in JSON format
logget https://example.com --show-logs --show-network --json

# Set custom timeout
logget https://example.com --show-logs --timeout 60

# Add custom headers
logget https://example.com --show-logs --header "Authorization: Bearer token" --header "X-Custom: value"
```

### Options

- `--show-logs`: Capture and display console logs
- `--show-network`: Capture and display network requests
- `--json`: Output results in JSON format
- `--timeout SECONDS`: Set timeout in seconds (default: 30)
- `--user-agent STRING`: Set User-Agent header (default: "logget/1.0")
- `--header "Key: Value"`: Add custom HTTP headers

## Use Cases

- **Web Development**: Debug JavaScript applications and API calls
- **Testing**: Verify network requests and console output
- **Monitoring**: Track application behavior and performance
- **Security Analysis**: Inspect network traffic and JavaScript execution
- **API Testing**: Monitor API calls and responses
- **AI Utility**: Allow your AI agents use this command for efficient debugging

## Comparison with curl

| Feature | curl | logget |
|---------|------|--------|
| HTTP requests | ✅ | ✅ |
| Custom headers | ✅ | ✅ |
| Response body | ✅ | ❌ |
| Console logs | ❌ | ✅ |
| Network monitoring | ❌ | ✅ |
| JavaScript execution | ❌ | ✅ |
| Browser automation | ❌ | ✅ |

## Requirements

- Go 1.21 or later (for building from source)
- No additional browser installation required
