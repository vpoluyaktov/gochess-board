# Chess UI Modular Structure

The chess UI has been refactored from a single large file (`chess-ui.js`, 2673 lines) into smaller, domain-specific modules for better maintainability and organization.

## Module Files (gochess-*.js)

### 1. **gochess-state.js** - Game State Management
- Global `gameState` object
- localStorage operations (currently disabled)
- State persistence functions

### 2. **gochess-players.js** - Player Configuration
- Player dropdown management
- Player name handling
- Info text updates
- Player controls visibility
- Game result functions (resign, draw)

### 3. **gochess-board.js** - Board Management
- Board initialization and configuration
- Drag-and-drop event handlers
- Move highlighting
- Board flip functionality

### 4. **gochess-engine.js** - Computer Move Logic
- Engine communication via `/api/computer-move`
- Computer move generation
- Move timing and clock integration
- Computer player detection

### 5. **gochess-clock.js** - Clock Management
- Time formatting and display
- Clock start/stop/pause
- Time control configuration
- Time warning indicators
- Timeout detection

### 6. **gochess-history.js** - Move History Display
- CodeMirror integration
- PGN notation with variants (tree-style)
- Current move highlighting
- Move history updates

### 7. **gochess-navigation.js** - Move Navigation
- Forward/backward navigation
- Go to start/end
- Position indicator
- Variant button management
- `openVariation()` function
- `newGame()` function

### 8. **gochess-opening.js** - Opening Display
- Opening name lookup via `/api/opening`
- Opening display updates
- SAN move conversion for API

### 9. **gochess-analysis.js** - Live Analysis
- WebSocket connection to analysis engine
- Principal variation (PV) display
- Arrow drawing on board
- Analysis depth tracking
- Score evaluation display

### 10. **gochess-pgn.js** - PGN Operations
- PGN file loading
- PGN file saving
- Multi-game PGN parsing
- Game selector modal
- Variant parsing (recursive)
- Standard PGN building with variants

### 11. **gochess-variants.js** - Variant Support
- Variant window creation
- Window messaging (postMessage API)
- Variant merging
- `startVariant()`, `mergeVariant()`, `closeVariant()`
- Variant mode detection

### 12. **gochess-main.js** - Initialization
- Board initialization
- Event listener setup
- CodeMirror initialization
- Keyboard shortcuts
- Application startup

## Load Order (in index.html)

The modules must be loaded in this specific order due to dependencies:

```html
<script src="/assets/js/gochess-state.js"></script>
<script src="/assets/js/gochess-players.js"></script>
<script src="/assets/js/gochess-board.js"></script>
<script src="/assets/js/gochess-engine.js"></script>
<script src="/assets/js/gochess-clock.js"></script>
<script src="/assets/js/gochess-history.js"></script>
<script src="/assets/js/gochess-navigation.js"></script>
<script src="/assets/js/gochess-opening.js"></script>
<script src="/assets/js/gochess-analysis.js"></script>
<script src="/assets/js/gochess-pgn.js"></script>
<script src="/assets/js/gochess-variants.js"></script>
<script src="/assets/js/gochess-main.js"></script>
```

## Key Dependencies

- **gochess-state.js** - Must load first (defines `gameState`)
- **gochess-players.js** - Used by board, engine, and PGN modules
- **gochess-board.js** - Defines `board`, `game`, and highlighting functions
- **gochess-engine.js** - Depends on players and board
- **gochess-history.js** - Defines `moveHistoryEditor` and display functions
- **gochess-navigation.js** - Uses all previous modules
- **gochess-analysis.js** - Defines `analysisActive` and `analysisWs`
- **gochess-main.js** - Must load last (initialization)

## Global Variables by Module

### gochess-state.js
- `gameState` (object)

### gochess-board.js
- `board` (Chessboard instance)
- `game` (Chess.js instance)
- `lastMoveSquares` (object)
- `squareClass` (string)

### gochess-engine.js
- `isComputerThinking` (boolean)

### gochess-history.js
- `moveHistoryEditor` (CodeMirror instance)

### gochess-analysis.js
- `analysisWs` (WebSocket)
- `analysisActive` (boolean)

### gochess-variants.js
- `isVariantMode` (boolean)
- `variantWindow` (Window)
- `mainWindow` (Window)
- `variantStartPosition` (number)

## Benefits of Modular Structure

1. **Maintainability** - Easier to find and modify specific functionality
2. **Readability** - Each file has a clear, focused purpose
3. **Collaboration** - Multiple developers can work on different modules
4. **Testing** - Easier to test individual components
5. **Reusability** - Modules can be reused or replaced independently
6. **Organization** - Clear separation of concerns

## Backup

The original monolithic file has been backed up as:
- `/home/ubuntu/git/gochess-board/server/assets/js/chess-ui.js.backup`
