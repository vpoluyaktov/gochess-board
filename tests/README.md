# Go-Chess Tests

This directory contains all tests for the go-chess application.

## Directory Structure

```
tests/
├── quality/          # Code quality tests
│   ├── test_code.sh
│   └── README.md
├── api/              # Backend API tests
│   ├── test_api.sh
│   └── README.md
└── ui/               # Frontend UI tests
    ├── test_ui.sh
    └── playwright/   # Playwright test files
        ├── package.json
        ├── playwright.config.js
        ├── README.md
        └── specs/    # Test specifications
            ├── 01-page-load.spec.js
            ├── 02-engine-selection.spec.js
            ├── 03-making-moves.spec.js
            ├── 04-computer-moves.spec.js
            ├── 05-opening-names.spec.js
            ├── 06-analysis.spec.js
            └── 07-responsive.spec.js
```

## Code Quality Tests

Go code quality checks including formatting, static analysis, and unit tests.

**Location:** `tests/quality/`

**Run:**
```bash
cd tests/quality
./test_code.sh
```

**Features:**
- Code formatting (go fmt)
- Static analysis (go vet)
- Race condition detection
- Module verification
- Unit tests

See [tests/quality/README.md](quality/README.md) for details.

## API Tests

Backend and API testing using bash, curl, and Go tools.

**Location:** `tests/api/`

**Run:**
```bash
cd tests/api
./test_api.sh
```

**Features:**
- HTTP endpoint testing
- WebSocket testing
- Performance benchmarks
- External engine integration

See [tests/api/README.md](api/README.md) for details.

## UI Tests

Frontend end-to-end testing using Playwright.

**Location:** `tests/ui/`

**Setup (first time):**
```bash
cd tests/ui
./test_ui.sh --setup
```

**Run:**
```bash
cd tests/ui
./test_ui.sh
```

**Features:**
- Page load and rendering
- User interactions (clicks, moves)
- Engine selection
- Computer moves
- Analysis features
- Responsive design

See [tests/ui/playwright/README.md](ui/playwright/README.md) for details.

## Running All Tests

From the project root:

```bash
# Code quality tests
./tests/quality/test_code.sh

# API tests
./tests/api/test_api.sh

# UI tests
./tests/ui/test_ui.sh
```

## CI/CD Integration

Both test suites support CI/CD environments:

```bash
# Code quality tests
cd tests/quality && ./test_code.sh

# API tests (always headless)
cd tests/api && ./test_api.sh

# UI tests (headless mode)
cd tests/ui && CI=true ./test_ui.sh
```

## Requirements

**API Tests:**
- Go 1.21+
- curl
- jq
- websocat (optional, for WebSocket tests)

**UI Tests:**
- Node.js 14+
- npm
- Playwright (installed via setup)

## Test Coverage

**Code Quality Tests:** 5 tests
- Code formatting
- Static analysis
- Race detection
- Module verification
- Unit tests

**API Tests:** 14 tests
- HTTP endpoint tests
- WebSocket tests
- Performance tests

**UI Tests:** 21 tests
- 7 test suites covering major UI features
- Page load, engine selection, moves, analysis, responsive design

**Total:** 40 comprehensive tests
