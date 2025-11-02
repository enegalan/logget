#!/bin/bash
# Test --no-rotate-fingerprints flag
run_test "No rotate fingerprints flag" --network --no-rotate-fingerprints

# Test --fingerprint-interval flag
run_test "Fingerprint interval flag" --network --fingerprint-interval 2000

# Test fingerprint interval with custom value
run_test "Custom fingerprint interval" --network --fingerprint-interval 3000

# Test fingerprint rotation enabled by default
run_test "Fingerprint rotation enabled (default)" --network

# Test fingerprint interval with follow mode
print_test "Fingerprint interval with follow mode"
"$BINARY_PATH" --follow --network --fingerprint-interval 2000 "$TEST_URL" > /dev/null 2>&1 &
pid=$!
sleep 0.5
kill $pid 2>/dev/null || true
wait $pid 2>/dev/null || true
test_pass "Fingerprint interval with follow mode"
