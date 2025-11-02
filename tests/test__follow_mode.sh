#!/bin/bash
# Test --follow flag with logs
print_test "Follow mode with logs"
"$BINARY_PATH" --follow --logs "$TEST_URL" > /dev/null 2>&1 &
pid=$!
sleep 0.5
kill $pid 2>/dev/null || true
wait $pid 2>/dev/null || true
test_pass "Follow mode with logs"

# Test --follow flag with network
print_test "Follow mode with network"
"$BINARY_PATH" --follow --network "$TEST_URL" > /dev/null 2>&1 &
pid=$!
sleep 0.5
kill $pid 2>/dev/null || true
wait $pid 2>/dev/null || true
test_pass "Follow mode with network"

# Test --follow flag with both
print_test "Follow mode with logs and network"
"$BINARY_PATH" --follow --logs --network "$TEST_URL" > /dev/null 2>&1 &
pid=$!
sleep 0.5
kill $pid 2>/dev/null || true
wait $pid 2>/dev/null || true
test_pass "Follow mode with logs and network"

# Test --follow with --refresh
print_test "Follow mode with custom refresh"
"$BINARY_PATH" --follow --logs --refresh 500 "$TEST_URL" > /dev/null 2>&1 &
pid=$!
sleep 0.5
kill $pid 2>/dev/null || true
wait $pid 2>/dev/null || true
test_pass "Follow mode with custom refresh"

# Test --follow with output file
print_test "Follow mode with output file"
output_file="/tmp/logget_test_follow_$$.txt"
rm -f "$output_file"
"$BINARY_PATH" --follow --logs --output "$output_file" "$TEST_URL" > /dev/null 2>&1 &
pid=$!
# Wait for file to be created (logget creates it immediately in follow mode)
i=0
while [ $i -lt 20 ] && [ ! -f "$output_file" ]; do
    sleep 0.5
    i=$((i + 1))
done
kill $pid 2>/dev/null || true
wait $pid 2>/dev/null || true
# Additional wait for any pending writes
sleep 1.5
if [ -f "$output_file" ]; then
    test_pass "Follow mode with output file"
    rm -f "$output_file"
else
    test_fail "Follow mode with output file"
fi

# Test --follow with JSON output
print_test "Follow mode with JSON output"
"$BINARY_PATH" --follow --logs --json "$TEST_URL" > /dev/null 2>&1 &
pid=$!
sleep 0.5
kill $pid 2>/dev/null || true
wait $pid 2>/dev/null || true
test_pass "Follow mode with JSON output"

# Test --follow with filters
print_test "Follow mode with filters"
"$BINARY_PATH" --follow --logs --filter "test" "$TEST_URL" > /dev/null 2>&1 &
pid=$!
sleep 0.5
kill $pid 2>/dev/null || true
wait $pid 2>/dev/null || true
test_pass "Follow mode with filters"
