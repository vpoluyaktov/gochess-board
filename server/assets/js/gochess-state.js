// Game State Management
// Handles game state variables and localStorage operations

var gameState = {
    moveHistory: [],           // UCI move list (e.g., ["e2e4", "e7e5"])
    variants: {},              // Variants at each position: { position: [[variant moves], ...] }
    whiteTimeMs: 300000,       // 5 minutes default
    blackTimeMs: 300000,
    timeControl: {
        initial: 5,            // minutes
        increment: 5           // seconds
    },
    clockRunning: false,
    clockInterval: null,
    lastClockUpdate: null,
    gameStartTime: Date.now(),
    gameId: generateGameId(),  // Unique ID for engine pooling
    whiteMoves: 0,
    blackMoves: 0,
    currentPosition: 0,        // Current position in move history (0 = start, moveHistory.length = end)
    isNavigating: false,       // True when viewing a historical position
    wasClockRunning: false,    // Remember if clock was running before navigation
    selectedVariant: null      // Selected variant info: { position: N, index: M, lineNum: L }
};

// Generate a unique game ID
function generateGameId() {
    return 'game-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
}

// Save game state to localStorage
function saveGameState() {
    // DISABLED: localStorage persistence disabled - always start fresh
    return;
}

// Load game state from localStorage
function loadGameState() {
    // DISABLED: localStorage persistence disabled - always start fresh
    return false;
}

// Clear saved game state
function clearGameState() {
    // DISABLED: localStorage persistence disabled
    return;
}
