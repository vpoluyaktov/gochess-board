// Opening Display
// Handles opening name lookup and display

async function updateOpeningDisplay() {
    // Get move history in SAN notation
    const sanMoves = [];
    const tempGame = new Chess();
    
    // Replay the game to get SAN notation
    // When navigating, only use moves up to currentPosition
    const movesToShow = gameState.isNavigating ? gameState.currentPosition : gameState.moveHistory.length;
    
    for (let i = 0; i < movesToShow; i++) {
        const uciMove = gameState.moveHistory[i];
        
        // Parse UCI move (e.g., "e2e4" or "e7e8q")
        const from = uciMove.substring(0, 2);
        const to = uciMove.substring(2, 4);
        const promotion = uciMove.length > 4 ? uciMove.substring(4) : undefined;
        
        // Get SAN notation before making the move
        const move = tempGame.move({
            from: from,
            to: to,
            promotion: promotion
        });
        
        if (move) {
            sanMoves.push(move.san);
        }
    }
    
    // Don't show opening info if no moves yet
    if (sanMoves.length === 0) {
        document.getElementById('openingDisplay').style.display = 'none';
        return;
    }
    
    try {
        // Call the opening API
        const response = await fetch('/api/opening', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                moves: sanMoves
            })
        });
        
        if (!response.ok) {
            console.error('Failed to fetch opening info');
            return;
        }
        
        const opening = await response.json();
        
        // Update the display (single line, no ECO)
        if (opening && opening.name) {
            document.getElementById('openingText').textContent = opening.name;
            document.getElementById('openingDisplay').style.display = 'block';
        } else {
            document.getElementById('openingDisplay').style.display = 'none';
        }
    } catch (error) {
        console.error('Error fetching opening info:', error);
    }
}
