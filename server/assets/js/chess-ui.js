// Chess UI Logic
// Handles game state, player interactions, computer moves, and live analysis

var board = null;
var game = new Chess();
var isComputerThinking = false;
var lastMoveSquares = { from: null, to: null };
var squareClass = 'square-55d63';
var lastPlayerMove = ''; // Track last move in UCI notation

// Analysis WebSocket connections
var whiteAnalysisWs = null;
var blackAnalysisWs = null;
var whiteAnalysisActive = false;
var blackAnalysisActive = false;

// Chess Clock variables
var clockInterval = null;
var clockRunning = false;
var whiteTimeMs = 300000; // 5 minutes in milliseconds
var blackTimeMs = 300000;
var currentTimeControl = { initial: 5, increment: 5 };

// -------------------------------------------------------------------------
// Last Move Highlighting
// -------------------------------------------------------------------------

function clearLastMoveHighlight() {
    $('#myBoard .' + squareClass).removeClass('highlight-last-move');
}

function highlightLastMove(from, to) {
    clearLastMoveHighlight();
    $('#myBoard .square-' + from).addClass('highlight-last-move');
    $('#myBoard .square-' + to).addClass('highlight-last-move');
    lastMoveSquares = { from: from, to: to };
}

// -------------------------------------------------------------------------
// Player Configuration
// -------------------------------------------------------------------------

function getWhitePlayer() {
    return document.getElementById('whitePlayer').value;
}

function getBlackPlayer() {
    return document.getElementById('blackPlayer').value;
}

function getCurrentPlayer() {
    return game.turn() === 'w' ? getWhitePlayer() : getBlackPlayer();
}

function isHuman(playerValue) {
    return playerValue === 'human';
}

function getPlayerName(playerValue) {
    if (playerValue === 'human') return 'Human';
    const select = document.getElementById('whitePlayer');
    const option = Array.from(select.options).find(opt => opt.value === playerValue);
    return option ? option.textContent : 'Engine';
}

function updateInfoText() {
    const whitePlayer = getWhitePlayer();
    const blackPlayer = getBlackPlayer();
    const infoDiv = document.getElementById('gameInfo');
    
    const whiteName = getPlayerName(whitePlayer);
    const blackName = getPlayerName(blackPlayer);
    
    if (isHuman(whitePlayer) && isHuman(blackPlayer)) {
        infoDiv.textContent = 'Human vs Human - Make your move!';
    } else if (!isHuman(whitePlayer) && !isHuman(blackPlayer)) {
        infoDiv.textContent = `${whiteName} vs ${blackName} - Watch the AI battle!`;
    } else if (isHuman(whitePlayer)) {
        infoDiv.textContent = `You play as White vs ${blackName}. Make your move!`;
    } else {
        infoDiv.textContent = `You play as Black vs ${whiteName}. Make your move!`;
    }
}

// -------------------------------------------------------------------------
// Chessboard Event Handlers
// -------------------------------------------------------------------------

function onDragStart(source, piece, position, orientation) {
    if (game.game_over()) return false;
    if (isComputerThinking) return false;

    const currentPlayer = getCurrentPlayer();
    if (!isHuman(currentPlayer)) return false;

    const turn = game.turn();
    if ((turn === 'w' && piece.search(/^b/) !== -1) ||
        (turn === 'b' && piece.search(/^w/) !== -1)) {
        return false;
    }
}

