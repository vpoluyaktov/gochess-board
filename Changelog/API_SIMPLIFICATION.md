# API Simplification - Clear Separation of Visualization Modes

## Problem

The `drawPrincipalVariation()` method was handling **three distinct use cases** in one method:

1. **Show single best move** - One arrow with score, no animation
2. **Show multiple best moves** - Multiple arrows (3 alternatives) with scores
3. **Show PV animation** - Looping animation with ghost pieces

This violated the **Single Responsibility Principle** and made the API confusing:

```javascript
// Old API - unclear what this does
board.drawPrincipalVariation(pvData, showPV, showBestMove);

// What does showPV=false, showBestMove=true do?
// What does showPV=true, showBestMove=false do?
// Confusing boolean flags!
```

---

## Solution

Split into **three focused methods**, one for each use case:

### 1. `drawBestMove(pvData)` - Single Best Move
```javascript
// Clear intent: Draw one arrow showing the best move
board.drawBestMove(pvData);
```

**Use case:** Show engine's top recommendation
- Draws single arrow from first move
- Includes evaluation score label
- No animation
- No ghost pieces

### 2. `drawMultipleBestMoves(multiPVLines, options)` - Multiple Alternatives
```javascript
// Clear intent: Draw multiple alternative moves
board.drawMultipleBestMoves(multiPVLines, {
    colors: ['#15781Bff', '#FFD700ff', '#DC3545ff'],
    maxLines: 3,
    opacity: 1.0
});
```

**Use case:** Show top 3 engine recommendations
- Draws multiple arrows (green, yellow, red)
- Each with its own score
- No animation
- No ghost pieces

### 3. `drawPVAnimation(pvData, options)` - Animated Principal Variation
```javascript
// Clear intent: Animate the principal variation
board.drawPVAnimation(pvData, {
    maxMoves: 6,
    firstMoveDelay: 2000,
    subsequentMoveDelay: 1500,
    pauseAfterLoop: 2000
});
```

**Use case:** Show animated sequence of moves
- Animates up to 6 moves (3 full moves)
- Shows ghost pieces
- Loops continuously
- Configurable timing

---

## API Comparison

### Before (Confusing)
```javascript
// Case 1: Single best move
board.drawPrincipalVariation(pvData, false, true);  // ❌ What do these booleans mean?

// Case 2: Multiple best moves
board.drawMultipleBestMoves(multiPVLines);  // ✅ This one was already clear

// Case 3: PV animation
board.drawPrincipalVariation(pvData, true, false);  // ❌ Confusing booleans again
```

### After (Clear)
```javascript
// Case 1: Single best move
board.drawBestMove(pvData);  // ✅ Crystal clear!

// Case 2: Multiple best moves
board.drawMultipleBestMoves(multiPVLines);  // ✅ Already clear

// Case 3: PV animation
board.drawPVAnimation(pvData);  // ✅ Intent is obvious!
```

---

## Benefits

### 1. **Self-Documenting Code**
```javascript
// Old - need to look up what parameters mean
board.drawPrincipalVariation(pvData, true, false);

// New - method name tells you exactly what it does
board.drawPVAnimation(pvData);
```

### 2. **Single Responsibility**
Each method does **one thing** and does it well:
- `drawBestMove()` - Shows one move
- `drawMultipleBestMoves()` - Shows multiple alternatives
- `drawPVAnimation()` - Animates a sequence

### 3. **Easier to Extend**
Want to add options to animation? Easy:
```javascript
board.drawPVAnimation(pvData, {
    maxMoves: 10,  // Show more moves
    speed: 'fast'  // Add speed control
});
```

### 4. **Better Type Safety**
```javascript
// Old - boolean flags are error-prone
board.drawPrincipalVariation(pvData, false, true);  // Easy to swap by mistake

// New - method names prevent mistakes
board.drawBestMove(pvData);  // Can't accidentally animate
board.drawPVAnimation(pvData);  // Can't accidentally show static arrow
```

---

## Application Code Changes

### Before
```javascript
if (showBestMoves && data.multiPV && data.multiPV.length > 0) {
    var multiPVLines = prepareMultiPVMoves(data.multiPV, game);
    board.drawMultipleBestMoves(multiPVLines);
} else if (data.pv && data.pv.length > 0) {
    var pvData = preparePVMoves(data, game);
    board.drawPrincipalVariation(pvData, showPV, showBestMove);  // ❌ Confusing
}
```

