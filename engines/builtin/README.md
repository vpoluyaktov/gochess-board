# Built-in Chess Engine

This package contains the GoChess built-in chess engine implementation.

## File Structure

The engine code is organized into separate files by functionality:

### Core Files

- **engine.go** - Main engine interface and entry point
  - `InternalEngine` struct
  - `NewEngine()` - Constructor
  - `GetBestMove()` - Main search interface
  - `GetBestMoveWithClock()` - Time management
  - `Analyze()` - Analysis mode with iterative deepening

- **uci.go** - UCI protocol implementation ⭐ NEW
  - `RunUCI()` - UCI protocol mode
  - Position parsing
  - Time management
  - Command handling
  - `SetOption()` and `Close()` interface methods

### Search & Algorithm Files

- **search.go** - Core search algorithms
  - `search()` - Minimax search with alpha-beta pruning
  - `quiescence()` - Quiescence search for tactical positions

- **analysis.go** - Analysis mode functionality
  - `AnalysisInfo` struct
  - `Analyze()` - Iterative deepening analysis with live updates
  - `searchWithStats()` - Search with node counting
  - `alphaBetaWithStop()` - Alpha-beta with stop channel support
  - `quiescenceWithStop()` - Quiescence with stop channel support

### Evaluation Files

- **evaluation.go** - Position evaluation
  - Piece values (constants)
  - Piece-square tables for positional evaluation
  - `evaluate()` - Main evaluation function
  - `evaluateKingSafety()` - King safety evaluation ⭐ NEW
  - `evaluatePawnStructure()` - Pawn structure analysis ⭐ NEW
  - `evaluateMobility()` - Mobility evaluation ⭐ NEW
  - `getPieceValue()` - Material value lookup
  - `getPieceSquareValue()` - Positional value lookup

- **move_ordering.go** - Move ordering for search optimization
  - `orderMoves()` - Sort moves for better alpha-beta pruning
  - `moveOrderScore()` - Assign scores to moves for ordering
  - MVV-LVA score tables ⭐ NEW

- **transposition.go** - Transposition table (hash table) ⭐ NEW
  - `TranspositionTable` - Hash table for position caching
  - `probe()` - Look up position in table
  - `store()` - Save position evaluation
  - `getZobristKey()` - Position hashing

- **killer_moves.go** - Killer move heuristic ⭐ NEW
  - `KillerMoves` - Killer move storage
  - `add()` - Add killer move
  - `isKiller()` - Check if move is killer
  - `getKillerScore()` - Get killer move bonus

## Engine Strength

**Current estimated ELO: ~1400-1600** (improved from 1000-1200)

The engine uses:
- Minimax search with alpha-beta pruning
- **Transposition table (64MB hash table)** ⭐ NEW
- **Killer move heuristic** ⭐ NEW
- **Enhanced move ordering (TT move, killer moves, MVV-LVA foundation)** ⭐ NEW
- Quiescence search for tactical stability
- **Enhanced evaluation with:**
  - Material and piece-square tables
  - **King safety (pawn shield)** ⭐ NEW
  - **Pawn structure (doubled pawns, passed pawns)** ⭐ NEW
  - **Mobility evaluation** ⭐ NEW
- Iterative deepening for time management

See [IMPROVEMENTS.md](IMPROVEMENTS.md) for detailed information about recent enhancements.

## Future Improvements

Areas for further enhancement to increase engine strength:

1. **Search Enhancements**
   - ✅ ~~Transposition tables~~ (DONE)
   - ✅ ~~Killer move heuristic~~ (DONE)
   - History heuristic
   - Null move pruning
   - Late move reductions (LMR)
   - Principal Variation Search (PVS)
   - Aspiration windows

2. **Evaluation Improvements**
   - ✅ ~~King safety evaluation~~ (DONE)
   - ✅ ~~Pawn structure analysis~~ (DONE)
   - ✅ ~~Mobility evaluation~~ (DONE)
   - Endgame-specific evaluation
   - Tapered evaluation (opening/middlegame/endgame)
   - King attack patterns
   - Piece coordination

3. **Move Ordering**
   - ✅ ~~MVV-LVA foundation~~ (DONE)
   - ✅ ~~Killer moves~~ (DONE)
   - ✅ ~~TT move priority~~ (DONE)
   - Complete MVV-LVA with piece detection
   - History heuristic
   - Counter moves

4. **Opening Book**
   - Integration with Polyglot opening book
   - Custom opening repertoire

5. **Endgame Tablebases**
   - Syzygy tablebase support
   - Gaviota tablebase support

## Testing

Run tests for the builtin engine:

```bash
go test ./engines/builtin -v
```

Run analysis tests:

```bash
go test -run TestInternalEngineAnalyze ./engines/builtin -v
```

## Usage

```go
import "gochess-board/engines/builtin"

// Create engine
engine := builtin.NewEngine()

// Get best move
move, err := engine.GetBestMove(fen, 5*time.Second)

// Or use with clock
move, err := engine.GetBestMoveWithClock(fen, moveHistory, 
    whiteTime, blackTime, whiteInc, blackInc)

// Analysis mode
stopCh := make(chan bool)
resultCh := make(chan builtin.AnalysisInfo, 10)

go engine.Analyze(fen, 10, stopCh, resultCh)

for info := range resultCh {
    fmt.Printf("Depth %d: %s (score: %d)\n", 
        info.Depth, info.BestMove, info.Score)
}
```
