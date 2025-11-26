# Go-Chess Scripts

This directory contains utility scripts for testing and managing the go-chess application.

## test_api.sh (Backend & API Tests)

Comprehensive test script for the go-chess server API.

### Features

**API Endpoint Tests:**
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
./test_api.sh
```

**With custom port:**
```bash
PORT=8080 ./test_api.sh
```

**With opening book:**
```bash
BOOK_FILE=/path/to/book.bin ./test_api.sh
```

**With debug logging:**
```bash
LOG_LEVEL=DEBUG ./test_api.sh
```

**All options:**
```bash
PORT=8080 \
LOG_LEVEL=DEBUG \
BOOK_FILE=/usr/share/games/gnuchess/book.bin \
./test_api.sh
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

**Code Quality (5 tests):**
1. **Code Formatting** - Ensures all Go files are properly formatted
2. **Static Analysis** - Runs `go vet` to catch common mistakes
3. **Race Detection** - Tests for data race conditions
4. **Module Verification** - Verifies go.mod integrity
5. **Unit Tests** - Runs all package unit tests

**API Endpoints (14 tests):**
6. **Health Check** - Verifies server is responding (GET /)
7. **Engine Discovery** - Lists all available engines (GET /api/engines)
8. **Built-in Engine** - Tests internal chess engine (POST /api/computer-move)
9. **Time Controls** - Tests clock-based time management (POST /api/computer-move)
10. **Opening Recognition** - Tests Italian Game recognition (POST /api/opening)
11. **Sicilian Defense** - Tests another opening (POST /api/opening)
12. **Error Handling** - Tests invalid FEN rejection (POST /api/computer-move)
13. **Polyglot Book** - Tests opening book integration (POST /api/computer-move)
14. **External Engine** - Tests Stockfish (if available) (POST /api/computer-move)
15. **Performance** - Tests response time for 5 moves (POST /api/computer-move)
16. **WebSocket Analysis** - Tests real-time analysis endpoint (WS /api/analysis)
17. **Server Start** - Verifies server starts successfully
18. **Server Stop** - Verifies clean shutdown

**Total: 19 comprehensive tests**

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
Code Quality Checks
========================================

TEST: Code formatting (go fmt)
✓ PASS: All files are properly formatted
TEST: Static analysis (go vet)
✓ PASS: No issues found
TEST: Race condition detection (go test -race)
✓ PASS: No race conditions detected
TEST: Module verification (go mod verify)
✓ PASS: All modules verified
TEST: Unit tests (go test ./...)
✓ PASS: All unit tests passed (8 packages)

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
  - Toga II 3.0 (uci)
  - Stockfish 16 (uci)
  - Crafty-23.4 (cecp)
  - GNU (cecp)

TEST: Computer move with built-in engine
✓ PASS: Built-in engine returned move: d2d4 (think time: 0ms)

TEST: Computer move with clock-based time management
✓ PASS: Clock-based move: b8c6 (think time: 1003ms)

TEST: Opening name recognition - Italian Game
✓ PASS: Opening recognized: Italian Game (C50)

TEST: Opening name recognition - Sicilian Defense
✓ PASS: Opening recognized: Sicilian Defense (B20)

TEST: Invalid FEN handling
✓ PASS: Invalid FEN properly rejected: Invalid FEN

TEST: Polyglot opening book - book move from starting position
✓ PASS: Book move used: d2d4 (instant response: 0ms)

TEST: External engine - Stockfish (if available)
✓ PASS: Stockfish returned move: d2d4 (think time: 0ms)

TEST: Performance test - 5 consecutive moves
✓ PASS: Completed 5 moves, average response time: 5ms

TEST: WebSocket analysis endpoint - connectivity
✓ PASS: WebSocket analysis working (depth: 12, score: 40)

========================================
Stopping Server
========================================

INFO: Stopping server (PID: 12345)
✓ PASS: Server stopped

========================================
Test Summary
========================================
Tests passed: 19
Tests failed: 0
Total tests:  19

