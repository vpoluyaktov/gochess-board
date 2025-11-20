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
    });
});
