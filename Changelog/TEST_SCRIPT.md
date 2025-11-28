# Server Test Script

## Summary

Created a comprehensive test script for the gochess-board server API that tests all major features and endpoints.

## Files Created

### `scripts/test_server.sh`
Comprehensive bash script that:
- Builds the application
- Starts the server automatically
- Runs 10 different API tests
- Provides colored output
- Shows test results summary
- Cleans up automatically

### `scripts/README.md`
Documentation for the scripts directory including:
- Usage instructions
- Configuration options
- Requirements
- Example output

## Features

### Tests Included

1. **Health Check** - Verifies server responds to GET /
2. **Engine Discovery** - Tests GET /api/engines
3. **Built-in Engine** - Tests computer move with internal engine
4. **Time Controls** - Tests clock-based time management
5. **Opening Recognition** - Tests Italian Game recognition
6. **Sicilian Defense** - Tests another opening pattern
7. **Error Handling** - Tests invalid FEN rejection
8. **Polyglot Book** - Tests opening book moves (if available)
9. **External Engine** - Tests Stockfish (if available)
10. **Performance** - Tests response time for 5 consecutive moves

### Configuration Options

The script supports environment variables:
- `PORT` - Server port (default: 35256)
- `LOG_LEVEL` - Logging level (default: INFO)
- `BOOK_FILE` - Path to Polyglot book file

### Output Features

- ✅ Colored output (blue/yellow/green/red)
- ✅ Clear test names and results
- ✅ Detailed error messages
- ✅ Test summary with pass/fail counts
- ✅ Automatic cleanup on exit

## Usage Examples

### Basic Usage
```bash
./scripts/test_server.sh
```

### With Custom Configuration
```bash
PORT=8080 LOG_LEVEL=DEBUG ./scripts/test_server.sh
```

### With Opening Book
```bash
BOOK_FILE=/usr/share/games/gnuchess/book.bin ./scripts/test_server.sh
```

## Example Output

```
========================================
Go-Chess Server API Test Suite
========================================

========================================
Building Application
========================================

INFO: Running: go build
✓ PASS: Application built successfully

========================================
Starting Server
========================================

INFO: Starting server on port 35256 with log level INFO
INFO: Server PID: 12345
✓ PASS: Server started successfully

========================================
Running API Tests
========================================

TEST: Health check - GET /
✓ PASS: Server is responding (HTTP 200)

TEST: Get engines list - GET /api/engines
✓ PASS: Retrieved 6 engines
✓ PASS: Built-in engine found: GoChess (Built-in)
  - GoChess (Built-in) (internal)
  - Fruit 2.1 (uci)
  - Toga II 3.0 (uci)
  - Stockfish 16 (uci)
  - Crafty-23.4 (cecp)
  - GNU (cecp)

TEST: Computer move with built-in engine
✓ PASS: Built-in engine returned move: b1c3 (think time: 48ms)

TEST: Computer move with clock-based time management
✓ PASS: Clock-based move: c7c6 (think time: 645ms)

TEST: Opening name recognition - Italian Game
✓ PASS: Opening recognized: Italian Game (C50)

TEST: Opening name recognition - Sicilian Defense
✓ PASS: Opening recognized: Sicilian Defense (B20)

TEST: Invalid FEN handling
✓ PASS: Invalid FEN properly rejected: invalid FEN

TEST: Polyglot opening book - book moves for e4
✓ PASS: Found 5 book moves
  - c7c5 (weight: 1234)
  - e7e5 (weight: 2345)
  - c7c6 (weight: 567)
  - e7e6 (weight: 890)
  - d7d5 (weight: 234)

TEST: External engine - Stockfish (if available)
✓ PASS: Stockfish returned move: e2e4 (think time: 123ms)

TEST: Performance test - 5 consecutive moves
✓ PASS: Completed 5 moves, average response time: 52ms

========================================
Stopping Server
========================================

INFO: Stopping server (PID: 12345)
✓ PASS: Server stopped

========================================
Test Summary
========================================
Tests passed: 13
Tests failed: 0
Total tests:  13

All tests passed! ✓
```

## Benefits

1. **Automated Testing** - No manual curl commands needed
2. **Comprehensive Coverage** - Tests all major API endpoints
3. **Easy to Run** - Single command execution
4. **Clear Output** - Color-coded results
5. **Automatic Cleanup** - Server stopped on exit
6. **Configurable** - Environment variables for customization
7. **CI/CD Ready** - Exit codes for automation
8. **Documentation** - Detailed README included

## Requirements

- `jq` - JSON processor (for parsing API responses)
- `curl` - HTTP client (usually pre-installed)
- `bash` - Shell (standard on Linux/macOS)

## Future Enhancements

Potential additions:
- WebSocket analysis endpoint testing
- Concurrent request testing
- Memory/CPU profiling
- Response time benchmarks
- Engine option testing
- Move validation testing
- Game state persistence testing
