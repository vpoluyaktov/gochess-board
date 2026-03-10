// PGN Loading and Saving
// Handles PGN file operations, parsing, game selection, and clipboard paste

// Count total games in PGN file (fast, doesn't parse content)
function countPGNGames(pgnText) {
    let count = 0;
    let inMoves = false;
    const lines = pgnText.split('\n');
    
    for (let i = 0; i < lines.length; i++) {
        const line = lines[i].trim();
        
        if (line.startsWith('[') && line.endsWith(']')) {
            if (inMoves) {
                // New game started
                count++;
                inMoves = false;
            }
        } else if (line && !line.startsWith('[')) {
            inMoves = true;
        }
    }
    
    // Count the last game
    if (inMoves) {
        count++;
    }
    
    return count;
}

// Parse PGN file and extract games (up to maxGames limit)
function parsePGNGames(pgnText, maxGames = Infinity) {
    const games = [];
    
    // Split by empty lines to find game boundaries
    // Each game starts with headers and ends with moves
    const lines = pgnText.split('\n');
    let currentGame = { headers: {}, moves: '' };
    let inHeaders = false;
    let inMoves = false;
    
    for (let i = 0; i < lines.length; i++) {
        const line = lines[i].trim();
        
        // Check if this is a header line
        if (line.startsWith('[') && line.endsWith(']')) {
            if (!inHeaders && inMoves) {
                // We've hit a new game, save the previous one
                if (currentGame.moves.trim()) {
                    games.push(currentGame);
                    
                    // Stop parsing if we've reached the limit
                    if (games.length >= maxGames) {
                        return games;
                    }
                }
                currentGame = { headers: {}, moves: '' };
                inMoves = false;
            }
            inHeaders = true;
            
            // Parse header
            const match = line.match(/\[(\w+)\s+"([^"]*)"\]/);
            if (match) {
                currentGame.headers[match[1]] = match[2];
            }
        } else if (line && !line.startsWith('[')) {
            // This is a move line
            inHeaders = false;
            inMoves = true;
            currentGame.moves += line + ' ';
        } else if (!line && inMoves) {
            // Empty line after moves might indicate end of game
            continue;
        }
    }
    
    // Don't forget the last game (if under limit)
    if (currentGame.moves.trim() && games.length < maxGames) {
        games.push(currentGame);
    }
    
    return games;
}

// Extract game metadata for display
function extractGameMetadata(game, index) {
    const headers = game.headers;
    
    // Extract first few moves for opening preview and API lookup
    const movesOnly = game.moves
        .replace(/\{[^}]*\}/g, '')  // Remove comments
        .replace(/\([^)]*\)/g, '')  // Remove variations
        .replace(/[!?]+/g, '')      // Remove annotations
        .replace(/\s*(1-0|0-1|1\/2-1\/2|\*)\s*$/, ''); // Remove result
    
    // Split by move numbers and extract moves
    // Handles both "1. e4 e5" and "1.e4 e5" formats
    const moves = [];
    const tokens = movesOnly.split(/\d+\./);  // Split by move numbers
    
    for (let i = 1; i < tokens.length && moves.length < 30; i++) {
        const moveText = tokens[i].trim();
        if (!moveText) continue;
        
        // Split the move text into individual moves (white and black)
        const moveParts = moveText.split(/\s+/);
        for (const move of moveParts) {
            if (move && !move.match(/^[!?]+$/) && !move.match(/^\d/)) {
                moves.push(move);
                if (moves.length >= 30) break;
            }
        }
    }
    
    // Try to get opening from headers first (only human-readable names, not ECO codes)
    let opening = headers.Opening || null;
    
    // Store moves for API lookup (API will find the deepest match)
    const movesForLookup = moves.slice(0, 15);
    
    return {
        index: index,
        white: headers.White || 'Unknown',
        black: headers.Black || 'Unknown',
        result: headers.Result || '*',
        date: headers.Date || '????.??.??',
        event: headers.Event || 'Unknown',
        opening: opening,
        moves: movesForLookup,
        needsLookup: !opening,  // Lookup if no Opening header found
        fullPGN: game
    };
}

