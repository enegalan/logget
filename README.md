# logget

A command-line tool to extract browser logs and network data from web pages using an embedded Chromium browser.

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

## Links

- [Installation](https://enegalan.github.io/logget-doc/docs/getting-started/installation)
- [Basic Usage](https://enegalan.github.io/logget-doc/docs/getting-started/quick-start#basic-usage)

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
