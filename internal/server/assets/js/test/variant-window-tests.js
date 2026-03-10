const assert = require('assert');

describe('Variant Window Messaging', function() {
    let mockWindow;
    let messageHandler;
    let sentMessages;
    
    beforeEach(function() {
        // Reset state
        sentMessages = [];
        window.variantDataLoaded = false;
        
        // Mock DOM elements that openVariation needs
        global.document.getElementById = function(id) {
            if (id === 'analysisEngine') {
                return { value: 'stockfish' };
            }
            if (id === 'whitePlayer' || id === 'blackPlayer') {
                return { value: 'human' };
            }
            return null;
        };
        
        // Mock player getter functions
        global.getWhitePlayer = function() { return 'human'; };
        global.getBlackPlayer = function() { return 'human'; };
        
        // Mock window.open
        mockWindow = {
            closed: false,
            postMessage: function(data, origin) {
                sentMessages.push({ data: data, origin: origin });
            },
            close: function() {
                this.closed = true;
            }
        };
        
        // Store original window.open
        global.originalWindowOpen = global.window.open;
        
        // Mock window.open
        global.window.open = function(url, target, features) {
            // Simulate variant window sending ready signal after a short delay
            setTimeout(function() {
                const readyEvent = {
                    origin: window.location.origin,
                    data: { type: 'variant-ready' }
                };
                window.dispatchEvent(new window.MessageEvent('message', readyEvent));
            }, 50);
            return mockWindow;
        };
    });
    
    afterEach(function() {
        // Restore original window.open
        if (global.originalWindowOpen) {
            global.window.open = global.originalWindowOpen;
        }
        
        // Clean up
        delete window.variantDataLoaded;
    });

    describe('Message Sending', function() {
        it('should send variant data to new window', function(done) {
            // Setup game state
            gameState.currentPosition = 2;
            gameState.moveHistory = ['e2e4', 'e7e5'];
            gameState.variants = {
                1: [['e7e6', 'g1f3']] // Variant after first move
            };
            gameState.selectedVariant = {
                position: 1,
                index: 0
            };
            
            // Call openVariation
            openVariation();
            
            // Wait a bit for async message sending
            setTimeout(function() {
                // Should have sent at least one message
                assert.ok(sentMessages.length > 0, 'Should send at least one message');
                
                // Check message content
                const firstMessage = sentMessages[0];
                assert.strictEqual(firstMessage.data.type, 'start-variant');
                assert.strictEqual(firstMessage.origin, window.location.origin);
                assert.ok(firstMessage.data.fen, 'Should include FEN');
                assert.ok(Array.isArray(firstMessage.data.moveHistory), 'Should include move history');
                assert.ok(Array.isArray(firstMessage.data.variantMoves), 'Should include variant moves');
                
                done();
            }, 250);
        });

        it('should stop sending after first successful send', function(done) {
            // Setup minimal game state
            gameState.currentPosition = 1;
            gameState.moveHistory = ['e2e4'];
            gameState.variants = {
                0: [['e7e5']]
            };
            gameState.selectedVariant = {
                position: 0,
                index: 0
            };
            
            // Call openVariation
            openVariation();
            
            // Wait for multiple intervals
            setTimeout(function() {
                // Should only send once (not multiple times)
                assert.strictEqual(sentMessages.length, 1, 'Should only send message once');
                
                done();
            }, 500);
        });

        it('should not send if window is closed', function(done) {
            // Override window.open to return a closed window
            global.window.open = function(url, target, features) {
                mockWindow.closed = true;
                return mockWindow;
            };
            
            // Setup game state
            gameState.currentPosition = 1;
            gameState.moveHistory = ['e2e4'];
            gameState.variants = {
                0: [['e7e5']]
            };
            gameState.selectedVariant = {
                position: 0,
                index: 0
            };
            
            // Call openVariation
            openVariation();
            
            // Wait a bit
            setTimeout(function() {
                // Should not send any messages (window was closed)
                assert.strictEqual(sentMessages.length, 0, 'Should not send to closed window');
                
                done();
            }, 250);
        });
    });

    describe('Message Receiving', function() {
        it('should process variant data message', function() {
            // Create mock event
            const mockEvent = {
                origin: window.location.origin,
                data: {
                    type: 'start-variant',
                    fen: 'rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1',
                    moveHistory: ['e2e4'],
                    currentPosition: 1,
                    variantStartPosition: 1,
                    variantIndex: 0,
                    variantMoves: ['e7e5', 'g1f3'],
                    whitePlayer: 'human',
                    blackPlayer: 'human',
                    timeControl: { type: 'unlimited' }
                }
            };
            
            // Process message
            handleVariantMessage(mockEvent);
            
            // Check that game state was updated
            // Should have loaded 1 move from moveHistory + 2 variant moves = 3 total
            assert.strictEqual(gameState.moveHistory.length, 3, 'Should load move history and variant moves');
            assert.strictEqual(gameState.currentPosition, 3, 'Should set current position after variant moves');
            assert.ok(window.variantDataLoaded, 'Should set loaded flag');
        });

        it('should ignore duplicate messages', function() {
            // Create mock event
            const mockEvent = {
                origin: window.location.origin,
                data: {
                    type: 'start-variant',
                    fen: 'rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1',
                    moveHistory: ['e2e4'],
                    currentPosition: 1,
                    variantStartPosition: 1,
                    variantIndex: 0,
                    variantMoves: ['e7e5'],
                    whitePlayer: 'human',
                    blackPlayer: 'human',
                    timeControl: { type: 'unlimited' }
                }
            };
            
            // Process message first time
            handleVariantMessage(mockEvent);
            const firstMoveCount = gameState.moveHistory.length;
            
            // Process same message again
            handleVariantMessage(mockEvent);
            
            // Should not process again
            assert.strictEqual(gameState.moveHistory.length, firstMoveCount, 
                'Should not process duplicate message');
        });

        it('should reject messages from different origin', function() {
            // Create mock event with wrong origin
            const mockEvent = {
                origin: 'https://evil.com',
                data: {
                    type: 'start-variant',
                    fen: 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1',
                    moveHistory: [],
                    currentPosition: 0
                }
            };
            
            // Store original move history
            const originalMoveHistory = gameState.moveHistory.slice();
            
            // Try to process message
            handleVariantMessage(mockEvent);
            
            // Should not have changed game state
            assert.deepStrictEqual(gameState.moveHistory, originalMoveHistory, 
                'Should reject message from different origin');
            assert.strictEqual(window.variantDataLoaded, false, 
                'Should not set loaded flag for rejected message');
        });

        it('should ignore non-variant messages', function() {
            // Create mock event with different message type
            const mockEvent = {
                origin: window.location.origin,
                data: {
                    type: 'some-other-message',
                    someData: 'test'
                }
            };
            
            // Store original state
            const originalMoveHistory = gameState.moveHistory.slice();
            
            // Process message
            handleVariantMessage(mockEvent);
            
            // Should not have changed game state
            assert.deepStrictEqual(gameState.moveHistory, originalMoveHistory, 
                'Should ignore non-variant messages');
        });
    });

    describe('Message Data Validation', function() {
        it('should include all required fields in variant data', function(done) {
            // Setup game state
            gameState.currentPosition = 2;
            gameState.moveHistory = ['e2e4', 'e7e5'];
            gameState.variants = {
                1: [['e7e6', 'g1f3']]
            };
            gameState.selectedVariant = {
                position: 1,
                index: 0
            };
            
            // Call openVariation
            openVariation();
            
            setTimeout(function() {
                const message = sentMessages[0].data;
                
                // Check all required fields
                assert.strictEqual(message.type, 'start-variant');
                assert.ok(message.fen, 'Should have FEN');
                assert.ok(Array.isArray(message.moveHistory), 'Should have moveHistory array');
                assert.ok(typeof message.currentPosition === 'number', 'Should have currentPosition');
                assert.ok(typeof message.variantStartPosition === 'number', 'Should have variantStartPosition');
                assert.ok(typeof message.variantIndex === 'number', 'Should have variantIndex');
                assert.ok(Array.isArray(message.variantMoves), 'Should have variantMoves array');
                assert.ok(message.whitePlayer, 'Should have whitePlayer');
                assert.ok(message.blackPlayer, 'Should have blackPlayer');
                assert.ok(message.timeControl, 'Should have timeControl');
                
                done();
            }, 250);
        });

        it('should correctly calculate FEN for variant position', function(done) {
            // Setup game state - variant after e2e4
            gameState.currentPosition = 1;
            gameState.moveHistory = ['e2e4'];
            gameState.variants = {
                0: [['e7e6']] // French Defense instead of e7e5
            };
            gameState.selectedVariant = {
                position: 0,
                index: 0
            };
            
            // Call openVariation
            openVariation();
            
            setTimeout(function() {
                const message = sentMessages[0].data;
                
                // FEN should be starting position (before e2e4)
                assert.ok(message.fen.includes('rnbqkbnr/pppppppp'), 
                    'FEN should be starting position for variant at move 0');
                
                done();
            }, 250);
        });
    });
});
