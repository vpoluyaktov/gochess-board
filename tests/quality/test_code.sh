#!/bin/bash

# Code quality test script for go-chess
# Tests code formatting, static analysis, race conditions, and unit tests

set -e  # Exit on error
set -o pipefail  # Exit on pipe failures

# Change to project root directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$PROJECT_ROOT"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

print_test() {
    echo -e "${YELLOW}TEST:${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓ PASS:${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

print_failure() {
    echo -e "${RED}✗ FAIL:${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

print_info() {
    echo -e "${BLUE}INFO:${NC} $1"
}

# Run code quality checks
run_code_checks() {
    print_header "Go Code Quality Checks"
    
    # Test 1: go fmt
    print_test "Code formatting (go fmt)"
    unformatted=$(gofmt -l . 2>/dev/null | grep -v ".git" | grep "\.go$" || true)
    if [ -z "$unformatted" ]; then
        print_success "All files are properly formatted"
    else
        print_failure "Found unformatted files:"
        echo "$unformatted" | sed 's/^/  - /'
        print_info "Run: gofmt -w ."
        exit 1
    fi
    
    # Test 2: go vet
    print_test "Static analysis (go vet)"
    if go vet ./... 2>&1 | tee /tmp/vet-output.log | grep -q ""; then
        vet_output=$(cat /tmp/vet-output.log)
        if [ -z "$vet_output" ]; then
            print_success "No issues found"
        else
            print_failure "go vet found issues:"
            cat /tmp/vet-output.log | sed 's/^/  /'
            exit 1
        fi
    else
        print_success "No issues found"
    fi
    
    # Test 3: go test -race (data race detection)
    print_test "Race condition detection (go test -race)"
    if go test -race -short ./... > /tmp/race-output.log 2>&1; then
        print_success "No race conditions detected"
    else
        print_failure "Race conditions detected:"
        cat /tmp/race-output.log | grep -A5 "WARNING: DATA RACE" | sed 's/^/  /'
        exit 1
    fi
    
    # Test 4: go mod verify
    print_test "Module verification (go mod verify)"
    if go mod verify > /tmp/mod-output.log 2>&1; then
        print_success "All modules verified"
    else
        print_failure "Module verification failed:"
        cat /tmp/mod-output.log | sed 's/^/  /'
        exit 1
    fi
    
    # Test 5: go test (all unit tests)
    print_test "Unit tests (go test ./...)"
    if go test ./... > /tmp/test-output.log 2>&1; then
        test_count=$(grep -c "^ok" /tmp/test-output.log || echo "0")
        print_success "All unit tests passed ($test_count packages)"
    else
        print_failure "Unit tests failed:"
        cat /tmp/test-output.log | grep -E "FAIL|Error" | sed 's/^/  /'
        exit 1
    fi
}

# Main execution
main() {
    print_header "Go-Chess Code Quality Test Suite"
    
    # Run code quality checks
    run_code_checks
    
    # Print summary
    print_header "Test Summary"
    echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
    echo -e "Total tests:  $((TESTS_PASSED + TESTS_FAILED))"
    echo ""
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}All tests passed! ✓${NC}\n"
        exit 0
    else
        echo -e "${RED}Some tests failed! ✗${NC}\n"
        exit 1
    fi
}

# Run main
main "$@"
