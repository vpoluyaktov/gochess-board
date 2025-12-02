// Chess Clock Functions
// Handles time control, clock display, and clock management

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
    document.getElementById('whiteClockTime').textContent = formatTime(gameState.whiteTimeMs);
    document.getElementById('blackClockTime').textContent = formatTime(gameState.blackTimeMs);
    
    // Update active clock styling - always show whose turn it is
    var whiteClock = document.getElementById('whiteClock');
    var blackClock = document.getElementById('blackClock');
    
    // Always indicate whose turn it is, regardless of clock running state
    if (game.turn() === 'w') {
        whiteClock.classList.add('active');
        blackClock.classList.remove('active');
    } else {
        blackClock.classList.add('active');
        whiteClock.classList.remove('active');
    }
    
    // Add time warnings - always show, not just when clock is running
    whiteClock.classList.remove('time-low', 'time-critical');
    blackClock.classList.remove('time-low', 'time-critical');
    
    // Show time warnings only for timed games (not unlimited/stopwatch mode)
    var isUnlimitedMode = gameState.timeControl.initial === 0;
    
    if (!isUnlimitedMode) {
        if (gameState.whiteTimeMs < 60000 && gameState.whiteTimeMs > 10000) {
            whiteClock.classList.add('time-low');
        } else if (gameState.whiteTimeMs <= 10000) {
            whiteClock.classList.add('time-critical');
        }
        
        if (gameState.blackTimeMs < 60000 && gameState.blackTimeMs > 10000) {
            blackClock.classList.add('time-low');
        } else if (gameState.blackTimeMs <= 10000) {
            blackClock.classList.add('time-critical');
        }
        
        // Check for time out (only in timed mode)
        if (gameState.whiteTimeMs <= 0) {
            stopClock();
            if (analysisActive) {
                stopAnalysis();
            }
            showTimeout('Time out! Black wins!');
        } else if (gameState.blackTimeMs <= 0) {
            stopClock();
            if (analysisActive) {
                stopAnalysis();
            }
            showTimeout('Time out! White wins!');
        }
    }
}

function startClockInterval() {
    if (gameState.clockInterval) return; // Already running
    
    gameState.lastClockUpdate = Date.now();
    
    // Start the interval
    gameState.clockInterval = setInterval(function() {
        var now = Date.now();
        var elapsed = now - gameState.lastClockUpdate;
        gameState.lastClockUpdate = now;
        
        // Check if we're in unlimited mode (stopwatch) or timed mode (countdown)
        var isUnlimitedMode = gameState.timeControl.initial === 0;
        
        if (game.turn() === 'w') {
            if (isUnlimitedMode) {
                gameState.whiteTimeMs += elapsed; // Count UP for stopwatch
            } else {
                gameState.whiteTimeMs -= elapsed; // Count DOWN for timer
            }
        } else {
            if (isUnlimitedMode) {
                gameState.blackTimeMs += elapsed; // Count UP for stopwatch
            } else {
                gameState.blackTimeMs -= elapsed; // Count DOWN for timer
            }
        }
        
        updateClockDisplay();
        saveGameState();
    }, 100); // Update every 100ms for smooth display
}

function startClock() {
    if (gameState.clockRunning) return; // Already running
    
    gameState.clockRunning = true;
    startClockInterval();
    updateClockDisplay();
    updateStartPauseButton();
    saveGameState();
    
    // Resume computer play if it's a computer's turn
    window.setTimeout(checkForComputerMove, 250);
}

function stopClock() {
    if (!gameState.clockRunning) return;
    
    gameState.clockRunning = false;
    if (gameState.clockInterval) {
        clearInterval(gameState.clockInterval);
        gameState.clockInterval = null;
    }
    updateClockDisplay();
    saveGameState();
}

function toggleClock() {
    if (gameState.clockRunning) {
        pauseClock();
    } else {
        startClock();
    }
}

function pauseClock() {
    stopClock();
    updateStartPauseButton();
    saveGameState();
}

function updateStartPauseButton() {
    const btn = document.getElementById('startPauseBtn');
    if (gameState.clockRunning) {
        btn.textContent = '⏸️ Pause Game';
        btn.style.background = '#ff9800';
    } else {
        btn.textContent = '▶️ Start Game';
        btn.style.background = '#4caf50';
    }
}

function setTimeControl(initialMinutes, incrementSeconds) {
    gameState.timeControl = { initial: initialMinutes, increment: incrementSeconds };
    gameState.whiteTimeMs = initialMinutes * 60 * 1000;
    gameState.blackTimeMs = initialMinutes * 60 * 1000;
    stopClock();
    updateClockDisplay();
    saveGameState();
}
