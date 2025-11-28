// Computer Move Logic
// Handles engine communication and computer move generation

var isComputerThinking = false;

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
        
        // Check if game is over
        if (checkGameOver()) {
            return; // Don't trigger another computer move if game is over
        }
        
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
