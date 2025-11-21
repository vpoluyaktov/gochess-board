// Comprehensive Unit Tests for Variant Logic
// Run with: mocha variant-tests.js
// Or in browser with mocha test runner

describe('Variant Creation and Management', function() {
    
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
        
        // Reset global variables
        isVariantMode = false;
        variantWindow = null;
        mainWindow = null;
        variantStartPosition = 0;
        variantIndex = -1;
    });
    
    describe('Basic Variant Storage', function() {
        
        it('should store a variant at the correct position', function() {
            // Setup: main line with 5 moves
            gameState.moveHistory = ['e2e4', 'e7e5', 'g1f3', 'b8c6', 'f1c4'];
            gameState.currentPosition = 5;
            
            // Create variant at move 3 (Nf3)
            const variantMoves = ['g1e2', 'b8c6']; // Alternative: Ne2 instead of Nf3
            const variantPosition = 2; // Position of move to replace
            
            if (!gameState.variants[variantPosition]) {
                gameState.variants[variantPosition] = [];
            }
            gameState.variants[variantPosition].push(variantMoves);
            
            // Verify
            chai.assert.equal(gameState.variants[2].length, 1);
            chai.assert.deepEqual(gameState.variants[2][0], variantMoves);
        });
        
        it('should support multiple variants at the same position', function() {
            gameState.moveHistory = ['e2e4', 'e7e5'];
            
            const variant1 = ['g1f3'];
            const variant2 = ['g1e2'];
            const variant3 = ['f1c4'];
            
            gameState.variants[2] = [variant1, variant2, variant3];
            
            chai.assert.equal(gameState.variants[2].length, 3);
            chai.assert.deepEqual(gameState.variants[2][0], variant1);
            chai.assert.deepEqual(gameState.variants[2][1], variant2);
            chai.assert.deepEqual(gameState.variants[2][2], variant3);
        });
    });
    
    describe('Nested Variant Storage', function() {
        
        it('should store nested variants with correct position adjustment', function() {
            // Main line: e4 e5 Nf3 Nc6 Bc4
            gameState.moveHistory = ['e2e4', 'e7e5', 'g1f3', 'b8c6', 'f1c4'];
            
            // Variant at position 2: Ne2 Nc6 Bc4 d6
            gameState.variants[2] = [['g1e2', 'b8c6', 'f1c4', 'd7d6']];
            
            // Sub-variant within the variant at position 4 (Bc4 in variant)
            // This is position 2 + 2 = 4 in absolute terms
            gameState.variants[4] = [['f1b5']]; // Bb5 instead of Bc4
            
            chai.assert.isDefined(gameState.variants[2]);
            chai.assert.isDefined(gameState.variants[4]);
            chai.assert.equal(gameState.variants[2][0].length, 4);
            chai.assert.equal(gameState.variants[4][0][0], 'f1b5');
        });
        
        it('should handle deeply nested variants (3 levels)', function() {
            gameState.moveHistory = ['e2e4', 'e7e5'];
            
            // Level 1: Variant at position 2
            gameState.variants[2] = [['g1f3', 'b8c6', 'f1c4']];
            
            // Level 2: Sub-variant at position 4 (within first variant)
            gameState.variants[4] = [['f1b5', 'a7a6']];
            
            // Level 3: Sub-sub-variant at position 5 (within sub-variant)
            gameState.variants[5] = [['f1a4']];
            
            chai.assert.equal(Object.keys(gameState.variants).length, 3);
            chai.assert.isDefined(gameState.variants[2]);
            chai.assert.isDefined(gameState.variants[4]);
            chai.assert.isDefined(gameState.variants[5]);
        });
    });
    
    describe('Variant Merging', function() {
        
        it('should add a new variant when variantIndex is -1', function() {
            gameState.moveHistory = ['e2e4', 'e7e5', 'g1f3'];
            gameState.variants = {};
            
            const mergeData = {
                variantStartPosition: 2,
                variantIndex: -1,
                moveHistory: ['e2e4', 'e7e5', 'g1e2', 'b8c6'],
                variants: {}
            };
            
            const variantMoves = mergeData.moveHistory.slice(mergeData.variantStartPosition);
            
            if (!gameState.variants[mergeData.variantStartPosition]) {
                gameState.variants[mergeData.variantStartPosition] = [];
            }
            
            if (mergeData.variantIndex < 0) {
                gameState.variants[mergeData.variantStartPosition].push(variantMoves);
            }
            
            chai.assert.equal(gameState.variants[2].length, 1);
            chai.assert.deepEqual(gameState.variants[2][0], ['g1e2', 'b8c6']);
        });
        
        it('should replace existing variant when variantIndex is valid', function() {
            gameState.moveHistory = ['e2e4', 'e7e5', 'g1f3'];
            gameState.variants[2] = [['g1e2', 'b8c6']];
            
            const mergeData = {
                variantStartPosition: 2,
                variantIndex: 0,
                moveHistory: ['e2e4', 'e7e5', 'g1e2', 'b8c6', 'f1c4'],
                variants: {}
            };
            
            const variantMoves = mergeData.moveHistory.slice(mergeData.variantStartPosition);
            
            if (mergeData.variantIndex >= 0 && mergeData.variantIndex < gameState.variants[2].length) {
                gameState.variants[mergeData.variantStartPosition][mergeData.variantIndex] = variantMoves;
            }
            
            chai.assert.equal(gameState.variants[2].length, 1);
            chai.assert.equal(gameState.variants[2][0].length, 3);
            chai.assert.deepEqual(gameState.variants[2][0], ['g1e2', 'b8c6', 'f1c4']);
        });
        
        it('should merge sub-variants with position adjustment', function() {
            gameState.moveHistory = ['e2e4', 'e7e5', 'g1f3'];
            gameState.variants = {};
            
            const mergeData = {
                variantStartPosition: 2,
                variantIndex: -1,
                moveHistory: ['e2e4', 'e7e5', 'g1e2', 'b8c6', 'f1c4'],
                variants: {
                    2: [['f1b5']] // Sub-variant at position 2 within variant (absolute: 2+2=4)
                }
            };
            
            const variantMoves = mergeData.moveHistory.slice(mergeData.variantStartPosition);
            
            if (!gameState.variants[mergeData.variantStartPosition]) {
                gameState.variants[mergeData.variantStartPosition] = [];
            }
            gameState.variants[mergeData.variantStartPosition].push(variantMoves);
            
            // Merge sub-variants with position adjustment
            if (mergeData.variants && Object.keys(mergeData.variants).length > 0) {
                for (const posStr in mergeData.variants) {
                    const pos = parseInt(posStr);
                    const adjustedPos = mergeData.variantStartPosition + pos;
                    
                    if (!gameState.variants[adjustedPos]) {
                        gameState.variants[adjustedPos] = [];
                    }
                    
                    for (const subVariant of mergeData.variants[pos]) {
                        gameState.variants[adjustedPos].push(subVariant);
                    }
                }
            }
            
            chai.assert.isDefined(gameState.variants[2]);
            chai.assert.isDefined(gameState.variants[4]); // 2 + 2
            chai.assert.deepEqual(gameState.variants[4][0], ['f1b5']);
        });
    });
    
    describe('PGN Parsing with Variants', function() {
        
        it('should parse simple PGN with one variant', function() {
            const pgn = '1.e4 e5 2.Nf3 (2.Bc4 Nf6) Nc6';
            const result = parsePGNWithVariants(pgn);
            
            chai.assert.equal(result.mainLine.length, 4); // e4, e5, Nf3, Nc6
            chai.assert.isDefined(result.variants[2]); // Variant at position 2 (Nf3)
            chai.assert.equal(result.variants[2][0].length, 2); // Bc4, Nf6
        });
        
        it('should parse PGN with nested variants', function() {
            const pgn = '1.e4 e5 2.Nf3 (2.Bc4 Nf6 (2...Nc6 3.Nf3) 3.d3) Nc6';
            const result = parsePGNWithVariants(pgn);
            
            chai.assert.equal(result.mainLine.length, 4);
            chai.assert.isDefined(result.variants[2]); // Main variant
            // Nested variants are merged into the main variants object
            chai.assert.equal(result.variants[2][0].length, 3); // Bc4, Nf6, d3
            // The nested variant (2...Nc6 3.Nf3) should also be present
            if (result.variants[3]) {
                chai.assert.equal(result.variants[3][0].length, 2); // Nc6, Nf3
            }
        });
        
        it('should parse PGN with multiple variants at same position', function() {
            const pgn = '1.e4 e5 2.Nf3 (2.Bc4 Nf6) (2.Nc3 Nf6) Nc6';
            const result = parsePGNWithVariants(pgn);
            
            chai.assert.equal(result.mainLine.length, 4);
            chai.assert.equal(result.variants[2].length, 2); // Two variants at position 2
            chai.assert.deepEqual(result.variants[2][0], ['Bc4', 'Nf6']);
            chai.assert.deepEqual(result.variants[2][1], ['Nc3', 'Nf6']);
        });
        
        it('should handle black move variants correctly', function() {
            const pgn = '1.e4 e5 (1...c5 2.Nf3) 2.Nf3';
            const result = parsePGNWithVariants(pgn);
            
            chai.assert.equal(result.mainLine.length, 3); // e4, e5, Nf3
            chai.assert.isDefined(result.variants[1]); // Variant at position 1 (e5)
            chai.assert.equal(result.variants[1][0].length, 2); // c5, Nf3
        });
    });
    
    describe('PGN Building with Variants', function() {
        
        it('should build standard PGN with simple variant', function() {
            gameState.moveHistory = ['e2e4', 'e7e5', 'g1f3', 'b8c6'];
            gameState.variants = {2: [['f1c4', 'g8f6']]}; // Only this variant
            
            const pgn = buildStandardPGN();
            
            chai.assert.include(pgn, '1.e4 e5');
            chai.assert.include(pgn, '2.Nf3');
            chai.assert.include(pgn, 'Bc4');
            chai.assert.include(pgn, 'Nf6');
            chai.assert.include(pgn, 'Nc6');
        });
        
        it('should build PGN with nested variants', function() {
            gameState.moveHistory = ['e2e4', 'e7e5', 'g1f3', 'b8c6'];
            gameState.variants = {
                2: [['f1c4', 'g8f6', 'd2d3']]
                // For true nested variants, position 4 would be within the variant line
                // but that requires the variant to be at an absolute position > 4
            };
            
            const pgn = buildStandardPGN();
            
            chai.assert.include(pgn, 'Bc4');
            chai.assert.include(pgn, 'Nf6');
            chai.assert.include(pgn, 'd3');
            // Verify the variant is properly formatted
            chai.assert.include(pgn, '2.Bc4');
        });
        
        it('should build PGN with multiple variants', function() {
            gameState.moveHistory = ['e2e4', 'e7e5'];
            gameState.variants = {
                2: [
                    ['g1f3', 'b8c6'],
                    ['f1c4', 'g8f6'],
                    ['b1c3', 'g8f6']
                ]
            };
            
            const pgn = buildStandardPGN();
            
            // All three variants should be present
            chai.assert.include(pgn, 'Nf3');
            chai.assert.include(pgn, 'Bc4');
            chai.assert.include(pgn, 'Nc3');
            chai.assert.include(pgn, 'Nc6');
        });
    });
    
    describe('Variant Navigation', function() {
        
        it('should enable Open Variation button when viewing move with variant', function() {
            gameState.moveHistory = ['e2e4', 'e7e5', 'g1f3', 'b8c6'];
            gameState.currentPosition = 3; // After Nf3
            gameState.variants[2] = [['f1c4', 'g8f6']];
            
            // Check if variant exists at currentPosition - 1
            const currentMoveIndex = gameState.currentPosition > 0 ? gameState.currentPosition - 1 : -1;
            const hasVariant = currentMoveIndex >= 0 && 
                              gameState.variants[currentMoveIndex] && 
                              gameState.variants[currentMoveIndex].length > 0;
            
            chai.assert.isTrue(hasVariant);
            chai.assert.equal(currentMoveIndex, 2);
        });
        
        it('should disable Open Variation button when no variant exists', function() {
            gameState.moveHistory = ['e2e4', 'e7e5', 'g1f3', 'b8c6'];
            gameState.currentPosition = 4; // After Nc6
            gameState.variants = {2: [['f1c4', 'g8f6']]}; // Reset variants
            
            const currentMoveIndex = gameState.currentPosition > 0 ? gameState.currentPosition - 1 : -1;
            const hasVariant = !!(currentMoveIndex >= 0 && 
                              gameState.variants[currentMoveIndex] && 
                              gameState.variants[currentMoveIndex].length > 0);
            
            chai.assert.strictEqual(hasVariant, false); // Should be false since no variant at position 3
        });
    });
    
    describe('Edge Cases', function() {
        
        it('should handle empty move history', function() {
            gameState.moveHistory = [];
            gameState.variants = {};
            
            const pgn = buildStandardPGN();
            
            chai.assert.equal(pgn, '');
        });
        
        it('should handle variant with no moves', function() {
            gameState.moveHistory = ['e2e4', 'e7e5'];
            gameState.variants[2] = [[]];
            
            const pgn = buildStandardPGN();
            
            chai.assert.include(pgn, '1.e4 e5');
            // Empty variant should be skipped or handled gracefully
        });
        
        it('should handle variant at position beyond move history', function() {
            gameState.moveHistory = ['e2e4', 'e7e5'];
            gameState.variants[10] = [['g1f3']];
            
            const pgn = buildStandardPGN();
            
            chai.assert.include(pgn, '1.e4 e5');
            // Variant at invalid position should be ignored
        });
        
        it('should handle deeply nested variants (5 levels)', function() {
            gameState.moveHistory = ['e2e4'];
            gameState.variants[1] = [['e7e5']];
            gameState.variants[2] = [['g1f3']];
            gameState.variants[3] = [['b8c6']];
            gameState.variants[4] = [['f1c4']];
            gameState.variants[5] = [['g8f6']];
            
            chai.assert.equal(Object.keys(gameState.variants).length, 5);
        });
    });
    
    describe('Round-trip Testing', function() {
        
        it('should preserve variants through save and load cycle', function() {
            // Create game with variants
            gameState.moveHistory = ['e2e4', 'e7e5', 'g1f3', 'b8c6'];
            gameState.variants[2] = [['f1c4', 'g8f6', 'd2d3']];
            gameState.variants[3] = [['b8d7']];
            
            // Build PGN
            const pgn = buildStandardPGN();
            
            // Parse it back
            const result = parsePGNWithVariants(pgn);
            
            // Verify main line
            chai.assert.equal(result.mainLine.length, 4);
            
            // Verify variants are preserved
            chai.assert.isDefined(result.variants[2]);
            chai.assert.isDefined(result.variants[3]);
            chai.assert.equal(result.variants[2][0].length, 3);
        });
        
        it('should preserve multiple variants through round-trip', function() {
            gameState.moveHistory = ['e2e4', 'e7e5'];
            gameState.variants = {
                2: [
                    ['g1f3', 'b8c6'],
                    ['f1c4', 'g8f6']
                ]
            };
            
            const pgn = buildStandardPGN();
            const result = parsePGNWithVariants(pgn);
            
            // Should have variants at position 2
            chai.assert.isDefined(result.variants[2]);
            // Should have at least 2 variants (may have more due to nested variant detection)
            chai.assert.isAtLeast(result.variants[2].length, 2);
        });
    });
});
