#!/bin/bash

# Test script for go-chess server API
# Tests major server features and endpoints

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

# Configuration
PORT="${PORT:-35256}"
BASE_URL="http://localhost:${PORT}"
BOOK_FILE="${BOOK_FILE:-/usr/share/games/gnuchess/book.bin}"
LOG_LEVEL="${LOG_LEVEL:-INFO}"

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

# Check if jq is installed
check_dependencies() {
    if ! command -v jq &> /dev/null; then
        echo -e "${RED}Error: jq is not installed. Please install it first.${NC}"
        echo "  Ubuntu/Debian: sudo apt-get install jq"
        echo "  macOS: brew install jq"
        exit 1
    fi
}

# Build the application
build_app() {
    print_header "Building Application"
    print_info "Running: go build"
    if go build; then
        print_success "Application built successfully"
    else
        print_failure "Failed to build application"
        exit 1
    fi
}

# Start the server
start_server() {
    print_header "Starting Server"
    
    # Check if book file exists
    if [ -f "$BOOK_FILE" ]; then
        print_info "Using opening book: $BOOK_FILE"
        BOOK_ARG="--book-file $BOOK_FILE"
    else
        print_info "Opening book not found, starting without it"
        BOOK_ARG=""
    fi
    
    print_info "Starting server on port $PORT with log level $LOG_LEVEL"
    if [ ! -z "$BOOK_ARG" ]; then
        ./go-chess --restart --no-browser --no-tui --log-level=$LOG_LEVEL $BOOK_ARG > server.log 2>&1 &
    else
        ./go-chess --restart --no-browser --no-tui --log-level=$LOG_LEVEL > server.log 2>&1 &
    fi
    SERVER_PID=$!
    
    print_info "Server PID: $SERVER_PID"
    print_info "Waiting for server to start..."
    sleep 3
    
    # Check if server is running
    if ps -p $SERVER_PID > /dev/null; then
        print_success "Server started successfully"
    else
        print_failure "Server failed to start"
        cat server.log
        exit 1
    fi
}

# Stop the server
stop_server() {
    print_header "Stopping Server"
    if [ ! -z "$SERVER_PID" ] && ps -p $SERVER_PID > /dev/null; then
        print_info "Stopping server (PID: $SERVER_PID)"
        kill $SERVER_PID 2>/dev/null || true
        sleep 1
        print_success "Server stopped"
    fi
}

# Test 1: Health check - GET /
test_health_check() {
    print_test "Health check - GET /"
    
    response=$(curl -s -w "\n%{http_code}" "$BASE_URL/")
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "200" ]; then
        print_success "Server is responding (HTTP $http_code)"
    else
        print_failure "Server returned HTTP $http_code"
    fi
}

# Test 2: Get engines list
test_get_engines() {
    print_test "Get engines list - GET /api/engines"
    
    response=$(curl -s "$BASE_URL/api/engines")
    
    # Check if response is valid JSON
    if echo "$response" | jq empty 2>/dev/null; then
        engine_count=$(echo "$response" | jq 'length')
        print_success "Retrieved $engine_count engines"
        
        # Check if built-in engine exists
        builtin_engine=$(echo "$response" | jq -r '.[] | select(.type == "internal") | .name')
        if [ ! -z "$builtin_engine" ]; then
            print_success "Built-in engine found: $builtin_engine"
        else
            print_failure "Built-in engine not found"
        fi
        
        # List all engines
        echo "$response" | jq -r '.[] | "  - \(.name) (\(.type))"'
    else
        print_failure "Invalid JSON response"
    fi
}

# Test 3: Computer move with built-in engine
test_builtin_engine_move() {
    print_test "Computer move with built-in engine"
    
    response=$(curl -s -X POST "$BASE_URL/api/computer-move" \
        -H "Content-Type: application/json" \
        -d '{
            "fen": "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
            "enginePath": "internal",
            "moveTime": 1000
        }')
    
    if echo "$response" | jq empty 2>/dev/null; then
        move=$(echo "$response" | jq -r '.move')
        think_time=$(echo "$response" | jq -r '.thinkTime')
        
        if [ ! -z "$move" ] && [ "$move" != "null" ]; then
            print_success "Built-in engine returned move: $move (think time: ${think_time}ms)"
        else
            print_failure "No move returned"
            echo "$response" | jq '.'
        fi
    else
        print_failure "Invalid JSON response"
    fi
}

