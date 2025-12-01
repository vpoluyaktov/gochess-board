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
    
    analysisActive = false;
    document.getElementById('analysisToggle').textContent = '🔍 Start Analysis';
    document.getElementById('analysisDepth').textContent = 'Depth: -';
    board.clearArrow();
}

// Draw principal variation (sequence of moves)
function drawPrincipalVariation(data, showPV, showBestMove) {
    // Create a temporary game instance to track positions
    var tempGame = new Chess(game.fen());
    
    // Get the starting full move number from FEN (6th field)
    var fenParts = tempGame.fen().split(' ');
    var startingMoveNumber = parseInt(fenParts[5]) || 1;
    
    // Determine max moves based on mode:
    // - showBestMove (default): show only 1 move
    // - showPV: show up to 6 moves (3 full moves) in the principal variation
    var maxMoves = showPV ? Math.min(6, data.pv.length) : 1;
    
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
            scoreLabel = 'M' + Math.abs(pvLine.score);
        }
        
        // Draw arrow (clear previous only on first arrow)
        var clearPrevious = (i === 0);
        board.drawArrow(from, to, arrowColor, scoreLabel, opacity, clearPrevious, null);
    }
}
