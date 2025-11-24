// Variant Support Functions
// Handles variant window creation, messaging, and merging

var isVariantMode = false;
var variantWindow = null;
var mainWindow = null;
var variantStartPosition = 0; // Track where the variant started
var variantIndex = -1; // Track which variant was opened (-1 for new variants)

function initializeVariantMode() {
    // Check if this window was opened as a variant
    if (window.opener && window.opener !== window) {
        isVariantMode = true;
        mainWindow = window.opener;
        
        // Show variant mode controls, hide start variant button
        document.getElementById('startVariantBtn').style.display = 'none';
        document.getElementById('variantModeControls').style.display = 'block';
        
        // Update page title to indicate variant mode
        document.title = 'Chess Board - Variant';
        
        // Add visual indicator
        const header = document.querySelector('h1');
        if (header) {
            header.textContent = '♟️ Go Chess Board - Variant Mode ♟️';
            header.style.color = '#ff9800';
        }
        
        // Listen for messages from parent window (main or another variant)
        window.addEventListener('message', handleVariantMessage);
        
        // Also listen for messages from child variant windows (for nested variants)
        window.addEventListener('message', handleMainWindowMessage);
        
        // Notify parent window that variant is ready
        console.log('Variant window initialized, sending ready signal to parent');
        mainWindow.postMessage({ type: 'variant-ready' }, window.location.origin);
        console.log('Ready signal sent');
    } else {
        // Main window mode - listen for messages from variant windows
        window.addEventListener('message', handleMainWindowMessage);
    }
}

