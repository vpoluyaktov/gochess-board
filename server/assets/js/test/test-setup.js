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
global.$ = require('jquery')(dom.window);

// Load Chess.js
const chessJsPath = path.join(__dirname, '..', 'chess.js');
const chessJsCode = fs.readFileSync(chessJsPath, 'utf8');
vm.runInThisContext(chessJsCode);

// Mock browser functions that aren't needed for tests
global.alert = function() {};
global.confirm = function() { return true; };

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

// Load modules in dependency order
loadAppCode('gochess-state.js');
loadAppCode('gochess-pgn.js');
loadAppCode('gochess-history.js');
loadAppCode('gochess-variants.js');
loadAppCode('gochess-navigation.js');

// Export chai
global.chai = require('chai');
