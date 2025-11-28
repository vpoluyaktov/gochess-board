# Built-in Chess Engine

This package contains the GoChess built-in chess engine implementation.

## File Structure

The engine code is organized into separate files by functionality:

### Core Files

- **engine.go** - Main engine struct and ChessEngine interface implementation
  - `InternalEngine` struct definition
  - `NewEngine()` constructor
  - `GetBestMove()` - Main move generation for games
  - `GetBestMoveWithClock()` - Time-controlled move generation
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
  - `getPieceValue()` - Material value lookup
  - `getPieceSquareValue()` - Positional value lookup

- **move_ordering.go** - Move ordering for search optimization
  - `orderMoves()` - Sort moves for better alpha-beta pruning
  - `moveOrderScore()` - Assign scores to moves for ordering

## Engine Strength

Current estimated ELO: ~1000-1200

The engine uses:
- Minimax search with alpha-beta pruning
- Quiescence search for tactical stability
- Piece-square tables for positional evaluation
- Basic move ordering (captures and checks first)
- Iterative deepening for time management

## Future Improvements

Areas for enhancement to increase engine strength:

1. **Search Enhancements**
   - Transposition tables
   - Killer move heuristic
   - History heuristic
   - Null move pruning
   - Late move reductions

2. **Evaluation Improvements**
   - King safety evaluation
   - Pawn structure analysis
   - Mobility evaluation
   - Endgame-specific evaluation
   - Tapered evaluation (opening/middlegame/endgame)

3. **Move Ordering**
   - MVV-LVA (Most Valuable Victim - Least Valuable Attacker)
   - Killer moves
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
