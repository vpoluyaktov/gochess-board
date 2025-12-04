// Live Analysis
// Handles WebSocket connection to analysis engine and arrow display

var analysisWs = null;
var analysisActive = false;

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
        
        // Verify the analysis is for the current position
        // Compare only the position part of FEN (first 4 fields), ignoring move counters
        if (data.fen) {
            var analysisFenParts = data.fen.split(' ').slice(0, 4).join(' ');
            var currentFenParts = game.fen().split(' ').slice(0, 4).join(' ');
            if (analysisFenParts !== currentFenParts) {
                // Silently discard stale analysis - this is expected during navigation
                return;
            }
        }
        
        // Update depth display in button
        if (data.depth !== undefined) {
            document.getElementById('analysisDepth').textContent = '(Depth: ' + data.depth + ')';
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

// Note: drawMultipleBestMoves is now handled inline with prepareMultiPVMoves()
// The application prepares the data, the library visualizes it
