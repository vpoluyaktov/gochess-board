# Go-Chess Code Quality Tests

This directory contains code quality and static analysis tests for the go-chess project.

## test_code.sh

Comprehensive Go code quality checks including formatting, static analysis, race detection, and unit tests.

### Features

**Code Quality Checks (5 tests):**
1. ✅ **Code Formatting** (`go fmt`) - Ensures all Go files are properly formatted
2. ✅ **Static Analysis** (`go vet`) - Catches common mistakes and suspicious code
3. ✅ **Race Detection** (`go test -race`) - Detects data race conditions
4. ✅ **Module Verification** (`go mod verify`) - Verifies go.mod integrity
5. ✅ **Unit Tests** (`go test ./...`) - Runs all package unit tests

### Usage

**Run all code quality checks:**
```bash
cd tests/quality
./test_code.sh
```

**From project root:**
```bash
./tests/quality/test_code.sh
```

### Requirements

- Go 1.21+
- All dependencies in `go.mod`

### Exit Codes

- `0` - All tests passed
- `1` - One or more tests failed

### Example Output

```
========================================
Go-Chess Code Quality Test Suite
========================================

========================================
Go Code Quality Checks
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
Test Summary
========================================
Tests passed: 5
Tests failed: 0
Total tests:  5

All tests passed! ✓
```

### Integration

This script is automatically called by:
- `tests/api/test_api.sh` - API test suite runs quality checks first
- Can be run standalone for quick quality verification

### CI/CD Integration

For continuous integration:

```bash
# Run quality checks
./tests/quality/test_code.sh

# Or as part of full test suite
./tests/api/test_api.sh
```

### Fixing Issues

**Formatting issues:**
```bash
gofmt -w .
```

**Vet issues:**
Review the reported issues and fix the code accordingly.

**Race conditions:**
Review the race detector output and fix synchronization issues.

**Module issues:**
```bash
go mod tidy
go mod verify
```

**Unit test failures:**
```bash
# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test -v ./package_name
```

### Notes

- All checks must pass for the script to exit successfully
- The script changes to the project root directory automatically
- Temporary output files are stored in `/tmp/`
- The script uses `set -e` to exit on first failure
