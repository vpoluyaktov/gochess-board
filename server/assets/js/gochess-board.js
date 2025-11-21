// Board Management
// Handles board initialization, event handlers, and move highlighting

var board = null;
var game = new Chess();
var lastMoveSquares = { from: null, to: null };
var squareClass = 'square-55d63';

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
// Board Utility Functions
// -------------------------------------------------------------------------

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
