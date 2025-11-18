// Chess UI Logic
// Handles game state, player interactions, computer moves, and live analysis

var board = null;
var game = new Chess();
var isComputerThinking = false;
var lastMoveSquares = { from: null, to: null };
var squareClass = 'square-55d63';
var moveHistoryEditor = null;

// Analysis WebSocket connection
var analysisWs = null;
var analysisActive = false;

// -------------------------------------------------------------------------
// Game State Management (Client-side)
// -------------------------------------------------------------------------

var gameState = {
    moveHistory: [],           // UCI move list (e.g., ["e2e4", "e7e5"])
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
    whiteMoves: 0,
    blackMoves: 0,
    currentPosition: 0,        // Current position in move history (0 = start, moveHistory.length = end)
    isNavigating: false,       // True when viewing a historical position
    wasClockRunning: false     // Remember if clock was running before navigation
};

// Save game state to localStorage
function saveGameState() {
    // DISABLED: localStorage persistence disabled - always start fresh
    return;
    
    /* DISABLED CODE:
    try {
        localStorage.setItem('chessGameState', JSON.stringify({
            fen: game.fen(),
            moveHistory: gameState.moveHistory,
            whiteTimeMs: gameState.whiteTimeMs,
            blackTimeMs: gameState.blackTimeMs,
            timeControl: gameState.timeControl,
            clockRunning: gameState.clockRunning,
            gameStartTime: gameState.gameStartTime
        }));
    } catch (e) {
        console.error('Failed to save game state:', e);
    }
    */
}

// Load game state from localStorage
function loadGameState() {
    // DISABLED: localStorage persistence disabled - always start fresh
    return false;
    
    /* DISABLED CODE:
    try {
        const saved = localStorage.getItem('chessGameState');
        if (saved) {
            const state = JSON.parse(saved);
            game.load(state.fen);
            board.position(state.fen);
            gameState.moveHistory = state.moveHistory || [];
            gameState.whiteTimeMs = state.whiteTimeMs || 300000;
            gameState.blackTimeMs = state.blackTimeMs || 300000;
            gameState.timeControl = state.timeControl || { initial: 5, increment: 5 };
            gameState.clockRunning = state.clockRunning || false;
            gameState.gameStartTime = state.gameStartTime || Date.now();
            updateMoveHistoryDisplay();
            updateClockDisplay();
            updateOpeningDisplay();
            
            // Restart clock if it was running
            if (gameState.clockRunning) {
                startClockInterval();
            }
            
            return true;
        }
    } catch (e) {
        console.error('Failed to load game state:', e);
    }
    return false;
    */
}

// Clear saved game state
function clearGameState() {
    // DISABLED: localStorage persistence disabled
    return;
    
    /* DISABLED CODE:
    try {
        localStorage.removeItem('chessGameState');
    } catch (e) {
        console.error('Failed to clear game state:', e);
    }
    */
}

// -------------------------------------------------------------------------
// Last Move Highlighting
// -------------------------------------------------------------------------

function clearLastMoveHighlight() {
    $('#myBoard .' + squareClass).removeClass('highlight-last-move');
}

function highlightLastMove(from, to) {
    clearLastMoveHighlight();
    $('#myBoard .square-' + from).addClass('highlight-last-move');
    $('#myBoard .square-' + to).addClass('highlight-last-move');
    lastMoveSquares = { from: from, to: to };
}

// -------------------------------------------------------------------------
// Player Configuration
// -------------------------------------------------------------------------

function resetPlayerDropdowns() {
    // Simply reset selection to default (Human vs Human)
    document.getElementById('whitePlayer').value = 'human';
    document.getElementById('blackPlayer').value = 'human';
}

function addPlayerToDropdown(dropdownId, playerName) {
    const dropdown = document.getElementById(dropdownId);
    
    // Check if this player name already exists
    for (let i = 0; i < dropdown.options.length; i++) {
        if (dropdown.options[i].value === playerName) {
            // Already exists, just select it
            dropdown.value = playerName;
            return;
        }
    }
    
    // Add new option at the beginning (after the first option which is "Human")
    const option = document.createElement('option');
    option.value = playerName;
    option.textContent = playerName;
    
    // Insert after "Human" option (index 1)
    if (dropdown.options.length > 1) {
        dropdown.insertBefore(option, dropdown.options[1]);
    } else {
        dropdown.appendChild(option);
    }
    
    // Select the new player
    dropdown.value = playerName;
}

function getWhitePlayer() {
    return document.getElementById('whitePlayer').value;
}

function getBlackPlayer() {
    return document.getElementById('blackPlayer').value;
}

function getCurrentPlayer() {
    return game.turn() === 'w' ? getWhitePlayer() : getBlackPlayer();
}

function isHuman(playerValue) {
    return playerValue === 'human';
}

function getPlayerName(playerValue) {
    if (playerValue === 'human') return 'Human';
    const select = document.getElementById('whitePlayer');
    const option = Array.from(select.options).find(opt => opt.value === playerValue);
    return option ? option.textContent : 'Engine';
}

function updateInfoText() {
    const whitePlayer = getWhitePlayer();
    const blackPlayer = getBlackPlayer();
    const infoDiv = document.getElementById('gameInfo');
    
    // Hide info text after first move to save space
    if (gameState.moveHistory.length > 0) {
        infoDiv.style.display = 'none';
        // Update player controls visibility
        updatePlayerControls();
        return;
    }
    
    infoDiv.style.display = 'block';
    const whiteName = getPlayerName(whitePlayer);
    const blackName = getPlayerName(blackPlayer);
    
    if (isHuman(whitePlayer) && isHuman(blackPlayer)) {
        infoDiv.textContent = 'Human vs Human - Make your move!';
    } else if (!isHuman(whitePlayer) && !isHuman(blackPlayer)) {
        infoDiv.textContent = `${whiteName} vs ${blackName} - Watch the AI battle!`;
    } else if (isHuman(whitePlayer)) {
        infoDiv.textContent = `You play as White vs ${blackName}. Make your move!`;
    } else {
        infoDiv.textContent = `You play as Black vs ${whiteName}. Make your move!`;
    }
    
    // Update player controls visibility
    updatePlayerControls();
}

function updatePlayerControls() {
    const whitePlayer = getWhitePlayer();
    const blackPlayer = getBlackPlayer();
    
    // Show/hide white player controls
    const whiteControls = document.getElementById('whitePlayerControls');
    if (whiteControls) {
        whiteControls.style.display = isHuman(whitePlayer) ? 'block' : 'none';
    }
    
    // Show/hide black player controls
    const blackControls = document.getElementById('blackPlayerControls');
    if (blackControls) {
        blackControls.style.display = isHuman(blackPlayer) ? 'block' : 'none';
    }
}

// -------------------------------------------------------------------------
// Chessboard Event Handlers
// -------------------------------------------------------------------------

function onDragStart(source, piece, position, orientation) {
    if (game.game_over()) return false;
    if (isComputerThinking) return false;

    const currentPlayer = getCurrentPlayer();
    if (!isHuman(currentPlayer)) return false;

    const turn = game.turn();
    if ((turn === 'w' && piece.search(/^b/) !== -1) ||
        (turn === 'b' && piece.search(/^w/) !== -1)) {
        return false;
    }
}

