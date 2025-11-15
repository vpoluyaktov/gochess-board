// Chess UI Logic
// Handles game state, player interactions, computer moves, and live analysis

var board = null;
var game = new Chess();
var isComputerThinking = false;
var lastMoveSquares = { from: null, to: null };
var squareClass = 'square-55d63';

// Analysis WebSocket connections
var whiteAnalysisWs = null;
var blackAnalysisWs = null;
var whiteAnalysisActive = false;
var blackAnalysisActive = false;

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
    blackMoves: 0
};

// Save game state to localStorage
function saveGameState() {
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
}

// Load game state from localStorage
function loadGameState() {
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
}

// Clear saved game state
function clearGameState() {
    try {
        localStorage.removeItem('chessGameState');
    } catch (e) {
        console.error('Failed to clear game state:', e);
    }
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

    // Clear arrow and update analyses with new position
    board.clearArrow();
    setTimeout(function() {
        if (whiteAnalysisActive && whiteAnalysisWs && whiteAnalysisWs.readyState === WebSocket.OPEN) {
            whiteAnalysisWs.send(JSON.stringify({
                action: 'update',
                fen: game.fen()
            }));
        }
        if (blackAnalysisActive && blackAnalysisWs && blackAnalysisWs.readyState === WebSocket.OPEN) {
            blackAnalysisWs.send(JSON.stringify({
                action: 'update',
                fen: game.fen()
            }));
        }
    }, 100);

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
    
    board.clearArrow();
    setTimeout(function() {
        if (whiteAnalysisActive && whiteAnalysisWs && whiteAnalysisWs.readyState === WebSocket.OPEN) {
            whiteAnalysisWs.send(JSON.stringify({
                action: 'update',
                fen: game.fen()
            }));
        }
        if (blackAnalysisActive && blackAnalysisWs && blackAnalysisWs.readyState === WebSocket.OPEN) {
            blackAnalysisWs.send(JSON.stringify({
                action: 'update',
                fen: game.fen()
            }));
        }
    }, 100);
};

// -------------------------------------------------------------------------
// Move History Update (Client-side)
// -------------------------------------------------------------------------

function updateMoveHistoryDisplay() {
    const textArea = document.getElementById('moveHistoryText');
    
    if (!gameState.moveHistory || gameState.moveHistory.length === 0) {
        textArea.value = '';
        return;
    }
    
    // Convert UCI moves to SAN notation for PGN format
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
        
        pgn += moveNum + '. ' + whiteMove;
        if (blackMove) {
            pgn += ' ' + blackMove;
        }
        
        // Add space between move pairs (let textarea handle line wrapping)
        if (i + 2 < sanMoves.length) {
            pgn += ' ';
        }
    }
    
    textArea.value = pgn;
    
    // Auto-scroll to bottom
    textArea.scrollTop = textArea.scrollHeight;
}

// Copy move history to clipboard
function copyMoveHistory() {
    const textArea = document.getElementById('moveHistoryText');
    
    if (!textArea.value) {
        return;
    }
    
    // Select and copy
    textArea.select();
    textArea.setSelectionRange(0, 99999); // For mobile devices
    
    try {
        document.execCommand('copy');
        
        // Visual feedback
        const btn = event.target;
        const originalText = btn.textContent;
        btn.textContent = '✓';
        setTimeout(function() {
            btn.textContent = originalText;
        }, 1000);
    } catch (err) {
        console.error('Failed to copy:', err);
    }
    
    // Deselect
    window.getSelection().removeAllRanges();
}

// -------------------------------------------------------------------------
// Opening Display
// -------------------------------------------------------------------------

