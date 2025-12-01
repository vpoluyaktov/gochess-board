// Live Analysis
// Handles WebSocket connection to analysis engine and arrow display

var analysisWs = null;
var analysisActive = false;
var pvAnimationTimeouts = []; // Track animation timeouts for cancellation
var currentPVSequence = null; // Track current PV sequence for comparison

// Initialize event listeners for analysis display mode changes
function initAnalysisDisplayListeners() {
    var radioButtons = document.querySelectorAll('input[name="analysisDisplay"]');
    radioButtons.forEach(function(radio) {
        radio.addEventListener('change', function() {
            // Cancel any ongoing PV animation when switching modes
            cancelPVAnimation();
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
            drawPrincipalVariation(data, showPV, showBestMove);
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
}

// Draw principal variation (sequence of moves)
function drawPrincipalVariation(data, showPV, showBestMove) {
    // Check if this is a new PV sequence
    var newPVSequence = data.pv.join(',');
    if (newPVSequence === currentPVSequence) {
        // Same PV, don't restart animation
        return;
    }
    
    // Cancel any ongoing animation for the previous PV
    cancelPVAnimation();
    currentPVSequence = newPVSequence;
    
    // Determine max moves based on mode:
    // - showBestMove (default): show only 1 move
    // - showPV: show up to 6 moves (3 full moves) in the principal variation
    var maxMoves = showPV ? Math.min(6, data.pv.length) : 1;
    
    // If showing only best move, draw immediately without animation
    if (!showPV) {
        drawPVArrowAtIndex(data, 0, true);
        return;
    }
    
    // For PV mode, animate arrows sequentially with looping
    var firstMoveDelay = 1500; // 1.5 seconds for first move (with score)
    var subsequentMoveDelay = 1000; // 1 second for subsequent moves
    
    function scheduleAnimation(loopIteration) {
        var cumulativeDelay = 0;
        
        for (var i = 0; i < maxMoves; i++) {
            (function(index) {
                var timeout = setTimeout(function() {
                    // Check if this animation is still valid
                    if (currentPVSequence === newPVSequence) {
                        // Always clear previous arrows to show only one at a time
                        var clearPrevious = true;
                        drawPVArrowAtIndex(data, index, clearPrevious);
                        
                        // If this is the last arrow, schedule the next loop
                        if (index === maxMoves - 1) {
                            var finalDelay = index === 0 ? firstMoveDelay : subsequentMoveDelay;
                            var loopTimeout = setTimeout(function() {
                                if (currentPVSequence === newPVSequence) {
                                    scheduleAnimation(loopIteration + 1);
                                }
                            }, finalDelay);
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
function drawPVArrowAtIndex(data, index, clearPrevious) {
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
            // Use different colors based on whose turn it is in the temp game
            var arrowColor = tempGame.turn() === 'w' ? '#f4f5f7ff' : '#605e5eff';
            
            // Use full opacity for all arrows
            var opacity = 1.0;
            
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
