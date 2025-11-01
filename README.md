# logget

A command-line tool similar to `curl` that extracts browser logs and network data from web pages using an embedded Chromium browser.

## Features

- **Console Log Collection**: Capture all console.log, console.error, console.warn, and console.info messages
- **Network Monitoring**: Track all HTTP requests (fetch, XMLHttpRequest) with headers, status codes, and timing
- **Cross-Platform**: Works on Windows, Linux, and macOS
- **No Chrome Required**: Uses embedded Chromium via `chromedp`
- **JSON Output**: Structured data output for easy parsing
- **File Output**: Write results to files instead of stdout
- **Append Mode**: Add output to existing files instead of overwriting
- **Custom Headers**: Add custom HTTP headers like curl (supports files)
- **Cookie Support**: Set cookies for authenticated requests (supports files)
- **Configurable Timeout**: Set custom timeout values
- **Fingerprint Rotation**: Rotate navigator fingerprints (userAgent, platform, language, screen properties, WebGL, Canvas) to prevent tracking
- **Performance Metrics**: Detailed timing metrics (Duration, TTFB, Connect Time, DNS, SSL, Send, Wait, Receive times, Content Download Time, Queued Time, Total)
- **HAR Export**: Export network data in HAR (HTTP Archive) format

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

# Output in CSV format
logget --logs --network --csv https://example.com

# Output in HAR format
logget --network --har https://example.com

# Write output to a file
logget --logs --output results.txt https://example.com

# Write JSON output to a file
logget --logs --json --output results.json https://example.com

# Write HAR output to a file
logget --network --har --output network.har https://example.com

# Append output to an existing file
logget --logs --output results.txt --append https://example.com

# Append JSON output to a file
logget --logs --json --output results.json --append https://example.com

# Set custom timeout
logget --logs --timeout 60 https://example.com

# Set custom wait time after page load (in milliseconds)
logget --logs --wait 5000 https://example.com

# Add custom headers
logget --logs --header "Authorization: Bearer token" --header "X-Custom: value" https://example.com

# Add headers from a file
logget --logs --header headers.txt https://example.com

# Set cookies for authenticated requests
logget --logs --cookie "session_id=abc123" --cookie "user_token=xyz789" https://example.com

# Set cookies from a file
logget --logs --cookie cookies.txt https://example.com

# Set cookies with additional attributes
logget --logs --cookie "session_id=abc123; domain=.example.com; secure" --cookie "user_pref=dark_mode; path=/settings" https://example.com

# Mix direct values and files
logget --logs --header "Authorization: Bearer token" --header headers.txt --cookie "session_id=abc123" --cookie cookies.txt https://example.com

# Suppress progress messages, only show data
logget --quiet --logs --network https://example.com
logget -q --logs --json https://example.com
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
logget --network --socket https://example.com

# Only JavaScript MIME types
logget --network --mime "^application/(javascript|json)$" https://example.com

# Only 2xx responses (regex)
logget --network --status "^2..$" https://example.com

# Only 200 or 204
logget --network --status "^(200|204)$" https://example.com

# Only requests to a specific domain
logget --network --domain "^api\\.example\\.com$" https://example.com

# Requests to subdomains of example.com
logget --network --domain "(.*\\.)?example\\.com$" https://example.com

# Only requests larger than 1KB
logget --network --min-size 1024 https://example.com

# Only requests smaller than 10KB
logget --network --max-size 10240 https://example.com

# Requests between 1KB and 100KB
logget --network --min-size 1024 --max-size 102400 https://example.com
```

### Fingerprint Rotation Examples

```bash
# Fingerprint rotation is enabled by default (every 5 seconds)
logget --network https://example.com

# Rotate fingerprints every 2 seconds
logget --fingerprint-interval 2000 --network https://example.com

# Rotate fingerprints with custom interval and stream to file
logget --fingerprint-interval 3000 -f --logs --output fingerprint.log https://example.com

