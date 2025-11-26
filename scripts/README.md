# Go-Chess Scripts

This directory contains utility scripts for testing and managing the go-chess application.

## test_server.sh

Comprehensive test script for the go-chess server API.

### Features

Tests all major server endpoints:
- ✅ Health check
- ✅ Engine discovery
- ✅ Built-in engine moves
- ✅ Clock-based time management
- ✅ Opening name recognition
- ✅ Invalid input handling
- ✅ Polyglot opening book (if available)
- ✅ External engines (if available)
- ✅ Performance testing

### Usage

**Basic usage:**
```bash
./scripts/test_server.sh
```

**With custom port:**
```bash
PORT=8080 ./scripts/test_server.sh
```

**With opening book:**
```bash
BOOK_FILE=/path/to/book.bin ./scripts/test_server.sh
```

**With debug logging:**
```bash
LOG_LEVEL=DEBUG ./scripts/test_server.sh
```

**All options:**
```bash
PORT=8080 \
LOG_LEVEL=DEBUG \
BOOK_FILE=/usr/share/games/gnuchess/book.bin \
./scripts/test_server.sh
```

### Requirements

**Required:**
- `jq` - JSON processor
  - Ubuntu/Debian: `sudo apt-get install jq`
  - macOS: `brew install jq`
- `curl` - HTTP client (usually pre-installed)

**Optional (for WebSocket test):**
- `websocat` - WebSocket client
  - Ubuntu/Debian: `sudo apt-get install websocat`
  - Cargo: `cargo install websocat`
  - macOS: `brew install websocat`

### What It Tests

1. **Health Check** - Verifies server is responding (GET /)
2. **Engine Discovery** - Lists all available engines (GET /api/engines)
3. **Built-in Engine** - Tests internal chess engine (POST /api/computer-move)
4. **Time Controls** - Tests clock-based time management (POST /api/computer-move)
5. **Opening Recognition** - Tests Italian Game recognition (POST /api/opening)
6. **Sicilian Defense** - Tests another opening (POST /api/opening)
7. **Error Handling** - Tests invalid FEN rejection (POST /api/computer-move)
8. **Polyglot Book** - Tests opening book integration (POST /api/computer-move)
9. **External Engine** - Tests Stockfish (if available) (POST /api/computer-move)
10. **Performance** - Tests response time for 5 moves (POST /api/computer-move)
11. **WebSocket Analysis** - Tests real-time analysis endpoint (WS /api/analysis)

### Output

The script provides colored output:
- 🔵 Blue: Headers and info
- 🟡 Yellow: Test names
- 🟢 Green: Passed tests
- 🔴 Red: Failed tests

### Exit Codes

- `0` - All tests passed
- `1` - One or more tests failed

### Example Output

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

INFO: Using opening book: /usr/share/games/gnuchess/book.bin
INFO: Starting server on port 35256 with log level INFO
INFO: Server PID: 12345
INFO: Waiting for server to start...
✓ PASS: Server started successfully

========================================
Running API Tests
========================================

TEST: Health check - GET /
✓ PASS: Server is responding (HTTP 200)

TEST: Get engines list - GET /api/engines
✓ PASS: Retrieved 6 engines
✓ PASS: Built-in engine found: GoChess Basic (Built-in)
  - GoChess Basic (Built-in) (internal)
  - Fruit 2.1 (uci)
  - Stockfish 16 (uci)
  - GNU Chess (cecp)

TEST: Computer move with built-in engine
✓ PASS: Built-in engine returned move: b1c3 (think time: 48ms)

...

========================================
Test Summary
========================================
Tests passed: 10
Tests failed: 0
Total tests:  10

All tests passed! ✓
```

### Notes

- The script automatically builds the application
- The server is started with `--restart` to kill any existing instance
- The server runs in the background and is automatically stopped when the script exits
- Server logs are saved to `server.log` in the current directory
- Some tests are skipped if dependencies are not available (e.g., Polyglot book, Stockfish)
