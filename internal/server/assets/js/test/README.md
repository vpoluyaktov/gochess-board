# Go Chess Test Suite

Comprehensive unit tests for variant logic, move history display, and chessboard library.

## Test Framework

Uses **Mocha** (test runner) + **Chai** (assertions)

## Test Files

- **variant-tests.js** - Variant creation, merging, parsing, and saving (22 tests)
- **history-tests.js** - PGN tree visualization and move history display (11 tests)
- **chessboard-tests.js** - Chessboard library utilities, ghost pieces, and PV animation (69 tests)
- **navigation-tests.js** - Position navigation and click handling (25 tests, 2 pending)
- **codemirror-integration-tests.js** - CodeMirror 5 API integration (23 tests)
- **variant-window-tests.js** - Window messaging for variant windows (9 tests)

**Total: 159 tests**
- **Browser**: 157 passing, 2 pending
- **Node.js**: 98 passing, 61 pending (59 ghost piece tests + 2 navigation tests require browser)

## Running Tests

### Option 1: Browser (Recommended for development)

1. Open `test-runner.html` in your browser
2. Tests will run automatically and display results
3. Refresh page to re-run tests

```bash
# From project root
open server/assets/js/test/test-runner.html
# or
firefox server/assets/js/test/test-runner.html
```

### Option 2: Command Line (Node.js)

1. Install dependencies:
```bash
cd server/assets/js/test
npm install
```

2. Run tests:
```bash
npm test
```

3. Watch mode (re-run on file changes):
```bash
npm run test:watch
```

## Test Coverage

### Variant Tests (variant-tests.js)

#### 1. Basic Variant Storage
- ✓ Store variant at correct position
- ✓ Support multiple variants at same position

#### 2. Nested Variant Storage
- ✓ Store nested variants with position adjustment
- ✓ Handle deeply nested variants (3+ levels)

#### 3. Variant Merging
- ✓ Add new variant (variantIndex = -1)
- ✓ Replace existing variant (variantIndex >= 0)
- ✓ Merge sub-variants with position adjustment

#### 4. PGN Parsing with Variants
- ✓ Parse simple PGN with one variant
- ✓ Parse PGN with nested variants
- ✓ Parse PGN with multiple variants at same position
- ✓ Handle black move variants correctly

#### 5. PGN Building with Variants
- ✓ Build standard PGN with simple variant
- ✓ Build PGN with nested variants
- ✓ Build PGN with multiple variants

#### 6. Variant Navigation
- ✓ Enable/disable Open Variation button correctly
- ✓ Detect variants at current position

#### 7. Edge Cases
- ✓ Empty move history
- ✓ Variant with no moves
- ✓ Variant at invalid position
- ✓ Deeply nested variants (5+ levels)

#### 8. Round-trip Testing
- ✓ Preserve variants through save/load cycle
- ✓ Preserve multiple variants through round-trip

### History Tests (history-tests.js)

#### 1. PGN Tree Visualization
- ✓ Display variant after white move on same line as move pair
- ✓ Keep variant move pairs on same line
- ✓ Handle variant after white move with black continuation
- ✓ Display variant after black move correctly
- ✓ Handle multiple variants at same position
- ✓ Handle nested variants
- ✓ Not break move pairs when variant exists
- ✓ Handle empty move history
- ✓ Handle game with no variants

#### 2. Move Pair Formatting
- ✓ Pad white moves for alignment
- ✓ Format move numbers correctly

### Chessboard Tests (chessboard-tests.js)

#### 1. Library Loading (3 tests)
- ✓ Chessboard constructor available
- ✓ ChessBoard alias available
- ✓ Utility functions available (fenToObj, objToFen)

#### 2. FEN Utility Functions (4 tests)
- ✓ Convert FEN to position object
- ✓ Convert position object to FEN
- ✓ Handle empty position
- ✓ Round-trip FEN conversion

#### 3. Custom Modifications (3 tests)
- ✓ Arrow drawing feature documented
- ✓ console.error instead of alert
- ✓ Ghost pieces and PV animation features documented

**Note:** Full DOM-dependent chessboard tests (board creation, piece movement, arrow drawing) 
require a real browser environment and should be tested with `test-runner.html`.

#### 4. Ghost Pieces and PV Animation (59 tests)

##### 4.1 Ghost Piece Management (10 tests)
- ✓ Have addGhostPiece method
- ✓ Have clearGhostPieces method
- ✓ Add ghost piece to destination square
- ✓ Hide original piece on source square
- ✓ Hide original piece on destination square for captures
- ✓ Remove existing ghost pieces before adding new one
- ✓ Clear all ghost pieces
- ✓ Restore original pieces when clearing
- ✓ Apply ghost-piece CSS class
- ✓ Apply ghost-fade-in animation class

