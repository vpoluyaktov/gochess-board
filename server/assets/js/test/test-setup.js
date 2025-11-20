// Test setup for Node.js environment
// This file loads dependencies and mocks browser globals

const { JSDOM } = require('jsdom');
const fs = require('fs');
const path = require('path');
const vm = require('vm');

// Create a browser-like environment
const dom = new JSDOM('<!DOCTYPE html><html><body></body></html>', {
    url: 'http://localhost',
    pretendToBeVisual: true
});

global.window = dom.window;
global.document = dom.window.document;
global.navigator = dom.window.navigator;

// Load jQuery and make it available globally
const jQuery = require('jquery')(dom.window);
global.$ = jQuery;
global.jQuery = jQuery;
global.window.jQuery = jQuery;
global.window['jQuery'] = jQuery;
global.window.$ = jQuery;
global.window['$'] = jQuery;

// Add isPlainObject polyfill for jQuery 3.x compatibility
if (!jQuery.isPlainObject) {
    jQuery.isPlainObject = function(obj) {
        return Object.prototype.toString.call(obj) === '[object Object]';
    };
    global.$.isPlainObject = jQuery.isPlainObject;
}

// Load Chess.js
const chessJsPath = path.join(__dirname, '..', 'chess.js');
const chessJsCode = fs.readFileSync(chessJsPath, 'utf8');
vm.runInThisContext(chessJsCode);

// Mock CodeMirror (not needed for tests)
global.CodeMirror = {
    fromTextArea: function() {
        return {
            setValue: function() {},
            getValue: function() { return ''; },
            on: function() {}
        };
    }
};

// Mock board and game objects
global.board = {
    position: function() {},
    resize: function() {}
};

global.game = new Chess();

// Load application code using vm to execute in current context
function loadAppCode(filename) {
    const filePath = path.join(__dirname, '..', filename);
    const code = fs.readFileSync(filePath, 'utf8');
    vm.runInThisContext(code);
}

// Load Chessboard library
loadAppCode('chessboard-1.0.1.js');

// Make Chessboard globally available
global.Chessboard = window.Chessboard || window['Chessboard'];
global.ChessBoard = window.ChessBoard || window['ChessBoard'];

// Load modules in dependency order
loadAppCode('gochess-state.js');
loadAppCode('gochess-pgn.js');
loadAppCode('gochess-history.js');
loadAppCode('gochess-variants.js');
loadAppCode('gochess-navigation.js');

// Export chai
global.chai = require('chai');