async function updateOpeningDisplay() {
    // Get move history in SAN notation
    const sanMoves = [];
    const tempGame = new Chess();
    
    // Replay the game to get SAN notation
    for (let i = 0; i < gameState.moveHistory.length; i++) {
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
}

function savePlayerSelections() {
    localStorage.setItem('whitePlayer', document.getElementById('whitePlayer').value);
    localStorage.setItem('blackPlayer', document.getElementById('blackPlayer').value);
    updatePlayerControls();
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
// Live Analysis - White
// -------------------------------------------------------------------------

function toggleWhiteAnalysis() {
    if (whiteAnalysisActive) {
        stopWhiteAnalysis();
    } else {
        startWhiteAnalysis();
    }
}

function startWhiteAnalysis() {
    var protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    var wsUrl = protocol + '//' + window.location.host + '/api/analysis';
    
    whiteAnalysisWs = new WebSocket(wsUrl);
    
    whiteAnalysisWs.onopen = function() {
        console.log('White Analysis WebSocket connected');
        whiteAnalysisActive = true;
        document.getElementById('whiteAnalysisToggle').textContent = 'Stop White';
        document.getElementById('whiteAnalysisInfo').style.display = 'block';
        
        whiteAnalysisWs.send(JSON.stringify({
            action: 'start',
            fen: game.fen(),
            enginePath: 'stockfish'
        }));
    };
    
    whiteAnalysisWs.onmessage = function(event) {
        var data = JSON.parse(event.data);
        
        if (data.error) {
            console.error('White Analysis error:', data.error);
            return;
        }
        
        document.getElementById('whiteAnalysisDepth').textContent = data.depth || '-';
        
        var scoreText = '';
        if (data.scoreType === 'cp') {
            var score = (data.score / 100).toFixed(2);
            scoreText = (data.score >= 0 ? '+' : '') + score;
        } else if (data.scoreType === 'mate') {
            scoreText = 'Mate in ' + Math.abs(data.score);
        }
        document.getElementById('whiteAnalysisScore').textContent = scoreText;
        
        var nodesText = data.nodes ? (data.nodes / 1000).toFixed(0) + 'k' : '-';
        document.getElementById('whiteAnalysisNodes').textContent = nodesText;
        
        if (data.bestMove && data.bestMove.length >= 4 && game.turn() === 'w') {
            var from = data.bestMove.substring(0, 2);
            var to = data.bestMove.substring(2, 4);
            var piece = game.get(from);
            if (piece && piece.color === 'w') {
                board.drawArrow(from, to, '#3296FF');
            }
        }
    };
    
    whiteAnalysisWs.onerror = function(error) {
        console.error('White WebSocket error:', error);
        stopWhiteAnalysis();
    };
    
    whiteAnalysisWs.onclose = function() {
        console.log('White Analysis WebSocket closed');
        if (whiteAnalysisActive) {
            stopWhiteAnalysis();
        }
    };
}

function stopWhiteAnalysis() {
    if (whiteAnalysisWs) {
        whiteAnalysisWs.send(JSON.stringify({ action: 'stop' }));
        whiteAnalysisWs.close();
        whiteAnalysisWs = null;
    }
    
    whiteAnalysisActive = false;
    document.getElementById('whiteAnalysisToggle').textContent = 'White Analysis';
    document.getElementById('whiteAnalysisInfo').style.display = 'none';
    
    if (!blackAnalysisActive) {
        board.clearArrow();
    }
}

// -------------------------------------------------------------------------
// Live Analysis - Black
// -------------------------------------------------------------------------

function toggleBlackAnalysis() {
    if (blackAnalysisActive) {
        stopBlackAnalysis();
    } else {
        startBlackAnalysis();
    }
}

function startBlackAnalysis() {
    var protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    var wsUrl = protocol + '//' + window.location.host + '/api/analysis';
    
    blackAnalysisWs = new WebSocket(wsUrl);
    
    blackAnalysisWs.onopen = function() {
        console.log('Black Analysis WebSocket connected');
        blackAnalysisActive = true;
        document.getElementById('blackAnalysisToggle').textContent = 'Stop Black';
        document.getElementById('blackAnalysisInfo').style.display = 'block';
        
        blackAnalysisWs.send(JSON.stringify({
            action: 'start',
            fen: game.fen(),
            enginePath: 'stockfish'
        }));
    };
    
    blackAnalysisWs.onmessage = function(event) {
        var data = JSON.parse(event.data);
        
        if (data.error) {
            console.error('Black Analysis error:', data.error);
            return;
        }
        
        document.getElementById('blackAnalysisDepth').textContent = data.depth || '-';
        
        var scoreText = '';
        if (data.scoreType === 'cp') {
            var score = (data.score / 100).toFixed(2);
            scoreText = (data.score >= 0 ? '+' : '') + score;
        } else if (data.scoreType === 'mate') {
            scoreText = 'Mate in ' + Math.abs(data.score);
        }
        document.getElementById('blackAnalysisScore').textContent = scoreText;
        
        var nodesText = data.nodes ? (data.nodes / 1000).toFixed(0) + 'k' : '-';
        document.getElementById('blackAnalysisNodes').textContent = nodesText;
        
        if (data.bestMove && data.bestMove.length >= 4 && game.turn() === 'b') {
            var from = data.bestMove.substring(0, 2);
            var to = data.bestMove.substring(2, 4);
            var piece = game.get(from);
            if (piece && piece.color === 'b') {
                board.drawArrow(from, to, '#FF6B6B');
            }
        }
    };
    
    blackAnalysisWs.onerror = function(error) {
        console.error('Black WebSocket error:', error);
        stopBlackAnalysis();
    };
    
    blackAnalysisWs.onclose = function() {
        console.log('Black Analysis WebSocket closed');
        if (blackAnalysisActive) {
            stopBlackAnalysis();
        }
    };
}

function stopBlackAnalysis() {
    if (blackAnalysisWs) {
        blackAnalysisWs.send(JSON.stringify({ action: 'stop' }));
        blackAnalysisWs.close();
        blackAnalysisWs = null;
    }
    
    blackAnalysisActive = false;
    document.getElementById('blackAnalysisToggle').textContent = 'Black Analysis';
    document.getElementById('blackAnalysisInfo').style.display = 'none';
    
    if (!whiteAnalysisActive) {
        board.clearArrow();
    }
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
    
    // Add time warnings
    whiteClock.classList.remove('time-low', 'time-critical');
    blackClock.classList.remove('time-low', 'time-critical');
    
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
    
    // Check for time out
    if (gameState.whiteTimeMs <= 0) {
        stopClock();
        alert('Time out! Black wins!');
    } else if (gameState.blackTimeMs <= 0) {
        stopClock();
        alert('Time out! White wins!');
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
    
    // Time control selector
    document.getElementById('timeControl').addEventListener('change', function() {
        var value = this.value;
        var parts = value.split('+');
        var initialMinutes = parseInt(parts[0]);
        var incrementSeconds = parseInt(parts[1]);
        
        setTimeControl(initialMinutes, incrementSeconds);
        stopClock();
        
        // Save preference
        localStorage.setItem('timeControl', value);
    });
    
    // Restore time control preference
    var savedTimeControl = localStorage.getItem('timeControl');
    if (savedTimeControl) {
        var select = document.getElementById('timeControl');
        if (Array.from(select.options).some(opt => opt.value === savedTimeControl)) {
            select.value = savedTimeControl;
            var parts = savedTimeControl.split('+');
            setTimeControl(parseInt(parts[0]), parseInt(parts[1]));
        }
    }
    
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
    
    // Check if computer should move
    window.setTimeout(checkForComputerMove, 500);
});