All tests passed! ✓
```

### Notes

- The script automatically builds the application
- The server is started with `--restart` to kill any existing instance
- The server runs in the background and is automatically stopped when the script exits
- Server logs are saved to `server.log` in the current directory
- Some tests are skipped if dependencies are not available (e.g., Polyglot book, Stockfish)

---

## test_ui.sh (Frontend & UI Tests)

End-to-end UI tests using Playwright to test the web interface.

### Features

**UI Test Coverage:**
- ✅ Page loading and rendering
- ✅ Chessboard display
- ✅ Engine selection dropdown
- ✅ Making moves (click interactions)
- ✅ Computer move requests
- ✅ Opening name recognition
- ✅ Position analysis
- ✅ Responsive design (mobile/tablet/desktop)

### Setup

**First time setup:**
```bash
./scripts/test_ui.sh --setup
```

This will:
1. Create `tests/ui/` directory structure
2. Install Playwright and dependencies
3. Download Chromium browser
4. Create test spec files
5. Generate configuration

### Usage

**Run all UI tests:**
```bash
./scripts/test_ui.sh
```

**Run tests with visible browser:**
```bash
cd tests/ui
npm run test:headed
```

**Run in debug mode:**
```bash
cd tests/ui
npm run test:debug
```

**Run with interactive UI:**
```bash
cd tests/ui
npm run test:ui
```

**View test report:**
```bash
cd tests/ui
npm run report
```

### Requirements

**Required:**
- Node.js (v14 or higher)
  - Ubuntu/Debian: `sudo apt-get install nodejs npm`
  - macOS: `brew install node`

**Installed by setup:**
- Playwright (`@playwright/test`)
- Chromium browser

### Test Files Created

After running `--setup`, the following structure is created:

```
tests/ui/
├── package.json              # Node.js dependencies
├── playwright.config.js      # Playwright configuration
├── README.md                 # UI test documentation
└── specs/
    ├── 01-page-load.spec.js       # Page loading tests
    ├── 02-engine-selection.spec.js # Engine dropdown tests
    ├── 03-making-moves.spec.js     # Move interaction tests
    ├── 04-computer-moves.spec.js   # Computer move tests
    ├── 05-opening-names.spec.js    # Opening recognition tests
    ├── 06-analysis.spec.js         # Analysis feature tests
    └── 07-responsive.spec.js       # Responsive design tests
```

### Customizing Tests

The test specs use common CSS selectors. Update them to match your actual HTML:

```javascript
// Example: Update board selector
const board = page.locator('.board-b72b1, #board, .chessboard').first();

// Example: Update square selector
const e2Square = page.locator('[data-square="e2"], .square-e2').first();
```

### CI/CD Integration

For continuous integration:

```bash
# Set CI mode
CI=true ./scripts/test_ui.sh

# Or in tests/ui directory
cd tests/ui
CI=true npm test
```

### Example Output

```
========================================
Go-Chess UI Test Suite
========================================

========================================
Starting Server
========================================

INFO: Server PID: 12345
✓ PASS: Server started successfully

========================================
Running Playwright Tests
========================================

Running 15 tests using 1 worker

  ✓ specs/01-page-load.spec.js:3:3 › Page Load › should load the chess board page (1.2s)
  ✓ specs/01-page-load.spec.js:10:3 › Page Load › should display the chessboard (0.8s)
  ✓ specs/01-page-load.spec.js:17:3 › Page Load › should load chess pieces (0.9s)
  ✓ specs/02-engine-selection.spec.js:3:3 › Engine Selection › should display engine selection dropdown (0.7s)
  ✓ specs/02-engine-selection.spec.js:10:3 › Engine Selection › should list available engines (0.8s)
  ✓ specs/02-engine-selection.spec.js:20:3 › Engine Selection › should include built-in engine in list (0.7s)
  ...

  15 passed (12.3s)

✓ PASS: All Playwright tests passed

========================================
Test Summary
========================================

All UI tests passed! ✓
```

### Notes

- Tests run in headless mode by default (no browser window)
- Screenshots are taken on test failures
- Test reports are generated in `tests/ui/playwright-report/`
- The server must be running for tests to work (script starts it automatically)
- Some tests may be skipped if UI elements don't match expected selectors
