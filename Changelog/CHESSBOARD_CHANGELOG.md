# Chessboard.js Changelog

## Version 2.0.1 (December 2024)

### Major Features Added

#### Ghost Pieces and PV Animation
- **Ghost Pieces**: Semi-transparent piece visualization for principal variation moves
  - `board.addGhostPiece(fromSquare, toSquare, piece)` - Add ghost piece with automatic source piece hiding
  - `board.clearGhostPieces()` - Remove all ghost pieces and restore original pieces
  - Ghost pieces use 0.5 opacity with fade-in animation
  - Automatically hides original pieces on source and destination squares

#### Principal Variation Animation
- **Animated PV Display**: Looping animation of principal variation moves
  - `board.drawPrincipalVariation(data, showPV, showBestMove, gameInstance)` - Animate PV with smart queuing
  - `board.drawPVArrowAtIndex(data, index, clearPrevious, showGhostPieces, gameInstance, clearGhosts)` - Draw single PV arrow
  - `board.cancelPVAnimation()` - Cancel ongoing animation
  - `board.setPositionChanged()` - Mark position change to force-start next PV
  - Timing: 2s for first move, 1.5s for subsequent moves, 2s pause between loops
  - Smart queuing: New PV during animation is queued and applied after current loop
  - Force-start on position changes (moves made)
  - Displays up to 6 moves (3 full moves) in PV mode

#### Arrow Enhancements
- **Move Numbers**: Display move numbers on arrows (e.g., "1", "1...", "2")
- **Score Labels**: Show evaluation scores on arrows
  - Centipawn scores: `+0.50`, `-1.25`
  - Mate scores: `+M3`, `-M5`
- **Opacity Control**: Configurable arrow opacity (0.8 for PV arrows)
- **Color Coding**: White (#FFFFFF) for white moves, black (#000000) for black moves

### API Changes

#### New Methods
```javascript
// Ghost Pieces
board.addGhostPiece(fromSquare, toSquare, piece)
board.clearGhostPieces()

// PV Animation (accepts pre-computed move data)
board.drawPrincipalVariation(pvData, showPV, showBestMove)
// pvData = { moves: [{from, to, piece, moveNumber, isBlackMove}, ...], scoreType, score }

board.drawPVArrowAtIndex(pvData, index, clearPrevious, showGhostPieces, clearGhosts)
board.cancelPVAnimation()
board.setPositionChanged()

// Multiple Best Moves (Multi-PV) - accepts pre-computed move data
board.drawMultipleBestMoves(multiPVLines, options)
// multiPVLines = [{from, to, piece, scoreType, score}, ...]
// Options: { colors: ['#color1', '#color2', ...], maxLines: 3, opacity: 1.0 }

// Score Formatting Helper
board.formatScoreLabel(scoreType, score)
// Returns formatted string like "+0.50", "-1.25", "+M3", "-M5"
```

#### Application Helper Functions (in gochess-analysis.js)
```javascript
// Prepare PV moves for visualization (handles Chess.js logic)
preparePVMoves(data, gameInstance)
// Returns: { moves: [...], scoreType, score }

// Prepare multi-PV moves for visualization
prepareMultiPVMoves(multiPV, gameInstance)
// Returns: [{from, to, piece, scoreType, score}, ...]
```

#### Enhanced Methods
```javascript
// Arrow drawing now supports move numbers
board.drawArrow(from, to, color, label, opacity, clearPrevious, moveNumber)
```

### Code Quality Improvements
- **Extracted score formatting logic** into reusable `formatScoreLabel()` helper function
- **Moved multi-PV visualization** from application to library for better separation of concerns
- **Eliminated code duplication** between PV animation and multi-PV display
- **Improved API consistency** with configurable options for multi-PV colors and opacity
- **Removed Chess.js dependency** - Library is now purely a visualization layer
  - Application prepares move data using Chess.js
  - Library accepts pre-computed data structures
  - Better separation of concerns: chess logic in app, visualization in library
  - Library can now be reused for other chess applications without modification

### Bug Fixes
- Fixed mate score formatting to include negative sign for negative mate scores
- Fixed arrow opacity rendering with proper `stroke-opacity` attribute
- Fixed ghost piece cleanup to properly restore original pieces

### CSS Changes
- Added `.ghost-piece` class with 0.5 opacity
- Added `.ghost-fade-in` animation class for smooth appearance
- Ghost pieces positioned absolutely to overlay on squares

### Testing
- Added 59 comprehensive tests for ghost pieces and PV animation
- Tests automatically skip in Node.js environment (require browser DOM)
- Total test suite: 159 tests (157 passing in browser, 98 passing in Node.js)

### Performance
- Efficient ghost piece management with cleanup tracking
- Minimal DOM manipulation for smooth animations
- Smart PV queuing prevents animation interruption

### Breaking Changes
None - All changes are additive and backward compatible

---

## Version 1.0.1 (Previous)

### Features
- Click-to-select and click-to-move functionality
- SVG arrow drawing for engine analysis visualization
  - `board.drawArrow(from, to, color, label, opacity, clearPrevious)`
  - `board.clearArrow()`
  - `board.getArrow()`
- Error handling improvements (console.error instead of alert)

---

## Version 1.0.0 (Original)

Original chessboard.js by Chris Oakman
- https://github.com/oakmac/chessboardjs/
- Released under MIT license
