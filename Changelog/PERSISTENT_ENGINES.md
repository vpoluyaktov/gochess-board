# Persistent Engine Pool Feature

## Overview

Added a new `--persistent-engines` flag that enables engine reuse between moves within the same game. This provides performance benefits by avoiding engine startup overhead on each move while maintaining warm hash tables.

## Usage

```bash
# Enable persistent engines
./gochess-board --persistent-engines

# Combine with other flags
./gochess-board --persistent-engines --book-file /path/to/book.bin
```

## How It Works

### Without `--persistent-engines` (Default)
- Each move request creates a new engine process
- Engine is closed after returning the move
- ~100-500ms overhead per move for engine initialization
- Hash tables start empty on each move

### With `--persistent-engines`
- Engines are cached in a pool, keyed by `gameID + enginePath`
- Same engine is reused for subsequent moves in the same game
- Engines are automatically closed after 10 minutes of inactivity
- Hash tables remain warm between moves (better analysis)

## Architecture

### Engine Pool (`engines/pool.go`)

```go
type EnginePool struct {
    engines        map[string]*PooledEngine  // Cached engines
    idleTimeout    time.Duration             // Default: 10 minutes
    cleanupTicker  *time.Ticker              // Runs every 1 minute
    engineFactory  EngineFactory             // Creates new engines
}

type PooledEngine struct {
    Engine     ChessEngine
    GameID     string
    EnginePath string
    EngineType string
    EngineName string
    LastUsed   time.Time
    Options    map[string]string
}
```

### Key Features

1. **Automatic Cleanup**: Background goroutine removes idle engines every minute
2. **Options Tracking**: If engine options change (e.g., ELO setting), old engine is closed and new one created
3. **Thread-Safe**: Mutex-protected for concurrent access
4. **Game Isolation**: Different games get different engine instances
5. **Graceful Shutdown**: All engines closed when server stops

## API Changes

### MoveRequest

Added `gameId` field to identify which game the move belongs to:

```json
{
  "fen": "...",
  "enginePath": "/usr/bin/stockfish",
  "gameId": "game-1701234567890-abc123def",
  "engineOptions": {...}
}
```

The Web UI automatically generates and sends a unique `gameId` for each game session. A new ID is generated when:
- The page loads (new session)
- User clicks "New Game"

**Note**: If `--persistent-engines` is not set, the `gameId` is ignored and per-request behavior is used.

## Benefits

### Performance
- **No startup overhead**: Engine already running for subsequent moves
- **Warm hash tables**: Engine remembers previously analyzed positions
- **Faster response**: Especially noticeable in rapid/bullet games

### Resource Efficiency
- **Fewer process spawns**: Reduces OS overhead
- **Less memory churn**: No repeated allocation/deallocation of hash tables
- **Cleaner process table**: Fewer zombie process risks

## Trade-offs

### Pros
- Faster move calculation after first move
- Better analysis quality (warm caches)
- Reduced system load

### Cons
- Higher memory usage (engines stay resident)
- Requires client to send `gameId`
- Slightly more complex server state

## Configuration

| Parameter | Value | Description |
|-----------|-------|-------------|
| Idle Timeout | 10 minutes | Time before unused engine is closed |
| Cleanup Interval | 1 minute | How often pool checks for idle engines |

## Files Changed

- `main.go` - Added `--persistent-engines` flag and pool initialization
- `engines/pool.go` - New engine pool implementation
- `engines/pool_test.go` - Comprehensive test suite
- `server/move_handler.go` - Updated to use pool when available, added `gameId` field
- `server/assets/js/gochess-state.js` - Added `gameId` to game state
- `server/assets/js/gochess-navigation.js` - Generate new `gameId` on new game
- `server/assets/js/gochess-engine.js` - Send `gameId` in move requests

## Testing

```bash
# Run pool tests
go test ./engines/... -v -run TestEnginePool

# Test with flag
./gochess-board --persistent-engines --no-browser --log-level DEBUG
```

## Future Enhancements

1. **Configurable timeout**: Add `--engine-idle-timeout` flag
2. **Pool size limit**: Maximum number of cached engines
3. **Engine prewarming**: Start engines before first move
4. **Statistics endpoint**: API to view pool status
