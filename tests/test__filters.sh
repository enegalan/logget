#!/bin/bash
# Test --filter flag
run_test "Filter pattern flag" --network --filter "google"

# Test --exclude flag
run_test "Exclude pattern flag" --network --exclude "favicon"

# Test --filter and --exclude together
run_test "Filter and exclude together" --network --filter "google" --exclude "favicon"

# Test --status filter
run_test "Status filter (2xx)" --network --status "^2..$"

# Test --status filter (specific status)
run_test "Status filter (200)" --network --status "200"

# Test --domain filter
run_test "Domain filter" --network --domain "google.com"

# Test --domain filter with regex
run_test "Domain filter regex" --network --domain "(.*\\.)?google\\.com$"

# Test --mime filter
run_test "MIME filter" --network --mime "text/html"

# Test --mime filter with regex
run_test "MIME filter regex" --network --mime "^text/"

# Test multiple filters together
run_test "Multiple filters" --network --status "^2..$" --domain "google.com" --mime "text/html"
