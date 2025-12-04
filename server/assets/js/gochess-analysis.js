// Live Analysis
// Handles WebSocket connection to analysis engine and arrow display

var analysisWs = null;
var analysisActive = false;
var lastAnalysisScore = null;  // Last known score from analysis (in centipawns, from White's perspective)

// -------------------------------------------------------------------------
// History Scoring - Analyze all moves in history to fill in scores
// -------------------------------------------------------------------------

var historyScoringWs = null;      // Separate WebSocket for history scoring
var historyScoringActive = false; // True when scoring history moves
var historyScoringQueue = [];     // Queue of {moveIndex, fen} to analyze
var historyScoringCurrent = -1;   // Current move index being analyzed
var historyScoringDepth = 12;     // Target depth for scoring (balance speed vs accuracy)
var historyScoringLastDepth = 0;  // Last depth received for current position

// Cache of FEN positions for each move (built lazily)
var movePositionCache = null;
var movePositionCacheLength = -1;

// Debounce timer for history display updates
var historyDisplayUpdateTimer = null;

// Schedule a debounced update to the history display and eval graph
// This prevents flicker when multiple scores arrive in quick succession
function scheduleHistoryDisplayUpdate() {
    if (historyDisplayUpdateTimer) {
        clearTimeout(historyDisplayUpdateTimer);
    }
    historyDisplayUpdateTimer = setTimeout(function() {
        historyDisplayUpdateTimer = null;
        updateMoveHistoryDisplay();
        // Also update eval graph if visible
        if (typeof updateEvalGraph === 'function') {
            updateEvalGraph();
        }
    }, 100); // 100ms debounce
}

// Find which move index corresponds to a given FEN position
// Returns the index of the move that resulted in this position, or -1 if not found
function findMoveIndexForFen(fen) {
    // Rebuild cache if move history has changed
    if (movePositionCache === null || movePositionCacheLength !== gameState.moveHistory.length) {
        movePositionCache = {};
        var tempGame = new Chess();
        
        for (var i = 0; i < gameState.moveHistory.length; i++) {
            var uciMove = gameState.moveHistory[i];
            var from = uciMove.substring(0, 2);
            var to = uciMove.substring(2, 4);
            var promotion = uciMove.length > 4 ? uciMove.substring(4) : undefined;
            
            try {
                tempGame.move({ from: from, to: to, promotion: promotion });
                // Store the position part of FEN (first 4 fields) as key
                var fenKey = tempGame.fen().split(' ').slice(0, 4).join(' ');
                movePositionCache[fenKey] = i;
            } catch (e) {
                // Invalid move, stop
                break;
            }
        }
        movePositionCacheLength = gameState.moveHistory.length;
    }
    
    // Look up the FEN in the cache
    var fenKey = fen.split(' ').slice(0, 4).join(' ');
    if (movePositionCache.hasOwnProperty(fenKey)) {
        return movePositionCache[fenKey];
    }
    return -1;
}

// -------------------------------------------------------------------------
// Move Preparation Functions (Chess.js Logic)
// -------------------------------------------------------------------------

// Prepare PV moves for visualization (computes all chess logic)
function preparePVMoves(data, gameInstance) {
    var moves = [];
    var tempGame = new Chess(gameInstance.fen());
    
    for (var i = 0; i < data.pv.length; i++) {
        var move = data.pv[i];
        
        // Parse UCI move format
        if (move.length < 4) break;
        
        var from = move.substring(0, 2);
        var to = move.substring(2, 4);
        var promotion = move.length > 4 ? move.substring(4) : undefined;
        
        // Verify piece exists at the from square
        var piece = tempGame.get(from);
        if (!piece) break;
        
        // Calculate move number and turn from current FEN
        var fenParts = tempGame.fen().split(' ');
        var moveNumber = parseInt(fenParts[5]) || 1;
        var isBlackMove = tempGame.turn() === 'b';
        
        // Store all computed data
        moves.push({
            from: from,
            to: to,
            piece: piece,
            moveNumber: moveNumber,
            isBlackMove: isBlackMove
        });
        
        // Apply move to temp game for next iteration
        try {
            tempGame.move({
                from: from,
                to: to,
                promotion: promotion || 'q'
            });
        } catch (e) {
            // Invalid move, stop processing
            break;
        }
    }
    
    return {
        moves: moves,
        scoreType: data.scoreType,
        score: data.score
    };
}