function onDrop(source, target) {
    // If navigating in history, return to end first
    if (gameState.isNavigating) {
        goToEnd();
        // After returning to end, the move will be processed normally
    }
    
    var move = game.move({
        from: source,
        to: target,
        promotion: 'q'
    });

    if (move === null) return 'snapback';

    highlightLastMove(source, target);
    
    // Track the move in UCI notation and add to history
    var uciMove = source + target;
    if (move.promotion) {
        uciMove += move.promotion;
    }
    gameState.moveHistory.push(uciMove);
    gameState.currentPosition = gameState.moveHistory.length;
    
    // Auto-start clock on first move if not already running
    if (!gameState.clockRunning && gameState.timeControl.initial > 0 && gameState.moveHistory.length === 1) {
        startClock();
    }
    
    // Add increment to the player who just moved (only if clock is running)
    if (gameState.clockRunning) {
        if (isWhiteMove) {
            gameState.whiteTimeMs += gameState.timeControl.increment * 1000;
        } else {
            gameState.blackTimeMs += gameState.timeControl.increment * 1000;
        }
        updateClockDisplay();
    }

    // Update display
    updateMoveHistoryDisplay();
    updateInfoText();
    updateOpeningDisplay();
    saveGameState();

    // Update analysis with new position
    updateAnalysisForCurrentPosition();

    window.setTimeout(checkForComputerMove, 250);
}

function onSnapEnd() {
    board.position(game.fen());
}

function onMoveEnd() {
    if (lastMoveSquares.from && lastMoveSquares.to) {
        $('#myBoard .square-' + lastMoveSquares.from).addClass('highlight-last-move');
        $('#myBoard .square-' + lastMoveSquares.to).addClass('highlight-last-move');
    }
}

// -------------------------------------------------------------------------
// Computer Move Logic
// -------------------------------------------------------------------------

async function makeComputerMove() {
    if (isComputerThinking) return;
    
    const currentPlayer = getCurrentPlayer();
    if (isHuman(currentPlayer)) return;

    if (game.game_over()) {
        console.log('Game over');
        return;
    }
    
    // Don't make computer moves if navigating in history
    if (gameState.isNavigating) {
        console.log('Navigating in history - computer waiting');
        return;
    }
    
    // Don't make computer moves if clock is paused (game is paused)
    // Exception: Allow first move even if clock not started yet
    if (!gameState.clockRunning && gameState.moveHistory.length > 0) {
        console.log('Game is paused - computer waiting');
        return;
    }

    isComputerThinking = true;
    var fen = game.fen();
    var isWhiteTurn = game.turn() === 'w';
    
    try {
        console.log('Making move with engine:', currentPlayer);
        
        // Build request with all game state
        const requestBody = {
            fen: fen,
            moves: gameState.moveHistory,
            enginePath: currentPlayer,
            moveTime: 0, // Use clock-based if available
            whiteTime: gameState.whiteTimeMs,
            blackTime: gameState.blackTimeMs,
            whiteIncrement: gameState.timeControl.increment * 1000,
            blackIncrement: gameState.timeControl.increment * 1000,
            engineOptions: {} // Can add ELO settings here later
        };
        
        const response = await fetch('/api/computer-move', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(requestBody)
        });

        const data = await response.json();
        console.log('Response:', data);
        
        if (data.error) {
            console.log('Game over or no legal moves:', data.error);
            isComputerThinking = false;
            return;
        }

        console.log('Applying move:', data.move, 'think time:', data.thinkTime, 'ms');
        
        // Update move history
        gameState.moveHistory.push(data.move);
        gameState.currentPosition = gameState.moveHistory.length;
        
        // Highlight move
        var moveStr = data.move;
        if (moveStr && moveStr.length >= 4) {
            var from = moveStr.substring(0, 2);
            var to = moveStr.substring(2, 4);
            highlightLastMove(from, to);
        }
        
        // Apply move to board
        game.load(data.fen);
        board.position(game.fen());
        
        // Auto-start clock on first move if not already running
        if (!gameState.clockRunning && gameState.timeControl.initial > 0 && gameState.moveHistory.length === 1) {
            startClock();
        }
        
        // Update clock (only if running)
        if (gameState.clockRunning) {
            // Deduct time spent thinking
            if (isWhiteTurn) {
                gameState.whiteTimeMs -= data.thinkTime;
                gameState.whiteTimeMs += gameState.timeControl.increment * 1000;
            } else {
                gameState.blackTimeMs -= data.thinkTime;
                gameState.blackTimeMs += gameState.timeControl.increment * 1000;
            }
            updateClockDisplay();
        }
        
        // Update display
        updateMoveHistoryDisplay();
        updateInfoText();
        updateOpeningDisplay();
        saveGameState();
        
        isComputerThinking = false;
        window.setTimeout(checkForComputerMove, 250);
        
    } catch (error) {
        console.error('Error getting computer move:', error);
        isComputerThinking = false;
    }
}

function checkForComputerMove() {
    const currentPlayer = getCurrentPlayer();
    console.log('checkForComputerMove - currentPlayer:', currentPlayer, 'isHuman:', isHuman(currentPlayer));
    if (!isHuman(currentPlayer) && !game.game_over()) {
        makeComputerMove();
    }
}

// Wrap makeComputerMove to update analysis
var originalMakeComputerMove = makeComputerMove;
makeComputerMove = async function() {
    await originalMakeComputerMove();
    updateAnalysisForCurrentPosition();
};

// -------------------------------------------------------------------------
// Move History Update (Client-side)
// -------------------------------------------------------------------------

function updateMoveHistoryDisplay() {
    if (!moveHistoryEditor) return;
    
    if (!gameState.moveHistory || gameState.moveHistory.length === 0) {
        moveHistoryEditor.setValue('');
        return;
    }
    
    // Convert UCI moves to SAN notation for PGN format
    // Always show all moves, but highlight the current position
    const sanMoves = [];
    const tempGame = new Chess();
    
    for (let i = 0; i < gameState.moveHistory.length; i++) {
        const uciMove = gameState.moveHistory[i];
        
        // Parse UCI move (e.g., "e2e4" or "e7e8q")
        const from = uciMove.substring(0, 2);
        const to = uciMove.substring(2, 4);
        const promotion = uciMove.length > 4 ? uciMove.substring(4) : undefined;
        
        // Get SAN notation
        const move = tempGame.move({
            from: from,
            to: to,
            promotion: promotion
        });
        
        if (move) {
            sanMoves.push(move.san);
        }
    }
    
    // Format as PGN (natural line wrapping)
    let pgn = '';
    for (let i = 0; i < sanMoves.length; i += 2) {
        const moveNum = Math.floor(i / 2) + 1;
        const whiteMove = sanMoves[i];
        const blackMove = sanMoves[i + 1] || '';
        
        pgn += moveNum + '.' + whiteMove;
        if (blackMove) {
            pgn += ' ' + blackMove;
        }
        
        // Add space between move pairs (let textarea handle line wrapping)
        if (i + 2 < sanMoves.length) {
            pgn += ' ';
        }
    }
    
    // Update CodeMirror content
    const currentValue = moveHistoryEditor.getValue();
    if (currentValue !== pgn) {
        moveHistoryEditor.setValue(pgn);
    }
    
    // Highlight current move position
    highlightCurrentMove();
    
    // Auto-scroll to current move (or bottom if at the end)
    if (gameState.currentPosition === gameState.moveHistory.length) {
        // At the end, scroll to bottom
        const lastLine = moveHistoryEditor.lineCount();
        moveHistoryEditor.scrollIntoView({line: lastLine, ch: 0});
    }
    
    // Update position indicator
    updatePositionIndicator();
}

