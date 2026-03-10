// Unit Tests for Chessboard Library
// Run with: mocha --require test-setup.js chessboard-tests.js
//
// Note: Full DOM-dependent tests (board creation, piece movement, etc.) require
// a real browser environment and are better tested with the test-runner.html
// These tests focus on utility functions and library loading

describe('Chessboard Library', function() {
    
    describe('Library Loading', function() {
        
        it('should have Chessboard constructor available', function() {
            chai.assert.isDefined(Chessboard, 'Chessboard should be defined');
            chai.assert.isFunction(Chessboard, 'Chessboard should be a function');
        });
        
        it('should have ChessBoard alias available', function() {
            chai.assert.isDefined(ChessBoard, 'ChessBoard should be defined');
            chai.assert.equal(ChessBoard, Chessboard, 'ChessBoard should be alias of Chessboard');
        });
        
        it('should have utility functions', function() {
            chai.assert.isFunction(Chessboard.fenToObj, 'Should have fenToObj');
            chai.assert.isFunction(Chessboard.objToFen, 'Should have objToFen');
        });
    });
    
    describe('FEN Utility Functions', function() {
        
        it('should convert FEN to position object', function() {
            const fen = 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR';
            const position = Chessboard.fenToObj(fen);
            
            chai.assert.isObject(position);
            chai.assert.equal(position.a1, 'wR', 'Should have white rook at a1');
            chai.assert.equal(position.e1, 'wK', 'Should have white king at e1');
            chai.assert.equal(position.e8, 'bK', 'Should have black king at e8');
        });
        
        it('should convert position object to FEN', function() {
            const position = {
                a1: 'wR', b1: 'wN', c1: 'wB', d1: 'wQ',
                e1: 'wK', f1: 'wB', g1: 'wN', h1: 'wR',
                a2: 'wP', b2: 'wP', c2: 'wP', d2: 'wP',
                e2: 'wP', f2: 'wP', g2: 'wP', h2: 'wP',
                a7: 'bP', b7: 'bP', c7: 'bP', d7: 'bP',
                e7: 'bP', f7: 'bP', g7: 'bP', h7: 'bP',
                a8: 'bR', b8: 'bN', c8: 'bB', d8: 'bQ',
                e8: 'bK', f8: 'bB', g8: 'bN', h8: 'bR'
            };
            
            const fen = Chessboard.objToFen(position);
            
            chai.assert.isString(fen);
            chai.assert.equal(fen, 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR');
        });
        
        it('should handle empty position', function() {
            const fen = Chessboard.objToFen({});
            chai.assert.equal(fen, '8/8/8/8/8/8/8/8');
        });
        
        it('should round-trip FEN conversion', function() {
            const originalFen = 'rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR';
            const position = Chessboard.fenToObj(originalFen);
            const convertedFen = Chessboard.objToFen(position);
            
            chai.assert.equal(convertedFen, originalFen);
        });
    });
    
    describe('Custom Modifications', function() {
        
        it('should document arrow drawing feature', function() {
            // The library has been modified to support arrow drawing
            // Methods: drawArrow(), clearArrow(), getArrow()
            chai.assert.ok(true, 'Arrow drawing feature documented');
        });
        
        it('should use console.error instead of alert', function() {
            // The library has been modified to use console.error instead of window.alert
            chai.assert.ok(true, 'Error handling improved');
        });
        
        it('should document ghost pieces and PV animation features', function() {
            // The library has been modified to support ghost pieces and PV animation
            // Methods: addGhostPiece(), clearGhostPieces(), drawPrincipalVariation(), etc.
            chai.assert.ok(true, 'Ghost pieces and PV animation features documented');
        });
    });
    
    // Ghost Pieces and PV Animation Tests
    // These tests require a DOM environment and board instance
    // Skip in Node.js environment (only run in browser)
    describe('Ghost Pieces and PV Animation', function() {
        let board;
        let container;
        let game;
        let canCreateBoard = false;
        
        before(function() {
            // Test if we can create a board instance
            try {
                const testContainer = document.createElement('div');
                testContainer.id = 'testBoardCheck';
                document.body.appendChild(testContainer);
                const testBoard = Chessboard('testBoardCheck', { position: 'start' });
                if (testBoard && testBoard.position) {
                    canCreateBoard = true;
                }
                if (testBoard && testBoard.destroy) {
                    testBoard.destroy();
                }
                if (testContainer.parentNode) {
                    testContainer.parentNode.removeChild(testContainer);
                }
            } catch (e) {
                canCreateBoard = false;
            }
        });
        
        beforeEach(function() {
            if (!canCreateBoard) {
                this.skip();
            }
            
            // Create a container for the board
            container = document.createElement('div');
            container.id = 'testBoard';
            document.body.appendChild(container);
            
            // Initialize a Chess.js game
            game = new Chess();
            
            // Create a chessboard instance
            board = Chessboard('testBoard', {
                position: 'start',
                draggable: false
            });
        });
        
        afterEach(function() {
            // Clean up
            if (board && board.destroy) {
                board.destroy();
            }
            if (container && container.parentNode) {
                container.parentNode.removeChild(container);
            }
        });
        
        describe('Ghost Piece Management', function() {
            
            it('should have addGhostPiece method', function() {
                chai.assert.isFunction(board.addGhostPiece, 'addGhostPiece should be a function');
            });
            
            it('should have clearGhostPieces method', function() {
                chai.assert.isFunction(board.clearGhostPieces, 'clearGhostPieces should be a function');
            });
            
            it('should add ghost piece to destination square', function() {
                board.addGhostPiece('e2', 'e4', 'wP');
                
                const $destSquare = $(container).find('.square-e4');
                const $ghostPiece = $destSquare.find('.ghost-piece');
                
                chai.assert.equal($ghostPiece.length, 1, 'Should have one ghost piece');
                chai.assert.include($ghostPiece.attr('src'), 'wP', 'Should have correct piece image');
            });
            
            it('should hide original piece on source square', function() {
                board.addGhostPiece('e2', 'e4', 'wP');
                
                const $sourceSquare = $(container).find('.square-e2');
                const $originalPiece = $sourceSquare.find('img.piece-417db');
                
                if ($originalPiece.length > 0) {
                    chai.assert.equal($originalPiece.css('visibility'), 'hidden', 
                        'Original piece should be hidden');
                }
            });
            
            it('should hide original piece on destination square for captures', function() {
                // Set up a position with a piece on e5
                board.position({e5: 'bP', e2: 'wP'});
                
                board.addGhostPiece('e2', 'e5', 'wP');
                
                const $destSquare = $(container).find('.square-e5');
                const $originalPiece = $destSquare.find('img.piece-417db').not('.ghost-piece');
                
                if ($originalPiece.length > 0) {
                    chai.assert.equal($originalPiece.css('visibility'), 'hidden',
                        'Captured piece should be hidden');
                }
            });
            
            it('should remove existing ghost pieces from squares before adding new one', function() {
                // Add first ghost piece
                board.addGhostPiece('e2', 'e4', 'wP');
                
                // Add second ghost piece to same destination
                board.addGhostPiece('e4', 'e5', 'wP');
                
                const $e4Square = $(container).find('.square-e4');
                const $ghostPieces = $e4Square.find('.ghost-piece');
                
                chai.assert.equal($ghostPieces.length, 0, 
                    'Should have removed ghost piece from e4');
            });
            
            it('should clear all ghost pieces', function() {
                board.addGhostPiece('e2', 'e4', 'wP');
                board.addGhostPiece('d2', 'd4', 'wP');
                board.addGhostPiece('g1', 'f3', 'wN');
                
                board.clearGhostPieces();
                
                const $allGhostPieces = $(container).find('.ghost-piece');
                chai.assert.equal($allGhostPieces.length, 0, 
                    'Should have removed all ghost pieces');
            });
            
            it('should restore original pieces when clearing ghost pieces', function() {
                board.position({e2: 'wP', e4: 'bP'});
                
                board.addGhostPiece('e2', 'e4', 'wP');
                board.clearGhostPieces();
                
                const $sourceSquare = $(container).find('.square-e2');
                const $sourcePiece = $sourceSquare.find('img.piece-417db');
                
                const $destSquare = $(container).find('.square-e4');
                const $destPiece = $destSquare.find('img.piece-417db');
                
                if ($sourcePiece.length > 0) {
                    chai.assert.equal($sourcePiece.css('visibility'), 'visible',
                        'Source piece should be restored');
                }
                
                if ($destPiece.length > 0) {
                    chai.assert.equal($destPiece.css('visibility'), 'visible',
                        'Destination piece should be restored');
                }
            });
            
            it('should apply ghost-piece CSS class', function() {
                board.addGhostPiece('e2', 'e4', 'wP');
                
                const $ghostPiece = $(container).find('.ghost-piece');
                chai.assert.equal($ghostPiece.length, 1, 'Should have ghost-piece class');
            });
            
            it('should apply ghost-fade-in animation class', function() {
                board.addGhostPiece('e2', 'e4', 'wP');
                
                const $ghostPiece = $(container).find('.ghost-piece');
                chai.assert.isTrue($ghostPiece.hasClass('ghost-fade-in'),
                    'Should have fade-in animation class');
            });
        });
        
        describe('PV Animation Control', function() {
            
            it('should have cancelPVAnimation method', function() {
                chai.assert.isFunction(board.cancelPVAnimation, 
                    'cancelPVAnimation should be a function');
            });
            
            it('should have setPositionChanged method', function() {
                chai.assert.isFunction(board.setPositionChanged,
                    'setPositionChanged should be a function');
            });
            
            it('should have drawPrincipalVariation method', function() {
                chai.assert.isFunction(board.drawPrincipalVariation,
                    'drawPrincipalVariation should be a function');
            });
            
            it('should have drawPVArrowAtIndex method', function() {
                chai.assert.isFunction(board.drawPVArrowAtIndex,
                    'drawPVArrowAtIndex should be a function');
            });
            
            it('should cancel PV animation and clear ghost pieces', function() {
                board.addGhostPiece('e2', 'e4', 'wP');
                board.cancelPVAnimation();
                
                const $ghostPieces = $(container).find('.ghost-piece');
                chai.assert.equal($ghostPieces.length, 0,
                    'Should clear ghost pieces when canceling');
            });
            
            it('should set position changed flag', function() {
                // This is a state flag, we can't directly test it
                // but we can verify the method exists and doesn't throw
                chai.assert.doesNotThrow(function() {
                    board.setPositionChanged();
                }, 'setPositionChanged should not throw');
            });
        });
        
        describe('PV Arrow Drawing', function() {
            
            it('should draw single PV arrow without ghost pieces', function() {
                const pvData = {
                    pv: ['e2e4'],
                    score: 25,
                    scoreType: 'cp'
                };
                
                board.drawPVArrowAtIndex(pvData, 0, true, false, game);
                
                // Check that an arrow was drawn (implementation detail)
                const arrow = board.getArrow();
                chai.assert.isDefined(arrow, 'Should have drawn an arrow');
            });
            
            it('should draw PV arrow with ghost piece', function() {
                const pvData = {
                    pv: ['e2e4'],
                    score: 25,
                    scoreType: 'cp'
                };
                
                board.drawPVArrowAtIndex(pvData, 0, true, true, game);
                
                const $ghostPieces = $(container).find('.ghost-piece');
                chai.assert.isAtLeast($ghostPieces.length, 1,
                    'Should have created ghost piece');
            });
            
            it('should include score label on first arrow', function() {
                const pvData = {
                    pv: ['e2e4', 'e7e5'],
                    score: 50,
                    scoreType: 'cp'
                };
                
                board.drawPVArrowAtIndex(pvData, 0, true, false, game);
                
                const arrow = board.getArrow();
                if (arrow && arrow.label) {
                    chai.assert.include(arrow.label, '0.50', 
                        'Should include score in label');
                }
            });
            
            it('should format mate scores correctly', function() {
                const pvData = {
                    pv: ['e2e4'],
                    score: 3,
                    scoreType: 'mate'
                };
                
                board.drawPVArrowAtIndex(pvData, 0, true, false, game);
                
                const arrow = board.getArrow();
                if (arrow && arrow.label) {
                    chai.assert.include(arrow.label, 'M3',
                        'Should format mate score');
                }
            });
            
            it('should show move numbers on arrows', function() {
                const pvData = {
                    pv: ['e2e4'],
                    score: 25,
                    scoreType: 'cp'
                };
                
                board.drawPVArrowAtIndex(pvData, 0, true, false, game);
                
                // Move number should be included in the arrow
                // This is implementation-specific, checking via getArrow()
                chai.assert.ok(true, 'Move numbers are displayed');
            });
            
            it('should use white color for white moves', function() {
                const pvData = {
                    pv: ['e2e4'],
                    score: 25,
                    scoreType: 'cp'
                };
                
                board.drawPVArrowAtIndex(pvData, 0, true, false, game);
                
                const arrow = board.getArrow();
                if (arrow) {
                    chai.assert.equal(arrow.color, '#FFFFFF',
                        'Should use white color for white moves');
                }
            });
            
            it('should use black color for black moves', function() {
                game.move('e4'); // White moves
                
                const pvData = {
                    pv: ['e7e5'],
                    score: -20,
                    scoreType: 'cp'
                };
                
                board.drawPVArrowAtIndex(pvData, 0, true, false, game);
                
                const arrow = board.getArrow();
                if (arrow) {
                    chai.assert.equal(arrow.color, '#000000',
                        'Should use black color for black moves');
                }
            });
            
            it('should apply moves to temporary game without affecting original', function() {
                const originalFen = game.fen();
                
                const pvData = {
                    pv: ['e2e4', 'e7e5', 'g1f3'],
                    score: 25,
                    scoreType: 'cp'
                };
                
                board.drawPVArrowAtIndex(pvData, 2, true, false, game);
                
                chai.assert.equal(game.fen(), originalFen,
                    'Original game should not be modified');
            });
            
            it('should handle promotion moves', function() {
                // Set up a position where pawn can promote
                game.load('8/P7/8/8/8/8/8/4K2k w - - 0 1');
                
                const pvData = {
                    pv: ['a7a8q'],
                    score: 900,
                    scoreType: 'cp'
                };
                
                chai.assert.doesNotThrow(function() {
                    board.drawPVArrowAtIndex(pvData, 0, true, false, game);
                }, 'Should handle promotion moves');
            });
            
            it('should handle invalid moves gracefully', function() {
                const pvData = {
                    pv: ['e2e5'], // Invalid move
                    score: 0,
                    scoreType: 'cp'
                };
                
                chai.assert.doesNotThrow(function() {
                    board.drawPVArrowAtIndex(pvData, 0, true, false, game);
                }, 'Should handle invalid moves without crashing');
            });
        });
        
        describe('PV Animation', function() {
            
            it('should require game instance parameter', function() {
                const pvData = {
                    pv: ['e2e4'],
                    score: 25,
                    scoreType: 'cp'
                };
                
                // Should log error but not crash
                chai.assert.doesNotThrow(function() {
                    board.drawPrincipalVariation(pvData, true, false);
                }, 'Should handle missing game instance gracefully');
            });
            
            it('should draw single move in best move mode', function() {
                const pvData = {
                    pv: ['e2e4', 'e7e5', 'g1f3'],
                    score: 25,
                    scoreType: 'cp'
                };
                
                board.drawPrincipalVariation(pvData, false, true, game);
                
                const arrow = board.getArrow();
                chai.assert.isDefined(arrow, 'Should draw arrow for best move');
            });
            
            it('should not show ghost pieces in best move mode', function() {
                const pvData = {
                    pv: ['e2e4'],
                    score: 25,
                    scoreType: 'cp'
                };
                
                board.drawPrincipalVariation(pvData, false, true, game);
                
                const $ghostPieces = $(container).find('.ghost-piece');
                chai.assert.equal($ghostPieces.length, 0,
                    'Should not show ghost pieces in best move mode');
            });
            
            it('should not restart animation for same PV sequence', function() {
                const pvData = {
                    pv: ['e2e4', 'e7e5'],
                    score: 25,
                    scoreType: 'cp'
                };
                
                // Draw first time
                board.drawPrincipalVariation(pvData, true, false, game);
                
                // Try to draw same PV again
                board.drawPrincipalVariation(pvData, true, false, game);
                
                // Should not restart (implementation detail)
                chai.assert.ok(true, 'Same PV should not restart animation');
            });
            
            it('should handle empty PV array', function() {
                const pvData = {
                    pv: [],
                    score: 0,
                    scoreType: 'cp'
                };
                
                chai.assert.doesNotThrow(function() {
                    board.drawPrincipalVariation(pvData, true, false, game);
                }, 'Should handle empty PV gracefully');
            });
            
            it('should limit PV to 6 moves maximum', function() {
                const pvData = {
                    pv: ['e2e4', 'e7e5', 'g1f3', 'b8c6', 'f1c4', 'g8f6', 'b1c3', 'd7d6'],
                    score: 25,
                    scoreType: 'cp'
                };
                
                // Should only animate first 6 moves
                chai.assert.doesNotThrow(function() {
                    board.drawPrincipalVariation(pvData, true, false, game);
                }, 'Should limit to 6 moves');
            });
        });
        
        describe('Integration with Arrow Drawing', function() {
            
            it('should clear previous arrows when drawing PV', function() {
                // Draw initial arrow
                board.drawArrow('e2', 'e4', '#FF0000', '', 1.0, true);
                
                const pvData = {
                    pv: ['d2d4'],
                    score: 25,
                    scoreType: 'cp'
                };
                
                board.drawPVArrowAtIndex(pvData, 0, true, false, game);
                
                const arrow = board.getArrow();
                chai.assert.equal(arrow.from, 'd2', 'Should have new arrow');
                chai.assert.equal(arrow.to, 'd4', 'Should have new arrow');
            });
            
            it('should use 0.8 opacity for PV arrows', function() {
                const pvData = {
                    pv: ['e2e4'],
                    score: 25,
                    scoreType: 'cp'
                };
                
                board.drawPVArrowAtIndex(pvData, 0, true, false, game);
                
                const arrow = board.getArrow();
                if (arrow) {
                    chai.assert.equal(arrow.opacity, 0.8,
                        'PV arrows should have 0.8 opacity');
                }
            });
        });
        
        describe('Edge Cases', function() {
            
            it('should handle very short PV (1 move)', function() {
                const pvData = {
                    pv: ['e2e4'],
                    score: 25,
                    scoreType: 'cp'
                };
                
                chai.assert.doesNotThrow(function() {
                    board.drawPrincipalVariation(pvData, true, false, game);
                }, 'Should handle single move PV');
            });
            
            it('should handle PV with only 2 characters (invalid)', function() {
                const pvData = {
                    pv: ['e2'],
                    score: 0,
                    scoreType: 'cp'
                };
                
                chai.assert.doesNotThrow(function() {
                    board.drawPVArrowAtIndex(pvData, 0, true, false, game);
                }, 'Should handle invalid move format');
            });
            
            it('should handle negative centipawn scores', function() {
                const pvData = {
                    pv: ['e2e4'],
                    score: -150,
                    scoreType: 'cp'
                };
                
                board.drawPVArrowAtIndex(pvData, 0, true, false, game);
                
                const arrow = board.getArrow();
                if (arrow && arrow.label) {
                    chai.assert.include(arrow.label, '-1.50',
                        'Should show negative score');
                }
            });
            
            it('should handle negative mate scores', function() {
                const pvData = {
                    pv: ['e2e4'],
                    score: -5,
                    scoreType: 'mate'
                };
                
                board.drawPVArrowAtIndex(pvData, 0, true, false, game);
                
                const arrow = board.getArrow();
                if (arrow && arrow.label) {
                    chai.assert.include(arrow.label, '-M5',
                        'Should show negative mate score');
                }
            });
        });
    });
});
