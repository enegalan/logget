# Tests for logget

This folder contains bash scripts to test all logget functionalities and achieve 100% coverage of all possible options.

## Simplified Structure

The test structure is very simple: each `test__*.sh` file is automatically executed when you run `run_tests.sh`.

### How to Add a New Test

1. Create a new file `test__my_new_test.sh` in this folder
2. Add tests directly (no wrapper functions needed)

Example:

```bash
#!/bin/bash
# Test 1
run_test "My test 1" --logs --network --timeout 10

# Test 2
run_test "My test 2" --network --filter "test" --timeout 10
```

That's it! The file will be executed automatically.

## Files

- `run_tests.sh` - Main script that executes all tests and generates coverage report
- `test__*.sh` - Individual test files (automatically executed)

## Available Functions

When writing a test, you have access to these functions:

- `run_test "name" [flags...]` - Runs a test with the given flags
- `test_pass "name"` - Marks a test as passed
- `test_fail "name"` - Marks a test as failed
- `print_test "name"` - Prints the test name

You also have access to these variables:
- `$BINARY_PATH` - Path to the logget binary
- `$TEST_URL` - Test URL (default: https://example.com)
- `$TEST_DIR` - Test directory
- `$PROJECT_ROOT` - Project root

## Usage

### Run all tests

```bash
./tests/run_tests.sh
```

The coverage report is automatically generated at the end of `run_tests.sh`.

### Run an individual test

You can run an individual test directly (will be slower because it doesn't have utility functions):

```bash
source ./tests/run_tests.sh  # Load utility functions
source ./tests/test__my_test.sh  # Execute the test
```

## Requirements

- The `logget` binary must be compiled in the project root
- Bash 4.0 or higher
- `timeout` command (available on most Unix systems)

## Coverage

Tests are designed to cover all possible logget options:

- All boolean flags (--logs, --network, --json, etc.)
- All value flags (--timeout, --wait, --header, etc.)
- All option combinations
- Error cases and edge cases
- Different output formats (JSON, CSV, HAR)
- Operation modes (normal, follow, verbose, etc.)