// Show game selector modal
async function showPGNGameSelector(games, wasLimited, totalGames) {
    const modal = document.getElementById('pgnGameSelectorModal');
    const gameCount = document.getElementById('gameCount');
    const tableBody = document.getElementById('gameTableBody');
    
    // Clear any analysis arrows
    board.clearArrow();
    
    // Update game count with limit info if applicable
    if (wasLimited) {
        gameCount.innerHTML = `<strong>${games.length}</strong> (limited from ${totalGames} total)`;
    } else {
        gameCount.textContent = games.length;
    }
    
    // Clear existing rows
    tableBody.innerHTML = '';
    
    // Populate table with games (initial display)
    games.forEach((gameData, index) => {
        const row = document.createElement('tr');
        row.id = 'game-row-' + index;
        row.onclick = function() {
            selectPGNGame(gameData.fullPGN);
        };
        
        const openingDisplay = gameData.opening || '<span style="color: #999;">Loading...</span>';
        
        row.innerHTML = `
            <td>${index + 1}</td>
            <td>${gameData.white}</td>
            <td>${gameData.black}</td>
            <td>${gameData.result}</td>
            <td>${gameData.date}</td>
            <td>${gameData.event}</td>
            <td id="opening-cell-${index}" style="font-family: 'Courier New', monospace; font-size: 12px;">${openingDisplay}</td>
        `;
        
        tableBody.appendChild(row);
    });
    
    // Show modal
    modal.style.display = 'flex';
    
    // Check if we need to lookup openings
    const gamesToLookup = games.filter(g => g.needsLookup);
    
    if (gamesToLookup.length > 0) {
        // Start lookup after a short delay (2 seconds) to avoid showing progress bar for fast loads
        const startTime = Date.now();
        await lookupOpeningNames(games, startTime);
    }
}

// Lookup opening names via API
async function lookupOpeningNames(games, startTime) {
    const loadingBar = document.getElementById('openingLoadingBar');
    const loadingProgress = document.getElementById('loadingProgress');
    const loadingBarFill = document.getElementById('loadingBarFill');
    
    const gamesToLookup = games.filter(g => g.needsLookup);
    const totalToLookup = gamesToLookup.length;
    let completed = 0;
    
    // Show progress bar immediately
    loadingBar.style.display = 'block';
    loadingProgress.textContent = `${completed}/${totalToLookup}`;
    
    // Lookup openings in batches to avoid overwhelming the server
    const batchSize = 5;
    for (let i = 0; i < gamesToLookup.length; i += batchSize) {
        const batch = gamesToLookup.slice(i, i + batchSize);
        
        await Promise.all(batch.map(async (gameData) => {
            try {
                const response = await fetch('/api/opening', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        moves: gameData.moves
                    })
                });
                
                if (response.ok) {
                    const opening = await response.json();
                    
                    if (opening && opening.name && opening.name.trim() !== '') {
                        gameData.opening = opening.name;
                        
                        // Update the cell
                        const cell = document.getElementById('opening-cell-' + gameData.index);
                        if (cell) {
                            cell.textContent = opening.name;
                        }
                    } else {
                        // No opening found, show first few moves
                        gameData.opening = gameData.moves.slice(0, 6).join(' ') || '-';
                        const cell = document.getElementById('opening-cell-' + gameData.index);
                        if (cell) {
                            cell.textContent = gameData.opening;
                        }
                    }
                }
            } catch (error) {
                Logger.game.error('Error fetching opening for game', { index: gameData.index, error: error.message });
                // Show moves as fallback
                gameData.opening = gameData.moves.slice(0, 6).join(' ') || '-';
                const cell = document.getElementById('opening-cell-' + gameData.index);
                if (cell) {
                    cell.textContent = gameData.opening;
                }
            }
            
            completed++;
            
            // Update progress
            loadingProgress.textContent = `${completed}/${totalToLookup}`;
            const percentage = (completed / totalToLookup) * 100;
            loadingBarFill.style.width = percentage + '%';
        }));
    }
    
    // Hide progress bar after completion
    if (loadingBar.style.display === 'block') {
        setTimeout(() => {
            loadingBar.style.display = 'none';
        }, 500);
    }
}

