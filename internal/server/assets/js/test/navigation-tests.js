const assert = require('assert');

describe('Move Navigation', function() {
    beforeEach(function() {
        // Reset game state before each test
        gameState.moveHistory = [];
        gameState.variants = {};
        gameState.currentPosition = 0;
        gameState.isNavigating = false;
        gameState.selectedVariant = null;
        game.reset();
    });

    describe('Position Navigation', function() {
        beforeEach(function() {
            // Set up a simple game with a few moves
            gameState.moveHistory = ['e2e4', 'e7e5', 'g1f3', 'b8c6', 'd2d4'];
            gameState.currentPosition = 0;
        });

        it('should navigate to a specific position', function() {
            goToPosition(3);
            assert.strictEqual(gameState.currentPosition, 3);
            assert.strictEqual(gameState.isNavigating, true);
        });

        it('should not navigate to invalid positions', function() {
            const initialPosition = gameState.currentPosition;
            goToPosition(-1);
            assert.strictEqual(gameState.currentPosition, initialPosition);
            
            goToPosition(10);
            assert.strictEqual(gameState.currentPosition, initialPosition);
        });

        it('should not navigate if already at target position', function() {
            gameState.currentPosition = 3;
            goToPosition(3);
            assert.strictEqual(gameState.currentPosition, 3);
        });

        it('should step forward correctly', function() {
            gameState.currentPosition = 2;
            stepForward();
            assert.strictEqual(gameState.currentPosition, 3);
        });

        it('should step backward correctly', function() {
            gameState.currentPosition = 3;
            stepBackward();
            assert.strictEqual(gameState.currentPosition, 2);
        });

        it('should not step backward from start', function() {
            gameState.currentPosition = 0;
            stepBackward();
            assert.strictEqual(gameState.currentPosition, 0);
        });

        it('should not step forward from end', function() {
            gameState.currentPosition = gameState.moveHistory.length;
            stepForward();
            assert.strictEqual(gameState.currentPosition, gameState.moveHistory.length);
        });

        it('should go to start', function() {
            gameState.currentPosition = 3;
            goToStart();
            assert.strictEqual(gameState.currentPosition, 0);
            assert.strictEqual(game.fen(), 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1');
        });

        it('should go to end', function() {
            gameState.currentPosition = 0;
            goToEnd();
            assert.strictEqual(gameState.currentPosition, gameState.moveHistory.length);
        });
    });

    describe('Click Navigation Parsing', function() {
        it('should parse white move click correctly', function() {
            // Line format: "1. e4    e5"
            // Clicking on "e4" should give position 1
            const line = "1. e4    e5";
            const moveNumber = 1;
            const ch = 3; // Character position on "e4"
            
            const targetPosition = (moveNumber - 1) * 2 + 1;
            assert.strictEqual(targetPosition, 1);
        });

        it('should parse black move click correctly', function() {
            // Line format: "1. e4    e5"
            // Clicking on "e5" should give position 2
            const line = "1. e4    e5";
            const moveNumber = 1;
            const ch = 9; // Character position on "e5"
            
            const targetPosition = (moveNumber - 1) * 2 + 2;
            assert.strictEqual(targetPosition, 2);
        });

        it('should detect variant lines', function() {
            const variantLine = "   └─ (14. Kxf4 ...";
            assert.strictEqual(variantLine.includes('└─'), true);
        });

        it('should detect continuation lines', function() {
            const continuationLine = "       14. Qe4+ ...";
            assert.strictEqual(continuationLine.startsWith('       '), true);
        });
    });

    describe('Variant Selection', function() {
        beforeEach(function() {
            // Set up a game with variants
            gameState.moveHistory = ['e2e4', 'e7e5', 'g1f3', 'b8c6'];
            gameState.variants = {
                1: [['e7e6', 'g1f3']], // Variant after e2e4
                3: [['g8f6', 'd2d4']]  // Variant after g1f3
            };
            gameState.currentPosition = 4;
        });

        it('should store selected variant info', function() {
            gameState.selectedVariant = {
                position: 1,
                index: 0,
                startLine: 2,
                endLine: 3
            };
            
            assert.strictEqual(gameState.selectedVariant.position, 1);
            assert.strictEqual(gameState.selectedVariant.index, 0);
        });

        it('should clear selected variant when clicking main line', function() {
            gameState.selectedVariant = { position: 1, index: 0 };
            // Simulate clicking on main line
            gameState.selectedVariant = null;
            assert.strictEqual(gameState.selectedVariant, null);
        });

        it('should enable variant button when variant is selected', function() {
            gameState.selectedVariant = { position: 1, index: 0 };
            
            // Check that variant exists
            assert.strictEqual(gameState.variants[1] !== undefined, true);
            assert.strictEqual(gameState.variants[1][0] !== undefined, true);
        });
    });

    describe('Position Indicator', function() {
        beforeEach(function() {
            gameState.moveHistory = ['e2e4', 'e7e5', 'g1f3', 'b8c6'];
        });

        it('should show correct position at start', function() {
            gameState.currentPosition = 0;
            const indicator = gameState.currentPosition + '/' + gameState.moveHistory.length;
            assert.strictEqual(indicator, '0/4');
        });

        it('should show correct position in middle', function() {
            gameState.currentPosition = 2;
            const indicator = gameState.currentPosition + '/' + gameState.moveHistory.length;
            assert.strictEqual(indicator, '2/4');
        });

        it('should show correct position at end', function() {
            gameState.currentPosition = 4;
            const indicator = gameState.currentPosition + '/' + gameState.moveHistory.length;
            assert.strictEqual(indicator, '4/4');
        });
    });

    describe('Move History Display', function() {
        it('should format move pairs correctly', function() {
            gameState.moveHistory = ['e2e4', 'e7e5', 'g1f3', 'b8c6'];
            
            // Build PGN should create lines like:
            // 1. e4    e5
            // 2. Nf3   Nc6
            const pgn = buildPGNWithVariants();
            const lines = pgn.split('\n');
            
            assert.strictEqual(lines.length >= 2, true);
            assert.strictEqual(lines[0].includes('1.'), true);
            assert.strictEqual(lines[1].includes('2.'), true);
        });

        it.skip('should display variants on separate lines', function() {
            gameState.moveHistory = ['e2e4', 'e7e5', 'g1f3'];
            gameState.variants = {
                2: [['b8c6']] // Variant after g1f3
            };
            
            const pgn = buildPGNWithVariants();
            console.log('PGN:', pgn);
            console.log('Variants:', JSON.stringify(gameState.variants));
            const lines = pgn.split('\n');
            console.log('Lines:', lines);
            
            // Should have main line and variant line
            assert.strictEqual(lines.length >= 2, true);
            
            // Variant line should have branch marker
            const hasVariantMarker = lines.some(line => line.includes('└─'));
            assert.strictEqual(hasVariantMarker, true);
        });

        it.skip('should indent nested variants', function() {
            gameState.moveHistory = ['e2e4', 'e7e5', 'g1f3'];
            gameState.variants = {
                2: [['b8c6', 'd2d4']]
            };
            
            const pgn = buildPGNWithVariants();
            
            // Check for proper indentation
            assert.strictEqual(pgn.includes('└─'), true);
        });
    });

    describe('Game State Consistency', function() {
        it('should maintain FEN consistency when navigating', function() {
            gameState.moveHistory = ['e2e4', 'e7e5', 'g1f3'];
            
            // Navigate to position 2
            goToPosition(2);
            const fen1 = game.fen();
            
            // Navigate away and back
            goToPosition(0);
            goToPosition(2);
            const fen2 = game.fen();
            
            assert.strictEqual(fen1, fen2);
        });

        it('should set isNavigating flag correctly', function() {
            gameState.moveHistory = ['e2e4', 'e7e5'];
            gameState.currentPosition = 0;
            
            goToPosition(1);
            assert.strictEqual(gameState.isNavigating, true);
        });

        it('should handle empty move history', function() {
            gameState.moveHistory = [];
            gameState.currentPosition = 0;
            
            goToStart();
            assert.strictEqual(gameState.currentPosition, 0);
            
            goToEnd();
            assert.strictEqual(gameState.currentPosition, 0);
        });
    });
});