function highlightCurrentMove() {
    if (!moveHistoryEditor) return;
    
    // Clear all current move markers
    moveHistoryEditor.getAllMarks().forEach(mark => mark.clear());
    
    if (gameState.moveHistory.length === 0) {
        return;
    }
    
    const text = moveHistoryEditor.getValue();
    
    // Find and highlight only the current move
    const movePattern = /(\d+)\.([^\s]+)(?:\s+([^\s]+))?/g;
    let match;
    
    while ((match = movePattern.exec(text)) !== null) {
        const moveNumber = parseInt(match[1]);
        const whiteMove = match[2];
        const blackMove = match[3];
        
        // Calculate move indices
        const whiteMoveIndex = (moveNumber - 1) * 2;
        const blackMoveIndex = whiteMoveIndex + 1;
        
        // Check white's move
        if (whiteMove && whiteMoveIndex === gameState.currentPosition - 1) {
            const whiteMoveStart = match.index + match[0].indexOf(whiteMove);
            const whiteMoveEnd = whiteMoveStart + whiteMove.length;
            const whiteStartPos = moveHistoryEditor.posFromIndex(whiteMoveStart);
            const whiteEndPos = moveHistoryEditor.posFromIndex(whiteMoveEnd);
            
            moveHistoryEditor.markText(whiteStartPos, whiteEndPos, {
                className: 'chess-current-move'
            });
            if (gameState.isNavigating) {
                moveHistoryEditor.scrollIntoView({from: whiteStartPos, to: whiteEndPos}, 100);
            }
            break;
        }
        
        // Check black's move
        if (blackMove && blackMoveIndex === gameState.currentPosition - 1) {
            const blackMoveStart = match.index + match[0].indexOf(blackMove);
            const blackMoveEnd = blackMoveStart + blackMove.length;
            const blackStartPos = moveHistoryEditor.posFromIndex(blackMoveStart);
            const blackEndPos = moveHistoryEditor.posFromIndex(blackMoveEnd);
            
            moveHistoryEditor.markText(blackStartPos, blackEndPos, {
                className: 'chess-current-move'
            });
            if (gameState.isNavigating) {
                moveHistoryEditor.scrollIntoView({from: blackStartPos, to: blackEndPos}, 100);
            }
            break;
        }
    }
}

// -------------------------------------------------------------------------
// Game History Navigation
// -------------------------------------------------------------------------

function updateAnalysisForCurrentPosition() {
    // Clear any existing arrows first
    board.clearArrow();
    
    // Update analysis engine with new position if active
    setTimeout(function() {
        if (analysisActive && analysisWs && analysisWs.readyState === WebSocket.OPEN) {
            analysisWs.send(JSON.stringify({
                action: 'update',
                fen: game.fen()
            }));
        }
    }, 100);
}

function updatePositionIndicator() {
    const indicator = document.getElementById('positionIndicator');
    if (indicator) {
        // Convert half-moves (plies) to full move notation
        // Standard chess notation shows whose turn it is to move
        if (gameState.currentPosition === 0) {
            indicator.textContent = 'Start';
        } else {
            const fullMoveNumber = Math.floor(gameState.currentPosition / 2) + 1;
            const isWhiteToMove = gameState.currentPosition % 2 === 0;
            
            if (isWhiteToMove) {
                // White to move (e.g., after 1...e5, show "2.")
                indicator.textContent = `${fullMoveNumber}.`;
            } else {
                // Black to move (e.g., after 1.e4, show "1...")
                indicator.textContent = `${fullMoveNumber}...`;
            }
        }
    }
    
    // Update button states
    const toStartBtn = document.querySelector('button[onclick="goToStart()"]');
    const stepBackBtn = document.querySelector('button[onclick="stepBackward()"]');
    const stepForwardBtn = document.querySelector('button[onclick="stepForward()"]');
    const toEndBtn = document.querySelector('button[onclick="goToEnd()"]');
    
    if (toStartBtn) toStartBtn.disabled = gameState.currentPosition === 0;
    if (stepBackBtn) stepBackBtn.disabled = gameState.currentPosition === 0;
    if (stepForwardBtn) stepForwardBtn.disabled = gameState.currentPosition >= gameState.moveHistory.length;
    if (toEndBtn) toEndBtn.disabled = gameState.currentPosition >= gameState.moveHistory.length;
}

function goToStart() {
    // Pause clock if running
    if (gameState.clockRunning && !gameState.isNavigating) {
        gameState.wasClockRunning = true;
        pauseClock();
    }
    
    gameState.isNavigating = true;
    gameState.currentPosition = 0;
    
    // Reset to starting position
    game.reset();
    board.position('start');
    lastMoveSquares = { from: null, to: null };
    clearLastMoveHighlight();
    
    updateMoveHistoryDisplay();
    updateInfoText();
    updateOpeningDisplay();
    updateAnalysisForCurrentPosition();
}

function stepBackward() {
    if (gameState.currentPosition === 0) return;
    
    // Pause clock if running
    if (gameState.clockRunning && !gameState.isNavigating) {
        gameState.wasClockRunning = true;
        pauseClock();
    }
    
    gameState.isNavigating = true;
    gameState.currentPosition--;
    
    // Rebuild game state up to current position
    game.reset();
    const tempGame = new Chess();
    
    for (let i = 0; i < gameState.currentPosition; i++) {
        const uciMove = gameState.moveHistory[i];
        const from = uciMove.substring(0, 2);
        const to = uciMove.substring(2, 4);
        const promotion = uciMove.length > 4 ? uciMove.substring(4) : undefined;
        
        tempGame.move({ from, to, promotion });
    }
    
    game.load(tempGame.fen());
    board.position(game.fen());
    
    // Highlight last move if not at start
    if (gameState.currentPosition > 0) {
        const lastMove = gameState.moveHistory[gameState.currentPosition - 1];
        const from = lastMove.substring(0, 2);
        const to = lastMove.substring(2, 4);
        highlightLastMove(from, to);
    } else {
        lastMoveSquares = { from: null, to: null };
        clearLastMoveHighlight();
    }
    
    updateMoveHistoryDisplay();
    updateInfoText();
    updateOpeningDisplay();
    updateAnalysisForCurrentPosition();
}

function stepForward() {
    if (gameState.currentPosition >= gameState.moveHistory.length) return;
    
    // Pause clock if running
    if (gameState.clockRunning && !gameState.isNavigating) {
        gameState.wasClockRunning = true;
        pauseClock();
    }
    
    gameState.isNavigating = true;
    
    // Apply the next move
    const uciMove = gameState.moveHistory[gameState.currentPosition];
    const from = uciMove.substring(0, 2);
    const to = uciMove.substring(2, 4);
    const promotion = uciMove.length > 4 ? uciMove.substring(4) : undefined;
    
    game.move({ from, to, promotion });
    board.position(game.fen());
    highlightLastMove(from, to);
    
    gameState.currentPosition++;
    
    updateMoveHistoryDisplay();
    updateInfoText();
    updateOpeningDisplay();
    updateAnalysisForCurrentPosition();
}

