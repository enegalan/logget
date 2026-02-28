#!/bin/bash
# Tests for --interact and --wait-after-interaction (and combinations with existing flags).
# Uses selectors that exist on typical test page (e.g. body, a).

# New parameters: interact wait only
run_test "interact wait only" --logs --interact "wait:100"

# interact click on element that exists (body)
run_test "interact click:body" --logs --interact "click:body"

# interact focus (first link is focusable)
run_test "interact focus" --logs --interact "focus:a"

# interact hover
run_test "interact hover" --logs --interact "hover:a"

# interact key
run_test "interact key:Tab" --logs --interact "key:Tab"

# interact wait + wait-after-interaction
run_test "interact wait and wait-after-interaction" --logs --interact "wait:200" --wait-after-interaction 500

# expect fail: selector does not exist (short timeout so it fails quickly)
run_test_expect_fail "interact click non-existent selector" --logs --timeout 5000 --interact "click:#selectorQueNoExiste"

# Mix: interact with --logs
run_test "interact with --logs" --logs --interact "wait:100"

# Mix: interact with --network
run_test "interact with --network" --network --interact "wait:100"

# Mix: interact with --logs --network --json --quiet --no-color
run_test "interact with logs network json quiet no-color" --logs --network --json --quiet --no-color --interact "wait:100"

# Mix: interact with JSON output and check structure
run_test_expect_output "interact with JSON has url or logs/network" "\"url\"" --logs --json --interact "wait:100"

# Mix: interact with --wait
run_test "interact with --wait" --logs --interact "wait:100" --wait 200

# Mix: interact with --timeout
run_test "interact with --timeout" --logs --interact "wait:100" --timeout 60000

# Mix: interact with --execute
run_test "interact with --execute" --logs --execute "1+1" --interact "wait:100"

# Mix: interact with --output
run_test "interact with --output" --logs --interact "wait:100" --output /tmp/logget_test_interact_out_$$.txt
rm -f /tmp/logget_test_interact_out_$$.txt

# Mix: interact with --filter
run_test "interact with --filter" --logs --interact "wait:100" --filter "."
