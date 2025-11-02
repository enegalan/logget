#!/bin/bash
# Test --user-agent flag
run_test "User-Agent flag" --network --user-agent "TestAgent/1.0"

# Test --header flag (single header)
run_test "Single header flag" --network --header "Authorization: Bearer test123"

# Test --header flag (multiple headers)
run_test "Multiple headers flag" --network --header "X-Custom: value1" --header "X-Another: value2"

# Test --cookie flag (single cookie)
run_test "Single cookie flag" --network --cookie "session_id=abc123"

# Test --cookie flag (cookie with attributes)
run_test "Cookie with attributes" --network --cookie "session_id=abc123; domain=.google.com"

# Test --cookie flag (multiple cookies)
run_test "Multiple cookies flag" --network --cookie "session_id=abc123" --cookie "user_token=xyz789"

# Test header file
print_test "Header from file"
header_file="/tmp/test_headers_$$.txt"
echo "Authorization: Bearer test123" > "$header_file"
echo "X-Custom-Header: test-value" >> "$header_file"
if "$BINARY_PATH" --network --header "$header_file" "$TEST_URL" > /dev/null 2>&1; then
    test_pass "Header from file"
else
    test_fail "Header from file"
fi
rm -f "$header_file"

# Test cookie file
print_test "Cookie from file"
cookie_file="/tmp/test_cookies_$$.txt"
echo "session_id=abc123" > "$cookie_file"
echo "user_token=xyz789" >> "$cookie_file"
if "$BINARY_PATH" --network --cookie "$cookie_file" "$TEST_URL" > /dev/null 2>&1; then
    test_pass "Cookie from file"
else
    test_fail "Cookie from file"
fi
rm -f "$cookie_file"

# Test mixed headers and cookies
run_test "Mixed headers and cookies" --network \
--header "X-Custom: value" \
--cookie "session_id=abc123" \
--user-agent "CustomAgent/1.0" \