// Close game selector modal
function closePGNGameSelector() {
    const modal = document.getElementById('pgnGameSelectorModal');
    modal.style.display = 'none';
}

// Select a game from the modal
function selectPGNGame(game) {
    closePGNGameSelector();
    
    // Combine headers and moves for loading
    let pgnText = '';
    for (const [key, value] of Object.entries(game.headers)) {
        pgnText += `[${key} "${value}"]\n`;
    }
    pgnText += '\n' + game.moves;
    
    loadPGNFromText(pgnText);
}

// Load PGN from file
function loadPGNFile() {
    const fileInput = document.getElementById('pgnFileInput');
    
    // Set up the file input handler
    fileInput.onchange = function(e) {
        const file = e.target.files[0];
        if (!file) {
            return;
        }
        
        // Stop any running analysis and clear arrows
        if (analysisActive) {
            stopAnalysis();
        }
        board.clearArrow();
        
        const reader = new FileReader();
        reader.onload = function(event) {
            const text = event.target.result;
            
            // First, count total games (fast operation)
            const totalGames = countPGNGames(text);
            
            if (totalGames === 0) {
                showDialog('No valid games found in the PGN file!', 'warning');
                return;
            } else if (totalGames === 1) {
                // Single game, load it directly
                loadPGNFromText(text);
            } else {
                // Limit to first 1000 games for memory protection
                const MAX_GAMES = 1000;
                const wasLimited = totalGames > MAX_GAMES;
                
                // Show warning if file was truncated
                if (wasLimited) {
                    showDialog(`This PGN file contains ${totalGames} games.\n\nFor memory protection, only the first ${MAX_GAMES} games will be loaded.`, 'warning');
                }
                
                // Parse only up to the limit (memory efficient)
                const games = parsePGNGames(text, MAX_GAMES);
                
                // Multiple games, show selector
                const gamesMetadata = games.map((game, index) => 
                    extractGameMetadata(game, index)
                );
                showPGNGameSelector(gamesMetadata, wasLimited, totalGames);
            }
        };
        
        reader.onerror = function() {
            showDialog('Failed to read file', 'error');
        };
        
        reader.readAsText(file);
        
        // Reset the input so the same file can be loaded again
        fileInput.value = '';
    };
    
    // Trigger the file picker
    fileInput.click();
}