##### 4.2 PV Animation Control (6 tests)
- ✓ Have cancelPVAnimation method
- ✓ Have setPositionChanged method
- ✓ Have drawPrincipalVariation method
- ✓ Have drawPVArrowAtIndex method
- ✓ Cancel PV animation and clear ghost pieces
- ✓ Set position changed flag

##### 4.3 PV Arrow Drawing (13 tests)
- ✓ Draw single PV arrow without ghost pieces
- ✓ Draw PV arrow with ghost piece
- ✓ Include score label on first arrow
- ✓ Format mate scores correctly
- ✓ Show move numbers on arrows
- ✓ Use white color for white moves
- ✓ Use black color for black moves
- ✓ Apply moves to temporary game without affecting original
- ✓ Handle promotion moves
- ✓ Handle invalid moves gracefully

##### 4.4 PV Animation (7 tests)
- ✓ Require game instance parameter
- ✓ Draw single move in best move mode
- ✓ Not show ghost pieces in best move mode
- ✓ Not restart animation for same PV sequence
- ✓ Handle empty PV array
- ✓ Limit PV to 6 moves maximum

##### 4.5 Integration with Arrow Drawing (2 tests)
- ✓ Clear previous arrows when drawing PV
- ✓ Use 0.8 opacity for PV arrows

##### 4.6 Edge Cases (5 tests)
- ✓ Handle very short PV (1 move)
- ✓ Handle PV with invalid format
- ✓ Handle negative centipawn scores
- ✓ Handle negative mate scores

**Purpose:** These tests verify the ghost piece visualization and principal variation 
animation features that were added to the chessboard library. They ensure that ghost 
pieces are properly created, positioned, and cleaned up, and that PV animations work 
correctly with proper timing, queuing, and cancellation behavior.

**Note:** Ghost piece tests require a real browser environment with full DOM support. 
They are automatically skipped when running in Node.js (command line) and will show 
as "pending". Run `test-runner.html` in a browser to execute these tests.

### Navigation Tests (navigation-tests.js)

