// Unit Tests for Move History Display and PGN Tree Visualization
// Run with: mocha --require test-setup.js history-tests.js

describe('Move History Display', function() {
    
    beforeEach(function() {
        // Reset game state before each test
        gameState = {
            moveHistory: [],
            currentPosition: 0,
            variants: {},
            timeControl: { initial: 0, increment: 0 },
            whiteTimeMs: 0,
            blackTimeMs: 0,
            clockRunning: false,
            isNavigating: false
        };
    });
    
    describe('PGN Tree Visualization', function() {
        
        it('should display variant after white move on same line as move pair', function() {
            // This is the critical test for the bug fix
            // Variant after white move should appear AFTER the black move completes the pair
            gameState.moveHistory = [
                'e2e4', 'e7e5',  // 1. e4 e5
                'g1f3', 'b8c6',  // 2. Nf3 Nc6
                'f1c4', 'f8c5',  // 3. Bc4 Bc5
                'e1g1', 'g8f6',  // 4. O-O Nf6
                'd2d3', 'd7d6',  // 5. d3 d6
                'c2c3', 'e8g8'   // 6. c3 O-O
            ];
            
            // Variant at position 10 (after 6. c3 - white move)
            gameState.variants[10] = [
                ['h2h3', 'a7a6', 'a2a4']  // (6. h3 a6 7. a4)
            ];
            
            const pgn = buildPGNWithVariants();
            const lines = pgn.split('\n');
            
            // Find the line with move 6
            const move6Line = lines.find(line => line.trim().startsWith('6. c3'));
            
            // The line should contain both white and black moves
            chai.assert.isDefined(move6Line, 'Move 6 line should exist');
            chai.assert.include(move6Line, '6. c3', 'Should contain white move');
            chai.assert.include(move6Line, 'O-O', 'Should contain black move on same line');
            
            // The variant should be on the NEXT line, not interrupting the move pair
            const move6Index = lines.indexOf(move6Line);
            const nextLine = lines[move6Index + 1];
            chai.assert.include(nextLine, '└─', 'Next line should be the variant');
            chai.assert.include(nextLine, '6. h3', 'Variant should start with 6. h3');
        });
        
        it('should keep variant move pairs on same line', function() {
            // Variant moves should also follow move-pair formatting
            gameState.moveHistory = [
                'e2e4', 'e7e5',
                'g1f3', 'b8c6'
            ];
            
            gameState.variants[2] = [
                ['f1c4', 'g8f6', 'd2d3', 'f8c5']  // (2. Bc4 Nf6 3. d3 Bc5)
            ];
            
            const pgn = buildPGNWithVariants();
            const lines = pgn.split('\n');
            
            // Find variant lines
            const variantStart = lines.find(line => line.includes('└─'));
            chai.assert.isDefined(variantStart);
            chai.assert.include(variantStart, '2. Bc4', 'Variant should start with white move');
            chai.assert.include(variantStart, 'Nf6', 'Black move should be on same line');
        });
        
        it('should handle variant after white move with black continuation', function() {
            // Simplified test for the critical bug fix
            // Variant after white move should appear AFTER black move completes the pair
            gameState.moveHistory = [
                'e2e4', 'e7e5',
                'g1f3', 'b8c6',
                'f1c4', 'f8c5',
                'e1g1', 'g8f6',  // 4. O-O Nf6
                'd2d3', 'd7d6',  // 5. d3 d6
                'c2c3', 'e8g8'   // 6. c3 O-O
            ];
            
            // Variant at position 10 (after 6. c3 - white move)
            gameState.variants[10] = [
                ['h2h3', 'a7a6']  // (6. h3 a6)
            ];
            
            const pgn = buildPGNWithVariants();
            const lines = pgn.split('\n');
            
            // Move 6 should have both white and black moves on same line
            const move6Line = lines.find(line => line.includes('c3') && line.includes('O-O'));
            chai.assert.isDefined(move6Line, 'Move 6 should be complete with both moves');
            
            // Variant should be on the next line
            const move6Index = lines.indexOf(move6Line);
            const variantLine = lines[move6Index + 1];
            chai.assert.include(variantLine, '└─', 'Variant should be on next line');
            chai.assert.include(variantLine, 'h3', 'Variant should contain h3');
            chai.assert.include(variantLine, 'a6', 'Variant should contain a6 on same line');
        });
        
        it('should display variant after black move correctly', function() {
            gameState.moveHistory = [
                'e2e4', 'e7e5',
                'g1f3', 'b8c6'
            ];
            
            // Variant after black move (position 3)
            gameState.variants[3] = [
                ['g8f6', 'f1c4']  // (2...Nf6 3. Bc4)
            ];
            
            const pgn = buildPGNWithVariants();
            
            // Should contain the variant marker
            chai.assert.include(pgn, '└─', 'Should have variant marker');
            // Should contain the variant moves
            chai.assert.include(pgn, 'Nf6', 'Should contain variant move');
            chai.assert.include(pgn, 'Bc4', 'Should contain variant continuation');
        });
        
        it('should handle multiple variants at same position', function() {
            gameState.moveHistory = [
                'e2e4', 'e7e5',
                'g1f3', 'b8c6'
            ];
            
            gameState.variants[2] = [
                ['f1c4', 'g8f6'],  // (2. Bc4 Nf6)
                ['b1c3', 'g8f6'],  // (2. Nc3 Nf6)
                ['f1b5', 'a7a6']   // (2. Bb5 a6)
            ];
            
            const pgn = buildPGNWithVariants();
            const lines = pgn.split('\n');
            
            // All three variants should be present
            const variantLines = lines.filter(line => line.includes('└─'));
            chai.assert.equal(variantLines.length, 3, 'Should have 3 variant lines');
            
            // Each should have proper formatting
            chai.assert.include(variantLines[0], '2. Bc4');
            chai.assert.include(variantLines[0], 'Nf6');
            chai.assert.include(variantLines[1], '2. Nc3');
            chai.assert.include(variantLines[2], '2. Bb5');
        });
        
        it('should handle nested variants', function() {
            // Note: Current architecture stores variants in a flat global object by position
            // True nested variants (variants within variants) would require hierarchical storage
            // This test verifies that variants at overlapping positions are displayed
            gameState.moveHistory = [
                'e2e4', 'e7e5',
                'g1f3', 'b8c6',
                'f1c4', 'f8c5'  // Add more moves so position 4 exists
            ];
            
            // Main variant at position 2
            gameState.variants[2] = [
                ['b1c3', 'g8f6']  // (2. Nc3 Nf6)
            ];
            
            // Another variant at position 4 (separate variant, not nested)
            gameState.variants[4] = [
                ['d2d3', 'e8g8']  // (3. d3 O-O)
            ];
            
            const pgn = buildPGNWithVariants();
            
            // Should contain both variants
            chai.assert.include(pgn, 'Nc3', 'Should contain first variant');
            chai.assert.include(pgn, 'd3', 'Should contain second variant');
            
            // Should have variant markers
            const variantCount = (pgn.match(/└─/g) || []).length;
            chai.assert.isAtLeast(variantCount, 2, 'Should have at least two variant markers');
        });
        
        it('should not break move pairs when variant exists', function() {
            // Regression test: ensure we never split white+black moves across lines
            // unless there's a variant AFTER the black move
            gameState.moveHistory = [
                'e2e4', 'e7e5',
                'g1f3', 'b8c6',
                'f1c4', 'f8c5',
                'd2d3', 'd7d6'
            ];
            
            // Variant after white move at position 4
            gameState.variants[4] = [
                ['b1c3', 'g8f6']
            ];
            
            const pgn = buildPGNWithVariants();
            const lines = pgn.split('\n');
            
            // Move 3 should have both white and black moves
            const move3Line = lines.find(line => line.trim().startsWith('3. Bc4'));
            chai.assert.include(move3Line, 'Bc5', 'Move pair should be complete');
            
            // Variant should be on next line, not interrupting the pair
            const move3Index = lines.indexOf(move3Line);
            const nextLine = lines[move3Index + 1];
            chai.assert.include(nextLine, '└─', 'Variant should be on next line');
        });
        
        it('should handle empty move history', function() {
            gameState.moveHistory = [];
            gameState.variants = {};
            
            const pgn = buildPGNWithVariants();
            chai.assert.equal(pgn, '', 'Empty history should produce empty string');
        });
        
        it('should handle game with no variants', function() {
            gameState.moveHistory = [
                'e2e4', 'e7e5',
                'g1f3', 'b8c6',
                'f1c4', 'f8c5'
            ];
            gameState.variants = {};
            
            const pgn = buildPGNWithVariants();
            const lines = pgn.split('\n');
            
            // Should have 3 lines (one per move pair)
            chai.assert.equal(lines.length, 3);
            
            // No variant markers
            const hasVariantMarkers = lines.some(line => line.includes('└─'));
            chai.assert.isFalse(hasVariantMarkers, 'Should have no variant markers');
        });
    });
    
    describe('Move Pair Formatting', function() {
        
        it('should pad white moves for alignment', function() {
            gameState.moveHistory = [
                'e2e4', 'e7e5',
                'g1f3', 'b8c6'
            ];
            
            const pgn = buildPGNWithVariants();
            const lines = pgn.split('\n');
            
            // Each line should have consistent spacing
            lines.forEach(line => {
                if (line.trim().length > 0) {
                    // White move should be padded
                    const match = line.match(/\d+\.\s+(\w+)\s+/);
                    if (match) {
                        // There should be spacing between white and black moves
                        chai.assert.match(line, /\d+\.\s+\w+\s+\w+/);
                    }
                }
            });
        });
        
        it('should format move numbers correctly', function() {
            gameState.moveHistory = [
                'e2e4', 'e7e5',
                'g1f3', 'b8c6',
                'f1c4', 'f8c5'
            ];
            
            const pgn = buildPGNWithVariants();
            const lines = pgn.split('\n');
            
            chai.assert.include(lines[0], '1. ');
            chai.assert.include(lines[1], '2. ');
            chai.assert.include(lines[2], '3. ');
        });
    });
});