// Build standard PGN format with variants in parentheses
function buildStandardPGN() {
    const tempGame = new Chess();
    
    function uciToSan(uciMove, game) {
        const from = uciMove.substring(0, 2);
        const to = uciMove.substring(2, 4);
        const promotion = uciMove.length > 4 ? uciMove.substring(4) : undefined;
        const move = game.move({ from, to, promotion });
        return move ? move.san : null;
    }
    
    function buildVariantPGN(variantMoves, startPosition, startGame) {
        const variantGame = new Chess(startGame.fen());
        let pgnParts = [];
        
        for (let i = 0; i < variantMoves.length; i++) {
            const moveIndex = startPosition + i;
            const isWhiteMove = moveIndex % 2 === 0;
            const moveNumber = Math.floor(moveIndex / 2) + 1;
            
            const san = uciToSan(variantMoves[i], variantGame);
            if (!san) break;
            
            if (isWhiteMove) {
                pgnParts.push(moveNumber + '.' + san);
            } else {
                if (i === 0) {
                    // First move is black's, need move number with ...
                    pgnParts.push(moveNumber + '...' + san);
                } else {
                    pgnParts.push(san);
                }
            }
            
            // Check for sub-variants at this position within the variant
            const subVariantPosition = moveIndex;
            if (gameState.variants[subVariantPosition]) {
                for (const subVariant of gameState.variants[subVariantPosition]) {
                    // Build sub-variant recursively
                    const subVariantGame = new Chess(startGame.fen());
                    // Replay variant moves up to this point
                    for (let j = 0; j <= i; j++) {
                        uciToSan(variantMoves[j], subVariantGame);
                    }
                    pgnParts.push(buildVariantPGN(subVariant, subVariantPosition, subVariantGame));
                }
            }
        }
        
        return '(' + pgnParts.join(' ') + ')';
    }
    
    // Build main line PGN
    let pgnParts = [];
    let lastWasVariant = false;
    
    for (let i = 0; i < gameState.moveHistory.length; i++) {
        const isWhiteMove = i % 2 === 0;
        const moveNumber = Math.floor(i / 2) + 1;
        
        const san = uciToSan(gameState.moveHistory[i], tempGame);
        if (!san) break;
        
        if (isWhiteMove) {
            pgnParts.push(moveNumber + '.' + san);
            lastWasVariant = false;
        } else {
            // Black move needs move number if it follows a variant
            if (lastWasVariant) {
                pgnParts.push(moveNumber + '...' + san);
            } else {
                pgnParts.push(san);
            }
            lastWasVariant = false;
        }
        
        // Check for variants AFTER this move (alternatives to this move)
        if (gameState.variants[i]) {
            for (const variant of gameState.variants[i]) {
                const variantGame = new Chess();
                // Replay moves up to the position before this move
                for (let j = 0; j < i; j++) {
                    uciToSan(gameState.moveHistory[j], variantGame);
                }
                pgnParts.push(buildVariantPGN(variant, i, variantGame));
            }
            lastWasVariant = true;
        }
    }
    
    // Check for variants after the last move
    if (gameState.variants[gameState.moveHistory.length]) {
        for (const variant of gameState.variants[gameState.moveHistory.length]) {
            const variantGame = new Chess();
            for (let j = 0; j < gameState.moveHistory.length; j++) {
                uciToSan(gameState.moveHistory[j], variantGame);
            }
            pgnParts.push(buildVariantPGN(variant, gameState.moveHistory.length, variantGame));
        }
    }
    
    return pgnParts.join(' ');
}