# Disable fingerprint rotation (if needed)
logget --no-rotate-fingerprints --network https://example.com
```

### Options

- `--logs`, `-L`: Capture and display console logs
- `--network`, `-N`: Capture and display network requests
- `--json`, `-J`: Output results in JSON format
- `--csv`: Output results in CSV format
- `--har`: Output results in HAR (HTTP Archive) format
- `--no-color`: Disable colored output
- `--quiet`, `-q`: Suppress progress messages, only show data (errors and warnings still displayed)
- `--output`, `-o` `<filename>`: Write to file instead of stdout
- `--append`, `-a`: Append to file instead of overwriting
- `--timeout`, `-T`: Set timeout in seconds (default: 60)
- `--wait`, `-W`: Wait time in milliseconds after page load (default: 3000)
- `--user-agent`, `-A` `<name>`: Set User-Agent header (default: "logget/1.0")
- `--header`, `-H` `<header|file>`: Add custom HTTP headers (format: 'Key: Value') or filename containing headers
- `--cookie`, `-C` `<data|filename>`: Set cookies (format: 'name=value' or 'name=value; domain=example.com') or filename containing cookies
- `--verbose`, `-V`: Show detailed HTTP protocol information
- `--version`, `-v`: Show version information
- `--follow`, `-f`: Stream logs and network requests in real-time
- `--filter` `<regex>`: Show only logs/requests matching this regex pattern
- `--exclude` `<regex>`: Exclude logs/requests matching this regex pattern
- `--status` `<regex>`: Only include requests whose HTTP status code matches this regex
- `--domain` `<regex>`: Only include requests whose domain matches this regex
- `--mime` `<regex>`: Only include requests whose MIME type matches this regex
- `--min-size` `<bytes>`: Only include requests whose size is at least this many bytes
- `--max-size` `<bytes>`: Only include requests whose size is at most this many bytes
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
- `--socket`: Only include WebSocket requests
- `--no-rotate-fingerprints`: Disable fingerprint rotation (default: enabled)
- `--fingerprint-interval`: Interval in milliseconds for fingerprint rotation (default: 5000)

### File Output

The `--output` and `--append` flags allow you to save results to files instead of displaying them on stdout:

```bash
# Save to a specific file (overwrites existing file)
logget https://example.com --logs --output results.txt

# Append to an existing file (creates file if it doesn't exist)
logget https://example.com --logs --output results.txt --append

# Works with JSON output too
logget https://example.com --logs --json --output data.json

# Works with HAR output too
logget --network --har --output network.har https://example.com

# Append JSON output to existing file
logget https://example.com --logs --json --output data.json --append