function goToEnd() {
    // If at the end and clock was running before navigation, resume it
    const shouldResumeClock = gameState.wasClockRunning && gameState.currentPosition < gameState.moveHistory.length;
    
    gameState.currentPosition = gameState.moveHistory.length;
    
    // Rebuild game state to the end
    game.reset();
    const tempGame = new Chess();
    
    for (let i = 0; i < gameState.moveHistory.length; i++) {
        const uciMove = gameState.moveHistory[i];
        const from = uciMove.substring(0, 2);
        const to = uciMove.substring(2, 4);
        const promotion = uciMove.length > 4 ? uciMove.substring(4) : undefined;
        
        tempGame.move({ from, to, promotion });
    }
    
    game.load(tempGame.fen());
    board.position(game.fen());
    
    // Highlight last move
    if (gameState.moveHistory.length > 0) {
        const lastMove = gameState.moveHistory[gameState.moveHistory.length - 1];
        const from = lastMove.substring(0, 2);
        const to = lastMove.substring(2, 4);
        highlightLastMove(from, to);
    }
    
    // We're back at the end, no longer navigating
    gameState.isNavigating = false;
    
    // Resume clock if it was running before navigation
    if (shouldResumeClock) {
        startClock();
        gameState.wasClockRunning = false;
    }
    
    updateMoveHistoryDisplay();
    updateInfoText();
    updateOpeningDisplay();
    updateAnalysisForCurrentPosition();
    
    // Check if computer should move
    window.setTimeout(checkForComputerMove, 250);
}

// Count total games in PGN file (fast, doesn't parse content)
function countPGNGames(pgnText) {
    let count = 0;
    let inMoves = false;
    const lines = pgnText.split('\n');
    
    for (let i = 0; i < lines.length; i++) {
        const line = lines[i].trim();
        
        if (line.startsWith('[') && line.endsWith(']')) {
            if (inMoves) {
                // New game started
                count++;
                inMoves = false;
            }
        } else if (line && !line.startsWith('[')) {
            inMoves = true;
        }
    }
    
    // Count the last game
    if (inMoves) {
        count++;
    }
    
    return count;
}

// Parse PGN file and extract games (up to maxGames limit)
function parsePGNGames(pgnText, maxGames = Infinity) {
    const games = [];
    
    // Split by empty lines to find game boundaries
    // Each game starts with headers and ends with moves
    const lines = pgnText.split('\n');
    let currentGame = { headers: {}, moves: '' };
    let inHeaders = false;
    let inMoves = false;
    
    for (let i = 0; i < lines.length; i++) {
        const line = lines[i].trim();
        
        // Check if this is a header line
        if (line.startsWith('[') && line.endsWith(']')) {
            if (!inHeaders && inMoves) {
                // We've hit a new game, save the previous one
                if (currentGame.moves.trim()) {
                    games.push(currentGame);
                    
                    // Stop parsing if we've reached the limit
                    if (games.length >= maxGames) {
                        return games;
                    }
                }
                currentGame = { headers: {}, moves: '' };
                inMoves = false;
            }
            inHeaders = true;
            
            // Parse header
            const match = line.match(/\[(\w+)\s+"([^"]*)"\]/);
            if (match) {
                currentGame.headers[match[1]] = match[2];
            }
        } else if (line && !line.startsWith('[')) {
            // This is a move line
            inHeaders = false;
            inMoves = true;
            currentGame.moves += line + ' ';
        } else if (!line && inMoves) {
            // Empty line after moves might indicate end of game
            continue;
        }
    }
    
    // Don't forget the last game (if under limit)
    if (currentGame.moves.trim() && games.length < maxGames) {
        games.push(currentGame);
    }
    
    return games;
}

// Extract game metadata for display
function extractGameMetadata(game, index) {
    const headers = game.headers;
    
    // Extract first few moves for opening preview and API lookup
    const movesOnly = game.moves
        .replace(/\{[^}]*\}/g, '')  // Remove comments
        .replace(/\([^)]*\)/g, '')  // Remove variations
        .replace(/[!?]+/g, '')      // Remove annotations
        .replace(/\s*(1-0|0-1|1\/2-1\/2|\*)\s*$/, ''); // Remove result
    
    // Split by move numbers and extract moves
    // Handles both "1. e4 e5" and "1.e4 e5" formats
    const moves = [];
    const tokens = movesOnly.split(/\d+\./);  // Split by move numbers
    
    for (let i = 1; i < tokens.length && moves.length < 30; i++) {
        const moveText = tokens[i].trim();
        if (!moveText) continue;
        
        // Split the move text into individual moves (white and black)
        const moveParts = moveText.split(/\s+/);
        for (const move of moveParts) {
            if (move && !move.match(/^[!?]+$/) && !move.match(/^\d/)) {
                moves.push(move);
                if (moves.length >= 30) break;
            }
        }
    }
    
    // Try to get opening from headers first (only human-readable names, not ECO codes)
    let opening = headers.Opening || null;
    
    // Store moves for API lookup (API will find the deepest match)
    const movesForLookup = moves.slice(0, 15);
    
    return {
        index: index,
        white: headers.White || 'Unknown',
        black: headers.Black || 'Unknown',
        result: headers.Result || '*',
        date: headers.Date || '????.??.??',
        event: headers.Event || 'Unknown',
        opening: opening,
        moves: movesForLookup,
        needsLookup: !opening,  // Lookup if no Opening header found
        fullPGN: game
    };
}

// Show game selector modal
async function showPGNGameSelector(games, wasLimited, totalGames) {
    const modal = document.getElementById('pgnGameSelectorModal');
    const gameCount = document.getElementById('gameCount');
    const tableBody = document.getElementById('gameTableBody');
    
    // Clear any analysis arrows
    board.clearArrow();
    
    // Update game count with limit info if applicable
    if (wasLimited) {
        gameCount.innerHTML = `<strong>${games.length}</strong> (limited from ${totalGames} total)`;
    } else {
        gameCount.textContent = games.length;
    }
    
    // Clear existing rows
    tableBody.innerHTML = '';
    
    // Populate table with games (initial display)
    games.forEach((gameData, index) => {
        const row = document.createElement('tr');
        row.id = 'game-row-' + index;
        row.onclick = function() {
            selectPGNGame(gameData.fullPGN);
        };
        
        const openingDisplay = gameData.opening || '<span style="color: #999;">Loading...</span>';
        
        row.innerHTML = `
            <td>${index + 1}</td>
            <td>${gameData.white}</td>
            <td>${gameData.black}</td>
            <td>${gameData.result}</td>
            <td>${gameData.date}</td>
            <td>${gameData.event}</td>
            <td id="opening-cell-${index}" style="font-family: 'Courier New', monospace; font-size: 12px;">${openingDisplay}</td>
        `;
        
        tableBody.appendChild(row);
    });
    
    // Show modal
    modal.style.display = 'flex';
    
    // Check if we need to lookup openings
    const gamesToLookup = games.filter(g => g.needsLookup);
    
    if (gamesToLookup.length > 0) {
        // Start lookup after a short delay (2 seconds) to avoid showing progress bar for fast loads
        const startTime = Date.now();
        await lookupOpeningNames(games, startTime);
    }
}

