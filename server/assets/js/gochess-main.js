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
        
        // DISABLED: Time control persistence disabled
        // localStorage.setItem('timeControl', value);
    });
    
    // Set default to Unlimited (0+0)
    var select = document.getElementById('timeControl');
    select.value = '0+0';
    setTimeControl(0, 0);
    
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
        
        // Add click handler for move navigation
        // Use the wrapper element to catch clicks
        moveHistoryEditor.getWrapperElement().addEventListener('click', function(e) {
            const pos = moveHistoryEditor.coordsChar({left: e.pageX, top: e.pageY});
            e.preventDefault();  // Prevent default click behavior
            e.stopPropagation();  // Stop event from bubbling
            navigateToMoveAtClick(pos.line, pos.ch);
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
        // Only handle arrow keys if not typing in an input/textarea
        if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA' || e.target.tagName === 'SELECT') {
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