function onDrop(source, target) {
    var move = game.move({
        from: source,
        to: target,
        promotion: 'q'
    });

    if (move === null) return 'snapback';

    highlightLastMove(source, target);
    
    // Track the move in UCI notation
    lastPlayerMove = source + target;
    if (move.promotion) {
        lastPlayerMove += move.promotion;
    }
    
    // Start clock on first move
    if (!clockRunning && currentTimeControl.initial > 0) {
        startClock();
    }
    
    // Add increment to the player who just moved
    if (clockRunning) {
        var wasWhiteTurn = move.color === 'w';
        addIncrement(wasWhiteTurn);
        updateClockDisplay();
    }

    // Update move history display
    updateMoveHistory();

    // Clear arrow and update analyses with new position
    board.clearArrow();
    setTimeout(function() {
        if (whiteAnalysisActive && whiteAnalysisWs && whiteAnalysisWs.readyState === WebSocket.OPEN) {
            whiteAnalysisWs.send(JSON.stringify({
                action: 'update',
                fen: game.fen()
            }));
        }
        if (blackAnalysisActive && blackAnalysisWs && blackAnalysisWs.readyState === WebSocket.OPEN) {
            blackAnalysisWs.send(JSON.stringify({
                action: 'update',
                fen: game.fen()
            }));
        }
    }, 100);

    window.setTimeout(checkForComputerMove, 250);
}

function onSnapEnd() {
    board.position(game.fen());
}

function onMoveEnd() {
    if (lastMoveSquares.from && lastMoveSquares.to) {
        $('#myBoard .square-' + lastMoveSquares.from).addClass('highlight-last-move');
        $('#myBoard .square-' + lastMoveSquares.to).addClass('highlight-last-move');
    }
}

// -------------------------------------------------------------------------
// Computer Move Logic
// -------------------------------------------------------------------------

async function makeComputerMove() {
    if (isComputerThinking) return;
    
    const currentPlayer = getCurrentPlayer();
    if (isHuman(currentPlayer)) return;

    if (game.game_over()) {
        console.log('Game over');
        return;
    }

    isComputerThinking = true;
    var fen = game.fen();
    
    try {
        console.log('Making move with engine:', currentPlayer);
        const response = await fetch('/api/computer-move', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ 
                fen: fen,
                enginePath: currentPlayer,
                lastMove: lastPlayerMove
            })
        });

        const data = await response.json();
        console.log('Response:', data);
        
        if (data.error) {
            console.log('Game over or no legal moves:', data.error);
            isComputerThinking = false;
            return;
        }

        console.log('Applying move:', data.move);
        
        var moveStr = data.move;
        if (moveStr && moveStr.length >= 4) {
            var from = moveStr.substring(0, 2);
            var to = moveStr.substring(2, 4);
            highlightLastMove(from, to);
        }
        
        game.load(data.fen);
        board.position(game.fen());
        
        // Add increment for computer's move
        if (clockRunning) {
            var wasBlackTurn = game.turn() === 'w'; // After move, turn switches
            addIncrement(wasBlackTurn);
            updateClockDisplay();
        }
        
        // Update move history
        updateMoveHistory();
        
        isComputerThinking = false;
        window.setTimeout(checkForComputerMove, 250);
        
    } catch (error) {
        console.error('Error getting computer move:', error);
        isComputerThinking = false;
    }
}

function checkForComputerMove() {
    const currentPlayer = getCurrentPlayer();
    console.log('checkForComputerMove - currentPlayer:', currentPlayer, 'isHuman:', isHuman(currentPlayer));
    if (!isHuman(currentPlayer) && !game.game_over()) {
        makeComputerMove();
    }
}

// Wrap makeComputerMove to update analysis
var originalMakeComputerMove = makeComputerMove;
makeComputerMove = async function() {
    await originalMakeComputerMove();
    
    board.clearArrow();
    setTimeout(function() {
        if (whiteAnalysisActive && whiteAnalysisWs && whiteAnalysisWs.readyState === WebSocket.OPEN) {
            whiteAnalysisWs.send(JSON.stringify({
                action: 'update',
                fen: game.fen()
            }));
        }
        if (blackAnalysisActive && blackAnalysisWs && blackAnalysisWs.readyState === WebSocket.OPEN) {
            blackAnalysisWs.send(JSON.stringify({
                action: 'update',
                fen: game.fen()
            }));
        }
    }, 100);
};

