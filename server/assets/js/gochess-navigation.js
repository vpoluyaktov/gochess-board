// Game History Navigation
// Handles navigation through move history (forward, backward, start, end)

function updateAnalysisForCurrentPosition() {
    // Cancel any ongoing PV animation
    board.cancelPVAnimation();
    
    // Set flag to force-start next PV animation (position changed)
    board.setPositionChanged();
    
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

// Navigate to a specific position in the move history
function goToPosition(targetPosition) {
    if (targetPosition < 0 || targetPosition > gameState.moveHistory.length) return;
    if (targetPosition === gameState.currentPosition) return; // Already at this position
    
    // Pause clock if running
    if (gameState.clockRunning && !gameState.isNavigating) {
        gameState.wasClockRunning = true;
        pauseClock();
    }
    
    gameState.isNavigating = true;
    gameState.currentPosition = targetPosition;
    
    // Rebuild game state up to target position
    game.reset();
    const tempGame = new Chess();
    
    for (let i = 0; i < targetPosition; i++) {
        const uciMove = gameState.moveHistory[i];
        const from = uciMove.substring(0, 2);
        const to = uciMove.substring(2, 4);
        const promotion = uciMove.length > 4 ? uciMove.substring(4) : undefined;
        
        tempGame.move({ from, to, promotion });
    }
    
    game.load(tempGame.fen());
    board.position(game.fen());
    
    // Highlight last move if not at start
    if (targetPosition > 0) {
        const lastMove = gameState.moveHistory[targetPosition - 1];
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
    // Determine which variant to open
    let variantPosition, variantIndex;
    
    if (gameState.selectedVariant) {
        // Use the clicked/selected variant
        variantPosition = gameState.selectedVariant.position;
        variantIndex = gameState.selectedVariant.index;
        console.log('Opening selected variant:', { variantPosition, variantIndex, selectedVariant: gameState.selectedVariant });
    } else {
        // Use variant at current position (legacy behavior)
        const currentMoveIndex = gameState.currentPosition > 0 ? gameState.currentPosition - 1 : -1;
        
        if (currentMoveIndex < 0 || !gameState.variants[currentMoveIndex] || 
            gameState.variants[currentMoveIndex].length === 0) {
            console.log('No variant at current position:', currentMoveIndex);
            return;
        }
        
        variantPosition = currentMoveIndex;
        variantIndex = 0; // Default to first variant
        console.log('Opening variant at current position:', { variantPosition, variantIndex });
    }
    
    // Validate variant exists
    console.log('Checking if variant exists:', { 
        variantPosition, 
        variantIndex,
        variantExists: gameState.variants[variantPosition] !== undefined,
        variantAtIndex: gameState.variants[variantPosition] ? gameState.variants[variantPosition][variantIndex] : undefined
    });
    
    if (!gameState.variants[variantPosition] || 
        !gameState.variants[variantPosition][variantIndex]) {
        console.log('Variant does not exist at position', variantPosition, 'index', variantIndex);
        console.log('Available variants:', gameState.variants);
        return;
    }
    
    // Get the selected variant
    const variant = gameState.variants[variantPosition][variantIndex];
    
    if (variant.length === 0) return;
    
    // Build FEN for the position BEFORE the variant branches
    const tempGame = new Chess();
    for (let i = 0; i < variantPosition; i++) {
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
        moveHistory: gameState.moveHistory.slice(0, variantPosition),
        currentPosition: variantPosition,
        variantStartPosition: variantPosition,
        variantIndex: variantIndex, // Track which variant was opened
        variantMoves: variant, // Include the variant moves to apply
        analysisActive: analysisActive,
        analysisEngine: document.getElementById('analysisEngine').value,
        whitePlayer: getWhitePlayer(),
        blackPlayer: getBlackPlayer(),
        timeControl: gameState.timeControl,
        whiteTimeMs: gameState.whiteTimeMs,
        blackTimeMs: gameState.blackTimeMs
    };
    
    // Open new tab with same URL
    variantWindow = window.open(window.location.href, '_blank');
    
    // Wait for variant window to signal it's ready
    let messageSent = false;
    const readyListener = function(event) {
        if (event.origin !== window.location.origin) return;
        if (event.data.type === 'variant-ready' && !messageSent) {
            console.log('Variant window is ready, sending data');
            messageSent = true;
            try {
                console.log('Sending variant data to child window:', variantData);
                variantWindow.postMessage(variantData, window.location.origin);
                console.log('Variant data sent successfully');
            } catch (e) {
                console.error('Error sending variant data:', e);
            }
            window.removeEventListener('message', readyListener);
        }
    };
    
    window.addEventListener('message', readyListener);
    
    // Fallback: if no ready signal after 5 seconds, try sending anyway
    setTimeout(function() {
        if (!messageSent && variantWindow && !variantWindow.closed) {
            console.log('Timeout waiting for ready signal, sending anyway');
            try {
                variantWindow.postMessage(variantData, window.location.origin);
                messageSent = true;
            } catch (e) {
                console.error('Error sending variant data:', e);
            }
        }
        window.removeEventListener('message', readyListener);
    }, 5000);
}

function updateVariantButtons() {
    const openVariationBtn = document.getElementById('openVariationBtn');
    
    if (openVariationBtn) {
        // Enable "Open variation" button in two cases:
        // 1. When a variant line is explicitly selected (clicked)
        // 2. When viewing a move that has variant alternatives
        
        if (gameState.selectedVariant) {
            // A variant line was clicked - enable button
            openVariationBtn.disabled = false;
        } else {
            // Check if current move has variants
            const currentMoveIndex = gameState.currentPosition > 0 ? gameState.currentPosition - 1 : -1;
            const hasVariant = currentMoveIndex >= 0 && 
                              gameState.variants[currentMoveIndex] && 
                              gameState.variants[currentMoveIndex].length > 0;
            openVariationBtn.disabled = !hasVariant;
        }
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
    gameState.gameId = generateGameId();  // New game ID for engine pooling
    
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
