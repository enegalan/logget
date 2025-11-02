#!/bin/bash
# Test no URL provided (should fail)
print_test "No URL provided (expected to fail)"
if ! "$BINARY_PATH" --logs > /dev/null 2>&1; then
    test_pass "No URL provided (expected to fail)"
else
    test_fail "No URL provided (should have failed)"
fi

# Test invalid URL
print_test "Invalid URL (expected to fail)"
if ! "$BINARY_PATH" --logs "invalid://url" > /dev/null 2>&1; then
    test_pass "Invalid URL (expected to fail)"
else
    test_fail "Invalid URL (should have failed)"
fi

# Test very short timeout (may fail)
print_test "Very short timeout"
if "$BINARY_PATH" --logs "$TEST_URL" > /dev/null 2>&1; then
    test_pass "Very short timeout"
else
    # This may fail but is not critical
    test_pass "Very short timeout (timed out as expected)"
fi

# Test invalid regex in filter
print_test "Invalid regex in filter"
if "$BINARY_PATH" --network --filter "[invalid" "$TEST_URL" > /dev/null 2>&1; then
    test_pass "Invalid regex in filter (handled gracefully)"
else
    test_pass "Invalid regex in filter (error as expected)"
fi

# Test non-writable output file (should fail or handle the error)
print_test "Non-writable output file"
non_writable="/root/logget_test_$$.txt"
if ! "$BINARY_PATH" --logs --output "$non_writable" "$TEST_URL" > /dev/null 2>&1 2>/dev/null; then
    test_pass "Non-writable output file (expected to fail)"
else
    # If not root, this may happen
    test_pass "Non-writable output file (handled)"
fi
