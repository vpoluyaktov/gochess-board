// Global Logger System
// Provides centralized logging with configurable log levels

/**
 * Log levels (higher number = more verbose)
 * OFF   = 0 - No logging
 * ERROR = 1 - Only errors
 * WARN  = 2 - Errors and warnings
 * INFO  = 3 - General information
 * DEBUG = 4 - Detailed debug information
 * TRACE = 5 - Very detailed trace information (e.g., every clock tick)
 */
var LogLevel = {
    OFF: 0,
    ERROR: 1,
    WARN: 2,
    INFO: 3,
    DEBUG: 4,
    TRACE: 5
};

// Current log level - change this to control verbosity
// Can be changed at runtime via browser console: setLogLevel(LogLevel.DEBUG)
var currentLogLevel = LogLevel.INFO;

// Log counter for tracking log entries
var logCounter = 0;

/**
 * Set the global log level
 * @param {number} level - One of LogLevel values
 */
function setLogLevel(level) {
    if (level >= LogLevel.OFF && level <= LogLevel.TRACE) {
        currentLogLevel = level;
        console.log('[LOGGER] Log level set to: ' + getLogLevelName(level));
    } else {
        console.error('[LOGGER] Invalid log level: ' + level);
    }
}

/**
 * Get the name of a log level
 * @param {number} level - Log level number
 * @returns {string} - Log level name
 */
function getLogLevelName(level) {
    switch (level) {
        case LogLevel.OFF: return 'OFF';
        case LogLevel.ERROR: return 'ERROR';
        case LogLevel.WARN: return 'WARN';
        case LogLevel.INFO: return 'INFO';
        case LogLevel.DEBUG: return 'DEBUG';
        case LogLevel.TRACE: return 'TRACE';
        default: return 'UNKNOWN';
    }
}

/**
 * Core logging function
 * @param {number} level - Log level
 * @param {string} category - Category/module name (e.g., 'CLOCK', 'ENGINE', 'BOARD')
 * @param {string} message - Log message
 * @param {object} data - Optional data object to log
 */
function log(level, category, message, data) {
    if (level > currentLogLevel) return;
    
    logCounter++;
    var timestamp = new Date().toISOString();
    var levelName = getLogLevelName(level);
    var prefix = '[#' + logCounter + '] [' + timestamp + '] [' + levelName + '] [' + category + ']';
    
    var logFn;
    switch (level) {
        case LogLevel.ERROR:
            logFn = console.error;
            break;
        case LogLevel.WARN:
            logFn = console.warn;
            break;
        default:
            logFn = console.log;
    }
    
    if (data !== undefined) {
        logFn(prefix + ' ' + message, data);
    } else {
        logFn(prefix + ' ' + message);
    }
}

// Convenience functions for each log level

function logError(category, message, data) {
    log(LogLevel.ERROR, category, message, data);
}

function logWarn(category, message, data) {
    log(LogLevel.WARN, category, message, data);
}

function logInfo(category, message, data) {
    log(LogLevel.INFO, category, message, data);
}

function logDebug(category, message, data) {
    log(LogLevel.DEBUG, category, message, data);
}

function logTrace(category, message, data) {
    log(LogLevel.TRACE, category, message, data);
}

// Category-specific loggers for common modules

var Logger = {
    clock: {
        error: function(msg, data) { logError('CLOCK', msg, data); },
        warn: function(msg, data) { logWarn('CLOCK', msg, data); },
        info: function(msg, data) { logInfo('CLOCK', msg, data); },
        debug: function(msg, data) { logDebug('CLOCK', msg, data); },
        trace: function(msg, data) { logTrace('CLOCK', msg, data); }
    },
    engine: {
        error: function(msg, data) { logError('ENGINE', msg, data); },
        warn: function(msg, data) { logWarn('ENGINE', msg, data); },
        info: function(msg, data) { logInfo('ENGINE', msg, data); },
        debug: function(msg, data) { logDebug('ENGINE', msg, data); },
        trace: function(msg, data) { logTrace('ENGINE', msg, data); }
    },
    board: {
        error: function(msg, data) { logError('BOARD', msg, data); },
        warn: function(msg, data) { logWarn('BOARD', msg, data); },
        info: function(msg, data) { logInfo('BOARD', msg, data); },
        debug: function(msg, data) { logDebug('BOARD', msg, data); },
        trace: function(msg, data) { logTrace('BOARD', msg, data); }
    },
    game: {
        error: function(msg, data) { logError('GAME', msg, data); },
        warn: function(msg, data) { logWarn('GAME', msg, data); },
        info: function(msg, data) { logInfo('GAME', msg, data); },
        debug: function(msg, data) { logDebug('GAME', msg, data); },
        trace: function(msg, data) { logTrace('GAME', msg, data); }
    },
    analysis: {
        error: function(msg, data) { logError('ANALYSIS', msg, data); },
        warn: function(msg, data) { logWarn('ANALYSIS', msg, data); },
        info: function(msg, data) { logInfo('ANALYSIS', msg, data); },
        debug: function(msg, data) { logDebug('ANALYSIS', msg, data); },
        trace: function(msg, data) { logTrace('ANALYSIS', msg, data); }
    },
    navigation: {
        error: function(msg, data) { logError('NAV', msg, data); },
        warn: function(msg, data) { logWarn('NAV', msg, data); },
        info: function(msg, data) { logInfo('NAV', msg, data); },
        debug: function(msg, data) { logDebug('NAV', msg, data); },
        trace: function(msg, data) { logTrace('NAV', msg, data); }
    }
};

// Log initialization
console.log('[LOGGER] GoChess Logger initialized. Current level: ' + getLogLevelName(currentLogLevel));
console.log('[LOGGER] To change log level, use: setLogLevel(LogLevel.DEBUG) or setLogLevel(LogLevel.TRACE)');
console.log('[LOGGER] Available levels: OFF=0, ERROR=1, WARN=2, INFO=3, DEBUG=4, TRACE=5');