// Lookup opening names via API
async function lookupOpeningNames(games, startTime) {
    const loadingBar = document.getElementById('openingLoadingBar');
    const loadingProgress = document.getElementById('loadingProgress');
    const loadingBarFill = document.getElementById('loadingBarFill');
    
    const gamesToLookup = games.filter(g => g.needsLookup);
    const totalToLookup = gamesToLookup.length;
    let completed = 0;
    
    // Show progress bar immediately
    loadingBar.style.display = 'block';
    loadingProgress.textContent = `${completed}/${totalToLookup}`;
    
    // Lookup openings in batches to avoid overwhelming the server
    const batchSize = 5;
    for (let i = 0; i < gamesToLookup.length; i += batchSize) {
        const batch = gamesToLookup.slice(i, i + batchSize);
        
        await Promise.all(batch.map(async (gameData) => {
            try {
                const response = await fetch('/api/opening', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        moves: gameData.moves
                    })
                });
                
                if (response.ok) {
                    const opening = await response.json();
                    
                    if (opening && opening.name && opening.name.trim() !== '') {
                        gameData.opening = opening.name;
                        
                        // Update the cell
                        const cell = document.getElementById('opening-cell-' + gameData.index);
                        if (cell) {
                            cell.textContent = opening.name;
                        }
                    } else {
                        // No opening found, show first few moves
                        gameData.opening = gameData.moves.slice(0, 6).join(' ') || '-';
                        const cell = document.getElementById('opening-cell-' + gameData.index);
                        if (cell) {
                            cell.textContent = gameData.opening;
                        }
                    }
                }
            } catch (error) {
                console.error('Error fetching opening for game', gameData.index, error);
                // Show moves as fallback
                gameData.opening = gameData.moves.slice(0, 6).join(' ') || '-';
                const cell = document.getElementById('opening-cell-' + gameData.index);
                if (cell) {
                    cell.textContent = gameData.opening;
                }
            }
            
            completed++;
            
            // Update progress
            loadingProgress.textContent = `${completed}/${totalToLookup}`;
            const percentage = (completed / totalToLookup) * 100;
            loadingBarFill.style.width = percentage + '%';
        }));
    }
    
    // Hide progress bar after completion
    if (loadingBar.style.display === 'block') {
        setTimeout(() => {
            loadingBar.style.display = 'none';
        }, 500);
    }
}

// Close game selector modal
function closePGNGameSelector() {
    const modal = document.getElementById('pgnGameSelectorModal');
    modal.style.display = 'none';
}

// Select a game from the modal
function selectPGNGame(game) {
    closePGNGameSelector();
    
    // Combine headers and moves for loading
    let pgnText = '';
    for (const [key, value] of Object.entries(game.headers)) {
        pgnText += `[${key} "${value}"]\n`;
    }
    pgnText += '\n' + game.moves;
    
    loadPGNFromText(pgnText);
}

// Load PGN from file
function loadPGNFile() {
    const fileInput = document.getElementById('pgnFileInput');
    
    // Set up the file input handler
    fileInput.onchange = function(e) {
        const file = e.target.files[0];
        if (!file) {
            return;
        }
        
        // Stop any running analysis and clear arrows
        if (analysisActive) {
            stopAnalysis();
        }
        board.clearArrow();
        
        const reader = new FileReader();
        reader.onload = function(event) {
            const text = event.target.result;
            
            // First, count total games (fast operation)
            const totalGames = countPGNGames(text);
            
            if (totalGames === 0) {
                alert('No valid games found in the PGN file!');
                return;
            } else if (totalGames === 1) {
                // Single game, load it directly
                loadPGNFromText(text);
            } else {
                // Limit to first 1000 games for memory protection
                const MAX_GAMES = 1000;
                const wasLimited = totalGames > MAX_GAMES;
                
                // Show warning if file was truncated
                if (wasLimited) {
                    alert(`This PGN file contains ${totalGames} games.\n\nFor memory protection, only the first ${MAX_GAMES} games will be loaded.`);
                }
                
                // Parse only up to the limit (memory efficient)
                const games = parsePGNGames(text, MAX_GAMES);
                
                // Multiple games, show selector
                const gamesMetadata = games.map((game, index) => 
                    extractGameMetadata(game, index)
                );
                showPGNGameSelector(gamesMetadata, wasLimited, totalGames);
            }
        };
        
        reader.onerror = function() {
            alert('Failed to read file');
        };
        
        reader.readAsText(file);
        
        // Reset the input so the same file can be loaded again
        fileInput.value = '';
    };
    
    // Trigger the file picker
    fileInput.click();
}