function startVariant() {
    // When viewing a move, we want to create a variant that's an alternative to that move
    // The variant should be stored AFTER the previous move (before the current move)
    // currentPosition is "after N moves", so the current move index is currentPosition - 1
    const currentMoveIndex = gameState.currentPosition - 1;
    
    // The variant will be stored at the position of the current move
    // This means it appears AFTER the previous move in PGN standard format
    const variantStartPos = currentMoveIndex;
    
    // Build FEN for the position before the current move (after the previous move)
    const tempGame = new Chess();
    for (let i = 0; i < currentMoveIndex; i++) {
        const uciMove = gameState.moveHistory[i];
        const from = uciMove.substring(0, 2);
        const to = uciMove.substring(2, 4);
        const promotion = uciMove.length > 4 ? uciMove.substring(4) : undefined;
        tempGame.move({ from, to, promotion });
    }
    
    // Collect current game state up to (but not including) the current move
    const movesBeforeVariant = gameState.moveHistory.slice(0, currentMoveIndex);
    const variantData = {
        type: 'start-variant',
        fen: tempGame.fen(),
        moveHistory: movesBeforeVariant,
        currentPosition: movesBeforeVariant.length,
        variantStartPosition: variantStartPos, // Track where variant starts
        variantIndex: -1, // -1 indicates this is a new variant
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
                console.log('Sending new variant data to child window:', variantData);
                variantWindow.postMessage(variantData, window.location.origin);
                console.log('New variant data sent successfully');
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

function handleVariantMessage(event) {
    // Verify origin for security
    if (event.origin !== window.location.origin) {
        console.log('Rejected message from wrong origin:', event.origin);
        return;
    }
    
    const data = event.data;
    
    // Filter out noise from browser extensions
    if (!data || !data.type || data.source === 'react-devtools-content-script') {
        return;
    }
    
    console.log('Received message:', data.type, data);
    
    if (data.type === 'start-variant') {
        // Prevent processing the same message multiple times
        if (window.variantDataLoaded) {
            console.log('Ignoring duplicate variant data message');
            return;
        }
        window.variantDataLoaded = true;
        
        // Check if moveHistoryEditor is initialized
        console.log('moveHistoryEditor initialized:', moveHistoryEditor !== null);
        
        // Load the game state from main window
        console.log('Loading variant data:', data);
        
        // Reset game to starting position first
        game.reset();
        
        // Replay moves up to current position
        for (let i = 0; i < data.moveHistory.length; i++) {
            const uciMove = data.moveHistory[i];
            const from = uciMove.substring(0, 2);
            const to = uciMove.substring(2, 4);
            const promotion = uciMove.length > 4 ? uciMove.substring(4) : undefined;
            
            game.move({ from, to, promotion });
        }
        
        // Update board position
        board.position(game.fen());
        
        // Set game state
        gameState.moveHistory = data.moveHistory.slice();
        gameState.currentPosition = data.currentPosition;
        gameState.timeControl = data.timeControl;
        gameState.whiteTimeMs = data.whiteTimeMs;
        gameState.blackTimeMs = data.blackTimeMs;
        
        // Store where the variant started and which variant it is
        variantStartPosition = data.variantStartPosition;
        variantIndex = data.variantIndex !== undefined ? data.variantIndex : -1;
        
        // Highlight last move if any
        if (data.moveHistory.length > 0) {
            const lastMove = data.moveHistory[data.moveHistory.length - 1];
            const from = lastMove.substring(0, 2);
            const to = lastMove.substring(2, 4);
            highlightLastMove(from, to);
        }
        
        // Set player selections
        document.getElementById('whitePlayer').value = data.whitePlayer;
        document.getElementById('blackPlayer').value = data.blackPlayer;
        
        // Apply variant moves if provided
        if (data.variantMoves && data.variantMoves.length > 0) {
            for (let i = 0; i < data.variantMoves.length; i++) {
                const uciMove = data.variantMoves[i];
                const from = uciMove.substring(0, 2);
                const to = uciMove.substring(2, 4);
                const promotion = uciMove.length > 4 ? uciMove.substring(4) : undefined;
                
                game.move({ from, to, promotion });
                gameState.moveHistory.push(uciMove);
                gameState.currentPosition++;
            }
            
            // Update board position
            board.position(game.fen());
            
            // Highlight last move
            if (data.variantMoves.length > 0) {
                const lastMove = data.variantMoves[data.variantMoves.length - 1];
                const from = lastMove.substring(0, 2);
                const to = lastMove.substring(2, 4);
                highlightLastMove(from, to);
            }
        }
        
        // Keep variants from the data if provided, otherwise preserve existing variants
        // This allows nested variants (variants within variant windows)
        if (data.variants) {
            gameState.variants = data.variants;
        }
        // Don't clear variants - allow sub-variants in variant windows
        
        // Update all displays
        console.log('Updating displays with gameState:', {
            moveHistory: gameState.moveHistory,
            currentPosition: gameState.currentPosition,
            variantStartPosition: variantStartPosition,
            variantIndex: variantIndex
        });
        updateMoveHistoryDisplay();
        updateClockDisplay();
        updateInfoText();
        updateOpeningDisplay();
        updatePositionIndicator();
        
        // Start analysis if it was active in main window
        if (data.analysisActive) {
            document.getElementById('analysisEngine').value = data.analysisEngine;
            setTimeout(function() {
                startAnalysis();
            }, 500);
        }
    } else if (data.type === 'apply-variant-moves') {
        // Apply variant moves to the current position
        console.log('Applying variant moves:', data.moves);
        
        for (let i = 0; i < data.moves.length; i++) {
            const uciMove = data.moves[i];
            const from = uciMove.substring(0, 2);
            const to = uciMove.substring(2, 4);
            const promotion = uciMove.length > 4 ? uciMove.substring(4) : undefined;
            
            game.move({ from, to, promotion });
            gameState.moveHistory.push(uciMove);
            gameState.currentPosition++;
        }
        
        // Update board position
        board.position(game.fen());
        
        // Highlight last move
        if (data.moves.length > 0) {
            const lastMove = data.moves[data.moves.length - 1];
            const from = lastMove.substring(0, 2);
            const to = lastMove.substring(2, 4);
            highlightLastMove(from, to);
        }
        
        // Update all displays
        updateMoveHistoryDisplay();
        updateInfoText();
        updateOpeningDisplay();
        updatePositionIndicator();
        updateAnalysisForCurrentPosition();
    }
}

function handleMainWindowMessage(event) {
    // Verify origin for security
    if (event.origin !== window.location.origin) {
        return;
    }
    
    const data = event.data;
    
    if (data.type === 'variant-ready') {
        console.log('Variant window is ready');
    } else if (data.type === 'merge-variant') {
        // Add or update variant as a branch in PGN notation
        console.log('Main window received merge request:', data);
        
        const variantStartPos = data.variantStartPosition;
        const variantMoves = data.moveHistory.slice(variantStartPos);
        const variantIdx = data.variantIndex;
        
        console.log('Variant start position:', variantStartPos);
        console.log('Variant index:', variantIdx);
        console.log('Extracted variant moves:', variantMoves);
        console.log('Current main line length:', gameState.moveHistory.length);
        
        if (variantMoves.length === 0) {
            alert('No variant moves to merge!');
            return;
        }
        
        // Initialize variants object if needed
        if (!gameState.variants) {
            gameState.variants = {};
        }
        
        // Initialize variants array at this position if needed
        if (!gameState.variants[variantStartPos]) {
            gameState.variants[variantStartPos] = [];
        }
        
        // If variantIdx is valid (>= 0), replace the existing variant
        // Otherwise, add as a new variant
        if (variantIdx >= 0 && variantIdx < gameState.variants[variantStartPos].length) {
            console.log('Replacing existing variant at index', variantIdx);
            gameState.variants[variantStartPos][variantIdx] = variantMoves;
            alert('Variant updated successfully!');
        } else {
            console.log('Adding new variant');
            gameState.variants[variantStartPos].push(variantMoves);
            alert('Variant added successfully!');
        }
        
        // Merge any sub-variants from the variant window
        // Sub-variants need to be adjusted: their positions are relative to the variant window's
        // move history, so we need to keep them as-is since they reference positions within
        // the variant line we just merged
        if (data.variants && Object.keys(data.variants).length > 0) {
            console.log('Merging sub-variants:', data.variants);
            
            // The sub-variants are stored with positions relative to the variant window
            // Since we're storing the variant moves starting at variantStartPos,
            // we need to adjust the sub-variant positions by adding variantStartPos
            for (const posStr in data.variants) {
                const pos = parseInt(posStr);
                const adjustedPos = variantStartPos + pos;
                
                if (!gameState.variants[adjustedPos]) {
                    gameState.variants[adjustedPos] = [];
                }
                
                // Add all sub-variants at this adjusted position
                for (const subVariant of data.variants[pos]) {
                    gameState.variants[adjustedPos].push(subVariant);
                }
            }
            
            console.log('Sub-variants merged');
        }
        
        console.log('Variants after merge:', gameState.variants);
        
        // Update display to show the variant
        updateMoveHistoryDisplay();
    }
}

function mergeVariant() {
    if (!isVariantMode || !mainWindow) {
        alert('This is not a variant window!');
        return;
    }
    
    console.log('Merging variant - Start position:', variantStartPosition);
    console.log('Total move history length:', gameState.moveHistory.length);
    console.log('Variant moves:', gameState.moveHistory.slice(variantStartPosition));
    console.log('Sub-variants:', gameState.variants);
    
    // Send variant data back to parent window (could be main window or another variant window)
    const variantData = {
        type: 'merge-variant',
        moveHistory: gameState.moveHistory,
        variantStartPosition: variantStartPosition,
        variantIndex: variantIndex, // Send which variant this is
        variants: gameState.variants, // Include any sub-variants created in this window
        fen: game.fen()
    };
    
    mainWindow.postMessage(variantData, window.location.origin);
    
    // Close this variant window
    setTimeout(function() {
        window.close();
    }, 500);
}

function closeVariant() {
    if (!isVariantMode) {
        alert('This is not a variant window!');
        return;
    }
    
    window.close();
}
