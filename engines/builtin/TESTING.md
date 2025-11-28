# Built-in Engine Test Suite

Comprehensive test coverage for the GoChess built-in chess engine.

## Test Organization

Tests are organized to match the source file structure:

| Test File | Source File | Tests | Coverage |
|-----------|-------------|-------|----------|
| `engine_test.go` | `engine.go` | 5 | Core engine interface |
| `evaluation_test.go` | `evaluation.go` | 8 | Position evaluation |
| `search_test.go` | `search.go` | 9 | Search algorithms |
| `move_ordering_test.go` | `move_ordering.go` | 6 | Move ordering |
| `analysis_test.go` | `analysis.go` | 3 | Analysis mode |

**Total: 31 tests** across 5 test files

## Running Tests

### Run all tests
```bash
go test ./engines/builtin -v
```

### Run specific test file
```bash
go test ./engines/builtin -v -run TestEvaluate
go test ./engines/builtin -v -run TestSearch
go test ./engines/builtin -v -run TestMoveOrder
```

### Run with coverage
```bash
go test ./engines/builtin -cover
go test ./engines/builtin -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Run benchmarks
```bash
go test ./engines/builtin -bench=.
go test ./engines/builtin -bench=BenchmarkSearch -benchmem
```

## Test Coverage Details

### Engine Tests (`engine_test.go`)
- ✅ `TestInternalEngineBasic` - Basic move generation
- ✅ `TestInternalEngineWithClock` - Time-controlled moves
- ✅ `TestInternalEngineCheckmate` - Mate position handling
- ✅ `TestInternalEngineSetOption` - Option setting
- ✅ `TestInternalEngineClose` - Cleanup

### Evaluation Tests (`evaluation_test.go`)
- ✅ `TestGetPieceValue` - Material values (13 subtests for all piece types)
- ✅ `TestGetPieceSquareValue` - Positional values and symmetry
- ✅ `TestEvaluateStartingPosition` - Starting position ~= 0
- ✅ `TestEvaluateMaterialAdvantage` - Material counting (3 subtests)
- ✅ `TestEvaluateCheckmate` - Checkmate detection
- ✅ `TestEvaluateStalemate` - Stalemate = 0
- ✅ `TestEvaluatePositionalAdvantage` - Piece-square tables
- ✅ `TestPieceSquareTablesSymmetry` - Table structure (6 subtests)
- ✅ `BenchmarkEvaluate` - Performance benchmark

### Search Tests (`search_test.go`)
- ✅ `TestSearchStartingPosition` - Basic search functionality
- ✅ `TestSearchFindsMate` - Tactical awareness
- ✅ `TestSearchDepthZero` - Quiescence at depth 0
- ✅ `TestSearchWithAlphaBetaPruning` - Alpha-beta bounds
- ✅ `TestQuiescence` - Quiescence search
- ✅ `TestQuiescenceDepthLimit` - Depth limiting
- ✅ `TestQuiescenceOnlySearchesCaptures` - Capture-only search
- ✅ `TestSearchIterativeDeepening` - Increasing depth search
- ✅ `TestSearchNoLegalMoves` - Stalemate/checkmate handling
- ✅ `BenchmarkSearch` - Search performance
- ✅ `BenchmarkQuiescence` - Quiescence performance

### Move Ordering Tests (`move_ordering_test.go`)
- ✅ `TestMoveOrderScore` - Scoring captures vs quiet moves
- ✅ `TestOrderMoves` - Move reordering
- ✅ `TestOrderMovesWithCaptures` - Captures first
- ✅ `TestOrderMovesEmptyList` - Edge case: empty list
- ✅ `TestOrderMovesSingleMove` - Edge case: single move
- ✅ `TestMoveOrderingConsistency` - Deterministic ordering
- ✅ `BenchmarkOrderMoves` - Ordering performance
- ✅ `BenchmarkMoveOrderScore` - Scoring performance

### Analysis Tests (`analysis_test.go`)
- ✅ `TestInternalEngineAnalyze` - Iterative deepening analysis
- ✅ `TestInternalEngineAnalyzeStop` - Stop signal handling
- ✅ `TestInternalEngineAnalyzeInvalidFEN` - Error handling

## Test Patterns

### Unit Tests
Tests focus on individual functions and methods:
- Material value calculations
- Piece-square value lookups
- Move scoring
- Position evaluation

### Integration Tests
Tests verify components working together:
- Search with evaluation
- Move ordering with search
- Analysis with all components

### Edge Cases
Tests cover boundary conditions:
- Empty move lists
- Single moves
- Checkmate/stalemate positions
- Invalid FEN strings

### Performance Tests
Benchmarks measure:
- Evaluation speed
- Search performance
- Move ordering overhead
- Quiescence search speed

## Adding New Tests

When adding new features, follow this pattern:

1. **Create test in matching file**
   - Evaluation feature → `evaluation_test.go`
   - Search improvement → `search_test.go`

2. **Test structure**
   ```go
   func TestNewFeature(t *testing.T) {
       engine := NewEngine()
       
       // Setup
       fen := "..."
       fenFunc, _ := chess.FEN(fen)
       game := chess.NewGame(fenFunc)
       
       // Execute
       result := engine.newFeature(...)
       
       // Verify
       if result != expected {
           t.Errorf("Expected %v, got %v", expected, result)
       }
   }
   ```

3. **Add benchmark if performance-critical**
   ```go
   func BenchmarkNewFeature(b *testing.B) {
       engine := NewEngine()
       // setup...
       
       b.ResetTimer()
       for i := 0; i < b.N; i++ {
           engine.newFeature(...)
       }
   }
   ```

## Continuous Integration

All tests must pass before merging:
```bash
go test ./engines/builtin
```

Expected output:
```
PASS
ok      gochess-board/engines/builtin   4.687s
```

## Future Test Additions

Areas that could use more coverage:
- [ ] Transposition table tests (when implemented)
- [ ] Opening book integration tests
- [ ] Endgame tablebase tests
- [ ] Time management edge cases
- [ ] Parallel search tests
- [ ] Memory leak tests
- [ ] Stress tests with long games