// -------------------------------------------------------------------------
// Move History Update
// -------------------------------------------------------------------------

async function updateMoveHistory() {
    try {
        const response = await fetch('/api/move-history');
        const history = await response.json();
        
        const listEl = document.getElementById('moveHistoryList');
        
        if (!history || history.length === 0) {
            listEl.innerHTML = '<div class="move-history-empty">No moves yet</div>';
            return;
        }
        
        let html = '';
        history.forEach(function(entry) {
            html += '<div class="move-pair">';
            html += '<div class="move-number">' + entry.moveNumber + '.</div>';
            html += '<div class="move-white">' + (entry.white || '') + '</div>';
            html += '<div class="move-black">' + (entry.black || '') + '</div>';
            html += '</div>';
        });
        
        listEl.innerHTML = html;
        
        // Auto-scroll to bottom
        listEl.scrollTop = listEl.scrollHeight;
    } catch (error) {
        console.error('Error fetching move history:', error);
    }
}

// -------------------------------------------------------------------------
// Game Control Functions
// -------------------------------------------------------------------------

async function newGame() {
    try {
        // Reset backend state
        await fetch('/api/reset', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            }
        });
        
        // Reload the page to reset frontend
        location.reload();
    } catch (error) {
        console.error('Error resetting game:', error);
        // Fallback to just reloading
        location.reload();
    }
}

// -------------------------------------------------------------------------
// Player Selection Management
// -------------------------------------------------------------------------

function restorePlayerSelections() {
    const savedWhite = localStorage.getItem('whitePlayer');
    const savedBlack = localStorage.getItem('blackPlayer');
    
    const whiteSelect = document.getElementById('whitePlayer');
    const blackSelect = document.getElementById('blackPlayer');
    
    if (savedWhite && Array.from(whiteSelect.options).some(opt => opt.value === savedWhite)) {
        whiteSelect.value = savedWhite;
    }
    if (savedBlack && Array.from(blackSelect.options).some(opt => opt.value === savedBlack)) {
        blackSelect.value = savedBlack;
    } else {
        const firstEngine = Array.from(blackSelect.options).find(opt => opt.value !== 'human');
        if (firstEngine) {
            blackSelect.value = firstEngine.value;
        }
    }
}

function savePlayerSelections() {
    localStorage.setItem('whitePlayer', document.getElementById('whitePlayer').value);
    localStorage.setItem('blackPlayer', document.getElementById('blackPlayer').value);
}

function flipBoard() {
    board.flip();
    
    const whitePlayer = document.getElementById('whitePlayer').value;
    const blackPlayer = document.getElementById('blackPlayer').value;
    
    document.getElementById('whitePlayer').value = blackPlayer;
    document.getElementById('blackPlayer').value = whitePlayer;
    
    savePlayerSelections();
    updateInfoText();
    window.setTimeout(checkForComputerMove, 250);
}

// -------------------------------------------------------------------------
// Live Analysis - White
// -------------------------------------------------------------------------

function toggleWhiteAnalysis() {
    if (whiteAnalysisActive) {
        stopWhiteAnalysis();
    } else {
        startWhiteAnalysis();
    }
}

