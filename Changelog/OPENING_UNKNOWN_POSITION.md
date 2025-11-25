# Opening Display - Unknown Position Handling

## Overview

Updated the opening lookup logic to hide the opening name when the game position leaves the opening book (unknown position). Previously, the last matched opening would persist even after players made moves outside the opening database.

## Problem

### Before
When players made moves beyond the opening book:
1. Opening name would continue to show the last matched opening
2. Example: "Italian Game" would display even after 20+ moves into the middlegame
3. Misleading - the position was no longer in the opening phase

### Example Scenario
```
Moves: e4 e5 Nf3 Nc6 Bc4 Bc5 h4
                              ^^^ Unknown move

Before: Still shows "Italian Game"
After:  Hides opening display (correct behavior)
```

## Solution

### Updated Logic (`server/opening.go`)

Modified the `Lookup` function to return `nil` when ANY move in the sequence is not found in the opening book:

```go
func (ob *OpeningBook) Lookup(moves []string) *OpeningInfo {
    // ... 
    
    for i, move := range moves {
        if current.Children == nil {
            // We've left the opening book - no more moves in database
            return nil
        }

        next, exists := current.Children[move]
        if !exists {
            // This move is not in the opening book
            return nil
        }
        
        // Continue traversing...
    }
    
    return lastOpening
}
```

**Key Changes**:
1. Returns `nil` if `current.Children` is `nil` (reached a leaf node)
2. Returns `nil` if the next move doesn't exist in the tree
3. Only returns an opening if ALL moves are in the opening book

## Behavior

### Known Opening Sequence
```
Moves: e4, e5, Nf3, Nc6, Bc4
Result: Shows "Italian Game"
Display: 📖 Italian Game
```

### Leaving the Opening Book
```
Moves: e4, e5, Nf3, Nc6, Bc4, Bc5, h4
                                    ^^^
Result: Returns nil (h4 not in book after this position)
Display: Hidden
```

### Unknown Opening from Start
```
Moves: a3, a6, b3, b6
       ^^^
Result: Returns nil (a3 followed by a6 not in book)
Display: Hidden
```

## User Experience

### Visual Flow

1. **Game Start**: Opening display hidden (no moves)
2. **Known Opening**: Display appears with opening name
   ```
   📖 Italian Game
   ```
3. **Still in Opening**: Display updates as opening evolves
   ```
   📖 Italian Game: Giuoco Piano
   ```
4. **Leaves Opening Book**: Display disappears immediately
   ```
   (hidden)
   ```
5. **New Game**: Display hidden again

### Benefits

1. **Accurate**: Only shows opening when position is actually in the opening phase
2. **Clear**: Players know when they've left known opening theory
3. **Educational**: Helps players understand opening boundaries
4. **Clean**: No stale information displayed

## Technical Details

### Return Conditions

**Returns Opening** (non-nil):
- All moves in the sequence exist in the opening book
- At least one node in the path has an opening name
- Returns the deepest (most specific) opening found

**Returns nil**:
- Any move in the sequence is not in the opening book
- The position has left known opening theory
- No opening name found in the traversed path

### Edge Cases

#### Case 1: Transpositions
```
Moves: Nf3, d5, d4, Nf6
Result: May return nil if this specific move order isn't in the book
Note: Same position as d4, d5, Nf3, Nf6 but different move order
```

#### Case 2: Deep Variations
```
Moves: e4, c5, Nf3, d6, d4, cxd4, Nxd4, Nf6, Nc3, a6, Be3, e5, Nb3, Be6
Result: Shows opening if all moves are in book (Sicilian Najdorf)
```

#### Case 3: Early Deviation
```
Moves: e4, a6
Result: nil (a6 after e4 not a standard opening)
```

## Testing

### Test Coverage

Created comprehensive tests in `opening_unknown_test.go`:

```go
✓ Italian Game - Known Opening (returns opening)
✓ Italian Game + Unknown Move (returns nil)
✓ Sicilian + Random Moves (returns nil)
✓ Unknown Opening From Start (returns nil)
✓ Single Known Move (returns opening)
✓ Known Opening Exact (returns opening)
```

### Test Results
```
=== RUN   TestOpeningBookUnknownPosition
    ✓ Found: Italian Game (C50)
    ✓ Correctly returned nil for unknown position
    ✓ Correctly returned nil for unknown position
    ✓ Correctly returned nil for unknown position
    ✓ Found: King's Pawn Game (B00)
    ✓ Found: King's Knight Opening (C40)
--- PASS: TestOpeningBookUnknownPosition
```

## API Behavior

### Request
```json
POST /api/opening
{
  "moves": ["e4", "e5", "Nf3", "Nc6", "Bc4", "Bc5", "h4"]
}
```

### Response (Unknown Position)
```json
{
  "eco": "",
  "name": "",
  "pgn": ""
}
```

### JavaScript Handling
```javascript
if (opening && opening.name) {
    // Show opening
    document.getElementById('openingText').textContent = opening.name;
    document.getElementById('openingDisplay').style.display = 'block';
} else {
    // Hide opening (unknown position)
    document.getElementById('openingDisplay').style.display = 'none';
}
```

## Performance

### Impact
- **Lookup Time**: Same O(m) complexity where m = number of moves
- **Early Exit**: May return faster when leaving opening book
- **Memory**: No change in memory usage
- **Network**: Same API call pattern

### Optimization
The function exits early when:
1. A move is not found in the tree
2. Reached a leaf node with no children
3. All moves processed successfully

## Comparison

### Old Behavior (Longest Prefix Match)
```
Input:  e4, e5, Nf3, Nc6, Bc4, Bc5, h4
Output: Italian Game (from e4, e5, Nf3, Nc6, Bc4)
Issue:  Misleading - still shows opening after leaving book
```

### New Behavior (Exact Match Only)
```
Input:  e4, e5, Nf3, Nc6, Bc4, Bc5, h4
Output: nil
Result: Opening display hidden (correct)
```

## Future Enhancements

Possible improvements:
1. **Transposition Detection**: Recognize same positions via different move orders
2. **Partial Match Option**: Add flag to enable old "longest prefix" behavior
3. **Opening Depth**: Show how many moves into the opening
4. **Deviation Indicator**: Show "Left opening after move 7"
5. **Alternative Continuations**: Suggest moves that stay in opening book

## Files Modified

- `server/opening.go` - Updated `Lookup()` function logic
- `server/opening_unknown_test.go` - Added comprehensive tests

## Summary

The opening display now accurately reflects whether the current position is in the opening book. When players make moves outside known opening theory, the display is hidden immediately, providing clear feedback about when they've left the opening phase.