#### 1. Position Navigation
- ✓ Navigate to specific position (goToPosition)
- ✓ Step forward/backward correctly
- ✓ Go to start/end positions
- ✓ Handle boundary conditions (can't go past start/end)
- ✓ Prevent navigation to invalid positions
- ✓ Skip navigation if already at target position

#### 2. Click Navigation Parsing
- ✓ Parse white move clicks correctly
- ✓ Parse black move clicks correctly
- ✓ Detect variant lines (└─ marker)
- ✓ Detect continuation lines (indentation)

#### 3. Variant Selection
- ✓ Store selected variant info (position, index, line range)
- ✓ Clear selected variant when clicking main line
- ✓ Enable variant button when variant is selected

#### 4. Position Indicator
- ✓ Show correct position at start (0/N)
- ✓ Show correct position in middle (M/N)
- ✓ Show correct position at end (N/N)

#### 5. Move History Display
- ✓ Format move pairs correctly
- ⏸ Display variants on separate lines (pending - edge case)
- ⏸ Indent nested variants (pending - edge case)

#### 6. Game State Consistency
- ✓ Maintain FEN consistency when navigating
- ✓ Set isNavigating flag correctly
- ✓ Handle empty move history

**Note:** 2 tests are pending (skipped) as they test edge cases in variant display 
that will be addressed in a future update.

### CodeMirror Integration Tests (codemirror-integration-tests.js)

#### 1. Text Content Management
- ✓ Set and get text content (setValue/getValue)
- ✓ Count lines correctly (lineCount)
- ✓ Handle empty content
- ✓ Get individual lines (getLine)

#### 2. Text Marking and Highlighting
- ✓ Create text marks (markText)
- ✓ Clear individual marks (mark.clear)
- ✓ Clear all marks (getAllMarks)
- ✓ Support multiple marks on same line

#### 3. Scrolling Behavior
- ✓ Scroll to specific position (scrollIntoView)
- ✓ Scroll to bottom when at end of game

#### 4. Event Handling
- ✓ Register paste event handler
- ✓ Register beforeChange event handler
- ✓ Access wrapper element for click handling

#### 5. Coordinate Conversion
- ✓ Convert mouse coordinates to character position (coordsChar)

#### 6. Integration with updateMoveHistoryDisplay
- ✓ Update editor content when move history changes
- ✓ Clear editor when move history is empty
- ✓ Skip update if content is unchanged

#### 7. Integration with highlightCurrentMove
- ✓ Highlight current move
- ✓ Clear previous highlights before adding new one
- ✓ No highlight when at start position
- ✓ Handle highlighting at end of game

#### 8. PGN Paste Handling
- ✓ Handle paste event structure

#### 9. Read-only Enforcement
- ✓ Register beforeChange handler to prevent edits

**Purpose:** These tests ensure all CodeMirror API calls used in the application 
are properly abstracted and will help identify breaking changes when migrating to CodeMirror 6.

---

### Variant Window Messaging Tests (variant-window-tests.js)

#### 1. Message Sending (3 tests)
- ✓ Send variant data to new window
- ✓ Stop sending after first successful send
- ✓ Not send if window is closed

#### 2. Message Receiving (4 tests)
- ✓ Process variant data message
- ✓ Ignore duplicate messages
- ✓ Reject messages from different origin
- ✓ Ignore non-variant messages

#### 3. Message Data Validation (2 tests)
- ✓ Include all required fields in variant data
- ✓ Correctly calculate FEN for variant position

**Purpose:** These tests verify the window.postMessage communication between the main 
window and variant windows. They ensure that variant data is sent correctly, received 
only once, and validated for security. This prevents the "chaotic piece movement" bug 
where messages were processed multiple times.

## Test Structure

```javascript
describe('Test Category', function() {
    beforeEach(function() {
        // Reset state before each test
    });
    
    it('should do something specific', function() {
        // Arrange
        // Act
        // Assert
    });
});
```

## Adding New Tests

1. Choose the appropriate test file:
   - `variant-tests.js` for variant logic
   - `history-tests.js` for PGN display
   - `chessboard-tests.js` for chessboard utilities

2. Add new test in appropriate `describe` block:

```javascript
it('should handle my new scenario', function() {
    // Setup
    gameState.moveHistory = ['e2e4', 'e7e5'];
    
    // Execute
    const result = myFunction();
    
    // Verify
    chai.assert.equal(result, expectedValue);
});
```

## Assertion Examples

```javascript
// Equality
chai.assert.equal(actual, expected);
chai.assert.deepEqual(actualArray, expectedArray);

// Existence
chai.assert.isDefined(variable);
chai.assert.isUndefined(variable);

// Boolean
chai.assert.isTrue(condition);
chai.assert.isFalse(condition);

// String contains
chai.assert.include(string, substring);

// Array/Object
chai.assert.lengthOf(array, 5);
chai.assert.property(object, 'key');
```

## Debugging Failed Tests

1. **Browser**: Open browser console to see detailed error messages
2. **Command line**: Look at stack trace in terminal
3. **Add logging**: Use `console.log()` in tests to debug

```javascript
it('should debug this', function() {
    console.log('Current state:', gameState);
    const result = myFunction();
    console.log('Result:', result);
    chai.assert.equal(result, expected);
});
```

## Continuous Integration

To run tests in CI/CD pipeline:

```yaml
# .github/workflows/test.yml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-node@v2
      - run: cd server/assets/js/test && npm install
      - run: cd server/assets/js/test && npm test
```

## Test Data Examples

### Simple Game with Variant
```javascript
gameState.moveHistory = ['e2e4', 'e7e5', 'g1f3', 'b8c6'];
gameState.variants[2] = [['f1c4', 'g8f6']];
// Main: 1.e4 e5 2.Nf3 Nc6
// Variant: (2.Bc4 Nf6)
```

### Nested Variants
```javascript
gameState.moveHistory = ['e2e4', 'e7e5'];
gameState.variants[2] = [['g1f3', 'b8c6', 'f1c4']];
gameState.variants[4] = [['f1b5']]; // Sub-variant
// Main: 1.e4 e5
// Variant: (2.Nf3 Nc6 3.Bc4 (3.Bb5))
```

## Performance Testing

For large games with many variants:

```javascript
it('should handle 100 variants efficiently', function() {
    this.timeout(5000); // Increase timeout
    
    for (let i = 0; i < 100; i++) {
        gameState.variants[i] = [['e2e4', 'e7e5']];
    }
    
    const start = Date.now();
    const pgn = buildStandardPGN();
    const duration = Date.now() - start;
    
    chai.assert.isBelow(duration, 1000); // Should complete in < 1s
});
```

## Troubleshooting

### Tests not running in browser
- Check browser console for errors
- Ensure all script paths are correct
- Verify jQuery and Chess.js are loaded

### Tests failing unexpectedly
- Check if global state is being reset in `beforeEach`
- Verify test isolation (tests should not depend on each other)
- Check for timing issues with async code

### Module not found errors
- Run `npm install` in test directory
- Check that all dependencies are listed in package.json
