#!/bin/bash
# Test -L flag (short for --logs)
run_test "Short -L flag for logs" -L

# Test -N flag (short for --network)
run_test "Short -N flag for network" -N

# Test -J flag (short for --json)
run_test "Short -J flag for JSON" --logs --network -J

# Test -W flag (short for --wait)
run_test "Short -W flag for wait" --logs -W 5000

# Test -V flag (short for --verbose)
run_test "Short -V flag for verbose" --logs --network -V

