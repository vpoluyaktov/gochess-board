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
    var prevWhite = document.getElementById('whiteClockTime').textContent;
    var prevBlack = document.getElementById('blackClockTime').textContent;
    var newWhite = formatTime(gameState.whiteTimeMs);
    var newBlack = formatTime(gameState.blackTimeMs);
    
    // Log significant changes (more than 2 seconds jump)
    var prevWhiteMs = parseTimeToMs(prevWhite);
    var prevBlackMs = parseTimeToMs(prevBlack);
    
    if (prevWhiteMs !== null && Math.abs(gameState.whiteTimeMs - prevWhiteMs) > 2000) {
        Logger.clock.debug('WHITE TIME JUMP DETECTED', {
            prevDisplay: prevWhite,
            prevMs: prevWhiteMs,
            newMs: gameState.whiteTimeMs,
            newDisplay: newWhite,
            jumpMs: gameState.whiteTimeMs - prevWhiteMs,
            turn: game.turn(),
            clockRunning: gameState.clockRunning,
            timeControl: gameState.timeControl
        });
    }
    
    if (prevBlackMs !== null && Math.abs(gameState.blackTimeMs - prevBlackMs) > 2000) {
        Logger.clock.debug('BLACK TIME JUMP DETECTED', {
            prevDisplay: prevBlack,
            prevMs: prevBlackMs,
            newMs: gameState.blackTimeMs,
            newDisplay: newBlack,
            jumpMs: gameState.blackTimeMs - prevBlackMs,
            turn: game.turn(),
            clockRunning: gameState.clockRunning,
            timeControl: gameState.timeControl
        });
    }
    
    document.getElementById('whiteClockTime').textContent = newWhite;
    document.getElementById('blackClockTime').textContent = newBlack;
    
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
        // Note: This is a backup check - the interval now handles timeout immediately
        if (gameState.whiteTimeMs <= 0) {
            handleGameEnd();
            showTimeout('Time out! Black wins!');
        } else if (gameState.blackTimeMs <= 0) {
            handleGameEnd();
            showTimeout('Time out! White wins!');
        }
    }
}

// Helper to parse display time back to ms for comparison
function parseTimeToMs(timeStr) {
    if (!timeStr || timeStr === '--:--') return null;
    var parts = timeStr.split(':');
    if (parts.length === 2) {
        return (parseInt(parts[0]) * 60 + parseInt(parts[1])) * 1000;
    } else if (parts.length === 3) {
        return (parseInt(parts[0]) * 3600 + parseInt(parts[1]) * 60 + parseInt(parts[2])) * 1000;
    }
    return null;
}

