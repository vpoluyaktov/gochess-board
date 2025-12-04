// Evaluation Graph
// Displays a history graph of evaluation scores throughout the game

var evalGraphCanvas = null;
var evalGraphCtx = null;

// Initialize the evaluation graph
function initEvalGraph() {
    evalGraphCanvas = document.getElementById('evalGraph');
    if (!evalGraphCanvas) return;
    
    evalGraphCtx = evalGraphCanvas.getContext('2d');
    
    // Handle click to navigate to move
    evalGraphCanvas.addEventListener('click', function(e) {
        if (!gameState.moveHistory || gameState.moveHistory.length === 0) return;
        
        var rect = evalGraphCanvas.getBoundingClientRect();
        var x = e.clientX - rect.left;
        var scaleX = evalGraphCanvas.width / rect.width;
        var canvasX = x * scaleX;
        
        // Calculate which move was clicked
        var totalMoves = gameState.moveHistory.length;
        if (totalMoves === 0) return;
        
        var padding = 10;
        var graphWidth = evalGraphCanvas.width - padding * 2;
        var moveIndex = Math.round((canvasX - padding) / graphWidth * (totalMoves - 1));
        moveIndex = Math.max(0, Math.min(totalMoves - 1, moveIndex));
        
        // Navigate to that position (moveIndex + 1 because position 0 is start)
        goToPosition(moveIndex + 1);
    });
    
    // Draw initial empty state
    drawEmptyGraph();
}

// Show/hide the evaluation graph (kept for compatibility but graph is always visible)
function showEvalGraph(show) {
    if (show) {
        updateEvalGraph();
    }
}

// Draw empty graph placeholder
function drawEmptyGraph() {
    if (!evalGraphCanvas || !evalGraphCtx) return;
    
    var width = evalGraphCanvas.width;
    var height = evalGraphCanvas.height;
    
    // Clear and draw background
    evalGraphCtx.fillStyle = '#e8e8e8';
    evalGraphCtx.fillRect(0, 0, width, height);
    
    // Draw center line
    var centerY = height / 2;
    evalGraphCtx.strokeStyle = '#ccc';
    evalGraphCtx.lineWidth = 1;
    evalGraphCtx.setLineDash([4, 4]);
    evalGraphCtx.beginPath();
    evalGraphCtx.moveTo(0, centerY);
    evalGraphCtx.lineTo(width, centerY);
    evalGraphCtx.stroke();
    evalGraphCtx.setLineDash([]);
}

// Clear the eval graph (called on new game or PGN load)
function clearEvalGraph() {
    drawEmptyGraph();
    updateEvalScoreDisplay(null);
}