function startWhiteAnalysis() {
    var protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    var wsUrl = protocol + '//' + window.location.host + '/api/analysis';
    
    whiteAnalysisWs = new WebSocket(wsUrl);
    
    whiteAnalysisWs.onopen = function() {
        console.log('White Analysis WebSocket connected');
        whiteAnalysisActive = true;
        document.getElementById('whiteAnalysisToggle').textContent = 'Stop White';
        document.getElementById('whiteAnalysisInfo').style.display = 'block';
        
        whiteAnalysisWs.send(JSON.stringify({
            action: 'start',
            fen: game.fen(),
            enginePath: 'stockfish'
        }));
    };
    
    whiteAnalysisWs.onmessage = function(event) {
        var data = JSON.parse(event.data);
        
        if (data.error) {
            console.error('White Analysis error:', data.error);
            return;
        }
        
        document.getElementById('whiteAnalysisDepth').textContent = data.depth || '-';
        
        var scoreText = '';
        if (data.scoreType === 'cp') {
            var score = (data.score / 100).toFixed(2);
            scoreText = (data.score >= 0 ? '+' : '') + score;
        } else if (data.scoreType === 'mate') {
            scoreText = 'Mate in ' + Math.abs(data.score);
        }
        document.getElementById('whiteAnalysisScore').textContent = scoreText;
        
        var nodesText = data.nodes ? (data.nodes / 1000).toFixed(0) + 'k' : '-';
        document.getElementById('whiteAnalysisNodes').textContent = nodesText;
        
        if (data.bestMove && data.bestMove.length >= 4 && game.turn() === 'w') {
            var from = data.bestMove.substring(0, 2);
            var to = data.bestMove.substring(2, 4);
            var piece = game.get(from);
            if (piece && piece.color === 'w') {
                board.drawArrow(from, to, '#3296FF');
            }
        }
    };
    
    whiteAnalysisWs.onerror = function(error) {
        console.error('White WebSocket error:', error);
        stopWhiteAnalysis();
    };
    
    whiteAnalysisWs.onclose = function() {
        console.log('White Analysis WebSocket closed');
        if (whiteAnalysisActive) {
            stopWhiteAnalysis();
        }
    };
}

function stopWhiteAnalysis() {
    if (whiteAnalysisWs) {
        whiteAnalysisWs.send(JSON.stringify({ action: 'stop' }));
        whiteAnalysisWs.close();
        whiteAnalysisWs = null;
    }
    
    whiteAnalysisActive = false;
    document.getElementById('whiteAnalysisToggle').textContent = 'White Analysis';
    document.getElementById('whiteAnalysisInfo').style.display = 'none';
    
    if (!blackAnalysisActive) {
        board.clearArrow();
    }
}

// -------------------------------------------------------------------------
// Live Analysis - Black
// -------------------------------------------------------------------------

function toggleBlackAnalysis() {
    if (blackAnalysisActive) {
        stopBlackAnalysis();
    } else {
        startBlackAnalysis();
    }
}

function startBlackAnalysis() {
    var protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    var wsUrl = protocol + '//' + window.location.host + '/api/analysis';
    
    blackAnalysisWs = new WebSocket(wsUrl);
    
    blackAnalysisWs.onopen = function() {
        console.log('Black Analysis WebSocket connected');
        blackAnalysisActive = true;
        document.getElementById('blackAnalysisToggle').textContent = 'Stop Black';
        document.getElementById('blackAnalysisInfo').style.display = 'block';
        
        blackAnalysisWs.send(JSON.stringify({
            action: 'start',
            fen: game.fen(),
            enginePath: 'stockfish'
        }));
    };
    
    blackAnalysisWs.onmessage = function(event) {
        var data = JSON.parse(event.data);
        
        if (data.error) {
            console.error('Black Analysis error:', data.error);
            return;
        }
        
        document.getElementById('blackAnalysisDepth').textContent = data.depth || '-';
        
        var scoreText = '';
        if (data.scoreType === 'cp') {
            var score = (data.score / 100).toFixed(2);
            scoreText = (data.score >= 0 ? '+' : '') + score;
        } else if (data.scoreType === 'mate') {
            scoreText = 'Mate in ' + Math.abs(data.score);
        }
        document.getElementById('blackAnalysisScore').textContent = scoreText;
        
        var nodesText = data.nodes ? (data.nodes / 1000).toFixed(0) + 'k' : '-';
        document.getElementById('blackAnalysisNodes').textContent = nodesText;
        
        if (data.bestMove && data.bestMove.length >= 4 && game.turn() === 'b') {
            var from = data.bestMove.substring(0, 2);
            var to = data.bestMove.substring(2, 4);
            var piece = game.get(from);
            if (piece && piece.color === 'b') {
                board.drawArrow(from, to, '#FF6B6B');
            }
        }
    };
    
    blackAnalysisWs.onerror = function(error) {
        console.error('Black WebSocket error:', error);
        stopBlackAnalysis();
    };
    
    blackAnalysisWs.onclose = function() {
        console.log('Black Analysis WebSocket closed');
        if (blackAnalysisActive) {
            stopBlackAnalysis();
        }
    };
}

