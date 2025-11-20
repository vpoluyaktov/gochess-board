// Live Analysis
// Handles WebSocket connection to analysis engine and arrow display

var analysisWs = null;
var analysisActive = false;

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
        
        // Draw arrows for principal variation
        if (data.pv && data.pv.length > 0) {
            // Check if PV display is enabled
            var showPV = document.getElementById('showPVArrows').checked;
            
            // Create a temporary game instance to track positions
            var tempGame = new Chess(game.fen());
            
            // Get the starting full move number from FEN (6th field)
            var fenParts = tempGame.fen().split(' ');
            var startingMoveNumber = parseInt(fenParts[5]) || 1;
            
            // Limit to 3 moves if PV is enabled, otherwise just 1 (best move only)
            var maxMoves = showPV ? Math.min(3, data.pv.length) : 1;
            
            for (var i = 0; i < maxMoves; i++) {
                var move = data.pv[i];
                
                // Parse UCI move format (e.g., "e2e4" or "e7e8q" for promotion)
                if (move.length < 4) continue;
                
                var from = move.substring(0, 2);
                var to = move.substring(2, 4);
                var promotion = move.length > 4 ? move.substring(4, 5) : undefined;
                
                // Verify the piece exists at the from square
                var piece = tempGame.get(from);
                if (!piece) continue;
                
                // Use different colors based on whose turn it is in the temp game
                var arrowColor = tempGame.turn() === 'w' ? '#f4f5f7ff' : '#605e5eff';
                
                // Calculate opacity: first arrow bright, subsequent ones dimmer
                var opacity = 1.0 - (i * 0.2);
                
                // Calculate actual chess move number from current FEN
                var currentFenParts = tempGame.fen().split(' ');
                var currentMoveNumber = parseInt(currentFenParts[5]) || 1;
                var isBlackMove = tempGame.turn() === 'b';
                
                // Only show score label on the first arrow
                var scoreLabel = '';
                if (i === 0) {
                    if (data.scoreType === 'cp' && data.score !== undefined) {
                        var score = (data.score / 100).toFixed(2);
                        scoreLabel = (data.score >= 0 ? '+' : '') + score;
                    } else if (data.scoreType === 'mate' && data.score !== undefined) {
                        scoreLabel = 'M' + Math.abs(data.score);
                    }
                }
                
                // Add move number to all arrows when PV is enabled
                var moveNumberLabel = null;
                if (showPV) {
                    // Format: "18" for white, "18..." for black
                    moveNumberLabel = isBlackMove ? currentMoveNumber + '...' : currentMoveNumber.toString();
                }
                
                // Draw arrow (clear previous only on first arrow)
                var clearPrevious = (i === 0);
                board.drawArrow(from, to, arrowColor, scoreLabel, opacity, clearPrevious, moveNumberLabel);
                
                // Apply move to temp game for next iteration
                try {
                    tempGame.move({
                        from: from,
                        to: to,
                        promotion: promotion || 'q'
                    });
                } catch (e) {
                    // Invalid move, stop drawing further arrows
                    break;
                }
            }
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
    
    analysisActive = false;
    document.getElementById('analysisToggle').textContent = '🔍 Start Analysis';
    document.getElementById('analysisDepth').textContent = 'Depth: -';
    board.clearArrow();
}