// Prepare multi-PV moves for visualization
function prepareMultiPVMoves(multiPV, gameInstance) {
    var lines = [];
    
    for (var i = 0; i < multiPV.length; i++) {
        var pvLine = multiPV[i];
        
        if (!pvLine.moves || pvLine.moves.length === 0) continue;
        
        var move = pvLine.moves[0]; // Only first move of each line
        
        // Parse UCI move format
        if (move.length < 4) continue;
        
        var from = move.substring(0, 2);
        var to = move.substring(2, 4);
        
        // Verify the piece exists at the from square
        var piece = gameInstance.get(from);
        if (!piece) continue;
        
        lines.push({
            from: from,
            to: to,
            piece: piece,
            scoreType: pvLine.scoreType,
            score: pvLine.score
        });
    }
    
    return lines;
}

// Initialize event listeners for analysis display mode changes
function initAnalysisDisplayListeners() {
    var radioButtons = document.querySelectorAll('input[name="analysisDisplay"]');
    radioButtons.forEach(function(radio) {
        radio.addEventListener('change', function() {
            // Cancel any ongoing PV animation when switching modes
            board.cancelPVAnimation();
            // Clear arrows when switching modes
            board.clearArrow();
        });
    });
}

// Call initialization when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initAnalysisDisplayListeners);
} else {
    initAnalysisDisplayListeners();
}

function toggleAnalysis() {
    if (analysisActive || historyScoringActive) {
        // Stop both if either is active
        stopAnalysis();
        stopHistoryScoring();
    } else {
        // Start regular analysis first (for current position arrows)
        startAnalysis();
        
        // Also start history scoring if there are missing scores
        if (hasMissingScores()) {
            startHistoryScoring();
        }
        
        // Update eval graph with any existing scores
        if (typeof updateEvalGraph === 'function') {
            updateEvalGraph();
        }
    }
}

// Check if any moves in history are missing scores
function hasMissingScores() {
    if (!gameState.moveHistory || gameState.moveHistory.length === 0) {
        return false;
    }
    
    for (var i = 0; i < gameState.moveHistory.length; i++) {
        if (!gameState.moveScores || gameState.moveScores[i] === null || gameState.moveScores[i] === undefined) {
            return true;
        }
    }
    return false;
}

