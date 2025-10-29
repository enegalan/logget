# logget

A command-line tool similar to `curl` that extracts browser logs and network data from web pages using an embedded Chromium browser.

## Features

- **Console Log Collection**: Capture all console.log, console.error, console.warn, and console.info messages
- **Network Monitoring**: Track all HTTP requests (fetch, XMLHttpRequest) with headers, status codes, and timing
- **Cross-Platform**: Works on Windows, Linux, and macOS
- **No Chrome Required**: Uses embedded Chromium via `go-rod`
- **JSON Output**: Structured data output for easy parsing
- **File Output**: Write results to files instead of stdout
- **Append Mode**: Add output to existing files instead of overwriting
- **Directory Organization**: Save files to specific directories with automatic directory creation
- **Custom Headers**: Add custom HTTP headers like curl
- **Cookie Support**: Set cookies for authenticated requests
- **Configurable Timeout**: Set custom timeout values

## Installation

### Pre-built Binaries

Download the appropriate binary for your platform from the [releases page](https://github.com/enegalan/logget/releases):

- **All platforms**: `logget-VERSION.zip` or `logget-VERSION.tar.gz`
- **Individual platforms**: Platform-specific `.zip` and `.tar.gz` files

### Build from Source

```bash
git clone https://github.com/enegalan/logget.git
cd logget
go mod tidy
go build -o logget .
```

### Cross-platform Build

```bash
# Using Makefile
make build-all

# Or using the build script directly
./scripts/build.sh
```

## Usage

```bash
logget [flags] <url>
```

### Basic Examples

```bash
# Show console logs
logget --logs https://example.com

# Show network requests
logget --network https://example.com

# Show both logs and network data
logget --logs --network https://example.com

# Output in JSON format
logget --logs --network --json https://example.com

# Write output to a file
logget --logs --output results.txt https://example.com

# Write JSON output to a file
logget --logs --json --output results.json https://example.com

# Append output to an existing file
logget --logs --output results.txt --append https://example.com

# Append JSON output to a file
logget --logs --json --output results.json --append https://example.com

# Save files to a specific directory
logget --logs --output results.txt --output-dir /tmp/logs https://example.com

# Create nested directories automatically
logget --logs --output data.txt --output-dir results/2024/october https://example.com

# Append to files in a specific directory
logget --logs --output data.txt --output-dir logs --append https://example.com

# Set custom timeout
logget --logs --timeout 60 https://example.com

# Set custom wait time after page load
logget --logs --wait 5 https://example.com

# Add custom headers
logget --logs --header "Authorization: Bearer token" --header "X-Custom: value" https://example.com

# Set cookies for authenticated requests
logget --logs --cookie "session_id=abc123" --cookie "user_token=xyz789" https://example.com

# Set cookies with additional attributes
logget --logs --cookie "session_id=abc123; domain=.example.com; secure" --cookie "user_pref=dark_mode; path=/settings" https://example.com
```

### Real-time Streaming Examples

```bash
# Stream browser logs from a URL in real-time
logget -f --logs https://example.com

# Stream network requests from a URL
logget -f --network https://example.com

# Stream both logs and network data
logget -f --logs --network https://example.com

# Stream with custom refresh interval (default: 100ms)
logget -f --logs --refresh 500 https://example.com

# Stream to a file
logget -f --logs --output stream.log https://example.com

# Stream with JSON output
logget -f --logs --json https://example.com

# Stream with custom headers and cookies
logget -f --logs --header "Authorization: Bearer token" --cookie "session=abc123" https://example.com

# Stream with filtering for ERROR messages only
logget -f --logs --filter "ERROR" https://example.com

# Stream with filtering for multiple patterns
logget -f --logs --filter "ERROR|WARN" https://example.com

# Stream with exclusion patterns
logget -f --logs --exclude "DEBUG" https://example.com

# Stream with filtering and JSON output
logget -f --logs --filter "ERROR|WARN" --json --output errors.json https://example.com

# Skip SSL verification for local development servers
logget -k --logs https://0.0.0.0:3030
logget -k --network https://localhost:8080
logget -k -f --logs --filter "ERROR" https://127.0.0.1:3000
```

### Request Type Filtering Examples

```bash
# Only XHR/fetch requests
logget --network --xhr https://example.com

# Only images
logget --network --img https://example.com

# Only scripts and CSS
logget --network --script --css https://example.com

# Only WebSocket traffic
logget --network --ws https://example.com

# Only WebAssembly requests
logget --network --wasm https://example.com
```

### Options

- `--logs`, `-L`: Capture and display console logs
- `--network`, `-N`: Capture and display network requests
- `--json`, `-J`: Output results in JSON format
- `--output`, `-o`: Write to file instead of stdout
- `--append`, `-a`: Append to file instead of overwriting
- `--output-dir`: Directory to save files in (creates directories automatically)
- `--timeout`, `-T`: Set timeout in seconds (default: 60)
- `--wait`, `-W`: Wait time in seconds after page load (default: 3)
- `--user-agent`, `-A`: Set User-Agent header (default: "logget/1.0")
- `--header`, `-H`: Add custom HTTP headers (format: 'Key: Value')
- `--cookie`, `-C`: Set cookies (format: 'name=value' or 'name=value; domain=example.com')
- `--verbose`, `-V`: Show detailed HTTP protocol information
- `--version`, `-v`: Show version information
- `--follow`, `-f`: Stream logs and network requests in real-time
- `--filter`: Show only logs/requests matching this regex pattern
- `--exclude`: Exclude logs/requests matching this regex pattern
- `--refresh`: Refresh interval in milliseconds for real-time streaming (default: 100)
- `--insecure`, `-k`: Skip SSL certificate verification (useful for self-signed certificates)
- `--xhr`: Only include fetch/XHR requests
- `--document`: Only include Document requests
- `--css`: Only include CSS requests
- `--script`: Only include Script requests
- `--font`: Only include Font requests
- `--img`: Only include Image requests
- `--media`: Only include Media requests
- `--manifest`: Only include Manifest requests
- `--ws`, `--websocket`: Only include WebSocket requests
- `--wasm`: Only include WebAssembly (application/wasm) requests

### File Output and Directory Organization

The `--output`, `--output-dir`, and `--append` flags allow you to save results to files instead of displaying them on stdout:

```bash
# Save to a specific file (overwrites existing file)
logget https://example.com --logs --output results.txt

# Append to an existing file (creates file if it doesn't exist)
logget https://example.com --logs --output results.txt --append

# Save to a directory (creates the directory if it doesn't exist)
logget https://example.com --logs --output results.txt --output-dir /tmp/logs

# Create nested directory structure automatically
logget https://example.com --logs --output data.txt --output-dir logs/2024/october/27

# Append to files in a specific directory
logget https://example.com --logs --output data.txt --output-dir logs --append

# Works with JSON output too
logget https://example.com --logs --json --output data.json --output-dir results

# Append JSON output to existing file
logget https://example.com --logs --json --output data.json --append
```

**Key Features:**
- **Automatic Directory Creation**: Creates directories and subdirectories as needed
- **Cross-Platform Paths**: Uses proper path separators for your operating system
- **Error Handling**: Shows clear error messages if file creation fails
- **Confirmation**: Displays the full path where the file was saved or appended

## Use Cases

- **Web Development**: Debug JavaScript applications and API calls
- **Testing**: Verify network requests and console output
- **Monitoring**: Track application behavior and performance
- **Security Analysis**: Inspect network traffic and JavaScript execution
- **API Testing**: Monitor API calls and responses
- **Automation**: Save results to files for batch processing and analysis
- **Logging**: Create organized log files with timestamps and structured data
- **Continuous Monitoring**: Append multiple runs to the same file for historical tracking
- **CI/CD Integration**: Generate reports and artifacts for continuous integration pipelines
- **Data Collection**: Accumulate data from multiple sources into single files
- **AI Utility**: Allow your AI agents use this command for efficient debugging
- **Real-time Monitoring**: Stream logs from web applications as they happen

## Comparison with curl

| Feature | curl | logget |
|---------|------|--------|
| HTTP requests | ✅ | ✅ |
| Custom headers | ✅ | ✅ |
| Cookie support | ✅ | ✅ |
| Response body | ✅ | ❌ |
| Console logs | ❌ | ✅ |
| Network monitoring | ❌ | ✅ |
| JavaScript execution | ❌ | ✅ |
| Browser automation | ❌ | ✅ |
| Real-time streaming | ❌ | ✅ |

## Development

### Available Scripts

The project includes some utility scripts in the `scripts/` directory:

- **`scripts/build.sh`**: Cross-platform compilation for all supported platforms
- **`scripts/release.sh`**: Create release packages (ZIP and TAR.GZ)

### Makefile Commands

```bash
make build        # Build for current platform
make build-all    # Build for all platforms (Linux, Windows, macOS)
make clean        # Clean build artifacts
make test         # Run tests
make deps         # Install dependencies
make install      # Install binary to system
make uninstall    # Remove binary from system
make release      # Create release packages (ZIP and TAR.GZ)
make help         # Show all available commands
```

### Creating Releases

```bash
# Create release packages
make release

# Or manually
cd scripts && ./release.sh
```

This will create:
- Individual platform packages (`.zip` and `.tar.gz`)
- Combined packages with all platforms

## Requirements

- Go 1.21 or later (for building from source)
- No additional browser installation required