// Save PGN to file
function savePGNFile() {
    if (gameState.moveHistory.length === 0) {
        showDialog('No moves to save!', 'warning');
        return;
    }
    
    // Get player names
    const whitePlayerValue = getWhitePlayer();
    const blackPlayerValue = getBlackPlayer();
    const whiteName = getPlayerName(whitePlayerValue);
    const blackName = getPlayerName(blackPlayerValue);
    
    // Create PGN content with headers
    const date = new Date();
    const dateStr = date.toISOString().split('T')[0].replace(/-/g, '.');
    
    const pgnMoves = buildStandardPGN();
    Logger.game.debug('Generated PGN', { moves: pgnMoves, variants: gameState.variants });
    
    let pgnWithHeaders = '[Event "Casual Game"]\n';
    pgnWithHeaders += '[Site "gochess-board"]\n';
    pgnWithHeaders += '[Date "' + dateStr + '"]\n';
    pgnWithHeaders += '[White "' + whiteName + '"]\n';
    pgnWithHeaders += '[Black "' + blackName + '"]\n';
    pgnWithHeaders += '[Result "*"]\n';
    pgnWithHeaders += '\n';
    pgnWithHeaders += pgnMoves;
    pgnWithHeaders += ' *\n';
    
    // Create filename with player names
    // Sanitize names for filename (remove special characters)
    const sanitizeFilename = function(name) {
        return name.replace(/[^a-zA-Z0-9_-]/g, '_');
    };
    const whiteFilename = sanitizeFilename(whiteName);
    const blackFilename = sanitizeFilename(blackName);
    const filename = dateStr + '_' + whiteFilename + '_vs_' + blackFilename + '.pgn';
    
    // Create a blob and download link
    const blob = new Blob([pgnWithHeaders], { type: 'application/x-chess-pgn' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    
    // Trigger download
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    
    // Visual feedback
    const btn = event.target;
    const originalText = btn.textContent;
    btn.textContent = '✓ Saved';
    setTimeout(function() {
        btn.textContent = originalText;
    }, 1500);
}

// Parse PGN with variants (recursive)
function parsePGNWithVariants(pgnText) {
    // Remove comments in braces
    pgnText = pgnText.replace(/\{[^}]*\}/g, '');
    
    // Remove annotations like !, ?, !!, ??, !?, ?!
    pgnText = pgnText.replace(/[!?]+/g, '');
    
    // Remove result markers
    pgnText = pgnText.replace(/\s*(1-0|0-1|1\/2-1\/2|\*)\s*$/, '');
    
    const result = {
        mainLine: [],
        variants: {}
    };
    
    let i = 0;
    let currentPosition = 0;
    
    function parseMovesRecursive() {
        const moves = [];
        const localVariants = {};
        
        while (i < pgnText.length) {
            const char = pgnText[i];
            
            // Skip whitespace
            if (char === ' ' || char === '\n' || char === '\t') {
                i++;
                continue;
            }
            
            // Start of variant
            if (char === '(') {
                i++; // Skip opening paren
                // Variant replaces the last move that was just parsed
                // currentPosition is now AFTER that move, so the variant position is currentPosition - 1
                const variantStartPos = currentPosition - 1;
                const savedPosition = currentPosition;
                
                // Parse variant recursively
                const variantData = parseMovesRecursive();
                currentPosition = savedPosition; // Restore position after variant
                
                // Store variant at the position it replaces
                if (!localVariants[variantStartPos]) {
                    localVariants[variantStartPos] = [];
                }
                localVariants[variantStartPos].push(variantData.mainLine);
                
                // Merge any nested variants
                for (const pos in variantData.variants) {
                    if (!localVariants[pos]) {
                        localVariants[pos] = [];
                    }
                    localVariants[pos].push(...variantData.variants[pos]);
                }
                continue;
            }
            
            // End of variant
            if (char === ')') {
                i++; // Skip closing paren
                return { mainLine: moves, variants: localVariants };
            }
            
            // Parse move number (e.g., "1." or "1...")
            const moveNumMatch = pgnText.substring(i).match(/^(\d+)\.(\.\.)?/);
            if (moveNumMatch) {
                i += moveNumMatch[0].length;
                continue;
            }
            
            // Parse move (SAN notation)
            const moveMatch = pgnText.substring(i).match(/^([NBRQK]?[a-h]?[1-8]?x?[a-h][1-8](?:=[NBRQ])?|O-O(?:-O)?)/);
            if (moveMatch) {
                moves.push(moveMatch[1]);
                currentPosition++;
                i += moveMatch[1].length;
                continue;
            }
            
            // Unknown character, skip it
            i++;
        }
        
        return { mainLine: moves, variants: localVariants };
    }
    
    const parsed = parseMovesRecursive();
    return parsed;
}

// Load PGN from text string
function loadPGNFromText(text) {
    if (!text || text.trim() === '') {
        showDialog('No PGN text provided!', 'warning');
        return;
    }
    
    try {
        // Parse PGN moves (extract just the moves, ignore headers and comments)
        const pgnText = text.trim();
        
        // Reset player dropdowns first to clear any previous custom names
        resetPlayerDropdowns();
        
        // Extract player names from PGN headers
        const whiteMatch = pgnText.match(/\[White\s+"([^"]+)"\]/);
        const blackMatch = pgnText.match(/\[Black\s+"([^"]+)"\]/);
        const whiteName = whiteMatch ? whiteMatch[1] : null;
        const blackName = blackMatch ? blackMatch[1] : null;
        
        // Add player names to dropdowns if found
        if (whiteName) {
            addPlayerToDropdown('whitePlayer', whiteName);
        }
        if (blackName) {
            addPlayerToDropdown('blackPlayer', blackName);
        }
        
        // Update UI to reflect the new player selections
        updatePlayerControls();
        updateInfoText();
        
        // Remove PGN headers (lines starting with [)
        let movesOnly = pgnText.split('\n')
            .filter(line => !line.startsWith('['))
            .join(' ')
            .trim();
        
        // Parse PGN with variants
        const parsed = parsePGNWithVariants(movesOnly);
        
        if (parsed.mainLine.length === 0) {
            showDialog('No valid moves found in PGN!', 'warning');
            return;
        }
        
        // Reset game and apply main line moves
        game.reset();
        board.position('start');
        gameState.moveHistory = [];
        gameState.moveScores = [];
        gameState.variants = {};
        clearLastMoveHighlight();
        
        // Clear any analysis arrows
        board.clearArrow();
        
        // Apply each main line move
        for (let i = 0; i < parsed.mainLine.length; i++) {
            const san = parsed.mainLine[i];
            
            try {
                const move = game.move(san);
                if (!move) {
                    showDialog('Invalid move at position ' + (i + 1) + ': ' + san, 'error');
                    break;
                }
                
                // Convert to UCI format for our history
                const uciMove = move.from + move.to + (move.promotion || '');
                gameState.moveHistory.push(uciMove);
                
                // Highlight last move
                if (i === parsed.mainLine.length - 1) {
                    highlightLastMove(move.from, move.to);
                }
            } catch (err) {
                showDialog('Error applying move ' + (i + 1) + ': ' + san + '\n' + err.message, 'error');
                break;
            }
        }
        
        // Convert variants from SAN to UCI
        for (const pos in parsed.variants) {
            const variantPosition = parseInt(pos);
            gameState.variants[variantPosition] = [];
            
            for (const variantMoves of parsed.variants[pos]) {
                // Create a game at the variant starting position
                const variantGame = new Chess();
                for (let j = 0; j < variantPosition; j++) {
                    const uciMove = gameState.moveHistory[j];
                    const from = uciMove.substring(0, 2);
                    const to = uciMove.substring(2, 4);
                    const promotion = uciMove.length > 4 ? uciMove.substring(4) : undefined;
                    variantGame.move({ from, to, promotion });
                }
                
                // Apply variant moves and convert to UCI
                const uciVariant = [];
                for (const san of variantMoves) {
                    const move = variantGame.move(san);
                    if (move) {
                        const uciMove = move.from + move.to + (move.promotion || '');
                        uciVariant.push(uciMove);
                    }
                }
                
                if (uciVariant.length > 0) {
                    gameState.variants[variantPosition].push(uciVariant);
                }
            }
        }
        
        // Update board and displays - start at beginning position
        game.reset();
        board.position('start');
        gameState.currentPosition = 0;
        gameState.isNavigating = false;
        clearLastMoveHighlight();
        updateMoveHistoryDisplay();
        updateOpeningDisplay();
        updateInfoText();
        updateAnalysisForCurrentPosition();
        saveGameState();
        
        // Clear eval graph for fresh start
        if (typeof clearEvalGraph === 'function') {
            clearEvalGraph();
        }
        
        // Visual feedback (only if called from a button click event)
        if (typeof event !== 'undefined' && event && event.target) {
            const btn = event.target;
            const originalText = btn.textContent;
            btn.textContent = '✓ Loaded';
            setTimeout(function() {
                btn.textContent = originalText;
            }, 1500);
        }
        
    } catch (err) {
        Logger.game.error('Failed to parse PGN', { error: err.message });
        showDialog('Failed to parse PGN. Please check the format and try again.', 'error');
    }
}

// Paste PGN from clipboard
async function pastePGN() {
    try {
        // Check if Clipboard API is available
        if (!navigator.clipboard || !navigator.clipboard.readText) {
            // Fallback for browsers without Clipboard API support
            showPasteDialog();
            return;
        }
        
        // Try to read from clipboard using the modern Clipboard API
        const text = await navigator.clipboard.readText();
        
        if (!text || text.trim() === '') {
            showDialog('Clipboard is empty. Copy a PGN first, then try again.', 'warning');
            return;
        }
        
        // Check if it looks like PGN (has move numbers or PGN headers)
        const looksLikePGN = /(\[.*\]|\d+\.\s*[NBRQK]?[a-h]?[1-8]?x?[a-h][1-8]|O-O)/i.test(text);
        
        if (!looksLikePGN) {
            showDialog('The clipboard content does not appear to be valid PGN notation.', 'warning');
            return;
        }
        
        // Stop any running analysis and clear arrows
        if (typeof analysisActive !== 'undefined' && analysisActive) {
            stopAnalysis();
        }
        if (typeof board !== 'undefined' && board.clearArrow) {
            board.clearArrow();
        }
        
        // Load the PGN
        loadPGNFromText(text);
        
    } catch (err) {
        // Handle permission denied or other errors
        if (err.name === 'NotAllowedError') {
            // Clipboard access denied - show fallback dialog
            showPasteDialog();
        } else {
            Logger.game.error('Failed to paste PGN', { error: err.message });
            showDialog('Failed to read clipboard. Please use the manual paste option.', 'error');
            showPasteDialog();
        }
    }
}

// Show a dialog for manual PGN paste (fallback for browsers without clipboard API)
function showPasteDialog() {
    // Create modal overlay
    const modal = document.createElement('div');
    modal.id = 'pasteModal';
    modal.className = 'modal';
    modal.style.display = 'flex';
    
    // Use font-size: 16px to prevent iOS Safari from zooming on focus
    modal.innerHTML = `
        <div class="modal-content" style="max-width: 600px;">
            <div class="modal-header">
                <h2>📋 Paste PGN</h2>
                <span class="modal-close" onclick="closePasteDialog()">&times;</span>
            </div>
            <div class="modal-body">
                <p class="modal-info">Paste your PGN notation below:</p>
                <textarea id="pasteTextarea" style="width: 100%; height: 200px; font-family: 'Courier New', monospace; font-size: 16px; padding: 10px; border: 2px solid #667eea; border-radius: 5px; resize: vertical;" placeholder="1. e4 e5 2. Nf3 Nc6 3. Bb5 a6..."></textarea>
            </div>
            <div class="modal-footer">
                <button onclick="closePasteDialog()" class="modal-btn modal-btn-cancel">Cancel</button>
                <button onclick="loadPastedPGN()" class="modal-btn" style="background: #667eea; color: white;">Load PGN</button>
            </div>
        </div>
    `;
    
    document.body.appendChild(modal);
    
    // Set up paste event listener to auto-load when user pastes
    const textarea = document.getElementById('pasteTextarea');
    if (textarea) {
        textarea.addEventListener('paste', function(e) {
            // Let the paste happen, then load after a short delay
            setTimeout(function() {
                if (textarea.value.trim()) {
                    loadPastedPGN();
                }
            }, 100);
        });
    }
}

// Close the paste dialog
function closePasteDialog() {
    const modal = document.getElementById('pasteModal');
    if (modal) {
        modal.remove();
    }
}

// Load PGN from the paste dialog textarea
function loadPastedPGN() {
    const textarea = document.getElementById('pasteTextarea');
    if (!textarea) {
        closePasteDialog();
        return;
    }
    
    const text = textarea.value.trim();
    
    if (!text) {
        showDialog('Please paste PGN notation first.', 'warning');
        return;
    }
    
    // Close the dialog
    closePasteDialog();
    
    // Stop any running analysis and clear arrows
    if (typeof analysisActive !== 'undefined' && analysisActive) {
        stopAnalysis();
    }
    if (typeof board !== 'undefined' && board.clearArrow) {
        board.clearArrow();
    }
    
    // Load the PGN
    loadPGNFromText(text);
}
