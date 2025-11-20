// Move History Display
// Handles move history display, PGN building with variants, and current move highlighting

var moveHistoryEditor = null;

function updateMoveHistoryDisplay() {
    if (!moveHistoryEditor) return;
    
    if (!gameState.moveHistory || gameState.moveHistory.length === 0) {
        moveHistoryEditor.setValue('');
        return;
    }
    
    // Convert UCI moves to SAN notation with variants
    const pgn = buildPGNWithVariants();
    
    // Update CodeMirror content
    const currentValue = moveHistoryEditor.getValue();
    if (currentValue !== pgn) {
        moveHistoryEditor.setValue(pgn);
    }
    
    // Highlight current move position
    highlightCurrentMove();
    
    // Auto-scroll to current move (or bottom if at the end)
    if (gameState.currentPosition === gameState.moveHistory.length) {
        // At the end, scroll to bottom
        const lastLine = moveHistoryEditor.lineCount();
        moveHistoryEditor.scrollIntoView({line: lastLine, ch: 0});
    }
    
    // Update position indicator
    updatePositionIndicator();
}

function buildPGNWithVariants() {
    // Build tree-style notation with variants on separate lines
    const tempGame = new Chess();
    let lines = [];
    
    function uciToSan(uciMove, game) {
        const from = uciMove.substring(0, 2);
        const to = uciMove.substring(2, 4);
        const promotion = uciMove.length > 4 ? uciMove.substring(4) : undefined;
        
        const move = game.move({ from, to, promotion });
        return move ? move.san : null;
    }
    
    function buildVariantLines(variantMoves, startPosition, startGame, firstLinePrefix, depth) {
        const variantGame = new Chess(startGame.fen());
        const variantLines = [];
        let currentLine = '';
        let isFirstLine = true;
        const indent = '       '; // Base indentation
        const subIndent = '          '; // Additional indent for sub-variants
        
        for (let i = 0; i < variantMoves.length; i++) {
            const moveIndex = startPosition + i;
            const isWhiteMove = moveIndex % 2 === 0;
            const moveNumber = Math.floor(moveIndex / 2) + 1;
            
            const san = uciToSan(variantMoves[i], variantGame);
            if (!san) break;
            
            // Start new line for each move pair
            if (isWhiteMove) {
                if (currentLine.length > 0) {
                    variantLines.push(currentLine);
                    currentLine = '';
                    isFirstLine = false;
                }
                
                // Add prefix only for first line (with branch symbol)
                if (isFirstLine) {
                    currentLine = firstLinePrefix + san.padEnd(6);
                    isFirstLine = false;
                } else {
                    // Align with the first move (after the opening parenthesis)
                    currentLine = indent + moveNumber + '. ' + san.padEnd(6);
                }
            } else {
                // Black move
                if (isFirstLine) {
                    // First move is black - need to start with the prefix
                    currentLine = firstLinePrefix + san;
                    isFirstLine = false;
                } else {
                    currentLine += san;
                }
            }
            
            // Check for sub-variants at this position within the variant
            // But skip if this is the first move of the variant (i === 0) because that would be
            // checking the same position where this variant itself is stored
            const subVariantPosition = moveIndex;
            if (i > 0 && gameState.variants[subVariantPosition]) {
                // Finish current line before adding sub-variants
                if (currentLine.length > 0) {
                    variantLines.push(currentLine);
                    currentLine = '';
                }
                
                // Add sub-variants with deeper indentation
                for (const subVariant of gameState.variants[subVariantPosition]) {
                    const subVariantGame = new Chess(startGame.fen());
                    // Replay variant moves up to this point
                    for (let j = 0; j <= i; j++) {
                        uciToSan(variantMoves[j], subVariantGame);
                    }
                    
                    const subVariantStartPos = subVariantPosition;
                    const isSubVariantWhiteMove = subVariantStartPos % 2 === 0;
                    const subVariantMoveNum = Math.floor(subVariantStartPos / 2) + 1;
                    
                    // Add branch symbol with more indentation for sub-variants
                    const subFirstMovePrefix = isSubVariantWhiteMove 
                        ? subIndent + '└─ (' + subVariantMoveNum + '. '
                        : subIndent + '└─ (' + subVariantMoveNum + '... ';
                    
                    const subVariantLines = buildVariantLines(subVariant, subVariantStartPos, subVariantGame, subFirstMovePrefix, depth + 1);
                    variantLines.push(...subVariantLines);
                }
                
                isFirstLine = false;
            }
        }
        
        if (currentLine.length > 0) {
            variantLines.push(currentLine + ')');
        }
        
        return variantLines;
    }
    
    // Build main line with move pairs
    let currentLine = '';
    let pendingVariantPosition = -1; // Track if we need to show variant after completing the pair
    
    for (let i = 0; i < gameState.moveHistory.length; i++) {
        const isWhiteMove = i % 2 === 0;
        const moveNumber = Math.floor(i / 2) + 1;
        
        const san = uciToSan(gameState.moveHistory[i], tempGame);
        if (!san) break;
        
        if (isWhiteMove) {
            // Start new line for each move pair
            if (currentLine.length > 0) {
                lines.push(currentLine);
            }
            currentLine = moveNumber + '. ' + san.padEnd(6);
        } else {
            // Black move - add to current line (completing the move pair)
            currentLine += san;
            
            // If there's a pending variant from the white move, show it now
            if (pendingVariantPosition >= 0) {
                // Push the completed line first
                if (currentLine.length > 0) {
                    lines.push(currentLine);
                    currentLine = '';
                }
                
                // Add the variant
                const variantPos = pendingVariantPosition;
                for (const variant of gameState.variants[variantPos]) {
                    const variantGame = new Chess();
                    for (let j = 0; j < variantPos; j++) {
                        uciToSan(gameState.moveHistory[j], variantGame);
                    }
                    
                    const isVariantWhiteMove = variantPos % 2 === 0;
                    const variantMoveNum = Math.floor(variantPos / 2) + 1;
                    
                    const firstMovePrefix = isVariantWhiteMove 
                        ? '   └─ (' + variantMoveNum + '. '
                        : '   └─ (' + variantMoveNum + '... ';
                    
                    const variantLines = buildVariantLines(variant, variantPos, variantGame, firstMovePrefix, 0);
                    lines.push(...variantLines);
                }
                pendingVariantPosition = -1;
            }
        }
        
        // Check for variants AFTER this move
        if (gameState.variants[i]) {
            const isVariantAfterWhiteMove = i % 2 === 0;
            
            if (isVariantAfterWhiteMove) {
                // Variant after white move - defer until black move completes the pair
                pendingVariantPosition = i;
            } else {
                // Variant after black move - show it now
                if (currentLine.length > 0) {
                    lines.push(currentLine);
                    currentLine = '';
                }
                
                for (const variant of gameState.variants[i]) {
                    const variantGame = new Chess();
                    for (let j = 0; j < i; j++) {
                        uciToSan(gameState.moveHistory[j], variantGame);
                    }
                    
                    const variantStartPos = i;
                    const isVariantWhiteMove = variantStartPos % 2 === 0;
                    const variantMoveNum = Math.floor(variantStartPos / 2) + 1;
                    
                    const firstMovePrefix = isVariantWhiteMove 
                        ? '   └─ (' + variantMoveNum + '. '
                        : '   └─ (' + variantMoveNum + '... ';
                    
                    const variantLines = buildVariantLines(variant, variantStartPos, variantGame, firstMovePrefix, 0);
                    lines.push(...variantLines);
                }
            }
        }
    }
    
    // Add last line if it exists
    if (currentLine.length > 0) {
        lines.push(currentLine);
        
        // Check for variants after the last move
        const lastMoveIndex = gameState.moveHistory.length;
        if (gameState.variants[lastMoveIndex]) {
            for (const variant of gameState.variants[lastMoveIndex]) {
                const variantGame = new Chess();
                for (let j = 0; j < gameState.moveHistory.length; j++) {
                    uciToSan(gameState.moveHistory[j], variantGame);
                }
                
                const variantStartPos = lastMoveIndex;
                const isVariantWhiteMove = variantStartPos % 2 === 0;
                const variantMoveNum = Math.floor(variantStartPos / 2) + 1;
                
                // Add branch symbol and opening parenthesis with first move
                const firstMovePrefix = isVariantWhiteMove 
                    ? '   └─ (' + variantMoveNum + '. '
                    : '   └─ (' + variantMoveNum + '... ';
                
                const variantLines = buildVariantLines(variant, variantStartPos, variantGame, firstMovePrefix, 0);
                lines.push(...variantLines);
            }
        }
    }
    
    return lines.join('\n');
}