// Start scoring all moves in history
function startHistoryScoring() {
    if (historyScoringActive) return;
    
    console.log('Starting history scoring...');
    
    // Build queue of positions to analyze (after each move)
    historyScoringQueue = [];
    var tempGame = new Chess();
    
    for (var i = 0; i < gameState.moveHistory.length; i++) {
        // Check if this move already has a score
        if (gameState.moveScores && gameState.moveScores[i] !== null && gameState.moveScores[i] !== undefined) {
            // Already has score, skip
            var uciMove = gameState.moveHistory[i];
            var from = uciMove.substring(0, 2);
            var to = uciMove.substring(2, 4);
            var promotion = uciMove.length > 4 ? uciMove.substring(4) : undefined;
            tempGame.move({ from: from, to: to, promotion: promotion });
            continue;
        }
        
        // Apply the move first, then get the resulting position
        var uciMove = gameState.moveHistory[i];
        var from = uciMove.substring(0, 2);
        var to = uciMove.substring(2, 4);
        var promotion = uciMove.length > 4 ? uciMove.substring(4) : undefined;
        tempGame.move({ from: from, to: to, promotion: promotion });
        
        // Queue the position AFTER the move for analysis
        historyScoringQueue.push({
            moveIndex: i,
            fen: tempGame.fen()
        });
    }
    
    if (historyScoringQueue.length === 0) {
        console.log('All moves already have scores');
        return;
    }
    
    console.log('Scoring ' + historyScoringQueue.length + ' positions...');
    
    // Ensure moveScores array exists and is the right size
    if (!gameState.moveScores) {
        gameState.moveScores = [];
    }
    while (gameState.moveScores.length < gameState.moveHistory.length) {
        gameState.moveScores.push(null);
    }
    
    historyScoringActive = true;
    historyScoringCurrent = -1;
    
    // Connect WebSocket
    var protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    var wsUrl = protocol + '//' + window.location.host + '/api/analysis';
    
    var engineSelect = document.getElementById('analysisEngine');
    var enginePath = engineSelect ? engineSelect.value : 'stockfish';
    
    historyScoringWs = new WebSocket(wsUrl);
    
    historyScoringWs.onopen = function() {
        console.log('History scoring WebSocket connected');
        // Start analyzing first position
        analyzeNextHistoryPosition(enginePath);
    };
    
    historyScoringWs.onmessage = function(event) {
        var data = JSON.parse(event.data);
        
        if (data.error) {
            console.error('History scoring error:', data.error);
            return;
        }
        
        // Check if this is for the current position we're analyzing
        if (historyScoringCurrent < 0 || historyScoringCurrent >= historyScoringQueue.length) {
            return;
        }
        
        var currentItem = historyScoringQueue[historyScoringCurrent];
        
        // Verify FEN matches
        if (data.fen) {
            var analysisFenParts = data.fen.split(' ').slice(0, 4).join(' ');
            var expectedFenParts = currentItem.fen.split(' ').slice(0, 4).join(' ');
            if (analysisFenParts !== expectedFenParts) {
                return; // Stale result
            }
        }
        
        // Track depth
        if (data.depth !== undefined) {
            historyScoringLastDepth = data.depth;
        }
        
        // Store score when we reach target depth
        if (data.score !== undefined && data.depth >= historyScoringDepth) {
            // Store score with type (for mate detection)
            gameState.moveScores[currentItem.moveIndex] = {
                score: data.score,
                scoreType: data.scoreType || 'cp'
            };
            var scoreStr = data.scoreType === 'mate' ? 'M' + data.score : data.score + ' cp';
            console.log('Scored move ' + (currentItem.moveIndex + 1) + ': ' + scoreStr);
            
            // Update display
            updateMoveHistoryDisplay();
            
            // Move to next position
            analyzeNextHistoryPosition(enginePath);
        }
    };
    
    historyScoringWs.onerror = function(error) {
        console.error('History scoring WebSocket error:', error);
        stopHistoryScoring();
    };
    
    historyScoringWs.onclose = function() {
        console.log('History scoring WebSocket closed');
        if (historyScoringActive) {
            stopHistoryScoring();
        }
    };
}

// Analyze next position in the queue
function analyzeNextHistoryPosition(enginePath) {
    historyScoringCurrent++;
    historyScoringLastDepth = 0;
    
    if (historyScoringCurrent >= historyScoringQueue.length) {
        // Done scoring all positions
        console.log('History scoring complete!');
        stopHistoryScoring();
        // Regular analysis is already running, no need to start it
        return;
    }
    
    var item = historyScoringQueue[historyScoringCurrent];
    console.log('Analyzing position ' + (historyScoringCurrent + 1) + '/' + historyScoringQueue.length + ' (move ' + (item.moveIndex + 1) + ')');
    
    // Send analysis request
    if (historyScoringCurrent === 0) {
        // First position - start engine
        historyScoringWs.send(JSON.stringify({
            action: 'start',
            fen: item.fen,
            enginePath: enginePath
        }));
    } else {
        // Subsequent positions - update
        historyScoringWs.send(JSON.stringify({
            action: 'update',
            fen: item.fen
        }));
    }
}