### After
```javascript
if (showBestMoves && data.multiPV && data.multiPV.length > 0) {
    // Case 2: Show 3 best moves (multiple arrows with scores)
    var multiPVLines = prepareMultiPVMoves(data.multiPV, game);
    board.drawMultipleBestMoves(multiPVLines);
} else if (data.pv && data.pv.length > 0) {
    var pvData = preparePVMoves(data, game);
    
    if (showPV) {
        // Case 3: Show PV animation (looping animation with ghost pieces)
        board.drawPVAnimation(pvData);  // ✅ Clear!
    } else {
        // Case 1: Show single best move (one arrow with score)
        board.drawBestMove(pvData);  // ✅ Clear!
    }
}
```

---

## Backward Compatibility

The old method is kept as a wrapper for backward compatibility:

```javascript
widget.drawPrincipalVariation = function (pvData, showPV, showBestMove) {
    // Deprecated: Use drawBestMove() or drawPVAnimation() instead
    if (!showPV) {
        widget.drawBestMove(pvData);
    } else {
        widget.drawPVAnimation(pvData);
    }
}
```

**Migration path:**
1. Old code continues to work
2. Developers can migrate at their own pace
3. New code uses clearer API

---

## Method Signatures

### `drawBestMove(pvData)`

**Parameters:**
- `pvData` (Object) - Pre-computed move data
  - `moves` (Array) - Array of move objects
  - `scoreType` (String) - 'cp' or 'mate'
  - `score` (Number) - Evaluation score

**Returns:** void

**Example:**
```javascript
var pvData = preparePVMoves(data, game);
board.drawBestMove(pvData);
```

---

### `drawMultipleBestMoves(multiPVLines, options)`

**Parameters:**
- `multiPVLines` (Array) - Array of move line objects
  - Each line: `{from, to, piece, scoreType, score}`
- `options` (Object, optional)
  - `colors` (Array) - Arrow colors (default: green, yellow, red)
  - `maxLines` (Number) - Max lines to show (default: 3)
  - `opacity` (Number) - Arrow opacity (default: 1.0)

**Returns:** void

**Example:**
```javascript
var multiPVLines = prepareMultiPVMoves(multiPV, game);
board.drawMultipleBestMoves(multiPVLines, {
    colors: ['#00FF00', '#FFFF00', '#FF0000'],
    maxLines: 3
});
```

---

### `drawPVAnimation(pvData, options)`

**Parameters:**
- `pvData` (Object) - Pre-computed move data
  - `moves` (Array) - Array of move objects
  - `scoreType` (String) - 'cp' or 'mate'
  - `score` (Number) - Evaluation score
- `options` (Object, optional)
  - `maxMoves` (Number) - Max moves to animate (default: 6)
  - `firstMoveDelay` (Number) - Delay for first move in ms (default: 2000)
  - `subsequentMoveDelay` (Number) - Delay for subsequent moves in ms (default: 1500)
  - `pauseAfterLoop` (Number) - Pause after last move in ms (default: 2000)

**Returns:** void

**Example:**
```javascript
var pvData = preparePVMoves(data, game);
board.drawPVAnimation(pvData, {
    maxMoves: 8,
    firstMoveDelay: 1500,
    subsequentMoveDelay: 1000
});
```

---

## Design Principles Applied

### 1. **Single Responsibility Principle**
Each method has one clear purpose:
- `drawBestMove()` → Show one move
- `drawMultipleBestMoves()` → Show alternatives
- `drawPVAnimation()` → Animate sequence

### 2. **Self-Documenting Code**
Method names clearly describe what they do:
- No need to read documentation
- No confusing boolean flags
- Intent is obvious from the code

### 3. **Open/Closed Principle**
Easy to extend with options:
```javascript
// Can add new options without breaking existing code
board.drawPVAnimation(pvData, {
    maxMoves: 10,
    speed: 'fast',
    showMoveNumbers: true  // New option
});
```

### 4. **Separation of Concerns**
- Application decides **what** to show
- Library handles **how** to show it
- Clear boundary between layers

---

## Testing

All tests pass with the new API:

```bash
✅ 98 passing (2s)
⏸ 40 pending (browser-only)
```

The backward compatibility wrapper ensures existing tests continue to work.

---

## Conclusion

This refactoring transforms a **confusing multi-purpose method** into **three focused, self-documenting methods**:

| Old API | New API | Clarity |
|---------|---------|---------|
| `drawPrincipalVariation(pvData, false, true)` | `drawBestMove(pvData)` | ✅ Much clearer |
| `drawPrincipalVariation(pvData, true, false)` | `drawPVAnimation(pvData)` | ✅ Much clearer |
| `drawMultipleBestMoves(lines)` | `drawMultipleBestMoves(lines)` | ✅ Already clear |

**Result: A cleaner, more intuitive API that's easier to use and maintain!** 🎯