function startClockInterval() {
    if (gameState.clockInterval) {
        Logger.clock.debug('startClockInterval called but interval already exists', {
            intervalId: gameState.clockInterval
        });
        return; // Already running
    }
    
    gameState.lastClockUpdate = Date.now();
    Logger.clock.debug('Starting clock interval', {
        lastClockUpdate: gameState.lastClockUpdate,
        whiteTimeMs: gameState.whiteTimeMs,
        blackTimeMs: gameState.blackTimeMs,
        timeControl: gameState.timeControl,
        turn: game.turn()
    });
    
    // Start the interval
    gameState.clockInterval = setInterval(function() {
        var now = Date.now();
        var elapsed = now - gameState.lastClockUpdate;
        
        // Detect abnormal elapsed time (more than 500ms between 100ms intervals)
        // This is normal when browser tab is in background (throttled to ~1s)
        // Only log at DEBUG level since it's expected behavior
        if (elapsed > 500) {
            Logger.clock.debug('Tab was throttled - catching up', {
                elapsed: elapsed,
                whiteTimeMs: gameState.whiteTimeMs,
                blackTimeMs: gameState.blackTimeMs
            });
        }
        
        gameState.lastClockUpdate = now;
        
        // Check if we're in unlimited mode (stopwatch) or timed mode (countdown)
        var isUnlimitedMode = gameState.timeControl.initial === 0;
        var turn = game.turn();
        var prevWhite = gameState.whiteTimeMs;
        var prevBlack = gameState.blackTimeMs;
        
        if (turn === 'w') {
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
        
        // Check for timeout IMMEDIATELY after decrementing (only in timed mode)
        if (!isUnlimitedMode) {
            if (gameState.whiteTimeMs <= 0) {
                gameState.whiteTimeMs = 0; // Clamp to zero
                Logger.clock.info('WHITE TIMEOUT - stopping clock immediately');
                // Clear interval first before handleGameEnd tries to stop clock
                clearInterval(gameState.clockInterval);
                gameState.clockInterval = null;
                gameState.clockRunning = false;
                // Now do full game end cleanup (analysis, buttons, etc.)
                handleGameEnd();
                updateClockDisplay();
                showTimeout('Time out! Black wins!');
                return; // Exit interval callback
            } else if (gameState.blackTimeMs <= 0) {
                gameState.blackTimeMs = 0; // Clamp to zero
                Logger.clock.info('BLACK TIMEOUT - stopping clock immediately');
                // Clear interval first before handleGameEnd tries to stop clock
                clearInterval(gameState.clockInterval);
                gameState.clockInterval = null;
                gameState.clockRunning = false;
                // Now do full game end cleanup (analysis, buttons, etc.)
                handleGameEnd();
                updateClockDisplay();
                showTimeout('Time out! White wins!');
                return; // Exit interval callback
            }
        }
        
        // Log clock tick at TRACE level (very verbose)
        Logger.clock.trace('Clock tick', {
            turn: turn,
            elapsed: elapsed,
            isUnlimitedMode: isUnlimitedMode,
            whiteTimeMs: gameState.whiteTimeMs,
            blackTimeMs: gameState.blackTimeMs
        });
        
        updateClockDisplay();
        saveGameState();
    }, 100); // Update every 100ms for smooth display
    
    Logger.clock.debug('Clock interval started', { intervalId: gameState.clockInterval });
}

function startClock() {
    Logger.clock.debug('startClock called', {
        clockRunning: gameState.clockRunning,
        whiteTimeMs: gameState.whiteTimeMs,
        blackTimeMs: gameState.blackTimeMs,
        timeControl: gameState.timeControl,
        turn: game.turn()
    });
    
    if (gameState.clockRunning) {
        Logger.clock.debug('Clock already running, returning early');
        return; // Already running
    }
    
    gameState.clockRunning = true;
    startClockInterval();
    updateClockDisplay();
    updateStartPauseButton();
    saveGameState();
    
    Logger.clock.info('Clock started', {
        whiteTimeMs: gameState.whiteTimeMs,
        blackTimeMs: gameState.blackTimeMs
    });
    
    // Resume computer play if it's a computer's turn
    window.setTimeout(checkForComputerMove, 250);
}

function stopClock() {
    Logger.clock.debug('stopClock called', {
        clockRunning: gameState.clockRunning,
        intervalId: gameState.clockInterval,
        whiteTimeMs: gameState.whiteTimeMs,
        blackTimeMs: gameState.blackTimeMs
    });
    
    if (!gameState.clockRunning) {
        Logger.clock.debug('Clock not running, returning early');
        return;
    }
    
    gameState.clockRunning = false;
    if (gameState.clockInterval) {
        clearInterval(gameState.clockInterval);
        Logger.clock.debug('Cleared interval', { intervalId: gameState.clockInterval });
        gameState.clockInterval = null;
    }
    updateClockDisplay();
    saveGameState();
    
    Logger.clock.info('Clock stopped');
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
    var prevTimeControl = JSON.stringify(gameState.timeControl);
    var prevWhite = gameState.whiteTimeMs;
    var prevBlack = gameState.blackTimeMs;
    
    gameState.timeControl = { initial: initialMinutes, increment: incrementSeconds };
    gameState.whiteTimeMs = initialMinutes * 60 * 1000;
    gameState.blackTimeMs = initialMinutes * 60 * 1000;
    
    Logger.clock.info('Time control changed', {
        prevTimeControl: prevTimeControl,
        newTimeControl: gameState.timeControl,
        prevWhiteMs: prevWhite,
        prevBlackMs: prevBlack,
        newWhiteMs: gameState.whiteTimeMs,
        newBlackMs: gameState.blackTimeMs
    });
    
    stopClock();
    updateClockDisplay();
    saveGameState();
}