function stopBlackAnalysis() {
    if (blackAnalysisWs) {
        blackAnalysisWs.send(JSON.stringify({ action: 'stop' }));
        blackAnalysisWs.close();
        blackAnalysisWs = null;
    }
    
    blackAnalysisActive = false;
    document.getElementById('blackAnalysisToggle').textContent = 'Black Analysis';
    document.getElementById('blackAnalysisInfo').style.display = 'none';
    
    if (!whiteAnalysisActive) {
        board.clearArrow();
    }
}

// -------------------------------------------------------------------------
// Chess Clock Functions
// -------------------------------------------------------------------------

function formatTime(milliseconds) {
    if (milliseconds < 0) milliseconds = 0;
    
    var totalSeconds = Math.floor(milliseconds / 1000);
    var minutes = Math.floor(totalSeconds / 60);
    var seconds = totalSeconds % 60;
    
    if (minutes >= 60) {
        var hours = Math.floor(minutes / 60);
        minutes = minutes % 60;
        return hours + ':' + pad(minutes) + ':' + pad(seconds);
    }
    
    return minutes + ':' + pad(seconds);
}

function pad(num) {
    return num < 10 ? '0' + num : num;
}

function updateClockDisplay() {
    document.getElementById('whiteClockTime').textContent = formatTime(whiteTimeMs);
    document.getElementById('blackClockTime').textContent = formatTime(blackTimeMs);
    
    // Update active clock styling (now using clock-compact class)
    var whiteClock = document.getElementById('whiteClock');
    var blackClock = document.getElementById('blackClock');
    
    if (clockRunning) {
        if (game.turn() === 'w') {
            whiteClock.classList.add('active');
            blackClock.classList.remove('active');
        } else {
            blackClock.classList.add('active');
            whiteClock.classList.remove('active');
        }
    } else {
        whiteClock.classList.remove('active');
        blackClock.classList.remove('active');
    }
    
    // Add time warnings
    whiteClock.classList.remove('time-low', 'time-critical');
    blackClock.classList.remove('time-low', 'time-critical');
    
    if (whiteTimeMs < 60000 && whiteTimeMs > 10000) {
        whiteClock.classList.add('time-low');
    } else if (whiteTimeMs <= 10000) {
        whiteClock.classList.add('time-critical');
    }
    
    if (blackTimeMs < 60000 && blackTimeMs > 10000) {
        blackClock.classList.add('time-low');
    } else if (blackTimeMs <= 10000) {
        blackClock.classList.add('time-critical');
    }
    
    // Check for time out
    if (whiteTimeMs <= 0) {
        stopClock();
        alert('Time out! Black wins!');
    } else if (blackTimeMs <= 0) {
        stopClock();
        alert('Time out! White wins!');
    }
}

function startClockInterval() {
    if (clockInterval) return; // Already running
    
    // Start the interval
    clockInterval = setInterval(function() {
        var elapsed = 100; // 100ms
        
        if (game.turn() === 'w') {
            whiteTimeMs -= elapsed;
        } else {
            blackTimeMs -= elapsed;
        }
        
        updateClockDisplay();
    }, 100); // Update every 100ms for smooth display
}

function startClock() {
    if (clockRunning) return; // Already running
    
    clockRunning = true;
    
    // Notify server
    fetch('/api/clock/start', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        }
    }).catch(err => console.error('Error starting clock:', err));
    
    startClockInterval();
}

