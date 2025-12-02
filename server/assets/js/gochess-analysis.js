// Live Analysis
// Handles WebSocket connection to analysis engine and arrow display

var analysisWs = null;
var analysisActive = false;
var pvAnimationTimeouts = []; // Track animation timeouts for cancellation
var currentPVSequence = null; // Track current PV sequence for comparison
var ghostPieces = []; // Track ghost pieces for cleanup
var pendingPVData = null; // Store pending PV to apply after current loop finishes
var isAnimating = false; // Track if animation is currently running
var positionChanged = false; // Track if position changed (move made) to force-start next PV

// Initialize event listeners for analysis display mode changes
function initAnalysisDisplayListeners() {
    var radioButtons = document.querySelectorAll('input[name="analysisDisplay"]');
    radioButtons.forEach(function(radio) {
        radio.addEventListener('change', function() {
            // Cancel any ongoing PV animation when switching modes
            cancelPVAnimation();
            // Clear arrows when switching modes
            board.clearArrow();
            // Clear ghost pieces when switching modes
            clearGhostPieces();
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
    if (analysisActive) {
        stopAnalysis();
    } else {
        startAnalysis();
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
        document.getElementById('analysisToggle').textContent = '⏹️ Stop Analysis';
        
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
        
        // Update depth display
        if (data.depth !== undefined) {
            document.getElementById('analysisDepth').textContent = 'Depth: ' + data.depth;
        }
        
        // Draw arrows based on display mode
        var showBestMove = document.getElementById('showBestMove').checked;
        var showPV = document.getElementById('showPVArrows').checked;
        var showBestMoves = document.getElementById('showBestMoves').checked;
        
        if (showBestMoves && data.multiPV && data.multiPV.length > 0) {
            // Show 3 best moves mode
            drawMultipleBestMoves(data.multiPV);
        } else if (data.pv && data.pv.length > 0) {
            // Show principal variation or single best move mode
            // Force start if position changed (move was made)
            var forceStart = positionChanged;
            if (positionChanged) {
                positionChanged = false; // Reset flag after using it
            }
            drawPrincipalVariation(data, showPV, showBestMove, forceStart);
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
    if (analysisWs) {
        analysisWs.send(JSON.stringify({ action: 'stop' }));
        analysisWs.close();
        analysisWs = null;
    }
    
    // Cancel any ongoing PV animation
    cancelPVAnimation();
    
    analysisActive = false;
    document.getElementById('analysisToggle').textContent = '🔍 Start Analysis';
    document.getElementById('analysisDepth').textContent = 'Depth: -';
    board.clearArrow();
}

// Cancel ongoing PV animation
function cancelPVAnimation() {
    // Clear all pending timeouts
    for (var i = 0; i < pvAnimationTimeouts.length; i++) {
        clearTimeout(pvAnimationTimeouts[i]);
    }
    pvAnimationTimeouts = [];
    currentPVSequence = null;
    pendingPVData = null;
    isAnimating = false;
    positionChanged = false;
    // Also clear ghost pieces
    clearGhostPieces();
}

// Add a ghost piece to the board
function addGhostPiece(fromSquare, toSquare, piece) {
    // Remove any existing ghost pieces from source and destination squares
    var $fromSquare = $('#myBoard .square-' + fromSquare);
    if ($fromSquare.length > 0) {
        $fromSquare.find('.ghost-piece').remove();
    }
    
    var $toSquare = $('#myBoard .square-' + toSquare);
    if ($toSquare.length === 0) return;
    $toSquare.find('.ghost-piece').remove();
    
    // Hide the piece on the source square
    var $sourcePiece = null;
    if ($fromSquare.length > 0) {
        $sourcePiece = $fromSquare.find('img.piece-417db');
        if ($sourcePiece.length > 0) {
            $sourcePiece.css('visibility', 'hidden');
        }
    }
    
    // Hide any piece on the destination square (for captures)
    var $destPiece = $toSquare.find('img.piece-417db');
    if ($destPiece.length > 0) {
        $destPiece.css('visibility', 'hidden');
    }
    
    // Create ghost piece image
    var pieceImage = '/assets/images/pieces/' + piece + '.png';
    var $ghostPiece = $('<img>')
        .attr('src', pieceImage)
        .addClass('ghost-piece')
        .addClass('ghost-fade-in')
        .css({
            width: '100%',
            height: '100%',
            position: 'absolute',
            top: 0,
            left: 0
        });
    
    // Add to destination square
    $toSquare.append($ghostPiece);
    
    // Track for cleanup (including both source and destination pieces for restoration)
    ghostPieces.push({
        fromSquare: fromSquare,
        toSquare: toSquare,
        element: $ghostPiece,
        sourcePiece: $sourcePiece,
        destPiece: $destPiece
    });
}

// Remove a ghost piece from a specific square
function removeGhostPiece(square) {
    var $square = $('#myBoard .square-' + square);
    $square.find('.ghost-piece').remove();
    
    // Remove from tracking array
    ghostPieces = ghostPieces.filter(function(gp) {
        return gp.square !== square;
    });
}

// Clear all ghost pieces
function clearGhostPieces() {
    for (var i = 0; i < ghostPieces.length; i++) {
        ghostPieces[i].element.remove();
        // Restore source piece visibility if it was hidden
        if (ghostPieces[i].sourcePiece && ghostPieces[i].sourcePiece.length > 0) {
            ghostPieces[i].sourcePiece.css('visibility', 'visible');
        }
        // Restore destination piece visibility if it was hidden (for captures)
        if (ghostPieces[i].destPiece && ghostPieces[i].destPiece.length > 0) {
            ghostPieces[i].destPiece.css('visibility', 'visible');
        }
    }
    ghostPieces = [];
}

// Draw principal variation (sequence of moves)
function drawPrincipalVariation(data, showPV, showBestMove, forceStart) {
    // forceStart = true when position changes (move made), false when engine sends new PV
    if (forceStart === undefined) forceStart = false;
    
    // Check if this is a new PV sequence
    var newPVSequence = data.pv.join(',');
    if (newPVSequence === currentPVSequence) {
        // Same PV, don't restart animation
        return;
    }
    
    // If showing only best move, draw immediately without animation (no ghost pieces)
    if (!showPV) {
        cancelPVAnimation();
        currentPVSequence = newPVSequence;
        drawPVArrowAtIndex(data, 0, true, false);
        return;
    }
    
    // If forceStart (move made), cancel everything and start immediately
    if (forceStart) {
        cancelPVAnimation();
        currentPVSequence = newPVSequence;
        isAnimating = true;
    } else {
        // Engine sent new PV during animation - queue it
        if (isAnimating) {
            pendingPVData = { data: data, showPV: showPV, showBestMove: showBestMove };
            return;
        }
        // No animation running, start new one
        currentPVSequence = newPVSequence;
        isAnimating = true;
    }
    
    // Determine max moves based on mode:
    // - showPV: show up to 6 moves (3 full moves) in the principal variation
    var maxMoves = Math.min(6, data.pv.length);
    
    // For PV mode, animate arrows sequentially with looping
    var firstMoveDelay = 2000; // 2 seconds for first move (with score)
    var subsequentMoveDelay = 1500; // 1.5 seconds for subsequent moves
    var pauseAfterLoop = 2000; // 2 seconds pause after last move
    
    function scheduleAnimation(loopIteration) {
        var cumulativeDelay = 0;
        
        for (var i = 0; i < maxMoves; i++) {
            (function(index) {
                var timeout = setTimeout(function() {
                    // Check if this animation is still valid
                    if (currentPVSequence === newPVSequence) {
                        // Clear previous arrows but keep ghost pieces
                        var clearPrevious = true;
                        // Only clear ghost pieces at the start of a new loop (first move)
                        var clearGhosts = (index === 0);
                        drawPVArrowAtIndex(data, index, clearPrevious, true, clearGhosts);
                        
                        // If this is the last arrow, check for pending PV or schedule next loop
                        if (index === maxMoves - 1) {
                            var finalDelay = index === 0 ? firstMoveDelay : subsequentMoveDelay;
                            var loopTimeout = setTimeout(function() {
                                if (currentPVSequence === newPVSequence) {
                                    // Check if there's a pending PV to start
                                    if (pendingPVData) {
                                        var pending = pendingPVData;
                                        pendingPVData = null;
                                        isAnimating = false;
                                        // Start the pending PV animation
                                        drawPrincipalVariation(pending.data, pending.showPV, pending.showBestMove, false);
                                    } else {
                                        // Continue looping with current PV
                                        scheduleAnimation(loopIteration + 1);
                                    }
                                }
                            }, finalDelay + pauseAfterLoop);
                            pvAnimationTimeouts.push(loopTimeout);
                        }
                    }
                }, cumulativeDelay);
                pvAnimationTimeouts.push(timeout);
                
                // Calculate delay for next arrow
                cumulativeDelay += (index === 0) ? firstMoveDelay : subsequentMoveDelay;
            })(i);
        }
    }
    
    // Start the first animation loop
    scheduleAnimation(0);
}

// Draw a single PV arrow at the specified index
function drawPVArrowAtIndex(data, index, clearPrevious, showGhostPieces, clearGhosts) {
    // Default parameters
    if (showGhostPieces === undefined) showGhostPieces = false;
    if (clearGhosts === undefined) clearGhosts = true;
    // Create a temporary game instance to track positions
    var tempGame = new Chess(game.fen());
    
    // Apply all moves up to the target index
    for (var i = 0; i <= index; i++) {
        var move = data.pv[i];
        
        // Parse UCI move format (e.g., "e2e4" or "e7e8q" for promotion)
        if (move.length < 4) return;
        
        var from = move.substring(0, 2);
        var to = move.substring(2, 4);
        var promotion = move.length > 4 ? move.substring(4, 5) : undefined;
        
        // Verify the piece exists at the from square
        var piece = tempGame.get(from);
        if (!piece) return;
        
        // If this is the arrow we want to draw
        if (i === index) {
            // Clear ghost pieces only if requested (at start of new loop)
            if (clearGhosts) {
                clearGhostPieces();
            }
            
            // Add ghost piece at destination square (only in PV mode)
            if (showGhostPieces) {
                var pieceCode = piece.color + piece.type.toUpperCase();
                addGhostPiece(from, to, pieceCode);
            }
            
            // Use different colors based on whose turn it is in the temp game
            var arrowColor = tempGame.turn() === 'w' ? '#FFFFFF' : '#000000';
            
            // Use 0.8 opacity for arrows
            var opacity = 0.8;
            
            // Calculate actual chess move number from current FEN
            var currentFenParts = tempGame.fen().split(' ');
            var currentMoveNumber = parseInt(currentFenParts[5]) || 1;
            var isBlackMove = tempGame.turn() === 'b';
            
            // Only show score label on the first arrow
            var scoreLabel = '';
            if (index === 0) {
                if (data.scoreType === 'cp' && data.score !== undefined) {
                    var score = (data.score / 100).toFixed(2);
                    scoreLabel = (data.score >= 0 ? '+' : '') + score;
                } else if (data.scoreType === 'mate' && data.score !== undefined) {
                    // Show sign for mate: positive = White mates, negative = Black mates
                    scoreLabel = (data.score >= 0 ? '+' : '') + 'M' + Math.abs(data.score);
                }
            }
            
            // Add move number to all arrows
            var moveNumberLabel = isBlackMove ? currentMoveNumber + '...' : currentMoveNumber.toString();
            
            // Draw arrow
            board.drawArrow(from, to, arrowColor, scoreLabel, opacity, clearPrevious, moveNumberLabel);
        }
        
        // Apply move to temp game for next iteration
        try {
            tempGame.move({
                from: from,
                to: to,
                promotion: promotion || 'q'
            });
        } catch (e) {
            // Invalid move, stop
            return;
        }
    }
}

// Draw multiple best moves (3 alternative first moves)
function drawMultipleBestMoves(multiPV) {
    // Define colors for the 3 best moves
    var colors = ['#15781Bff', '#FFD700ff', '#DC3545ff']; // Green (best), Yellow (middle), Red (worst)
    
    for (var i = 0; i < Math.min(3, multiPV.length); i++) {
        var pvLine = multiPV[i];
        
        if (!pvLine.moves || pvLine.moves.length === 0) continue;
        
        var move = pvLine.moves[0]; // Only show first move of each line
        
        // Parse UCI move format (e.g., "e2e4" or "e7e8q" for promotion)
        if (move.length < 4) continue;
        
        var from = move.substring(0, 2);
        var to = move.substring(2, 4);
        
        // Verify the piece exists at the from square
        var piece = game.get(from);
        if (!piece) continue;
        
        var arrowColor = colors[i];
        
        // Full opacity for all arrows
        var opacity = 1.0;
        
        // Format score label
        var scoreLabel = '';
        if (pvLine.scoreType === 'cp' && pvLine.score !== undefined) {
            var score = (pvLine.score / 100).toFixed(2);
            scoreLabel = (pvLine.score >= 0 ? '+' : '') + score;
        } else if (pvLine.scoreType === 'mate' && pvLine.score !== undefined) {
            // Show sign for mate: positive = White mates, negative = Black mates
            scoreLabel = (pvLine.score >= 0 ? '+' : '') + 'M' + Math.abs(pvLine.score);
        }
        
        // Draw arrow (clear previous only on first arrow)
        var clearPrevious = (i === 0);
        board.drawArrow(from, to, arrowColor, scoreLabel, opacity, clearPrevious, null);
    }
}
