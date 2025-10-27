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
git clone https://github.com/yourusername/logget.git
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
logget <url> [flags]
```

### Basic Examples

```bash
# Show console logs
logget https://example.com --logs

# Show network requests
logget https://example.com --network

# Show both logs and network data
logget https://example.com --logs --network

# Output in JSON format
logget https://example.com --logs --network --json

# Write output to a file
logget https://example.com --logs --output results.txt

# Write JSON output to a file
logget https://example.com --logs --json --output results.json

# Append output to an existing file
logget https://example.com --logs --output results.txt --append

# Append JSON output to a file
logget https://example.com --logs --json --output results.json --append

# Save files to a specific directory
logget https://example.com --logs --output results.txt --output-dir /tmp/logs

# Create nested directories automatically
logget https://example.com --logs --output data.txt --output-dir results/2024/october

# Append to files in a specific directory
logget https://example.com --logs --output data.txt --output-dir logs --append

# Set custom timeout
logget https://example.com --logs --timeout 60

# Set custom wait time after page load
logget https://example.com --logs --wait 5

# Add custom headers
logget https://example.com --logs --header "Authorization: Bearer token" --header "X-Custom: value"

# Set cookies for authenticated requests
logget https://example.com --logs --cookie "session_id=abc123" --cookie "user_token=xyz789"

# Set cookies with additional attributes
logget https://example.com --logs --cookie "session_id=abc123; domain=.example.com; secure" --cookie "user_pref=dark_mode; path=/settings"
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

## Comparison with curl

| Feature | curl | logget |
|---------|------|--------|
| HTTP requests | ✅ | ✅ |
| Custom headers | ✅ | ✅ |
| Cookie support | ✅ | ✅ |
| File output | ✅ | ✅ |
| Directory organization | ✅ | ✅ |
| Response body | ✅ | ❌ |
| Console logs | ❌ | ✅ |
| Network monitoring | ❌ | ✅ |
| JavaScript execution | ❌ | ✅ |
| Browser automation | ❌ | ✅ |

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