function stopClock() {
    if (!clockRunning) return;
    
    clockRunning = false;
    if (clockInterval) {
        clearInterval(clockInterval);
        clockInterval = null;
    }
    updateClockDisplay();
}

function addIncrement(isWhite) {
    var incrementMs = currentTimeControl.increment * 1000;
    if (isWhite) {
        whiteTimeMs += incrementMs;
    } else {
        blackTimeMs += incrementMs;
    }
}

function setTimeControl(initialMinutes, incrementSeconds) {
    currentTimeControl = { initial: initialMinutes, increment: incrementSeconds };
    whiteTimeMs = initialMinutes * 60 * 1000;
    blackTimeMs = initialMinutes * 60 * 1000;
    
    // Send to backend
    fetch('/api/clock/set', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            initialMinutes: initialMinutes,
            incrementSeconds: incrementSeconds
        })
    }).catch(err => console.error('Error setting time control:', err));
    
    updateClockDisplay();
}

function syncClockWithServer() {
    fetch('/api/clock/get')
        .then(response => response.json())
        .then(data => {
            whiteTimeMs = data.whiteTimeLeft;
            blackTimeMs = data.blackTimeLeft;
            
            // Sync clock running state
            if (data.clockRunning && !clockRunning) {
                // Backend started the clock, start frontend too
                clockRunning = true;
                startClockInterval();
            } else if (!data.clockRunning && clockRunning) {
                // Backend stopped the clock
                stopClock();
            }
            
            updateClockDisplay();
        })
        .catch(err => console.error('Error syncing clock:', err));
}

// -------------------------------------------------------------------------
// Initialization
// -------------------------------------------------------------------------

$(document).ready(function() {
    // Initialize chessboard
    var config = {
        draggable: true,
        position: 'start',
        onDragStart: onDragStart,
        onDrop: onDrop,
        onSnapEnd: onSnapEnd,
        onMoveEnd: onMoveEnd,
        pieceTheme: '/assets/chess/pieces/{piece}.png'
    };
    
    board = Chessboard('myBoard', config);
    
    // Make board responsive
    $(window).resize(function() {
        board.resize();
    });
    
    // Update move history every 2 seconds
    setInterval(updateMoveHistory, 2000);
    updateMoveHistory();
    
    // Player selection event handlers
    document.getElementById('whitePlayer').addEventListener('change', function() {
        savePlayerSelections();
        updateInfoText();
        window.setTimeout(checkForComputerMove, 250);
    });

    document.getElementById('blackPlayer').addEventListener('change', function() {
        savePlayerSelections();
        updateInfoText();
        window.setTimeout(checkForComputerMove, 250);
    });
    
    // Time control selector
    document.getElementById('timeControl').addEventListener('change', function() {
        var value = this.value;
        var parts = value.split('+');
        var initialMinutes = parseInt(parts[0]);
        var incrementSeconds = parseInt(parts[1]);
        
        setTimeControl(initialMinutes, incrementSeconds);
        stopClock();
        
        // Save preference
        localStorage.setItem('timeControl', value);
    });
    
    // Restore time control preference
    var savedTimeControl = localStorage.getItem('timeControl');
    if (savedTimeControl) {
        var select = document.getElementById('timeControl');
        if (Array.from(select.options).some(opt => opt.value === savedTimeControl)) {
            select.value = savedTimeControl;
            var parts = savedTimeControl.split('+');
            setTimeControl(parseInt(parts[0]), parseInt(parts[1]));
        }
    }
    
    // Initialize
    restorePlayerSelections();
    updateInfoText();
    updateClockDisplay();
    
    // Sync clock with server periodically (every 2 seconds)
    setInterval(syncClockWithServer, 2000);
    syncClockWithServer(); // Initial sync
    
    window.setTimeout(checkForComputerMove, 500);
});
