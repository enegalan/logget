#!/bin/bash
# Test --min-size filter
run_test "Min-size filter" --network --min-size 100

# Test --max-size filter
run_test "Max-size filter" --network --max-size 10000

# Test --min-size and --max-size together
run_test "Min-size and max-size together" --network --min-size 100 --max-size 10000

# Test with larger min-size
run_test "Large min-size filter" --network --min-size 1024

# Test with smaller max-size
run_test "Small max-size filter" --network --max-size 100

# Test with size range
run_test "Size range filter" --network --min-size 500 --max-size 5000