function highlightCurrentMove() {
    if (!moveHistoryEditor) return;
    
    // Clear all current move markers
    moveHistoryEditor.getAllMarks().forEach(mark => mark.clear());
    
    if (gameState.moveHistory.length === 0) {
        return;
    }
    
    const text = moveHistoryEditor.getValue();
    const lines = text.split('\n');
    
    // Main line highlighting
    if (gameState.currentPosition === 0) {
        return;
    }
    
    const moveIndex = gameState.currentPosition - 1;
    const isWhiteMove = moveIndex % 2 === 0;
    const moveNumber = Math.floor(moveIndex / 2) + 1;
    
    // Find the line with this move number in the main line (not in variants)
    for (let lineNum = 0; lineNum < lines.length; lineNum++) {
        const line = lines[lineNum];
        
        // Skip variant lines (those with └─ or indentation)
        if (line.includes('└─') || line.startsWith('       ')) {
            continue;
        }
        
        // Match main line format: "1. e4    e5"
        const mainLinePattern = new RegExp('^' + moveNumber + '\\.\\s+(\\S+)(?:\\s+(\\S+))?');
        const match = line.match(mainLinePattern);
        
        if (match) {
            const whiteMove = match[1];
            const blackMove = match[2];
            
            if (isWhiteMove && whiteMove) {
                // Highlight white's move
                const moveStart = line.indexOf(whiteMove);
                const moveEnd = moveStart + whiteMove.length;
                
                moveHistoryEditor.markText(
                    {line: lineNum, ch: moveStart},
                    {line: lineNum, ch: moveEnd},
                    {className: 'chess-current-move'}
                );
                
                if (gameState.isNavigating) {
                    moveHistoryEditor.scrollIntoView({line: lineNum, ch: moveStart}, 100);
                }
                break;
            } else if (!isWhiteMove && blackMove) {
                // Highlight black's move
                const moveStart = line.indexOf(blackMove, line.indexOf(whiteMove) + whiteMove.length);
                const moveEnd = moveStart + blackMove.length;
                
                moveHistoryEditor.markText(
                    {line: lineNum, ch: moveStart},
                    {line: lineNum, ch: moveEnd},
                    {className: 'chess-current-move'}
                );
                
                if (gameState.isNavigating) {
                    moveHistoryEditor.scrollIntoView({line: lineNum, ch: moveStart}, 100);
                }
                break;
            }
        }
    }
}

