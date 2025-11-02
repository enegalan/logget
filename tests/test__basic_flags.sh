#!/bin/bash
# Test --logs flag
run_test "Basic --logs flag" --logs

# Test --network flag
run_test "Basic --network flag" --network

# Test --logs and --network together
run_test "Combined --logs --network flags" --logs --network

# Test --verbose flag
run_test "Verbose flag" --verbose

# Test --version flag
print_test "Version flag"
if "$BINARY_PATH" --version > /tmp/logget_test_output_$$.txt 2>&1; then
    test_pass "Version flag"
else
    test_fail "Version flag"
fi

# Test --quiet flag
run_test "Quiet flag" --logs --quiet

# Test --no-color flag
run_test "No-color flag" --logs --no-color

# Test timeout flag
run_test "Custom timeout" --logs

# Test wait flag
run_test "Custom wait time" --logs --wait 5000