// Stop history scoring
function stopHistoryScoring() {
    historyScoringActive = false;
    historyScoringQueue = [];
    historyScoringCurrent = -1;
    
    if (historyScoringWs) {
        try {
            historyScoringWs.send(JSON.stringify({ action: 'stop' }));
            historyScoringWs.close();
        } catch (e) {
            // Ignore errors when closing
        }
        historyScoringWs = null;
    }
    
    // Only reset button if regular analysis is not running
    if (!analysisActive) {
        document.getElementById('analysisToggle').innerHTML = '🔍 Start Analysis <span id="analysisDepth" style="opacity: 0.7; font-size: 0.85em;"></span>';
    } else {
        // Regular analysis is still running, show its button state
        document.getElementById('analysisToggle').innerHTML = '⏹️ Stop Analysis <span id="analysisDepth" style="opacity: 0.7; font-size: 0.85em;"></span>';
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
        document.getElementById('analysisToggle').innerHTML = '⏹️ Stop Analysis <span id="analysisDepth" style="opacity: 0.7; font-size: 0.85em;"></span>';
        
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
        
        // Store score for the position this analysis is for (even if it's not the current position)
        // This ensures we capture scores for moves even when the game moves on quickly
        if (data.score !== undefined && data.fen) {
            // Find which move this FEN corresponds to (the move that led to this position)
            var moveIndex = findMoveIndexForFen(data.fen);
            if (moveIndex >= 0) {
                if (!gameState.moveScores) {
                    gameState.moveScores = [];
                }
                while (gameState.moveScores.length < gameState.moveHistory.length) {
                    gameState.moveScores.push(null);
                }
                
                var oldScore = gameState.moveScores[moveIndex];
                // Only update if we don't have a score yet
                if (oldScore === null || oldScore === undefined) {
                    // Store score with type (for mate detection)
                    gameState.moveScores[moveIndex] = {
                        score: data.score,
                        scoreType: data.scoreType || 'cp'
                    };
                    // Debounce display updates to avoid flicker
                    scheduleHistoryDisplayUpdate();
                }
            }
        }
        
        // Verify the analysis is for the current position (for arrow display)
        // Compare only the position part of FEN (first 4 fields), ignoring move counters
        var isCurrentPosition = true;
        if (data.fen) {
            var analysisFenParts = data.fen.split(' ').slice(0, 4).join(' ');
            var currentFenParts = game.fen().split(' ').slice(0, 4).join(' ');
            if (analysisFenParts !== currentFenParts) {
                isCurrentPosition = false;
            }
        }
        
        // Only update arrows and depth display for current position
        if (!isCurrentPosition) {
            return;
        }
        
        // Update depth display in button
        if (data.depth !== undefined) {
            document.getElementById('analysisDepth').textContent = '(Depth: ' + data.depth + ')';
        }
        
        // Store the last analysis score for reference
        if (data.score !== undefined) {
            lastAnalysisScore = data.score;
        }
        
        // Draw arrows based on display mode
        var showPV = document.getElementById('showPVArrows').checked;
        var showBestMoves = document.getElementById('showBestMoves').checked;
        
        // Determine visualization mode
        var mode = 'none';
        if (showBestMoves && data.multiPV && data.multiPV.length > 0) {
            mode = 'multi-pv';
        } else if (data.pv && data.pv.length > 0) {
            mode = showPV ? 'pv-animation' : 'best-move';
        }
        
        // Visualize based on mode
        switch (mode) {
            case 'best-move':
                // Case 1: Show single best move (one arrow with score)
                var pvData = preparePVMoves(data, game);
                board.drawBestMove(pvData);
                break;
                
            case 'multi-pv':
                // Case 2: Show 3 best moves (multiple arrows with scores)
                var multiPVLines = prepareMultiPVMoves(data.multiPV, game);
                board.drawMultipleBestMoves(multiPVLines);
                break;
                
            case 'pv-animation':
                // Case 3: Show PV animation (looping animation with ghost pieces)
                var pvData = preparePVMoves(data, game);
                board.drawPVAnimation(pvData);
                break;
                
            case 'none':
                // No visualization data available
                break;
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
    // Cancel any ongoing PV animation
    board.cancelPVAnimation();
    
    analysisActive = false;
    document.getElementById('analysisToggle').innerHTML = '🔍 Start Analysis <span id="analysisDepth" style="opacity: 0.7; font-size: 0.85em;"></span>';
    board.clearArrow();
    
    if (analysisWs) {
        analysisWs.send(JSON.stringify({ action: 'stop' }));
        analysisWs.close();
        analysisWs = null;
    }
}

// -------------------------------------------------------------------------
// Score capture for new moves
// -------------------------------------------------------------------------

// Called after a move is made to ensure the moveScores array has a slot for this move
// The actual score will be filled in by the analysis engine after analyzing the new position
function captureScoreForLastMove() {
    var moveIndex = gameState.moveHistory.length - 1;
    if (moveIndex < 0) return;
    
    // Ensure moveScores array exists and is the right size
    if (!gameState.moveScores) {
        gameState.moveScores = [];
    }
    while (gameState.moveScores.length < gameState.moveHistory.length) {
        gameState.moveScores.push(null);
    }
}

// Note: drawMultipleBestMoves is now handled inline with prepareMultiPVMoves()
// The application prepares the data, the library visualizes it
