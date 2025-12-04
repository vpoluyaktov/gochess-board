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
    
    // Don't make computer moves if clock is not running
    // Computer must wait for user to click "Start Game" button
    if (!gameState.clockRunning) {
        console.log('Clock not running - computer waiting for Start Game');
        return;
    }

    isComputerThinking = true;
    var fen = game.fen();
    var isWhiteTurn = game.turn() === 'w';
    
    try {
        console.log('Making move with engine:', currentPlayer);
        
        // Get move time for unlimited mode
        var moveTimeMs = 0;
        if (gameState.timeControl.initial === 0 && gameState.timeControl.increment === 0) {
            // Unlimited mode: use configured move time
            var moveTimeSelect = document.getElementById('moveTime');
            var moveTimeSeconds = moveTimeSelect ? parseInt(moveTimeSelect.value) : 5;
            moveTimeMs = moveTimeSeconds * 1000;
        }
        
        // Build request with all game state
        const requestBody = {
            fen: fen,
            moves: gameState.moveHistory,
            enginePath: currentPlayer,
            moveTime: moveTimeMs, // Use configured time for unlimited mode
            whiteTime: gameState.whiteTimeMs,
            blackTime: gameState.blackTimeMs,
            whiteIncrement: gameState.timeControl.increment * 1000,
            blackIncrement: gameState.timeControl.increment * 1000,
            isUnlimited: gameState.timeControl.initial === 0 && gameState.timeControl.increment === 0,
            engineOptions: {}, // Can add ELO settings here later
            gameId: gameState.gameId  // For persistent engine pooling
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
        // Ensure moveScores array slot exists (score will be filled by analysis)
        captureScoreForLastMove();
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
        
        // Note: Clock is NOT auto-started for computer moves.
        // User must click "Start Game" button to begin computer vs computer games.
        
        // Update clock (only if running)
        if (gameState.clockRunning) {
            var isUnlimitedMode = gameState.timeControl.initial === 0 && gameState.timeControl.increment === 0;
            
            if (!isUnlimitedMode) {
                // Timed mode: deduct time spent thinking and add increment
                if (isWhiteTurn) {
                    gameState.whiteTimeMs -= data.thinkTime;
                    gameState.whiteTimeMs += gameState.timeControl.increment * 1000;
                } else {
                    gameState.blackTimeMs -= data.thinkTime;
                    gameState.blackTimeMs += gameState.timeControl.increment * 1000;
                }
            }
            // In unlimited mode, the clock interval already handles counting up
            // so we don't need to adjust the time here
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
