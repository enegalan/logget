# logget

A command-line tool to extract browser logs and network data from web pages using an embedded Chromium browser.

## Features

- **Console Log Collection**: Capture all console.log, console.error, console.warn, and console.info messages
- **Network Monitoring**: Track all HTTP requests (fetch, XMLHttpRequest) with headers, status codes, and timing
- **JavaScript Execution**: Execute JavaScript code in the page context for debugging and testing (supports inline code or file path)
- **Interactions**: Run page actions (click, focus, hover, type, select, key) via repeatable `--interact` for testing flows and capturing resulting logs/network; key names follow `event.key` (e.g. Escape, Tab, Control+Enter)
- **Cross-Platform**: Works on Windows, Linux, and macOS
- **No Chrome Required**: Automatically downloads and uses Chromium via `rod`
- **Custom Headers**: Add custom HTTP headers like curl (supports files)
- **Cookie Support**: Set cookies for authenticated requests (supports files)
- **Configurable Timeout**: Set custom timeout values

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
