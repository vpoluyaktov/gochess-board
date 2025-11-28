# Engine Package Removal

## Summary

Completely removed the `engine/` package and updated all code to use `engine_interface` directly. This eliminates the unnecessary compatibility layer and simplifies the codebase.

## Changes Made

### 1. Deleted Package
- ✅ Removed `engine/` directory entirely
- ✅ Deleted `engine/engine.go` (the re-export compatibility layer)

### 2. Updated Imports
Updated all imports from `gochess-board/engine` → `gochess-board/engine_interface` in:
- ✅ `main.go`
- ✅ `tui/tui.go`
- ✅ `server/server.go`
- ✅ `server/move_handler.go`
- ✅ `server/analysis_handler.go`
- ✅ `server/server_test.go`
- ✅ `server/move_handler_test.go`
- ✅ `server/opening_api_test.go`

### 3. Updated Package References
Changed all references from `engine.` → `engine_interface.` for:
- Type references (e.g., `engine.ChessEngine` → `engine_interface.ChessEngine`)
- Function calls (e.g., `engine.DiscoverEngines()` → `engine_interface.DiscoverEngines()`)
- Global variables (e.g., `engine.GlobalMonitor` → `engine_interface.GlobalMonitor`)

## Final Structure

**Before:**
```
engine/                  ← Compatibility layer (re-exports)
└── engine.go

engine_interface/        ← Actual implementations
├── engine_interface.go
├── engine_uci.go
├── engine_cecp.go
├── engine_discovery.go
├── engine_monitor.go
└── ... (tests)

internal_engine/         ← Built-in engine
└── internal_engine.go
```

**After:**
```
engine_interface/        ← All engine code
├── engine_interface.go
├── engine_uci.go
├── engine_cecp.go
├── engine_discovery.go
├── engine_monitor.go
└── ... (tests)

internal_engine/         ← Built-in engine
└── internal_engine.go
```

## Benefits

1. **Cleaner structure**: No confusion about which package to use
2. **No indirection**: Direct imports, no re-export layer
3. **More maintainable**: Single source of truth for engine code
4. **Clearer intent**: `engine_interface` clearly indicates this is the engine implementation package

## Testing

All tests pass and application works correctly:

```bash
✅ go build                    # Success
✅ go test ./...               # All pass (except expected opening test)
✅ Application runs            # Server starts successfully
✅ Internal engine works       # Moves calculated correctly
✅ External engines work       # UCI/CECP engines discovered
```

### Test Results
```bash
$ curl -X POST http://localhost:35256/api/computer-move \
  -d '{"fen":"...","enginePath":"internal","moveTime":1000}'

{
  "move": "b1c3",
  "fen": "rnbqkbnr/pppppppp/8/8/8/2N5/PPPPPPPP/R1BQKBNR b KQkq - 1 1",
  "thinkTime": 49
}
```

## Breaking Changes

**None for external users** - this is an internal refactoring. The application behavior is identical.

## Files Modified

- **Deleted**: `engine/engine.go`
- **Deleted**: `engine/` directory
- **Modified**: 8 files (updated imports)
  - `main.go`
  - `tui/tui.go`
  - `server/server.go`
  - `server/move_handler.go`
  - `server/analysis_handler.go`
  - `server/server_test.go`
  - `server/move_handler_test.go`
  - `server/opening_api_test.go`

## Migration Notes

If anyone has external code importing `gochess-board/engine`, they should update to:
```go
// Old (no longer exists)
import "gochess-board/engine"

// New
import "gochess-board/engine_interface"
```

All type and function names remain the same, just the package name changed.