// Save PGN to file
function savePGNFile() {
    if (!moveHistoryEditor) return;
    
    const pgnContent = moveHistoryEditor.getValue();
    
    if (!pgnContent) {
        alert('No moves to save!');
        return;
    }
    
    // Get player names
    const whitePlayerValue = getWhitePlayer();
    const blackPlayerValue = getBlackPlayer();
    const whiteName = getPlayerName(whitePlayerValue);
    const blackName = getPlayerName(blackPlayerValue);
    
    // Create PGN content with headers
    const date = new Date();
    const dateStr = date.toISOString().split('T')[0].replace(/-/g, '.');
    
    let pgnWithHeaders = '[Event "Casual Game"]\n';
    pgnWithHeaders += '[Site "go-chess"]\n';
    pgnWithHeaders += '[Date "' + dateStr + '"]\n';
    pgnWithHeaders += '[White "' + whiteName + '"]\n';
    pgnWithHeaders += '[Black "' + blackName + '"]\n';
    pgnWithHeaders += '[Result "*"]\n';
    pgnWithHeaders += '\n';
    pgnWithHeaders += pgnContent;
    pgnWithHeaders += ' *\n';
    
    // Create filename with player names
    // Sanitize names for filename (remove special characters)
    const sanitizeFilename = function(name) {
        return name.replace(/[^a-zA-Z0-9_-]/g, '_');
    };
    const whiteFilename = sanitizeFilename(whiteName);
    const blackFilename = sanitizeFilename(blackName);
    const filename = dateStr + '_' + whiteFilename + '_vs_' + blackFilename + '.pgn';
    
    // Create a blob and download link
    const blob = new Blob([pgnWithHeaders], { type: 'application/x-chess-pgn' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    
    // Trigger download
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    
    // Visual feedback
    const btn = event.target;
    const originalText = btn.textContent;
    btn.textContent = '✓ Saved';
    setTimeout(function() {
        btn.textContent = originalText;
    }, 1500);
}

// Load PGN from text string
function loadPGNFromText(text) {
    if (!text || text.trim() === '') {
        alert('No PGN text provided!');
        return;
    }
    
    try {
        // Parse PGN moves (extract just the moves, ignore headers and comments)
        const pgnText = text.trim();
        
        // Reset player dropdowns first to clear any previous custom names
        resetPlayerDropdowns();
        
        // Extract player names from PGN headers
        const whiteMatch = pgnText.match(/\[White\s+"([^"]+)"\]/);
        const blackMatch = pgnText.match(/\[Black\s+"([^"]+)"\]/);
        const whiteName = whiteMatch ? whiteMatch[1] : null;
        const blackName = blackMatch ? blackMatch[1] : null;
        
        // Add player names to dropdowns if found
        if (whiteName) {
            addPlayerToDropdown('whitePlayer', whiteName);
        }
        if (blackName) {
            addPlayerToDropdown('blackPlayer', blackName);
        }
        
        // Update UI to reflect the new player selections
        updatePlayerControls();
        updateInfoText();
        
        // Remove PGN headers (lines starting with [)
        let movesOnly = pgnText.split('\n')
            .filter(line => !line.startsWith('['))
            .join(' ')
            .trim();
        
        // Remove comments in braces and parentheses
        movesOnly = movesOnly.replace(/\{[^}]*\}/g, '');
        movesOnly = movesOnly.replace(/\([^)]*\)/g, '');
        
        // Remove result markers
        movesOnly = movesOnly.replace(/\s*(1-0|0-1|1\/2-1\/2|\*)\s*$/, '');
        
        // Remove annotations like !, ?, !!, ??, !?, ?!
        movesOnly = movesOnly.replace(/[!?]+/g, '');
        
        // Extract moves (format: 1. e4 e5 2. Nf3 Nc6...)
        const movePattern = /\d+\.\s*([^\s]+)(?:\s+([^\s]+))?/g;
        const moves = [];
        let match;
        
        while ((match = movePattern.exec(movesOnly)) !== null) {
            if (match[1]) moves.push(match[1]);
            if (match[2]) moves.push(match[2]);
        }
        
        if (moves.length === 0) {
            alert('No valid moves found in clipboard!');
            return;
        }
        
        // Reset game and apply moves
        game.reset();
        board.position('start');
        gameState.moveHistory = [];
        clearLastMoveHighlight();
        
        // Clear any analysis arrows
        board.clearArrow();
        
        // Apply each move
        for (let i = 0; i < moves.length; i++) {
            const san = moves[i];
            
            try {
                const move = game.move(san);
                if (!move) {
                    alert('Invalid move at position ' + (i + 1) + ': ' + san);
                    break;
                }
                
                // Convert to UCI format for our history
                const uciMove = move.from + move.to + (move.promotion || '');
                gameState.moveHistory.push(uciMove);
                
                // Highlight last move
                if (i === moves.length - 1) {
                    highlightLastMove(move.from, move.to);
                }
            } catch (err) {
                alert('Error applying move ' + (i + 1) + ': ' + san + '\n' + err.message);
                break;
            }
        }
        
        // Update board and displays
        board.position(game.fen());
        gameState.currentPosition = gameState.moveHistory.length;
        gameState.isNavigating = false;
        updateMoveHistoryDisplay();
        updateOpeningDisplay();
        updateInfoText();
        updateAnalysisForCurrentPosition();
        saveGameState();
        
        // Visual feedback
        const btn = event.target;
        const originalText = btn.textContent;
        btn.textContent = '✓ Loaded';
        setTimeout(function() {
            btn.textContent = originalText;
        }, 1500);
        
    } catch (err) {
        console.error('Failed to parse PGN:', err);
        alert('Failed to parse PGN. Please check the format and try again.');
    }
}


// -------------------------------------------------------------------------
// Opening Display
// -------------------------------------------------------------------------

async function updateOpeningDisplay() {
    // Get move history in SAN notation
    const sanMoves = [];
    const tempGame = new Chess();
    
    // Replay the game to get SAN notation
    // When navigating, only use moves up to currentPosition
    const movesToShow = gameState.isNavigating ? gameState.currentPosition : gameState.moveHistory.length;
    
    for (let i = 0; i < movesToShow; i++) {
        const uciMove = gameState.moveHistory[i];
        
        // Parse UCI move (e.g., "e2e4" or "e7e8q")
        const from = uciMove.substring(0, 2);
        const to = uciMove.substring(2, 4);
        const promotion = uciMove.length > 4 ? uciMove.substring(4) : undefined;
        
        // Get SAN notation before making the move
        const move = tempGame.move({
            from: from,
            to: to,
            promotion: promotion
        });
        
        if (move) {
            sanMoves.push(move.san);
        }
    }
    
    // Don't show opening info if no moves yet
    if (sanMoves.length === 0) {
        document.getElementById('openingDisplay').style.display = 'none';
        return;
    }
    
    try {
        // Call the opening API
        const response = await fetch('/api/opening', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                moves: sanMoves
            })
        });
        
        if (!response.ok) {
            console.error('Failed to fetch opening info');
            return;
        }
        
        const opening = await response.json();
        
        // Update the display (single line, no ECO)
        if (opening && opening.name) {
            document.getElementById('openingText').textContent = opening.name;
            document.getElementById('openingDisplay').style.display = 'block';
        } else {
            document.getElementById('openingDisplay').style.display = 'none';
        }
    } catch (error) {
        console.error('Error fetching opening info:', error);
    }
}

// -------------------------------------------------------------------------
// Game Control Functions
// -------------------------------------------------------------------------

function newGame() {
    // Reset game state
    game.reset();
    board.position('start');
    gameState.moveHistory = [];
    gameState.currentPosition = 0;
    gameState.isNavigating = false;
    gameState.wasClockRunning = false;
    gameState.whiteTimeMs = gameState.timeControl.initial * 60 * 1000;
    gameState.blackTimeMs = gameState.timeControl.initial * 60 * 1000;
    gameState.clockRunning = false;
    gameState.gameStartTime = Date.now();
    
    // Stop clock if running
    if (gameState.clockInterval) {
        clearInterval(gameState.clockInterval);
        gameState.clockInterval = null;
    }
    
    // Clear highlights
    clearLastMoveHighlight();
    
    // Reset player dropdowns to default (Human vs Human)
    resetPlayerDropdowns();
    
    // Update displays
    updateMoveHistoryDisplay();
    updateClockDisplay();
    updateInfoText();
    updateOpeningDisplay();
    
    // Clear saved state
    clearGameState();
    
    // Update start/pause button
    updateStartPauseButton();
    
    // Check if computer should move first
    window.setTimeout(checkForComputerMove, 500);
}

// -------------------------------------------------------------------------
// Clock Control Functions (FIDE Rules)
// -------------------------------------------------------------------------

function toggleClock() {
    if (gameState.clockRunning) {
        pauseClock();
    } else {
        startClock();
    }
}

function pauseClock() {
    stopClock();
    updateStartPauseButton();
    saveGameState();
}

function updateStartPauseButton() {
    const btn = document.getElementById('startPauseBtn');
    if (gameState.clockRunning) {
        btn.textContent = '⏸️ Pause Game';
        btn.style.background = '#ff9800';
    } else {
        btn.textContent = '▶️ Start Game';
        btn.style.background = '#4caf50';
    }
}

// -------------------------------------------------------------------------
// Game Result Functions
// -------------------------------------------------------------------------

function resignWhite() {
    if (confirm('White resigns. Black wins!')) {
        stopClock();
        alert('Game Over: Black wins by resignation');
        // Optionally start a new game or disable moves
    }
}

function resignBlack() {
    if (confirm('Black resigns. White wins!')) {
        stopClock();
        alert('Game Over: White wins by resignation');
        // Optionally start a new game or disable moves
    }
}

function offerDraw() {
    if (confirm('Offer a draw?')) {
        stopClock();
        alert('Game Over: Draw by agreement');
        // Optionally start a new game
    }
}

// -------------------------------------------------------------------------
// Player Selection Management
// -------------------------------------------------------------------------

