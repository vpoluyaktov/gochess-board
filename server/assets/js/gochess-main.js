// Main Initialization
// Handles board initialization, event listeners, and application startup

$(document).ready(function() {
    // Initialize chessboard
    var config = {
        draggable: true,
        position: 'start',
        onDragStart: onDragStart,
        onDrop: onDrop,
        onSnapEnd: onSnapEnd,
        onMoveEnd: onMoveEnd,
        pieceTheme: '/assets/images/pieces/{piece}.png'
    };
    
    board = Chessboard('myBoard', config);
    
    // Make board responsive
    $(window).resize(function() {
        board.resize();
    });
    
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
    
    // Analysis engine selector - restart analysis if active
    document.getElementById('analysisEngine').addEventListener('change', function() {
        if (analysisActive) {
            // Stop current analysis
            stopAnalysis();
            // Restart with new engine after a brief delay
            setTimeout(function() {
                startAnalysis();
            }, 100);
        }
    });
    
    // Time control selector
    document.getElementById('timeControl').addEventListener('change', function() {
        var value = this.value;
        var parts = value.split('+');
        var initialMinutes = parseInt(parts[0]);
        var incrementSeconds = parseInt(parts[1]);
        
        setTimeControl(initialMinutes, incrementSeconds);
        stopClock();
        
        // Show/hide move time selector for unlimited mode
        var moveTimeSelection = document.getElementById('moveTimeSelection');
        if (value === '0+0') {
            moveTimeSelection.style.display = 'flex';
        } else {
            moveTimeSelection.style.display = 'none';
        }
        
        // DISABLED: Time control persistence disabled
        // localStorage.setItem('timeControl', value);
    });
    
    // Set default to Unlimited (0+0)
    var select = document.getElementById('timeControl');
    select.value = '0+0';
    setTimeControl(0, 0);
    
    // Show move time selector for unlimited mode
    document.getElementById('moveTimeSelection').style.display = 'flex';
    
    /* DISABLED CODE: Restore time control preference
    var savedTimeControl = localStorage.getItem('timeControl');
    if (savedTimeControl) {
        var select = document.getElementById('timeControl');
        if (Array.from(select.options).some(opt => opt.value === savedTimeControl)) {
            select.value = savedTimeControl;
            var parts = savedTimeControl.split('+');
            setTimeControl(parseInt(parts[0]), parseInt(parts[1]));
        }
    }
    */
    
    // Initialize
    restorePlayerSelections();
    
    // Try to load saved game state
    var loaded = loadGameState();
    if (!loaded) {
        // No saved state, start fresh
        updateMoveHistoryDisplay();
        updateClockDisplay();
    }
    
    updateInfoText();
    updateStartPauseButton();
    updatePositionIndicator();
    
    // Initialize CodeMirror for move history
    const moveHistoryTextArea = document.getElementById('moveHistoryText');
    if (moveHistoryTextArea) {
        moveHistoryEditor = CodeMirror.fromTextArea(moveHistoryTextArea, {
            mode: 'chess',
            lineNumbers: false,
            lineWrapping: true,
            readOnly: 'nocursor',  // Prevent cursor and editing
            theme: 'default',
            viewportMargin: Infinity
        });
        
        // Add click handler for move navigation with double-click detection
        // Use the wrapper element to catch clicks
        let clickTimer = null;
        let lastClickLine = null;
        
        moveHistoryEditor.getWrapperElement().addEventListener('click', function(e) {
            const pos = moveHistoryEditor.coordsChar({left: e.pageX, top: e.pageY});
            e.preventDefault();  // Prevent default click behavior
            e.stopPropagation();  // Stop event from bubbling
            
            const isVariantLine = lineToVariantMap[pos.line] !== undefined;
            
            // Check for double-click (two clicks on same line within 300ms)
            if (clickTimer && lastClickLine === pos.line) {
                // Double-click detected
                clearTimeout(clickTimer);
                clickTimer = null;
                lastClickLine = null;
                
                console.log('Double-click detected at line:', pos.line);
                console.log('Variant info at line:', lineToVariantMap[pos.line]);
                
                // Check if this is a variant line
                if (isVariantLine) {
                    console.log('Opening variant at line', pos.line);
                    // Select the variant first
                    selectVariantLine(pos.line);
                    // Then open it
                    openVariation();
                } else {
                    console.log('Not a variant line, navigating normally');
                    navigateToMoveAtClick(pos.line, pos.ch);
                }
            } else {
                // Single click
                if (isVariantLine) {
                    // On variant line - delay to detect potential double-click
                    lastClickLine = pos.line;
                    clickTimer = setTimeout(function() {
                        // Single click confirmed
                        navigateToMoveAtClick(pos.line, pos.ch);
                        clickTimer = null;
                        lastClickLine = null;
                    }, 300);
                } else {
                    // Not a variant line - execute immediately (no delay)
                    navigateToMoveAtClick(pos.line, pos.ch);
                }
            }
        });
        
        // Add paste event handler
        moveHistoryEditor.on('paste', function(cm, e) {
            e.preventDefault();
            const pastedText = (e.clipboardData || window.clipboardData).getData('text');
            
            // Load the pasted PGN
            loadPGNFromText(pastedText);
        });
        
        // Prevent manual editing - restore from game state
        moveHistoryEditor.on('beforeChange', function(cm, change) {
            if (change.origin !== 'setValue') {
                change.cancel();
            }
        });
    }
    
    // Add keyboard navigation
    document.addEventListener('keydown', function(e) {
        // Only handle arrow keys if not typing in an input/textarea/select
        // Exception: allow arrow keys for radio buttons (they should only respond to mouse/touch)
        if (e.target.tagName === 'TEXTAREA' || e.target.tagName === 'SELECT') {
            return;
        }
        if (e.target.tagName === 'INPUT' && e.target.type !== 'radio') {
            return;
        }
        
        switch(e.key) {
            case 'ArrowLeft':
                e.preventDefault();
                stepBackward();
                break;
            case 'ArrowRight':
                e.preventDefault();
                stepForward();
                break;
            case 'Home':
                e.preventDefault();
                goToStart();
                break;
            case 'End':
                e.preventDefault();
                goToEnd();
                break;
        }
    });
    
    // Check if computer should move
    window.setTimeout(checkForComputerMove, 500);
    
    // Initialize variant mode detection
    initializeVariantMode();
});