// Update the evaluation graph with current scores
function updateEvalGraph() {
    if (!evalGraphCanvas || !evalGraphCtx) {
        initEvalGraph();
    }
    if (!evalGraphCanvas || !evalGraphCtx) return;
    
    var scores = gameState.moveScores || [];
    var totalMoves = gameState.moveHistory ? gameState.moveHistory.length : 0;
    
    // Clear canvas
    var width = evalGraphCanvas.width;
    var height = evalGraphCanvas.height;
    evalGraphCtx.clearRect(0, 0, width, height);
    
    // Draw background
    evalGraphCtx.fillStyle = '#e8e8e8';
    evalGraphCtx.fillRect(0, 0, width, height);
    
    // Draw center line (0 evaluation)
    var centerY = height / 2;
    evalGraphCtx.strokeStyle = '#999';
    evalGraphCtx.lineWidth = 1;
    evalGraphCtx.setLineDash([4, 4]);
    evalGraphCtx.beginPath();
    evalGraphCtx.moveTo(0, centerY);
    evalGraphCtx.lineTo(width, centerY);
    evalGraphCtx.stroke();
    evalGraphCtx.setLineDash([]);
    
    if (totalMoves === 0) {
        // No moves yet
        updateEvalScoreDisplay(null);
        return;
    }
    
    // Calculate points
    var padding = 10;
    var graphWidth = width - padding * 2;
    var graphHeight = height - padding * 2;
    var maxEval = 500; // 5 pawns max scale (in centipawns)
    
    var points = [];
    var lastValidScore = null;
    
    for (var i = 0; i < totalMoves; i++) {
        var x = padding + (i / Math.max(1, totalMoves - 1)) * graphWidth;
        if (totalMoves === 1) x = padding + graphWidth / 2;
        
        var scoreData = scores[i];
        var evalCp = 0;
        var isMate = false;
        
        if (scoreData !== null && scoreData !== undefined) {
            if (typeof scoreData === 'number') {
                evalCp = scoreData;
            } else if (scoreData.scoreType === 'mate') {
                // Mate score - show at max height
                // scoreData.score is moves to mate: positive = white wins, negative = black wins
                // For mate, we show at maximum evaluation
                if (scoreData.score > 0) {
                    evalCp = maxEval;  // White is winning (mate in N)
                } else if (scoreData.score < 0) {
                    evalCp = -maxEval; // Black is winning (mate in N)
                } else {
                    evalCp = maxEval;  // Checkmate (score = 0 means immediate mate)
                }
                isMate = true;
            } else {
                evalCp = scoreData.score;
            }
            lastValidScore = scoreData;
        } else if (lastValidScore !== null) {
            // Use last valid score for continuity
            if (typeof lastValidScore === 'number') {
                evalCp = lastValidScore;
            } else if (lastValidScore.scoreType === 'mate') {
                if (lastValidScore.score > 0) {
                    evalCp = maxEval;
                } else if (lastValidScore.score < 0) {
                    evalCp = -maxEval;
                } else {
                    evalCp = maxEval;
                }
            } else {
                evalCp = lastValidScore.score;
            }
        }
        
        // Clamp to max scale
        evalCp = Math.max(-maxEval, Math.min(maxEval, evalCp));
        
        // Convert to Y coordinate (positive = up = white winning)
        var y = centerY - (evalCp / maxEval) * (graphHeight / 2);
        
        points.push({ x: x, y: y, isMate: isMate, hasScore: scoreData !== null && scoreData !== undefined });
    }
    
    if (points.length === 0) {
        updateEvalScoreDisplay(null);
        return;
    }
    
    // Draw filled area
    evalGraphCtx.beginPath();
    evalGraphCtx.moveTo(points[0].x, centerY);
    
    for (var i = 0; i < points.length; i++) {
        evalGraphCtx.lineTo(points[i].x, points[i].y);
    }
    
    evalGraphCtx.lineTo(points[points.length - 1].x, centerY);
    evalGraphCtx.closePath();
    
    // Create gradient fill - white above center, dark below
    var gradient = evalGraphCtx.createLinearGradient(0, 0, 0, height);
    gradient.addColorStop(0, 'rgba(255, 255, 255, 1)');    // Pure white at top
    gradient.addColorStop(0.4, 'rgba(255, 255, 255, 0.95)'); // Still white near center
    gradient.addColorStop(0.5, 'rgba(180, 180, 180, 0.7)'); // Light gray at center
    gradient.addColorStop(0.6, 'rgba(80, 80, 80, 0.85)');  // Dark gray below center
    gradient.addColorStop(1, 'rgba(40, 40, 40, 1)');       // Dark at bottom
    evalGraphCtx.fillStyle = gradient;
    evalGraphCtx.fill();
    
    // Draw the line
    evalGraphCtx.strokeStyle = '#667eea';
    evalGraphCtx.lineWidth = 2;
    evalGraphCtx.beginPath();
    evalGraphCtx.moveTo(points[0].x, points[0].y);
    
    for (var i = 1; i < points.length; i++) {
        evalGraphCtx.lineTo(points[i].x, points[i].y);
    }
    evalGraphCtx.stroke();
    
    // Draw current position indicator
    var currentPos = gameState.currentPosition;
    if (currentPos > 0 && currentPos <= totalMoves) {
        var currentIndex = currentPos - 1;
        var currentX = padding + (currentIndex / Math.max(1, totalMoves - 1)) * graphWidth;
        if (totalMoves === 1) currentX = padding + graphWidth / 2;
        
        evalGraphCtx.strokeStyle = '#667eea';
        evalGraphCtx.lineWidth = 2;
        evalGraphCtx.beginPath();
        evalGraphCtx.moveTo(currentX, padding);
        evalGraphCtx.lineTo(currentX, height - padding);
        evalGraphCtx.stroke();
        
        // Draw dot at current position
        if (points[currentIndex]) {
            evalGraphCtx.fillStyle = '#667eea';
            evalGraphCtx.beginPath();
            evalGraphCtx.arc(currentX, points[currentIndex].y, 4, 0, Math.PI * 2);
            evalGraphCtx.fill();
        }
    }
    
    // Update score display
    var currentScore = null;
    if (currentPos > 0 && currentPos <= scores.length) {
        currentScore = scores[currentPos - 1];
    }
    updateEvalScoreDisplay(currentScore);
}

// Update the score display text
function updateEvalScoreDisplay(scoreData) {
    var scoreEl = document.getElementById('evalGraphScore');
    if (!scoreEl) return;
    
    if (scoreData === null || scoreData === undefined) {
        scoreEl.textContent = '';
        return;
    }
    
    var text = '';
    if (typeof scoreData === 'number') {
        var pawns = scoreData / 100;
        text = (pawns >= 0 ? '+' : '') + pawns.toFixed(2);
    } else if (scoreData.scoreType === 'mate') {
        text = (scoreData.score >= 0 ? '+' : '-') + 'M' + Math.abs(scoreData.score);
    } else {
        var pawns = scoreData.score / 100;
        text = (pawns >= 0 ? '+' : '') + pawns.toFixed(2);
    }
    
    scoreEl.textContent = text;
    
    // Color based on who's winning
    if (typeof scoreData === 'number') {
        scoreEl.style.color = scoreData >= 0 ? '#333' : '#333';
        scoreEl.style.background = scoreData >= 50 ? 'rgba(255,255,255,0.95)' : 
                                   scoreData <= -50 ? 'rgba(80,80,80,0.9)' : 'rgba(200,200,200,0.9)';
        scoreEl.style.color = scoreData <= -50 ? '#fff' : '#333';
    } else {
        var score = scoreData.score;
        scoreEl.style.background = score >= 50 ? 'rgba(255,255,255,0.95)' : 
                                   score <= -50 ? 'rgba(80,80,80,0.9)' : 'rgba(200,200,200,0.9)';
        scoreEl.style.color = score <= -50 ? '#fff' : '#333';
    }
}

// Initialize when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initEvalGraph);
} else {
    initEvalGraph();
}
