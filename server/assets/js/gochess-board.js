// Board Management
// Handles board initialization, event handlers, and move highlighting

var board = null;
var game = new Chess();
var lastMoveSquares = { from: null, to: null };
var squareClass = 'square-55d63';

// -------------------------------------------------------------------------
// Game End Helper
// -------------------------------------------------------------------------

/**
 * Handles all cleanup when a game ends (checkmate, stalemate, draw, resign, timeout)
 * This ensures consistent behavior across all game-ending scenarios.
 */
function handleGameEnd() {
    // Stop the clock
    stopClock();
    
    // Stop analysis if running
    if (typeof analysisActive !== 'undefined' && analysisActive) {
        stopAnalysis();
    }
    
    // Stop history scoring if running
    if (typeof historyScoringActive !== 'undefined' && historyScoringActive) {
        stopHistoryScoring();
    }
    
    // Update buttons
    updateStartPauseButton();
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
    // Ensure moveScores array slot exists (score will be filled by analysis)
    captureScoreForLastMove();
    gameState.currentPosition = gameState.moveHistory.length;
    
    // Auto-start clock on first move if not already running
    // Start clock for any first move, regardless of time control setting
    if (!gameState.clockRunning && gameState.moveHistory.length === 1) {
        Logger.board.info('Auto-starting clock on first move');
        startClock();
    }
    
    // Add increment to the player who just moved (only if clock is running)
    // Note: move.color is the color that just moved ('w' or 'b')
    if (gameState.clockRunning) {
        var prevWhite = gameState.whiteTimeMs;
        var prevBlack = gameState.blackTimeMs;
        
        if (move.color === 'w') {
            gameState.whiteTimeMs += gameState.timeControl.increment * 1000;
        } else {
            gameState.blackTimeMs += gameState.timeControl.increment * 1000;
        }
        
        Logger.board.debug('Human move completed - adding increment', {
            moveColor: move.color,
            move: uciMove,
            increment: gameState.timeControl.increment * 1000,
            whiteChange: gameState.whiteTimeMs - prevWhite,
            blackChange: gameState.blackTimeMs - prevBlack
        });
        
        updateClockDisplay();
    }

    // Update display
    updateMoveHistoryDisplay();
    updateInfoText();
    updateOpeningDisplay();
    saveGameState();

    // Update analysis with new position
    updateAnalysisForCurrentPosition();

    // Check if game is over
    if (checkGameOver()) {
        return; // Don't trigger computer move if game is over
    }

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
// Board Utility Functions
// -------------------------------------------------------------------------

function checkGameOver() {
    // Check for threefold repetition first (even if game_over() doesn't catch it)
    if (game.in_threefold_repetition()) {
        Logger.game.info('Threefold repetition detected');
        handleGameEnd();
        showGameOver('Draw by threefold repetition.');
        return true;
    }
    
    if (game.game_over()) {
        // Handle all game end cleanup
        handleGameEnd();
        
        // Determine the game result
        let message = '';
        if (game.in_checkmate()) {
            const winner = game.turn() === 'w' ? 'Black' : 'White';
            message = `Checkmate! ${winner} wins!`;
        } else if (game.in_stalemate()) {
            message = 'Stalemate! Game is a draw.';
        } else if (game.in_draw()) {
            if (game.in_threefold_repetition()) {
                message = 'Draw by threefold repetition.';
            } else if (game.insufficient_material()) {
                message = 'Draw by insufficient material.';
            } else {
                message = 'Draw by fifty-move rule.';
            }
        } else {
            message = 'Game Over!';
        }
        
        // Show game over dialog
        showGameOver(message);
        return true;
    }
    return false;
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
