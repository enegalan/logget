#!/bin/bash

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$TEST_DIR/.." && pwd)"

COLORS_GO="$PROJECT_ROOT/src/colors.go"
COLORS_SH="$PROJECT_ROOT/scripts/colors.sh"
if [ ! -f "$COLORS_SH" ] || [ "$COLORS_GO" -nt "$COLORS_SH" ]; then
    "$PROJECT_ROOT/scripts/generate_colors.sh"
fi
source "$COLORS_SH"

BINARY_PATH="$PROJECT_ROOT/logget"
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0


if [ ! -f "$BINARY_PATH" ]; then
    echo -e "${RED}Error: logget binary not found at $BINARY_PATH${NC}"
    echo -e "${YELLOW}Please build logget first: go build -o logget .${NC}"
    exit 1
fi

TEST_URL="https://google.com"

# Function to count expected total tests by analyzing test files
count_expected_tests() {
    local count=0
    for test_file in "$TEST_DIR"/test__*.sh; do
        if [ -f "$test_file" ] && [ "$(basename "$test_file")" != "run_tests.sh" ]; then
            local run_test_count=0
            if grep -q "^run_test " "$test_file" 2>/dev/null; then
                run_test_count=$(grep -c "^run_test " "$test_file" 2>/dev/null)
                run_test_count=$(echo "$run_test_count" | tr -d ' ')
                [ -z "$run_test_count" ] && run_test_count=0
            fi
            local print_test_count=0
            if grep -q "^print_test " "$test_file" 2>/dev/null; then
                print_test_count=$(grep -c "^print_test " "$test_file" 2>/dev/null)
                print_test_count=$(echo "$print_test_count" | tr -d ' ')
                [ -z "$print_test_count" ] && print_test_count=0
            fi
            count=$((count + run_test_count + print_test_count))
        fi
    done
    echo $count
}

print_test() {
    local next_test=$((TESTS_TOTAL + 1))
    echo -e "${BLUE}▶ Testing: $1${NC} ${GRAY}[$next_test/$EXPECTED_TOTAL_TESTS]${NC}"
}

test_pass() {
    ((TESTS_PASSED++))
    ((TESTS_TOTAL++))
    echo -e "${GREEN}✓ PASS: $1${NC} ${GRAY}[$TESTS_TOTAL/$EXPECTED_TOTAL_TESTS]${NC}"
}

test_fail() {
    ((TESTS_FAILED++))
    ((TESTS_TOTAL++))
    echo -e "${RED}✗ FAIL: $1${NC} ${GRAY}[$TESTS_TOTAL/$EXPECTED_TOTAL_TESTS]${NC}"
}

test_result() {
    if [ $? -eq 0 ]; then
        test_pass "$1"
    else
        test_fail "$1"
    fi
}

cleanup() {
    rm -f /tmp/logget_test_*.txt /tmp/logget_test_*.json /tmp/logget_test_*.csv /tmp/logget_test_*.har
    rm -f /tmp/test_headers.txt /tmp/test_cookies.txt
}

run_test() {
    local test_name="$1"
    shift
    print_test "$test_name"
    local output_file="/tmp/logget_test_output_$$.txt"
    if "$BINARY_PATH" "$@" "$TEST_URL" > "$output_file" 2>&1; then
        test_pass "$test_name"
        return 0
    else
        test_fail "$test_name"
        echo -e "${YELLOW}  Output:${NC}"
        head -n 5 "$output_file" | sed 's/^/    /'
        return 1
    fi
}

run_test_expect_fail() {
    local test_name="$1"
    shift
    print_test "$test_name (expected to fail)"
    local output_file="/tmp/logget_test_output_$$.txt"
    if ! "$BINARY_PATH" "$@" "$TEST_URL" > "$output_file" 2>&1; then
        test_pass "$test_name"
        return 0
    else
        test_fail "$test_name (should have failed)"
        return 1
    fi
}

run_test_expect_output() {
    local test_name="$1"
    local expected="$2"
    shift 2
    print_test "$test_name"
    local output_file="/tmp/logget_test_output_$$.txt"
    if "$BINARY_PATH" "$@" "$TEST_URL" > "$output_file" 2>&1 && grep -q "$expected" "$output_file"; then
        test_pass "$test_name"
        return 0
    else
        test_fail "$test_name"
        echo -e "${YELLOW}  Expected to find: $expected${NC}"
        return 1
    fi
}


echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  logget Test Suite${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

EXPECTED_TOTAL_TESTS=$(count_expected_tests)
START_TIME=$(date +%s)

cleanup

echo -e "${BLUE}Running all test files...${NC}"
echo -e "${GRAY}Expected total tests: $EXPECTED_TOTAL_TESTS${NC}"
echo ""

for test_file in "$TEST_DIR"/test__*.sh; do
    if [ -f "$test_file" ] && [ "$(basename "$test_file")" != "run_tests.sh" ]; then
        test_name=$(basename "$test_file" .sh | sed 's/^test__//' | tr '_' ' ')
        echo -e "${BLUE}→ Running: ${test_name}${NC}"
        source "$test_file"
        echo ""
    fi
done

cleanup

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Test Summary${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${GREEN}Total Tests:  $TESTS_TOTAL${NC}"
echo -e "${GREEN}Passed:       $TESTS_PASSED${NC}"
echo -e "${RED}Failed:       $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_TOTAL -gt 0 ]; then
    SUCCESS_RATE=$((TESTS_PASSED * 100 / TESTS_TOTAL))
    echo -e "${BLUE}Success Rate: ${SUCCESS_RATE}%${NC}"
    END_TIME=$(date +%s)
    ELAPSED_TIME=$((END_TIME - START_TIME))
    MINUTES=$((ELAPSED_TIME / 60))
    SECONDS=$((ELAPSED_TIME % 60))
    if [ $MINUTES -gt 0 ]; then
        echo -e "${BLUE}Total Time: ${MINUTES}m ${SECONDS}s${NC}"
    else
        echo -e "${BLUE}Total Time: ${SECONDS}s${NC}"
    fi
    if [ $TESTS_FAILED -gt 0 ]; then
        echo -e "${RED}✗ Some tests failed${NC}"
        EXIT_CODE=1
    else
        echo -e "${GREEN}✓ All tests passed!${NC}"
        EXIT_CODE=0
    fi
else
    echo -e "${RED}No tests executed!${NC}"
    EXIT_CODE=1
fi

exit ${EXIT_CODE:-1}

