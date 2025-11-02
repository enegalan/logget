#!/bin/bash
output_file="/tmp/logget_test_output_$$.txt"

# Test JSON output
print_test "JSON output format"
if "$BINARY_PATH" --logs --network --json "$TEST_URL" > "$output_file" 2>&1; then
    if grep -q "\"url\"" "$output_file" || grep -q "\"logs\"" "$output_file" || grep -q "\"network\"" "$output_file"; then
    test_pass "JSON output format"
    else
        test_fail "JSON output format (no JSON structure found)"
    fi
else
    test_fail "JSON output format"
fi

# Test CSV output
print_test "CSV output format"
if "$BINARY_PATH" --logs --network --csv "$TEST_URL" > "$output_file" 2>&1; then
    if grep -q "," "$output_file"; then
        test_pass "CSV output format"
    else
        test_fail "CSV output format (no CSV structure found)"
    fi
else
    test_fail "CSV output format"
fi

# Test HAR output
print_test "HAR output format"
if "$BINARY_PATH" --network --har "$TEST_URL" > "$output_file" 2>&1; then
    if grep -q "\"log\"" "$output_file" || grep -q "\"entries\"" "$output_file"; then
        test_pass "HAR output format"
    else
        test_fail "HAR output format (no HAR structure found)"
    fi
else
    test_fail "HAR output format"
fi

# Test output to file
print_test "Output to file"
test_output="/tmp/logget_test_file_$$.txt"
if "$BINARY_PATH" --logs --output "$test_output" "$TEST_URL" > /dev/null 2>&1; then
    if [ -f "$test_output" ]; then
        test_pass "Output to file"
        rm -f "$test_output"
    else
        test_fail "Output to file (file not created)"
    fi
else
    test_fail "Output to file"
fi

# Test append to file
print_test "Append to file"
test_output="/tmp/logget_test_append_$$.txt"
echo "existing content" > "$test_output"
if "$BINARY_PATH" --logs --output "$test_output" --append "$TEST_URL" > /dev/null 2>&1; then
    if [ -f "$test_output" ] && grep -q "existing content" "$test_output"; then
        test_pass "Append to file"
        rm -f "$test_output"
    else
        test_fail "Append to file (content not appended)"
    fi
else
    test_fail "Append to file"
fi

# Test JSON output to file
print_test "JSON output to file"
test_output="/tmp/logget_test_json_$$.json"
if "$BINARY_PATH" --logs --json --output "$test_output" "$TEST_URL" > /dev/null 2>&1; then
    if [ -f "$test_output" ]; then
        test_pass "JSON output to file"
        rm -f "$test_output"
    else
        test_fail "JSON output to file (file not created)"
    fi
else
    test_fail "JSON output to file"
fi

# Test HAR output to file
print_test "HAR output to file"
test_output="/tmp/logget_test_har_$$.har"
if "$BINARY_PATH" --network --har --output "$test_output" "$TEST_URL" > /dev/null 2>&1; then
    if [ -f "$test_output" ]; then
        test_pass "HAR output to file"
        rm -f "$test_output"
    else
        test_fail "HAR output to file (file not created)"
    fi
else
    test_fail "HAR output to file"
fi