# Quiet mode: save without confirmation messages
logget -q --logs --output results.txt https://example.com
```

### File-Based Input

Both `--header` and `--cookie` flags support reading from files in addition to direct values:

**Header File Format:**
```
Authorization: Bearer token123
X-Custom-Header: value
Content-Type: application/json
# Comments (lines starting with # are ignored)
```

**Cookie File Format:**
```
session_id=abc123
user_token=xyz789
pref=dark_mode; domain=example.com
# Comments (lines starting with # are ignored)
```

The tools automatically detect whether the argument is a file path or a direct value:
- If it contains `:` (for headers) or `=` (for cookies), it's treated as direct data
- Otherwise, it attempts to read from the file
- You can mix direct values and files in the same command

## Output Schemas

### JSON Schema

When using `--json`, logget outputs structured JSON data. The format depends on whether you're in follow mode (streaming) or batch mode.

#### Batch Mode (Full Output)

In batch mode, the complete JSON output contains a single `OutputData` object:

```json
{
  "url": "https://example.com",
  "logs": [
    {
      "level": "INFO",
      "message": "Console log message",
      "time": "2024-10-31T23:00:00Z",
      "source": "console"
    }
  ],
  "network": [
    {
      "url": "https://example.com/api/data",
      "method": "GET",
      "status": 200,
      "headers": {
        "Content-Type": "application/json",
        "Content-Length": "1234"
      },
      "timestamp": "2024-10-31T23:00:00Z",
      "type": "application/json",
      "size": 1234,
      "resourceType": "XHR"
    }
  ],
  "duration": "3.456789s"
}
```

#### Follow Mode (Streaming)

In follow mode (`-f`), each log entry and network request is output as a separate JSON object (one per line):

```json
{"level":"INFO","message":"Log message","time":"2024-10-31T23:00:00Z","source":"console"}
{"url":"https://example.com","method":"GET","status":200,"headers":{"Content-Type":"text/html"},"timestamp":"2024-10-31T23:00:01Z","type":"text/html","size":184,"resourceType":"Document"}
```

#### Field Descriptions

**OutputData (Batch Mode Only):**
- `url` (string): Visited URL
- `logs` (array, optional): Array of `LogEntry` objects
- `network` (array, optional): Array of `NetworkEntry` objects
- `duration` (string): Total time taken to load the page

**LogEntry:**
- `level` (string): Log level (`DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`, `LOG`, `TRACE`)
- `message` (string): The log message content
- `time` (string): Timestamp in RFC3339 format (ISO 8601)
- `source` (string): Source of the log (`"browser"` or `"console"`)

**NetworkEntry:**
- `url` (string): Request URL
- `method` (string): HTTP method (e.g., `GET`, `POST`, `PUT`, `DELETE`)
- `status` (integer): HTTP status code (e.g., 200, 404, 500)
- `headers` (object): Response headers as key-value pairs (string keys and values)
- `timestamp` (string): Timestamp in RFC3339 format (ISO 8601)
- `type` (string): MIME type of the response (e.g., `"text/html"`, `"application/json"`)
- `size` (integer): Size of the response in bytes
- `resourceType` (string): Resource type (e.g., `"Document"`, `"XHR"`, `"Image"`, `"Script"`, `"Stylesheet"`, `"Font"`, `"Media"`, `"Manifest"`, `"WebSocket"`, `"Other"`)
- `duration` (float, optional): Total request duration in milliseconds
- `durationFormatted` (string, optional): Formatted duration (e.g., `"123.45ms"` or `"1.23s"`)
- `timeToFirstByte` (float, optional): Time to first byte (TTFB) in milliseconds
- `timeToFirstByteFormatted` (string, optional): Formatted TTFB (e.g., `"50.23ms"` or `"0.50s"`)
- `connectTime` (float, optional): Connection establishment time in milliseconds
- `connectTimeFormatted` (string, optional): Formatted connect time
- `dnsLookupTime` (float, optional): DNS lookup time in milliseconds
- `dnsLookupTimeFormatted` (string, optional): Formatted DNS lookup time
- `sslTime` (float, optional): SSL/TLS handshake time in milliseconds (HTTPS only)
- `sslTimeFormatted` (string, optional): Formatted SSL time
- `sendTime` (float, optional): Time to send request in milliseconds
- `sendTimeFormatted` (string, optional): Formatted send time
- `waitTime` (float, optional): Wait time (server processing) in milliseconds
- `waitTimeFormatted` (string, optional): Formatted wait time
- `receiveTime` (float, optional): Time to receive response headers in milliseconds
- `receiveTimeFormatted` (string, optional): Formatted receive time
- `contentDownloadTime` (float, optional): Full content download time (body + headers) in milliseconds
- `contentDownloadTimeFormatted` (string, optional): Formatted content download time
- `queuedTime` (float, optional): Time from navigation start to request start in milliseconds
- `queuedTimeFormatted` (string, optional): Formatted queued time
- `requestStartTime` (float, optional): Request start time relative to navigation start in milliseconds
- `requestStartTimeFormatted` (string, optional): Formatted request start time
- `responseStartTime` (float, optional): Response start time relative to navigation start in milliseconds
- `responseStartTimeFormatted` (string, optional): Formatted response start time
- `total` (float, optional): Sum of all timing phases (Queued + DNS + Connect + SSL + Send + Wait + Receive + ContentDownload) in milliseconds
- `totalFormatted` (string, optional): Formatted total time (e.g., `"863.12ms"` or `"0.86s"`)

### CSV Schema

When using `--csv`, logget outputs CSV data with headers. In follow mode, the header is written once at the beginning.

#### Logs CSV

**Columns:**
1. `timestamp` (string): Timestamp in RFC3339 format (ISO 8601)
2. `level` (string): Log level in uppercase (`DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`, `LOG`, `TRACE`)
3. `source` (string): Source of the log (`browser` or `console`)
4. `message` (string): The log message content (may contain commas, properly escaped)

**Example:**
```csv
timestamp,level,source,message
2024-10-31T23:00:00Z,INFO,console,"Application started"
2024-10-31T23:00:01Z,ERROR,browser,"Failed to load resource"
```

#### Network CSV

**Columns:**
1. `timestamp` (string): Timestamp in RFC3339 format (ISO 8601)
2. `method` (string): HTTP method (`GET`, `POST`, `PUT`, `DELETE`, etc.)
3. `url` (string): Request URL (may contain commas, properly escaped)
4. `status` (string): HTTP status code as string (e.g., `"200"`, `"404"`)
5. `resourceType` (string): Resource type (e.g., `Document`, `XHR`, `Image`, `Script`, `Stylesheet`, `Font`, `Media`, `Manifest`, `WebSocket`, `Other`)
6. `mimeType` (string): MIME type of the response (e.g., `text/html`, `application/json`)
7. `size` (string): Size of the response in bytes as string (e.g., `"1234"`)
8. `duration` (string): Total request duration in milliseconds (e.g., `"123.45"`)
9. `ttfb` (string): Time to first byte in milliseconds (e.g., `"50.23"`)
10. `connectTime` (string): Connection time in milliseconds (e.g., `"10.12"`)
11. `dnsTime` (string): DNS lookup time in milliseconds (e.g., `"5.34"`)
12. `sslTime` (string): SSL/TLS handshake time in milliseconds (e.g., `"20.56"`)
13. `sendTime` (string): Time to send request in milliseconds (e.g., `"2.10"`)
14. `waitTime` (string): Wait time (server processing) in milliseconds (e.g., `"30.45"`)
15. `receiveTime` (string): Time to receive response in milliseconds (e.g., `"15.67"`)

**Example:**
```csv
timestamp,method,url,status,resourceType,mimeType,size,duration,ttfb,connectTime,dnsTime,sslTime,sendTime,waitTime,receiveTime
2024-10-31T23:00:00Z,GET,https://example.com/,200,Document,text/html,184,123.45,50.23,10.12,5.34,20.56,2.10,30.45,15.67
2024-10-31T23:00:01Z,GET,https://example.com/api/data,200,XHR,application/json,1234,234.56,100.12,15.23,8.45,25.67,3.20,50.34,20.45
2024-10-31T23:00:02Z,GET,https://example.com/favicon.ico,404,Other,text/html,230,45.67,20.34,5.12,3.45,0.00,1.50,10.23,5.12
```

**CSV Format Notes:**
- Headers are always included in the first line
- Values containing commas, quotes, or newlines are properly escaped according to RFC 4180
- Timestamps use RFC3339 format (e.g., `2024-10-31T23:00:00Z`)
- In follow mode (`-f`), CSV rows are streamed one at a time with the header written once at the start

### HAR Schema

When using `--har`, logget outputs network data in HAR (HTTP Archive) format. HAR files can be analyzed with browser DevTools, HAR viewers, or other analysis tools.

**Usage:**
```bash
# Output HAR to stdout
logget --network --har https://example.com

# Output HAR to file
logget --network --har --output network.har https://example.com
```

**Example HAR Structure:**
```json
{
  "log": {
    "version": "1.2",
    "creator": {
      "name": "logget",
      "version": "2.0"
    },
    "pages": [...],
    "entries": [
      {
        "request": {...},
        "response": {...},
        "timings": {...}
      }
    ]
  }
}
```

## Use Cases

- **Web Development**: Debug JavaScript applications and API calls
- **Testing**: Verify network requests and console output
- **Monitoring**: Track application behavior and performance
- **Security Analysis**: Inspect network traffic and JavaScript execution
- **Automation**: Save results to files for batch processing and analysis
- **Logging**: Create organized log files with timestamps and structured data
- **AI Utility**: Allow your AI agents use this command for efficient debugging

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