// Navigate to a specific move position by clicking on it
function navigateToMoveAtClick(lineNum, ch) {
    const text = moveHistoryEditor.getValue();
    const lines = text.split('\n');
    
    if (lineNum >= lines.length) return;
    
    const line = lines[lineNum];
    
    // Skip variant lines (those with └─ or indentation)
    if (line.includes('└─') || line.startsWith('       ')) {
        return;
    }
    
    // Parse main line format: "1. e4    e5"
    // Extract move number and moves
    const mainLinePattern = /^(\d+)\.\s+(\S+)(?:\s+(\S+))?/;
    const match = line.match(mainLinePattern);
    
    if (!match) return;
    
    const moveNumber = parseInt(match[1]);
    const whiteMove = match[2];
    const blackMove = match[3];
    
    // Determine which move was clicked based on character position
    const whiteMoveStart = line.indexOf(whiteMove);
    const whiteMoveEnd = whiteMoveStart + whiteMove.length;
    
    let targetPosition;
    
    if (ch >= whiteMoveStart && ch <= whiteMoveEnd) {
        // Clicked on white's move
        targetPosition = (moveNumber - 1) * 2 + 1;
    } else if (blackMove) {
        const blackMoveStart = line.indexOf(blackMove, whiteMoveEnd);
        const blackMoveEnd = blackMoveStart + blackMove.length;
        
        if (ch >= blackMoveStart && ch <= blackMoveEnd) {
            // Clicked on black's move
            targetPosition = (moveNumber - 1) * 2 + 2;
        } else {
            // Clicked elsewhere on the line, default to white's move
            targetPosition = (moveNumber - 1) * 2 + 1;
        }
    } else {
        // Only white's move exists
        targetPosition = (moveNumber - 1) * 2 + 1;
    }
    
    // Navigate to the target position
    if (targetPosition !== undefined && targetPosition >= 0 && targetPosition <= gameState.moveHistory.length) {
        goToPosition(targetPosition);
    }
}
