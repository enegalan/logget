#!/bin/bash
# Test --xhr flag
run_test "XHR only filter" --network --xhr

# Test --document flag
run_test "Document only filter" --network --document

# Test --css flag
run_test "CSS only filter" --network --css

# Test --script flag
run_test "Script only filter" --network --script

# Test --font flag
run_test "Font only filter" --network --font

# Test --img flag
run_test "Image only filter" --network --img

# Test --media flag
run_test "Media only filter" --network --media

# Test --manifest flag
run_test "Manifest only filter" --network --manifest

# Test --socket flag
run_test "WebSocket only filter" --network --socket

# Test multiple resource types together
run_test "Multiple resource types" --network --xhr --document