function restorePlayerSelections() {
    // DISABLED: Always start with Human vs Human (default dropdown values)
    return;
    
    /* DISABLED CODE:
    const savedWhite = localStorage.getItem('whitePlayer');
    const savedBlack = localStorage.getItem('blackPlayer');
    
    const whiteSelect = document.getElementById('whitePlayer');
    const blackSelect = document.getElementById('blackPlayer');
    
    if (savedWhite && Array.from(whiteSelect.options).some(opt => opt.value === savedWhite)) {
        whiteSelect.value = savedWhite;
    }
    if (savedBlack && Array.from(blackSelect.options).some(opt => opt.value === savedBlack)) {
        blackSelect.value = savedBlack;
    } else {
        const firstEngine = Array.from(blackSelect.options).find(opt => opt.value !== 'human');
        if (firstEngine) {
            blackSelect.value = firstEngine.value;
        }
    }
    */
}

function savePlayerSelections() {
    // DISABLED: Player selection persistence disabled
    updatePlayerControls();
    
    /* DISABLED CODE:
    localStorage.setItem('whitePlayer', document.getElementById('whitePlayer').value);
    localStorage.setItem('blackPlayer', document.getElementById('blackPlayer').value);
    */
}

function flipBoard() {
    board.flip();
    
    const whitePlayer = document.getElementById('whitePlayer').value;
    const blackPlayer = document.getElementById('blackPlayer').value;
    
    document.getElementById('whitePlayer').value = blackPlayer;
    document.getElementById('blackPlayer').value = whitePlayer;
    
    savePlayerSelections();
    updateInfoText();
    window.setTimeout(checkForComputerMove, 250);
}

// -------------------------------------------------------------------------
// Live Analysis
// -------------------------------------------------------------------------

function toggleAnalysis() {
    if (analysisActive) {
        stopAnalysis();
    } else {
        startAnalysis();
    }
}

function startAnalysis() {
    var protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    var wsUrl = protocol + '//' + window.location.host + '/api/analysis';
    
    // Get selected engine from dropdown
    var engineSelect = document.getElementById('analysisEngine');
    var enginePath = engineSelect ? engineSelect.value : 'stockfish';
    
    analysisWs = new WebSocket(wsUrl);
    
    analysisWs.onopen = function() {
        console.log('Analysis WebSocket connected');
        analysisActive = true;
        document.getElementById('analysisToggle').textContent = '⏹️ Stop Analysis';
        
        analysisWs.send(JSON.stringify({
            action: 'start',
            fen: game.fen(),
            enginePath: enginePath
        }));
    };
    
    analysisWs.onmessage = function(event) {
        var data = JSON.parse(event.data);
        
        if (data.error) {
            console.error('Analysis error:', data.error);
            return;
        }
        
        // Update depth display
        if (data.depth !== undefined) {
            document.getElementById('analysisDepth').textContent = 'Depth: ' + data.depth;
        }
        
        // Draw arrows for principal variation
        if (data.pv && data.pv.length > 0) {
            // Check if PV display is enabled
            var showPV = document.getElementById('showPVArrows').checked;
            
            // Create a temporary game instance to track positions
            var tempGame = new Chess(game.fen());
            
            // Get the starting full move number from FEN (6th field)
            var fenParts = tempGame.fen().split(' ');
            var startingMoveNumber = parseInt(fenParts[5]) || 1;
            
            // Limit to 3 moves if PV is enabled, otherwise just 1 (best move only)
            var maxMoves = showPV ? Math.min(3, data.pv.length) : 1;
            
            for (var i = 0; i < maxMoves; i++) {
                var move = data.pv[i];
                
                // Parse UCI move format (e.g., "e2e4" or "e7e8q" for promotion)
                if (move.length < 4) continue;
                
                var from = move.substring(0, 2);
                var to = move.substring(2, 4);
                var promotion = move.length > 4 ? move.substring(4, 5) : undefined;
                
                // Verify the piece exists at the from square
                var piece = tempGame.get(from);
                if (!piece) continue;
                
                // Use different colors based on whose turn it is in the temp game
                var arrowColor = tempGame.turn() === 'w' ? '#3296FF' : '#FF6B6B';
                
                // Calculate opacity: first arrow bright, subsequent ones dimmer
                var opacity = 1.0 - (i * 0.2);
                
                // Calculate actual chess move number from current FEN
                var currentFenParts = tempGame.fen().split(' ');
                var currentMoveNumber = parseInt(currentFenParts[5]) || 1;
                var isBlackMove = tempGame.turn() === 'b';
                
                // Only show score label on the first arrow
                var scoreLabel = '';
                if (i === 0) {
                    if (data.scoreType === 'cp' && data.score !== undefined) {
                        var score = (data.score / 100).toFixed(2);
                        scoreLabel = (data.score >= 0 ? '+' : '') + score;
                    } else if (data.scoreType === 'mate' && data.score !== undefined) {
                        scoreLabel = 'M' + Math.abs(data.score);
                    }
                }
                
                // Add move number to all arrows when PV is enabled
                var moveNumberLabel = null;
                if (showPV) {
                    // Format: "18" for white, "18..." for black
                    moveNumberLabel = isBlackMove ? currentMoveNumber + '...' : currentMoveNumber.toString();
                }
                
                // Draw arrow (clear previous only on first arrow)
                var clearPrevious = (i === 0);
                board.drawArrow(from, to, arrowColor, scoreLabel, opacity, clearPrevious, moveNumberLabel);
                
                // Apply move to temp game for next iteration
                try {
                    tempGame.move({
                        from: from,
                        to: to,
                        promotion: promotion || 'q'
                    });
                } catch (e) {
                    // Invalid move, stop drawing further arrows
                    break;
                }
            }
        }
    };
    
    analysisWs.onerror = function(error) {
        console.error('Analysis WebSocket error:', error);
        stopAnalysis();
    };
    
    analysisWs.onclose = function() {
        console.log('Analysis WebSocket closed');
        if (analysisActive) {
            stopAnalysis();
        }
    };
}

function stopAnalysis() {
    if (analysisWs) {
        analysisWs.send(JSON.stringify({ action: 'stop' }));
        analysisWs.close();
        analysisWs = null;
    }
    
    analysisActive = false;
    document.getElementById('analysisToggle').textContent = '🔍 Start Analysis';
    document.getElementById('analysisDepth').textContent = 'Depth: -';
    board.clearArrow();
}

// -------------------------------------------------------------------------
// Chess Clock Functions
// -------------------------------------------------------------------------

function formatTime(milliseconds) {
    if (milliseconds < 0) milliseconds = 0;
    
    var totalSeconds = Math.floor(milliseconds / 1000);
    var minutes = Math.floor(totalSeconds / 60);
    var seconds = totalSeconds % 60;
    
    if (minutes >= 60) {
        var hours = Math.floor(minutes / 60);
        minutes = minutes % 60;
        return hours + ':' + pad(minutes) + ':' + pad(seconds);
    }
    
    return minutes + ':' + pad(seconds);
}

function pad(num) {
    return num < 10 ? '0' + num : num;
}

