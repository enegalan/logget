# Contribution Guidelines

Thank you for your interest in contributing to `logget`! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)

## Code of Conduct

Please note that this project is released with a [Contributor Code of Conduct](https://github.com/enegalan/logget/blob/main/CODE_OF_CONDUCT.md). By participating in this project you agree to abide by its terms.

## Getting Started

Before you begin:

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/your-username/logget.git
   cd logget
   ```
3. **Add the upstream remote**:
   ```bash
   git remote add upstream https://github.com/egalan/logget.git
   ```

## Development Setup

### Prerequisites

- **Go 1.24 or higher** - Check with `go version`
- **Make** - For running build commands
- **Bash 4.0+** - For running tests

### Initial Setup

1. **Install dependencies**:
   ```bash
   make deps
   ```

2. **Build the project**:
   ```bash
   make build
   ```

3. **Run tests** to verify everything works:
   ```bash
   make test
   ```

### Available Commands

- `make build` - Build for current platform
- `make build-all` - Build for all platforms (Linux, Windows, macOS)
- `make clean` - Clean build artifacts
- `make test` - Run all tests
- `make deps` - Install/update dependencies
- `make install` - Install binary to system
- `make uninstall` - Remove binary from system
- `make release` - Create release packages
- `make help` - Show all available commands

## Project Structure

```
logget/
├── main.go              # Entry point and CLI flag definitions
├── go.mod               # Go module definition
├── Makefile             # Build automation
├── VERSION              # Current version number
├── src/                 # Source code
│   ├── chrome/          # Chromium browser integration
│   ├── colors/          # Color output utilities
│   ├── command/         # Command execution and state management
│   ├── core/            # Core functionality (logger, HAR, headers, etc.)
│   ├── flags/           # Flag definitions and parsing
│   ├── helpers/         # Helper utilities (filters, regex, URL, etc.)
│   └── io/              # File I/O operations
├── scripts/             # Build and release scripts
├── tests/               # Test suite (bash scripts)
└── build/               # Build output directory
```

## Coding Standards

### Language and Style

- **Use Go exclusively** for source code changes
- **Follow [`gofmt`](https://go.dev/blog/gofmt) style** - The codebase uses standard Go formatting
- **Use tabs `\t` for indentation** in `*.go` files and `Makefile`
- **Use clear, descriptive names** - Avoid abbreviations
- **Keep functions small** - Favor early returns and avoid deep nesting
- **Error handling** - Return errors explicitly; do not panic except in `main()` startup failures

### Code Organization

- **Do not move or rename files** unless explicitly requested
- **Do not modify** `LICENSE` and `VERSION`
- **Keep helpers under `helpers/`** - Avoid cross-cutting utilities outside this folder
- **Use English** for code, comments, commit messages, and documentation

### Logging and Output

- **Use helpers in `src/core/logger.go`** - Respect existing logging levels and formats
- **Use color helpers in `src/colors/`** - Respect existing coloring structure
- **Honor existing formatting options** in `src/command/formatter.go` and `src/command/output.go`

### Performance and Safety

- **Prefer streaming/iterative processing** over loading entire data into memory when feasible
- **Avoid global mutable state** - Pass dependencies explicitly
- **Do not add network calls or telemetry** without explicit approval

### Documentation

- **Update [documentation](https://github.com/enegalan/logget-doc)** when user-facing behavior or flags change
- **Add brief package-level or function doc comments** for non-obvious logic
- **Keep comments brief** and only where they aid maintainability

## Testing

### Running Tests

Run all tests:
```bash
make test
```

Or run tests directly:
```bash
cd tests && ./run_tests.sh
```

### Writing Tests

Tests are located in the `tests/` directory and are written as bash scripts:

1. **Create a new test file** following the pattern `test__*.sh`
2. **Use the `run_test` function** to execute tests:
   ```bash
   #!/bin/bash
   run_test "Test name" --logs --network --timeout 10000
   ```
3. **Test files are automatically executed** by `run_tests.sh`

See [`tests/README.md`](https://github.com/enegalan/logget/blob/main/tests/README.md) for detailed testing guidelines and available functions.

### Test Coverage

Tests should cover:
- All boolean flags and their combinations
- All value flags with valid and invalid inputs
- Error cases and edge cases
- Different output formats (JSON, YAML, CSV, HAR)
- Operation modes (normal, follow, verbose, etc.)

## Submitting Changes

### Before You Submit

1. **Ensure your code compiles**:
   ```bash
   make build
   ```

2. **Run all tests with 100% coverage**:
   ```bash
   make test
   ```

3. **Format your code (in case you didn't)**:
   ```bash
   gofmt -w .
   ```

4. **Update documentation** if you've changed user-facing behavior

### Creating a Branch

Create a branch from `main`:
```bash
git checkout main
git pull upstream main
git checkout -b feature/your-feature-name
```

Use descriptive branch names:
- `feature/add-new-filter`
- `fix/header-parsing-bug`
- `docs/update-readme`

## Commit Guidelines

### Commit Message Format

Write concise, imperative commit messages:

- ✅ Good: `Add filter for header matching`
- ✅ Good: `Fix timeout handling for long-running requests`
- ✅ Good: `Update README with new flag documentation`
- ❌ Bad: `fixed some bugs`
- ❌ Bad: `WIP`
- ❌ Bad: `changes`

### Commit Best Practices

- **Group related changes** - Avoid large, mixed edits
- **Make atomic commits** - Each commit should represent a single logical change
- **Write clear commit messages** - Explain what and why, not how
- **Use English** for all commit messages

## Pull Request Process

1. **Push your branch** to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create a Pull Request** on GitHub:
   - Provide a clear title and description
   - Reference any related issues
   - Explain what changes you made and why

3. **Respond to feedback** - Be open to suggestions and ready to make changes

### PR Checklist

Before submitting, ensure:

- [ ] Code compiles without errors
- [ ] All tests pass
- [ ] Code follows project style guidelines
- [ ] Documentation is updated (if needed)
- [ ] Commit messages are clear and follow guidelines
- [ ] No unnecessary files are included

## What to Contribute

Contributions are welcome in many forms:

- **Bug fixes** - Report and fix issues
- **New features** - Propose and implement new functionality
- **Documentation** - Improve docs, add examples. Check it out [here](https://github.com/enegalan/logget-doc).
- **Tests** - Add more tests and increase test coverage
- **Code quality** - Refactoring and improvements
- **Performance** - Optimizations

### Feature Requests

Before implementing a major feature:
1. Open an issue to discuss the feature
2. Wait for feedback and approval
3. Then proceed with implementation

## Questions?

If you have questions or need help:

- Open an issue on GitHub
- Check existing issues and discussions
- Review the codebase and documentation

Thank you for contributing to `logget`! 🎉
