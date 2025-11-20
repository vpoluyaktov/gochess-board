// Game History Navigation
// Handles navigation through move history (forward, backward, start, end)

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
    
    // Update variant navigation buttons
    updateVariantButtons();
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

// -------------------------------------------------------------------------
// Variant Navigation Functions
// -------------------------------------------------------------------------

function openVariation() {
    // Open variant of the currently displayed move
    // If at position N (after move N-1), the variant is stored at index N-1
    const currentMoveIndex = gameState.currentPosition > 0 ? gameState.currentPosition - 1 : -1;
    
    if (currentMoveIndex < 0 || !gameState.variants[currentMoveIndex] || 
        gameState.variants[currentMoveIndex].length === 0) {
        return;
    }
    
    // Get the first variant (could be extended to choose between multiple variants)
    const variant = gameState.variants[currentMoveIndex][0];
    
    if (variant.length === 0) return;
    
    // Build FEN for the position BEFORE the current move (where the variant branches)
    const tempGame = new Chess();
    for (let i = 0; i < currentMoveIndex; i++) {
        const uciMove = gameState.moveHistory[i];
        const from = uciMove.substring(0, 2);
        const to = uciMove.substring(2, 4);
        const promotion = uciMove.length > 4 ? uciMove.substring(4) : undefined;
        tempGame.move({ from, to, promotion });
    }
    
    // Prepare variant data for the new window
    const variantData = {
        type: 'start-variant',
        fen: tempGame.fen(),
        moveHistory: gameState.moveHistory.slice(0, currentMoveIndex),
        currentPosition: currentMoveIndex,
        variantStartPosition: currentMoveIndex,
        variantIndex: 0, // Track which variant was opened (always first one for now)
        variantMoves: variant, // Include the variant moves to apply
        analysisActive: analysisActive,
        analysisEngine: document.getElementById('analysisEngine').value,
        whitePlayer: getWhitePlayer(),
        blackPlayer: getBlackPlayer(),
        timeControl: gameState.timeControl,
        whiteTimeMs: gameState.whiteTimeMs,
        blackTimeMs: gameState.blackTimeMs
    };
    
    // Open new window with same URL
    const windowFeatures = 'width=1400,height=900,menubar=no,toolbar=no,location=no,status=no';
    variantWindow = window.open(window.location.href, '_blank', windowFeatures);
    
    // Wait for variant window to be ready, then send data
    const sendDataInterval = setInterval(function() {
        if (variantWindow && !variantWindow.closed) {
            try {
                variantWindow.postMessage(variantData, window.location.origin);
            } catch (e) {
                console.error('Error sending variant data:', e);
            }
        } else {
            clearInterval(sendDataInterval);
        }
    }, 100);
    
    // Stop trying after 5 seconds
    setTimeout(function() {
        clearInterval(sendDataInterval);
    }, 5000);
}

function updateVariantButtons() {
    const openVariationBtn = document.getElementById('openVariationBtn');
    
    if (openVariationBtn) {
        // Enable "Open variation" button when viewing a move that has variant alternatives
        // currentPosition is the number of moves made (position after the last move)
        // To check if the CURRENT DISPLAYED move has variants, we check currentPosition - 1
        // Example: at position 27 (after move 14 White "Kh3"), check variants[26] 
        // because the variant "14. Kxf4" is stored as an alternative at that position
        const currentMoveIndex = gameState.currentPosition > 0 ? gameState.currentPosition - 1 : -1;
        const hasVariant = currentMoveIndex >= 0 && 
                          gameState.variants[currentMoveIndex] && 
                          gameState.variants[currentMoveIndex].length > 0;
        openVariationBtn.disabled = !hasVariant;
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
    gameState.variants = {};
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
