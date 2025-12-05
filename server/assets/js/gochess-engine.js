// Computer Move Logic
// Handles engine communication and computer move generation

var isComputerThinking = false;

async function makeComputerMove() {
    if (isComputerThinking) return;
    
    const currentPlayer = getCurrentPlayer();
    if (isHuman(currentPlayer)) return;

    if (game.game_over()) {
        Logger.engine.debug('Game over');
        return;
    }
    
    // Don't make computer moves if navigating in history
    if (gameState.isNavigating) {
        Logger.engine.debug('Navigating in history - computer waiting');
        return;
    }
    
    // Don't make computer moves if clock is not running
    // Computer must wait for user to click "Start Game" button
    if (!gameState.clockRunning) {
        Logger.engine.debug('Clock not running - computer waiting for Start Game');
        return;
    }

    isComputerThinking = true;
    var fen = game.fen();
    var isWhiteTurn = game.turn() === 'w';
    
    try {
        Logger.engine.info('Making move with engine:', { engine: currentPlayer });
        
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
        Logger.engine.debug('Response received', data);
        
        if (data.error) {
            Logger.engine.info('Game over or no legal moves', { error: data.error });
            isComputerThinking = false;
            return;
        }

        Logger.engine.info('Applying move', { move: data.move, thinkTime: data.thinkTime });
        
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
        // NOTE: The clock interval already counts down time in real-time while the engine thinks.
        // We should NOT deduct thinkTime again here - that would cause double deduction!
        // We only need to add the increment (if any) for the player who just moved.
        if (gameState.clockRunning) {
            var isUnlimitedMode = gameState.timeControl.initial === 0 && gameState.timeControl.increment === 0;
            var prevWhite = gameState.whiteTimeMs;
            var prevBlack = gameState.blackTimeMs;
            
            Logger.engine.debug('Computer move completed - before clock update', {
                move: data.move,
                thinkTime: data.thinkTime,
                isWhiteTurn: isWhiteTurn,
                isUnlimitedMode: isUnlimitedMode,
                whiteTimeMs: gameState.whiteTimeMs,
                blackTimeMs: gameState.blackTimeMs,
                increment: gameState.timeControl.increment * 1000
            });
            
            if (!isUnlimitedMode) {
                // Timed mode: only add increment (time already counted down by interval)
                // DO NOT deduct thinkTime - the clock interval already did that in real-time!
                if (isWhiteTurn) {
                    gameState.whiteTimeMs += gameState.timeControl.increment * 1000;
                } else {
                    gameState.blackTimeMs += gameState.timeControl.increment * 1000;
                }
                
                Logger.engine.debug('Computer move - after adding increment', {
                    isWhiteTurn: isWhiteTurn,
                    incrementAdded: gameState.timeControl.increment * 1000,
                    whiteChange: gameState.whiteTimeMs - prevWhite,
                    blackChange: gameState.blackTimeMs - prevBlack
                });
            } else {
                Logger.engine.trace('Computer move - unlimited mode, no clock adjustment');
            }
            // In unlimited mode, the clock interval already handles counting up
            // so we don't need to adjust the time here
            updateClockDisplay();
        } else {
            Logger.engine.debug('Computer move completed - clock not running, no update', {
                move: data.move
            });
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
        Logger.engine.error('Error getting computer move', { error: error.message });
        isComputerThinking = false;
    }
}

function checkForComputerMove() {
    const currentPlayer = getCurrentPlayer();
    Logger.engine.trace('checkForComputerMove', { currentPlayer: currentPlayer, isHuman: isHuman(currentPlayer) });
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