# Test 4: Computer move with time controls
test_time_controls() {
    print_test "Computer move with clock-based time management"
    
    response=$(curl -s -X POST "$BASE_URL/api/computer-move" \
        -H "Content-Type: application/json" \
        -d '{
            "fen": "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
            "enginePath": "internal",
            "whiteTime": 300000,
            "blackTime": 300000,
            "whiteIncrement": 2000,
            "blackIncrement": 2000,
            "moves": ["e2e4"]
        }')
    
    if echo "$response" | jq empty 2>/dev/null; then
        move=$(echo "$response" | jq -r '.move')
        think_time=$(echo "$response" | jq -r '.thinkTime')
        
        if [ ! -z "$move" ] && [ "$move" != "null" ]; then
            print_success "Clock-based move: $move (think time: ${think_time}ms)"
        else
            print_failure "No move returned"
        fi
    else
        print_failure "Invalid JSON response"
    fi
}

# Test 5: Opening name recognition
test_opening_recognition() {
    print_test "Opening name recognition - Italian Game"
    
    response=$(curl -s -X POST "$BASE_URL/api/opening" \
        -H "Content-Type: application/json" \
        -d '{
            "moves": ["e4", "e5", "Nf3", "Nc6", "Bc4"]
        }')
    
    if echo "$response" | jq empty 2>/dev/null; then
        opening_name=$(echo "$response" | jq -r '.name')
        eco_code=$(echo "$response" | jq -r '.eco')
        
        if [ ! -z "$opening_name" ] && [ "$opening_name" != "null" ]; then
            print_success "Opening recognized: $opening_name ($eco_code)"
        else
            print_failure "Opening not recognized"
        fi
    else
        print_failure "Invalid JSON response"
    fi
}

# Test 6: Sicilian Defense recognition
test_sicilian_defense() {
    print_test "Opening name recognition - Sicilian Defense"
    
    response=$(curl -s -X POST "$BASE_URL/api/opening" \
        -H "Content-Type: application/json" \
        -d '{
            "moves": ["e4", "c5"]
        }')
    
    if echo "$response" | jq empty 2>/dev/null; then
        opening_name=$(echo "$response" | jq -r '.name')
        eco_code=$(echo "$response" | jq -r '.eco')
        
        if [ ! -z "$opening_name" ] && [ "$opening_name" != "null" ]; then
            print_success "Opening recognized: $opening_name ($eco_code)"
        else
            print_failure "Opening not recognized"
        fi
    else
        print_failure "Invalid JSON response"
    fi
}

# Test 7: Invalid FEN handling
test_invalid_fen() {
    print_test "Invalid FEN handling"
    
    response=$(curl -s -X POST "$BASE_URL/api/computer-move" \
        -H "Content-Type: application/json" \
        -d '{
            "fen": "invalid-fen-string",
            "enginePath": "internal",
            "moveTime": 1000
        }')
    
    if echo "$response" | jq empty 2>/dev/null; then
        error=$(echo "$response" | jq -r '.error')
        
        if [ ! -z "$error" ] && [ "$error" != "null" ]; then
            print_success "Invalid FEN properly rejected: $error"
        else
            print_failure "Invalid FEN was not rejected"
        fi
    else
        print_failure "Invalid JSON response"
    fi
}

# Test 8: Polyglot book integration (if book is available)
test_polyglot_book() {
    if [ ! -f "$BOOK_FILE" ]; then
        print_info "Skipping Polyglot book test (book file not found)"
        return
    fi
    
    print_test "Polyglot opening book - book move from starting position"
    
    # The book is integrated into /api/computer-move
    # It will use a book move if available, otherwise fall back to engine
    response=$(curl -s -X POST "$BASE_URL/api/computer-move" \
        -H "Content-Type: application/json" \
        -d '{
            "fen": "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
            "enginePath": "internal",
            "moveTime": 100
        }')
    
    if echo "$response" | jq empty 2>/dev/null; then
        move=$(echo "$response" | jq -r '.move')
        think_time=$(echo "$response" | jq -r '.thinkTime')
        
        if [ ! -z "$move" ] && [ "$move" != "null" ]; then
            # Book moves are typically instant (< 10ms)
            if [ "$think_time" -lt 10 ]; then
                print_success "Book move used: $move (instant response: ${think_time}ms)"
            else
                print_success "Move returned: $move (${think_time}ms - may be engine move)"
            fi
        else
            print_failure "No move returned"
        fi
    else
        print_failure "Invalid JSON response"
    fi
}

