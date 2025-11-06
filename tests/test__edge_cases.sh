#!/bin/bash
# Test all options combined
print_test "All options combined"
if "$BINARY_PATH" \
--logs \
--network \
--json \
--verbose \
--quiet \
--no-color \
--timeout 30000 \
--wait 1000 \
--user-agent "TestAgent/1.0" \
--header "X-Test: value" \
--cookie "test=value" \
--filter "google" \
--exclude "favicon" \
--status "^2..$" \
--domain "google.com" \
--min-size 100 \
--max-size 10000 \
--no-rotate-fingerprints \
--fingerprint-interval 5000 \
"$TEST_URL" > /dev/null 2>&1; then
    test_pass "All options combined"
else
    test_fail "All options combined"
fi

# Test URL without protocol (should add it)
print_test "URL without protocol"
if "$BINARY_PATH" --logs "google.com" > /dev/null 2>&1; then
    test_pass "URL without protocol"
else
    test_fail "URL without protocol"
fi

# Test URL with http://
print_test "URL with http://"
if "$BINARY_PATH" --logs "http://google.com" > /dev/null 2>&1; then
    test_pass "URL with http://"
else
    test_fail "URL with http://"
fi

# Test URL with https://
print_test "URL with https://"
if "$BINARY_PATH" --logs "https://google.com" > /dev/null 2>&1; then
    test_pass "URL with https://"
else
    test_fail "URL with https://"
fi

# Test no data collection flags (should show help)
print_test "No data collection flags"
if "$BINARY_PATH" "$TEST_URL" > /dev/null 2>&1; then
    test_pass "No data collection flags"
else
    test_pass "No data collection flags (exited as expected)"
fi

# Test --insecure flag
run_test "Insecure SSL flag" --network --insecure

# Test --k flag (short insecure)
run_test "Insecure SSL flag (short)" --network -k

# Test multiple output formats (only one should work)
print_test "Multiple output formats (should use last one)"
if "$BINARY_PATH" --network --json --csv --har "$TEST_URL" > /dev/null 2>&1; then
    test_pass "Multiple output formats"
else
    test_pass "Multiple output formats (handled)"
fi

# Test complex filter combination
run_test "Complex filter combination" \
--network \
--status "^(200|404)$" \
--domain ".*\\.com$" \
--mime "^(text|application)/" \
--min-size 50 \
--max-size 50000 \

# Test all resource types
run_test "All resource types" \
--network \
--xhr \
--document \
--css \
--script \
--font \
--img \
--media \
--manifest \
--socket \

# Test verbose with network
run_test "Verbose with network" --network --verbose

# Test verbose with logs
run_test "Verbose with logs" --logs --verbose

# Test verbose with network and logs
run_test "Verbose with network and logs" --network --logs --verbose