function updateClockDisplay() {
    document.getElementById('whiteClockTime').textContent = formatTime(gameState.whiteTimeMs);
    document.getElementById('blackClockTime').textContent = formatTime(gameState.blackTimeMs);
    
    // Update active clock styling
    var whiteClock = document.getElementById('whiteClock');
    var blackClock = document.getElementById('blackClock');
    
    if (gameState.clockRunning) {
        if (game.turn() === 'w') {
            whiteClock.classList.add('active');
            blackClock.classList.remove('active');
        } else {
            blackClock.classList.add('active');
            whiteClock.classList.remove('active');
        }
    } else {
        whiteClock.classList.remove('active');
        blackClock.classList.remove('active');
    }
    
    // Add time warnings (skip if unlimited time mode)
    whiteClock.classList.remove('time-low', 'time-critical');
    blackClock.classList.remove('time-low', 'time-critical');
    
    if (gameState.timeControl.initial > 0) {
        if (gameState.whiteTimeMs < 60000 && gameState.whiteTimeMs > 10000) {
            whiteClock.classList.add('time-low');
        } else if (gameState.whiteTimeMs <= 10000) {
            whiteClock.classList.add('time-critical');
        }
        
        if (gameState.blackTimeMs < 60000 && gameState.blackTimeMs > 10000) {
            blackClock.classList.add('time-low');
        } else if (gameState.blackTimeMs <= 10000) {
            blackClock.classList.add('time-critical');
        }
    }
    
    // Check for time out (skip if unlimited time mode)
    if (gameState.timeControl.initial > 0) {
        if (gameState.whiteTimeMs <= 0) {
            stopClock();
            alert('Time out! Black wins!');
        } else if (gameState.blackTimeMs <= 0) {
            stopClock();
            alert('Time out! White wins!');
        }
    }
}

function startClockInterval() {
    if (gameState.clockInterval) return; // Already running
    
    gameState.lastClockUpdate = Date.now();
    
    // Start the interval
    gameState.clockInterval = setInterval(function() {
        var now = Date.now();
        var elapsed = now - gameState.lastClockUpdate;
        gameState.lastClockUpdate = now;
        
        if (game.turn() === 'w') {
            gameState.whiteTimeMs -= elapsed;
        } else {
            gameState.blackTimeMs -= elapsed;
        }
        
        updateClockDisplay();
        saveGameState();
    }, 100); // Update every 100ms for smooth display
}

function startClock() {
    if (gameState.clockRunning) return; // Already running
    
    gameState.clockRunning = true;
    startClockInterval();
    updateClockDisplay();
    updateStartPauseButton();
    saveGameState();
    
    // Resume computer play if it's a computer's turn
    window.setTimeout(checkForComputerMove, 250);
}

function stopClock() {
    if (!gameState.clockRunning) return;
    
    gameState.clockRunning = false;
    if (gameState.clockInterval) {
        clearInterval(gameState.clockInterval);
        gameState.clockInterval = null;
    }
    updateClockDisplay();
    saveGameState();
}

function setTimeControl(initialMinutes, incrementSeconds) {
    gameState.timeControl = { initial: initialMinutes, increment: incrementSeconds };
    gameState.whiteTimeMs = initialMinutes * 60 * 1000;
    gameState.blackTimeMs = initialMinutes * 60 * 1000;
    stopClock();
    updateClockDisplay();
    saveGameState();
}

// -------------------------------------------------------------------------
// Initialization
// -------------------------------------------------------------------------

$(document).ready(function() {
    // Initialize chessboard
    var config = {
        draggable: true,
        position: 'start',
        onDragStart: onDragStart,
        onDrop: onDrop,
        onSnapEnd: onSnapEnd,
        onMoveEnd: onMoveEnd,
        pieceTheme: '/assets/images/pieces/{piece}.png'
    };
    
    board = Chessboard('myBoard', config);
    
    // Make board responsive
    $(window).resize(function() {
        board.resize();
    });
    
    // Player selection event handlers
    document.getElementById('whitePlayer').addEventListener('change', function() {
        savePlayerSelections();
        updateInfoText();
        window.setTimeout(checkForComputerMove, 250);
    });

    document.getElementById('blackPlayer').addEventListener('change', function() {
        savePlayerSelections();
        updateInfoText();
        window.setTimeout(checkForComputerMove, 250);
    });
    
    // Analysis engine selector - restart analysis if active
    document.getElementById('analysisEngine').addEventListener('change', function() {
        if (analysisActive) {
            // Stop current analysis
            stopAnalysis();
            // Restart with new engine after a brief delay
            setTimeout(function() {
                startAnalysis();
            }, 100);
        }
    });
    
    // Time control selector
    document.getElementById('timeControl').addEventListener('change', function() {
        var value = this.value;
        var parts = value.split('+');
        var initialMinutes = parseInt(parts[0]);
        var incrementSeconds = parseInt(parts[1]);
        
        setTimeControl(initialMinutes, incrementSeconds);
        stopClock();
        
        // DISABLED: Time control persistence disabled
        // localStorage.setItem('timeControl', value);
    });
    
    // Set default to Unlimited (0+0)
    var select = document.getElementById('timeControl');
    select.value = '0+0';
    setTimeControl(0, 0);
    
    /* DISABLED CODE: Restore time control preference
    var savedTimeControl = localStorage.getItem('timeControl');
    if (savedTimeControl) {
        var select = document.getElementById('timeControl');
        if (Array.from(select.options).some(opt => opt.value === savedTimeControl)) {
            select.value = savedTimeControl;
            var parts = savedTimeControl.split('+');
            setTimeControl(parseInt(parts[0]), parseInt(parts[1]));
        }
    }
    */
    
    // Initialize
    restorePlayerSelections();
    
    // Try to load saved game state
    var loaded = loadGameState();
    if (!loaded) {
        // No saved state, start fresh
        updateMoveHistoryDisplay();
        updateClockDisplay();
    }
    
    updateInfoText();
    updateStartPauseButton();
    updatePositionIndicator();
    
    // Initialize CodeMirror for move history
    const moveHistoryTextArea = document.getElementById('moveHistoryText');
    if (moveHistoryTextArea) {
        moveHistoryEditor = CodeMirror.fromTextArea(moveHistoryTextArea, {
            mode: 'chess',
            lineNumbers: false,
            lineWrapping: true,
            readOnly: true,
            theme: 'default',
            viewportMargin: Infinity
        });
        
        // Add paste event handler
        moveHistoryEditor.on('paste', function(cm, e) {
            e.preventDefault();
            const pastedText = (e.clipboardData || window.clipboardData).getData('text');
            
            // Load the pasted PGN
            loadPGNFromText(pastedText);
        });
        
        // Prevent manual editing - restore from game state
        moveHistoryEditor.on('beforeChange', function(cm, change) {
            if (change.origin !== 'setValue') {
                change.cancel();
            }
        });
    }
    
    // Add keyboard navigation
    document.addEventListener('keydown', function(e) {
        // Only handle arrow keys if not typing in an input/textarea
        if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA' || e.target.tagName === 'SELECT') {
            return;
        }
        
        switch(e.key) {
            case 'ArrowLeft':
                e.preventDefault();
                stepBackward();
                break;
            case 'ArrowRight':
                e.preventDefault();
                stepForward();
                break;
            case 'Home':
                e.preventDefault();
                goToStart();
                break;
            case 'End':
                e.preventDefault();
                goToEnd();
                break;
        }
    });
    
    // Check if computer should move
    window.setTimeout(checkForComputerMove, 500);
});