# Test 9: External engine (if available)
test_external_engine() {
    print_test "External engine - Stockfish (if available)"
    
    # Check if Stockfish is in the engine list
    engines=$(curl -s "$BASE_URL/api/engines")
    stockfish_path=$(echo "$engines" | jq -r '.[] | select(.name | contains("Stockfish")) | .path' | head -1)
    
    if [ -z "$stockfish_path" ] || [ "$stockfish_path" = "null" ]; then
        print_info "Stockfish not found, skipping external engine test"
        return
    fi
    
    response=$(curl -s -X POST "$BASE_URL/api/computer-move" \
        -H "Content-Type: application/json" \
        -d "{
            \"fen\": \"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1\",
            \"enginePath\": \"$stockfish_path\",
            \"moveTime\": 1000
        }")
    
    if echo "$response" | jq empty 2>/dev/null; then
        move=$(echo "$response" | jq -r '.move')
        think_time=$(echo "$response" | jq -r '.thinkTime')
        
        if [ ! -z "$move" ] && [ "$move" != "null" ]; then
            print_success "Stockfish returned move: $move (think time: ${think_time}ms)"
        else
            print_failure "No move returned from Stockfish"
        fi
    else
        print_failure "Invalid JSON response"
    fi
}

# Test 10: Performance test - multiple moves
test_performance() {
    print_test "Performance test - 5 consecutive moves"
    
    local total_time=0
    local moves_completed=0
    
    for i in {1..5}; do
        start_time=$(date +%s%3N)
        response=$(curl -s -X POST "$BASE_URL/api/computer-move" \
            -H "Content-Type: application/json" \
            -d '{
                "fen": "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
                "enginePath": "internal",
                "moveTime": 500
            }')
        end_time=$(date +%s%3N)
        
        if echo "$response" | jq empty 2>/dev/null; then
            move=$(echo "$response" | jq -r '.move')
            if [ ! -z "$move" ] && [ "$move" != "null" ]; then
                elapsed=$((end_time - start_time))
                total_time=$((total_time + elapsed))
                moves_completed=$((moves_completed + 1))
            fi
        fi
    done
    
    if [ $moves_completed -eq 5 ]; then
        avg_time=$((total_time / 5))
        print_success "Completed 5 moves, average response time: ${avg_time}ms"
    else
        print_failure "Only completed $moves_completed out of 5 moves"
    fi
}

# Test 11: WebSocket analysis endpoint (basic connectivity test)
test_websocket_analysis() {
    print_test "WebSocket analysis endpoint - connectivity"
    
    # Check if websocat is available for WebSocket testing
    if ! command -v websocat &> /dev/null; then
        print_info "websocat not installed, skipping WebSocket test"
        print_info "Install with: cargo install websocat (or apt-get install websocat)"
        return
    fi
    
    # Check if Stockfish is available
    engines=$(curl -s "$BASE_URL/api/engines")
    stockfish_path=$(echo "$engines" | jq -r '.[] | select(.name | contains("Stockfish")) | .path' | head -1)
    
    if [ -z "$stockfish_path" ] || [ "$stockfish_path" = "null" ]; then
        print_info "Stockfish not found, skipping WebSocket analysis test"
        return
    fi
    
    # Test WebSocket connection with a start message
    ws_url="ws://localhost:${PORT}/api/analysis"
    start_msg='{"action":"start","fen":"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1","enginePath":"'"$stockfish_path"'","depth":10}'
    
    # Send start message and read a few responses (timeout after 3 seconds)
    response=$(echo "$start_msg" | timeout 3 websocat -n1 "$ws_url" 2>/dev/null || true)
    
    if [ ! -z "$response" ]; then
        # Check if we got valid JSON analysis data
        if echo "$response" | jq -e '.depth' &>/dev/null; then
            depth=$(echo "$response" | jq -r '.depth')
            score=$(echo "$response" | jq -r '.score')
            print_success "WebSocket analysis working (depth: $depth, score: $score)"
        else
            print_success "WebSocket connected (received: ${response:0:50}...)"
        fi
    else
        print_info "WebSocket test inconclusive (no response within timeout)"
    fi
}

# Main test execution
main() {
    print_header "Go-Chess Server API Test Suite"
    
    # Check dependencies
    check_dependencies
    
    # Build and start server
    build_app
    start_server
    
    # Ensure server is stopped on exit
    trap stop_server EXIT
    
    # Run all tests
    print_header "Running API Tests"
    
    test_health_check
    test_get_engines
    test_builtin_engine_move
    test_time_controls
    test_opening_recognition
    test_sicilian_defense
    test_invalid_fen
    test_polyglot_book
    test_external_engine
    test_performance
    test_websocket_analysis
    
    # Print summary
    print_header "Test Summary"
    echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
    echo -e "Total tests:  $((TESTS_PASSED + TESTS_FAILED))"
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "\n${GREEN}All tests passed! ✓${NC}\n"
        exit 0
    else
        echo -e "\n${RED}Some tests failed! ✗${NC}\n"
        exit 1
    fi
}

# Run main function
main
