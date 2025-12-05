// Player Configuration
// Handles player selection, dropdowns, and player-related UI

function resetPlayerDropdowns() {
    // Simply reset selection to default (Human vs Human)
    document.getElementById('whitePlayer').value = 'human';
    document.getElementById('blackPlayer').value = 'human';
}

function addPlayerToDropdown(dropdownId, playerName) {
    const dropdown = document.getElementById(dropdownId);
    
    // Check if this player name already exists
    for (let i = 0; i < dropdown.options.length; i++) {
        if (dropdown.options[i].value === playerName) {
            // Already exists, just select it
            dropdown.value = playerName;
            return;
        }
    }
    
    // Add new option at the beginning (after the first option which is "Human")
    const option = document.createElement('option');
    option.value = playerName;
    option.textContent = playerName;
    
    // Insert after "Human" option (index 1)
    if (dropdown.options.length > 1) {
        dropdown.insertBefore(option, dropdown.options[1]);
    } else {
        dropdown.appendChild(option);
    }
    
    // Select the new player
    dropdown.value = playerName;
}

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
    
    // Hide info text after first move to save space
    if (gameState.moveHistory.length > 0) {
        infoDiv.style.display = 'none';
        // Update player controls visibility
        updatePlayerControls();
        return;
    }
    
    infoDiv.style.display = 'block';
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
    
    // Update player controls visibility
    updatePlayerControls();
}

function updatePlayerControls() {
    const whitePlayer = getWhitePlayer();
    const blackPlayer = getBlackPlayer();
    
    // Show/hide white player controls
    const whiteControls = document.getElementById('whitePlayerControls');
    if (whiteControls) {
        whiteControls.style.display = isHuman(whitePlayer) ? 'flex' : 'none';
    }
    
    // Show/hide black player controls
    const blackControls = document.getElementById('blackPlayerControls');
    if (blackControls) {
        blackControls.style.display = isHuman(blackPlayer) ? 'flex' : 'none';
    }
    
    // Highlight draw buttons if threefold repetition is available
    updateDrawButtonState();
}

function updateDrawButtonState() {
    const drawButtons = document.querySelectorAll('.draw-btn');
    const canClaimDraw = game.in_threefold_repetition();
    
    drawButtons.forEach(btn => {
        if (canClaimDraw) {
            btn.classList.add('draw-available');
            btn.title = 'Threefold repetition - Claim draw!';
        } else {
            btn.classList.remove('draw-available');
            btn.title = 'Offer draw';
        }
    });
}

function restorePlayerSelections() {
    // DISABLED: Always start with Human vs Human (default dropdown values)
    // But still need to update controls visibility on page load
    updatePlayerControls();
}

function savePlayerSelections() {
    // DISABLED: Player selection persistence disabled
    updatePlayerControls();
}

// -------------------------------------------------------------------------
// Game Result Functions
// -------------------------------------------------------------------------

function resignWhite() {
    showConfirm('White resigns. Black wins!', function() {
        handleGameEnd();
        showGameOver('Game Over: Black wins by resignation');
    });
}

function resignBlack() {
    showConfirm('Black resigns. White wins!', function() {
        handleGameEnd();
        showGameOver('Game Over: White wins by resignation');
    });
}

function offerDraw() {
    // Check if draw can be claimed by threefold repetition
    if (game.in_threefold_repetition()) {
        showConfirm('Threefold repetition detected! Claim draw?', function() {
            handleGameEnd();
            showGameOver('Game Over: Draw by threefold repetition');
        });
    } else {
        showConfirm('Offer a draw?', function() {
            handleGameEnd();
            showGameOver('Game Over: Draw by agreement');
        });
    }
}
