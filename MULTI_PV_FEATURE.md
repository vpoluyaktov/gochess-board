# Multi-PV (3 Best Moves) Feature

## Overview

The analysis feature now supports displaying 3 best moves simultaneously, allowing users to see alternative candidate moves with their evaluations.

## UI Controls

Two radio button options are available in the analysis panel:

1. **Show principal variation** - Displays a sequence of best moves (the main line)
2. **Show 3 best moves** - Displays 3 alternative first moves with different colored arrows

## Engine Support

### UCI Engines (Full Support) ✅

**Examples:** Stockfish, Komodo, Leela Chess Zero

- **Implementation:** Uses native UCI `MultiPV` option
- **Configuration:** Automatically sets `setoption name MultiPV value 3`
- **Behavior:** Engine calculates 3 independent variations simultaneously
- **Performance:** Efficient - engine optimizes search across multiple lines

### Built-in Engine (Full Support) ✅

**Engine:** GoChess Board built-in engine

- **Implementation:** Custom multi-PV search in `searchWithStatsMultiPV()`
- **Behavior:** Searches all legal moves, ranks them, returns top 3
- **Performance:** Slightly slower than single-PV as it searches all moves without pruning
- **Quality:** Provides accurate top 3 moves with proper evaluations

### CECP/XBoard Engines (Limited Support) ⚠️

**Examples:** GNU Chess, Crafty

- **Implementation:** Returns single best move only
- **Limitation:** CECP protocol doesn't have standard multi-PV support
- **Behavior:** When "Show 3 best moves" is selected, falls back to showing single best move
- **MultiPV Array:** Contains single entry for consistency

## Visual Display

### Principal Variation Mode
- Shows sequence of moves in the main line
- White arrows for white moves, dark arrows for black moves
- Move numbers displayed on arrows
- Score shown on first arrow

### 3 Best Moves Mode
- **1st best move:** Green arrow (#15781B)
- **2nd best move:** Blue arrow (#1E88E5)
- **3rd best move:** Orange arrow (#FFA726)
- Each arrow shows its evaluation score
- Opacity slightly decreases for 2nd and 3rd moves

## Technical Implementation

### Data Structure

```go
type PVLine struct {
    Score     int      `json:"score"`
    ScoreType string   `json:"scoreType"` // "cp" or "mate"
    Moves     []string `json:"moves"`
}

type AnalysisInfo struct {
    // ... existing fields ...
    MultiPV   []PVLine `json:"multiPV"`   // Multiple PV lines
}
```

### Backend Flow

1. **UCI Engine:** 
   - Sets `MultiPV=3` during initialization
   - Parses `multipv` field from engine output
   - Combines 3 PV lines into single `AnalysisInfo`

2. **Builtin Engine:**
   - Searches all legal moves at root
   - Sorts by evaluation score
   - Returns top 3 with full variations

3. **CECP Engine:**
   - Returns single PV line
   - Wraps in single-entry `MultiPV` array

### Frontend Flow

```javascript
if (showBestMoves && data.multiPV && data.multiPV.length > 0) {
    // Show 3 best moves mode
    drawMultipleBestMoves(data.multiPV);
} else if (data.pv && data.pv.length > 0) {
    // Show principal variation mode (fallback)
    drawPrincipalVariation(data, showPV);
}
```

## Files Modified

- `analysis/analysis.go` - Added `PVLine` struct and `MultiPV` field
- `analysis/analysis_uci.go` - UCI multi-PV support
- `analysis/analysis_builtin.go` - MultiPV conversion for builtin engine
- `analysis/analysis_cecp.go` - Single-entry MultiPV for CECP engines
- `engines/builtin/analysis.go` - Multi-PV search implementation
- `server/assets/js/gochess-analysis.js` - Frontend display logic

## Usage

1. Start analysis with any engine
2. Select "Show 3 best moves" radio button
3. See colored arrows for top 3 candidate moves
4. Switch back to "Show principal variation" to see the main line sequence

## Performance Notes

- **UCI engines:** No performance impact (native support)
- **Builtin engine:** ~2-3x slower than single-PV (searches all moves)
- **CECP engines:** No performance impact (single-PV only)

## Future Enhancements

Potential improvements:
- Configurable number of PV lines (1-5)
- Display full variations for each PV line (not just first move)
- CECP multi-PV via sequential analysis with excluded moves
- Persistent user preference for display mode
