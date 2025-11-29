# Multi-PV Feature Tests

## Overview

Comprehensive test suite for the 3 best moves (Multi-PV) feature covering all engine types and edge cases.

## Test Files

### 1. `analysis/analysis_multipv_test.go`

Tests for the analysis package Multi-PV functionality.

#### Data Structure Tests
- **TestPVLine_Structure** - Verifies PVLine struct fields
- **TestPVLine_MateScore** - Tests PVLine with mate scores
- **TestAnalysisInfo_MultiPV** - Tests AnalysisInfo with 3 PV lines
- **TestAnalysisInfo_SinglePV** - Tests backward compatibility with single PV
- **TestAnalysisInfo_EmptyMultiPV** - Tests empty MultiPV array handling

#### Integration Tests
- **TestBuiltinAnalysisEngine_MultiPV** - Tests builtin engine multi-PV generation
- **TestBuiltinAnalysisEngine_MultiPV_TacticalPosition** - Tests multi-PV on tactical position
- **TestParseCECPAnalysis_MultiPV** - Tests CECP engine single-entry MultiPV creation

#### Validation Tests
- **TestMultiPV_ScoreOrdering** - Verifies PV lines are sorted by score
- **TestMultiPV_MixedScoreTypes** - Tests handling of mixed score types (cp/mate)

### 2. `engines/builtin/analysis_multipv_test.go`

Tests for the builtin engine's multi-PV search implementation.

#### Core Functionality Tests
- **TestSearchWithStatsMultiPV_StartingPosition** - Tests multi-PV from starting position
- **TestSearchWithStatsMultiPV_MiddleGame** - Tests multi-PV in middlegame
- **TestSearchWithStatsMultiPV_TacticalPosition** - Tests multi-PV with tactical shots
- **TestSearchWithStatsMultiPV_EndgamePosition** - Tests multi-PV in endgame
- **TestSearchWithStatsMultiPV_FewMoves** - Tests when few legal moves available

#### Edge Case Tests
- **TestSearchWithStatsMultiPV_StopChannel** - Tests stop channel handling
- **TestSearchWithStatsMultiPV_ScoreTypes** - Tests centipawn vs mate score handling

#### Data Structure Tests
- **TestPVLine_Structure** - Verifies builtin PVLine struct
- **TestAnalysisInfo_MultiPVField** - Tests MultiPV field in builtin AnalysisInfo

#### Performance Tests
- **BenchmarkSearchWithStatsMultiPV** - Benchmarks multi-PV at depth 3
- **BenchmarkSearchWithStatsMultiPV_Depth4** - Benchmarks multi-PV at depth 4

## Running the Tests

### Run all Multi-PV tests
```bash
go test ./analysis/... ./engines/builtin/... -v -run "MultiPV|PVLine"
```

### Run specific test categories

**Data structure tests:**
```bash
go test ./analysis/... -v -run TestPVLine
go test ./analysis/... -v -run TestAnalysisInfo_MultiPV
```

**Builtin engine tests:**
```bash
go test ./engines/builtin/... -v -run TestSearchWithStatsMultiPV
```

**Integration tests:**
```bash
go test ./analysis/... -v -run TestBuiltinAnalysisEngine_MultiPV
```

**Validation tests:**
```bash
go test ./analysis/... -v -run TestMultiPV_ScoreOrdering
```

### Run benchmarks
```bash
go test ./engines/builtin/... -bench=BenchmarkSearchWithStatsMultiPV -benchmem
```

## Test Coverage

### Engine Types Covered
- ✅ **UCI Engines** - Tested via integration (requires Stockfish)
- ✅ **Builtin Engine** - Comprehensive unit and integration tests
- ✅ **CECP Engines** - Tested via parsing function

### Positions Tested
- ✅ Starting position
- ✅ Middlegame positions
- ✅ Tactical positions (mate threats)
- ✅ Endgame positions
- ✅ Positions with few legal moves

### Features Tested
- ✅ Multi-PV generation (3 lines)
- ✅ Score ordering (descending)
- ✅ Centipawn scores
- ✅ Mate scores
- ✅ PV move sequences
- ✅ Backward compatibility (single PV)
- ✅ Stop channel handling
- ✅ Empty/edge cases

## Expected Test Results

All tests should pass with output similar to:

```
=== RUN   TestSearchWithStatsMultiPV_StartingPosition
    analysis_multipv_test.go:60: Depth 3: Found 3 PV lines, best move: a2a4, score: 54, nodes: 5699
--- PASS: TestSearchWithStatsMultiPV_StartingPosition (0.10s)

=== RUN   TestSearchWithStatsMultiPV_MiddleGame
    analysis_multipv_test.go:100: Middlegame: Found 3 PV lines, best: c4d5 (score: 12), nodes: 63559
    analysis_multipv_test.go:103:   PV 1: c4d5 (score: 12 cp)
    analysis_multipv_test.go:103:   PV 2: c4b5 (score: 6 cp)
    analysis_multipv_test.go:103:   PV 3: d1e2 (score: -19 cp)
--- PASS: TestSearchWithStatsMultiPV_MiddleGame (1.54s)

=== RUN   TestSearchWithStatsMultiPV_TacticalPosition
    analysis_multipv_test.go:132: Found mate in 2
    analysis_multipv_test.go:142: Tactical: Found 3 PV lines, best: h5f7 (score: 9997 mate), nodes: 93049
--- PASS: TestSearchWithStatsMultiPV_TacticalPosition (2.85s)
```

## Performance Characteristics

### Builtin Engine Multi-PV Performance
- **Starting position (depth 3)**: ~5,700 nodes, ~0.10s
- **Middlegame (depth 4)**: ~63,000 nodes, ~1.5s
- **Tactical position (depth 4)**: ~93,000 nodes, ~2.8s
- **Endgame (depth 5)**: ~1,600 nodes, ~0.04s

Multi-PV search is approximately 2-3x slower than single-PV search because it evaluates all root moves without alpha-beta pruning at the root level.

## Test Maintenance

### Adding New Tests

When adding new multi-PV functionality:

1. Add unit tests for new data structures
2. Add integration tests for engine behavior
3. Add validation tests for correctness
4. Update this documentation

### Common Test Patterns

**Testing PV line structure:**
```go
if len(multiPV) != 3 {
    t.Errorf("Expected 3 PV lines, got %d", len(multiPV))
}
```

**Testing score ordering:**
```go
for i := 0; i < len(multiPV)-1; i++ {
    if multiPV[i].Score < multiPV[i+1].Score {
        t.Error("PV lines not sorted by score")
    }
}
```

**Testing backward compatibility:**
```go
if info.BestMove != info.MultiPV[0].Moves[0] {
    t.Error("BestMove should match first PV line")
}
```

## Known Limitations

1. **CECP engines** - Only single PV tested (protocol limitation)
2. **UCI engines** - Requires actual engine binary for full integration tests
3. **Mate scores** - Display scores differ from internal scores (tested separately)
4. **Performance** - Multi-PV is slower than single-PV (expected behavior)

## Future Test Enhancements

- [ ] Add tests for configurable PV count (1-5 lines)
- [ ] Add tests for PV line display in UI
- [ ] Add tests for engine switching with multi-PV
- [ ] Add stress tests with deep searches
- [ ] Add tests for concurrent multi-PV requests
