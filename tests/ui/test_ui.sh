#!/bin/bash

# UI Test script for go-chess using Playwright
# Tests the web interface and user interactions

set -e
set -o pipefail

# Change to script directory (tests/ui/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Project root is two levels up
PROJECT_ROOT="$(cd ../.. && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PORT="${PORT:-35256}"
BASE_URL="http://localhost:${PORT}"
PLAYWRIGHT_DIR="playwright"

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
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

print_warning() {
    echo -e "${YELLOW}WARNING:${NC} $1"
}

# Check if Node.js is installed
check_nodejs() {
    if ! command -v node &> /dev/null; then
        echo -e "${RED}Error: Node.js is not installed.${NC}"
        echo "Please install Node.js first:"
        echo "  Ubuntu/Debian: sudo apt-get install nodejs npm"
        echo "  macOS: brew install node"
        exit 1
    fi
    
    node_version=$(node --version)
    print_info "Node.js version: $node_version"
}

# Check if Playwright is installed
check_playwright() {
    if [ ! -d "$PLAYWRIGHT_DIR/node_modules" ]; then
        print_warning "Playwright not installed"
        return 1
    fi
    
    if [ ! -f "$PLAYWRIGHT_DIR/node_modules/.bin/playwright" ]; then
        print_warning "Playwright binary not found"
        return 1
    fi
    
    return 0
}

# Setup Playwright
setup_playwright() {
    print_header "Setting Up Playwright"
    
    if [ ! -f "$PLAYWRIGHT_DIR/package.json" ]; then
        echo -e "${RED}Error: package.json not found in $PLAYWRIGHT_DIR${NC}"
        echo "The Playwright configuration files should be in tests/ui/playwright/"
        exit 1
    fi
    
    # Install Playwright
    print_info "Installing Playwright (this may take a few minutes)..."
    cd "$PLAYWRIGHT_DIR"
    npm install
    npx playwright install chromium
    cd - > /dev/null
    
    print_success "Playwright installed successfully"
}

# Start the server
start_server() {
    print_header "Starting Server"
    
    # Check if server is already running
    if curl -s "$BASE_URL" > /dev/null 2>&1; then
        print_info "Server already running at $BASE_URL"
        return
    fi
    
    # Build if needed
    if [ ! -f "$PROJECT_ROOT/go-chess" ]; then
        print_info "Building application..."
        (cd "$PROJECT_ROOT" && go build)
    fi
    
    print_info "Starting server on port $PORT..."
    (cd "$PROJECT_ROOT" && ./go-chess --restart --no-browser --no-tui > /tmp/ui-test-server.log 2>&1) &
    SERVER_PID=$!
    
    print_info "Server PID: $SERVER_PID"
    print_info "Waiting for server to start..."
    sleep 3
    
    # Check if server is running
    if ps -p $SERVER_PID > /dev/null; then
        print_success "Server started successfully"
    else
        print_failure "Server failed to start"
        cat /tmp/ui-test-server.log
        exit 1
    fi
}

# Stop the server
stop_server() {
    if [ ! -z "$SERVER_PID" ] && ps -p $SERVER_PID > /dev/null; then
        print_info "Stopping server (PID: $SERVER_PID)"
        kill $SERVER_PID 2>/dev/null || true
        sleep 1
    fi
}

# Run Playwright tests
run_playwright_tests() {
    print_header "Running Playwright Tests"
    
    cd "$PLAYWRIGHT_DIR"
    
    if npx playwright test; then
        print_success "All Playwright tests passed"
        cd - > /dev/null
        return 0
    else
        print_failure "Some Playwright tests failed"
        print_info "View detailed report: cd tests/ui/playwright && npm run report"
        cd - > /dev/null
        return 1
    fi
}

# Main execution
main() {
    print_header "Go-Chess UI Test Suite"
    
    # Parse arguments
    if [ "$1" = "--setup" ] || [ "$1" = "-s" ]; then
        check_nodejs
        setup_playwright
        print_success "UI test setup complete!"
        print_info "Run tests with: ./test_ui.sh"
        exit 0
    fi
    
    # Check dependencies
    check_nodejs
    
    if ! check_playwright; then
        print_warning "Playwright not installed. Run with --setup to install:"
        print_info "  ./test_ui.sh --setup"
        exit 1
    fi
    
    # Start server
    start_server
    trap stop_server EXIT
    
    # Run tests
    if run_playwright_tests; then
        print_header "Test Summary"
        echo -e "${GREEN}All UI tests passed! ✓${NC}\n"
        exit 0
    else
        print_header "Test Summary"
        echo -e "${RED}Some UI tests failed! ✗${NC}\n"
        exit 1
    fi
}

# Run main
main "$@"
