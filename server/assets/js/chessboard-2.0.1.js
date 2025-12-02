// chessboard.js v1.0.0 (original)
// https://github.com/oakmac/chessboardjs/
//
// Copyright (c) 2019, Chris Oakman
// Released under the MIT license
// https://github.com/oakmac/chessboardjs/blob/master/LICENSE.md
//
// -------------------------------------------------------------
// v2.0.1 (modified by https://github.com/vpoluyaktov)
// Modifications:
// - Added click-to-select and click-to-move functionality
// - Added SVG arrow drawing for engine analysis visualization
//   - board.drawArrow(from, to, color, label, opacity, clearPrevious, moveNumber) - Draw arrow with optional label and opacity
//   - board.clearArrow() - Remove all arrows and labels
//   - board.getArrow() - Get current arrow info
// - Added ghost pieces and PV animation for principal variation visualization
//   - board.addGhostPiece(fromSquare, toSquare, piece) - Add semi-transparent ghost piece
//   - board.clearGhostPieces() - Remove all ghost pieces and restore original pieces
//   
//   Three clear visualization modes:
//   - board.drawBestMove(pvData) - Draw single best move arrow with score (no animation)
//   - board.drawMultipleBestMoves(multiPVLines, options) - Draw multiple best move arrows (3 alternatives)
//   - board.drawPVAnimation(pvData, options) - Animate principal variation with ghost pieces (looping)
//   
//   Helper methods:
//   - board.drawPVArrowAtIndex(pvData, index, clearPrevious, showGhostPieces, clearGhosts) - Draw single PV arrow
//   - board.cancelPVAnimation() - Cancel ongoing PV animation
//   - board.setPositionChanged() - Mark position as changed to force-start next PV animation
//   - board.formatScoreLabel(scoreType, score) - Format evaluation score as label string
//   
//   Deprecated (backward compatibility):
//   - board.drawPrincipalVariation(pvData, showPV, showBestMove) - Use drawBestMove() or drawPVAnimation() instead
//   
// - Pure visualization library - NO Chess.js dependency (application provides pre-computed move data)

// start anonymous scope
;(function () {
  'use strict'

  var $ = window['jQuery']

  // ---------------------------------------------------------------------------
  // Constants
  // ---------------------------------------------------------------------------

  var COLUMNS = 'abcdefgh'.split('')
  var DEFAULT_DRAG_THROTTLE_RATE = 20
  var ELLIPSIS = '…'
  var MINIMUM_JQUERY_VERSION = '1.8.3'
  var RUN_ASSERTS = false
  var START_FEN = 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR'
  var START_POSITION = fenToObj(START_FEN)

  // default animation speeds
  var DEFAULT_APPEAR_SPEED = 200
  var DEFAULT_MOVE_SPEED = 200
  var DEFAULT_SNAPBACK_SPEED = 60
  var DEFAULT_SNAP_SPEED = 30
  var DEFAULT_TRASH_SPEED = 100

  // use unique class names to prevent clashing with anything else on the page
  // and simplify selectors
  // NOTE: these should never change
  var CSS = {}
  CSS['alpha'] = 'alpha-d2270'
  CSS['black'] = 'black-3c85d'
  CSS['board'] = 'board-b72b1'
  CSS['chessboard'] = 'chessboard-63f37'
  CSS['clearfix'] = 'clearfix-7da63'
  CSS['highlight1'] = 'highlight1-32417'
  CSS['highlight2'] = 'highlight2-9c5d2'
  CSS['notation'] = 'notation-322f9'
  CSS['numeric'] = 'numeric-fc462'
  CSS['piece'] = 'piece-417db'
  CSS['row'] = 'row-5277c'
  CSS['sparePieces'] = 'spare-pieces-7492f'
  CSS['sparePiecesBottom'] = 'spare-pieces-bottom-ae20f'
  CSS['sparePiecesTop'] = 'spare-pieces-top-4028b'
  CSS['square'] = 'square-55d63'
  CSS['white'] = 'white-1e1d7'

  // ---------------------------------------------------------------------------
  // Misc Util Functions
  // ---------------------------------------------------------------------------

  function throttle (f, interval, scope) {
    var timeout = 0
    var shouldFire = false
    var args = []

    var handleTimeout = function () {
      timeout = 0
      if (shouldFire) {
        shouldFire = false
        fire()
      }
    }

    var fire = function () {
      timeout = window.setTimeout(handleTimeout, interval)
      f.apply(scope, args)
    }

    return function (_args) {
      args = arguments
      if (!timeout) {
        fire()
      } else {
        shouldFire = true
      }
    }
  }

  // function debounce (f, interval, scope) {
  //   var timeout = 0
  //   return function (_args) {
  //     window.clearTimeout(timeout)
  //     var args = arguments
  //     timeout = window.setTimeout(function () {
  //       f.apply(scope, args)
  //     }, interval)
  //   }
  // }

  function uuid () {
    return 'xxxx-xxxx-xxxx-xxxx-xxxx-xxxx-xxxx-xxxx'.replace(/x/g, function (c) {
      var r = (Math.random() * 16) | 0
      return r.toString(16)
    })
  }

  function deepCopy (thing) {
    return JSON.parse(JSON.stringify(thing))
  }

  function parseSemVer (version) {
    var tmp = version.split('.')
    return {
      major: parseInt(tmp[0], 10),
      minor: parseInt(tmp[1], 10),
      patch: parseInt(tmp[2], 10)
    }
  }

  // returns true if version is >= minimum
  function validSemanticVersion (version, minimum) {
    version = parseSemVer(version)
    minimum = parseSemVer(minimum)

    var versionNum = (version.major * 100000 * 100000) +
                     (version.minor * 100000) +
                     version.patch
    var minimumNum = (minimum.major * 100000 * 100000) +
                     (minimum.minor * 100000) +
                     minimum.patch

    return versionNum >= minimumNum
  }

  function interpolateTemplate (str, obj) {
    for (var key in obj) {
      if (!obj.hasOwnProperty(key)) continue
      var keyTemplateStr = '{' + key + '}'
      var value = obj[key]
      while (str.indexOf(keyTemplateStr) !== -1) {
        str = str.replace(keyTemplateStr, value)
      }
    }
    return str
  }

  if (RUN_ASSERTS) {
    console.assert(interpolateTemplate('abc', {a: 'x'}) === 'abc')
    console.assert(interpolateTemplate('{a}bc', {}) === '{a}bc')
    console.assert(interpolateTemplate('{a}bc', {p: 'q'}) === '{a}bc')
    console.assert(interpolateTemplate('{a}bc', {a: 'x'}) === 'xbc')
    console.assert(interpolateTemplate('{a}bc{a}bc', {a: 'x'}) === 'xbcxbc')
    console.assert(interpolateTemplate('{a}{a}{b}', {a: 'x', b: 'y'}) === 'xxy')
  }

  // ---------------------------------------------------------------------------
  // Predicates
  // ---------------------------------------------------------------------------

  function isString (s) {
    return typeof s === 'string'
  }

  function isFunction (f) {
    return typeof f === 'function'
  }

  function isInteger (n) {
    return typeof n === 'number' &&
           isFinite(n) &&
           Math.floor(n) === n
  }

  function validAnimationSpeed (speed) {
    if (speed === 'fast' || speed === 'slow') return true
    if (!isInteger(speed)) return false
    return speed >= 0
  }

  function validThrottleRate (rate) {
    return isInteger(rate) &&
           rate >= 1
  }

  function validMove (move) {
    // move should be a string
    if (!isString(move)) return false

    // move should be in the form of "e2-e4", "f6-d5"
    var squares = move.split('-')
    if (squares.length !== 2) return false

    return validSquare(squares[0]) && validSquare(squares[1])
  }

  function validSquare (square) {
    return isString(square) && square.search(/^[a-h][1-8]$/) !== -1
  }

  if (RUN_ASSERTS) {
    console.assert(validSquare('a1'))
    console.assert(validSquare('e2'))
    console.assert(!validSquare('D2'))
    console.assert(!validSquare('g9'))
    console.assert(!validSquare('a'))
    console.assert(!validSquare(true))
    console.assert(!validSquare(null))
    console.assert(!validSquare({}))
  }

  function validPieceCode (code) {
    return isString(code) && code.search(/^[bw][KQRNBP]$/) !== -1
  }

  if (RUN_ASSERTS) {
    console.assert(validPieceCode('bP'))
    console.assert(validPieceCode('bK'))
    console.assert(validPieceCode('wK'))
    console.assert(validPieceCode('wR'))
    console.assert(!validPieceCode('WR'))
    console.assert(!validPieceCode('Wr'))
    console.assert(!validPieceCode('a'))
    console.assert(!validPieceCode(true))
    console.assert(!validPieceCode(null))
    console.assert(!validPieceCode({}))
  }

  function validFen (fen) {
    if (!isString(fen)) return false

    // cut off any move, castling, etc info from the end
    // we're only interested in position information
    fen = fen.replace(/ .+$/, '')

    // expand the empty square numbers to just 1s
    fen = expandFenEmptySquares(fen)

    // FEN should be 8 sections separated by slashes
    var chunks = fen.split('/')
    if (chunks.length !== 8) return false

    // check each section
    for (var i = 0; i < 8; i++) {
      if (chunks[i].length !== 8 ||
          chunks[i].search(/[^kqrnbpKQRNBP1]/) !== -1) {
        return false
      }
    }

    return true
  }

  if (RUN_ASSERTS) {
    console.assert(validFen(START_FEN))
    console.assert(validFen('8/8/8/8/8/8/8/8'))
    console.assert(validFen('r1bqkbnr/pppp1ppp/2n5/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R'))
    console.assert(validFen('3r3r/1p4pp/2nb1k2/pP3p2/8/PB2PN2/p4PPP/R4RK1 b - - 0 1'))
    console.assert(!validFen('3r3z/1p4pp/2nb1k2/pP3p2/8/PB2PN2/p4PPP/R4RK1 b - - 0 1'))
    console.assert(!validFen('anbqkbnr/8/8/8/8/8/PPPPPPPP/8'))
    console.assert(!validFen('rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/'))
    console.assert(!validFen('rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBN'))
    console.assert(!validFen('888888/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR'))
    console.assert(!validFen('888888/pppppppp/74/8/8/8/PPPPPPPP/RNBQKBNR'))
    console.assert(!validFen({}))
  }

  function validPositionObject (pos) {
    if (!$.isPlainObject(pos)) return false

    for (var i in pos) {
      if (!pos.hasOwnProperty(i)) continue

      if (!validSquare(i) || !validPieceCode(pos[i])) {
        return false
      }
    }

    return true
  }

  if (RUN_ASSERTS) {
    console.assert(validPositionObject(START_POSITION))
    console.assert(validPositionObject({}))
    console.assert(validPositionObject({e2: 'wP'}))
    console.assert(validPositionObject({e2: 'wP', d2: 'wP'}))
    console.assert(!validPositionObject({e2: 'BP'}))
    console.assert(!validPositionObject({y2: 'wP'}))
    console.assert(!validPositionObject(null))
    console.assert(!validPositionObject('start'))
    console.assert(!validPositionObject(START_FEN))
  }

  function isTouchDevice () {
    return 'ontouchstart' in document.documentElement
  }

  function validJQueryVersion () {
    return typeof window.$ &&
           $.fn &&
           $.fn.jquery &&
           validSemanticVersion($.fn.jquery, MINIMUM_JQUERY_VERSION)
  }

  // ---------------------------------------------------------------------------
  // Chess Util Functions
  // ---------------------------------------------------------------------------

  // convert FEN piece code to bP, wK, etc
  function fenToPieceCode (piece) {
    // black piece
    if (piece.toLowerCase() === piece) {
      return 'b' + piece.toUpperCase()
    }

    // white piece
    return 'w' + piece.toUpperCase()
  }

  // convert bP, wK, etc code to FEN structure
  function pieceCodeToFen (piece) {
    var pieceCodeLetters = piece.split('')

    // white piece
    if (pieceCodeLetters[0] === 'w') {
      return pieceCodeLetters[1].toUpperCase()
    }

    // black piece
    return pieceCodeLetters[1].toLowerCase()
  }

  // convert FEN string to position object
  // returns false if the FEN string is invalid
  function fenToObj (fen) {
    if (!validFen(fen)) return false

    // cut off any move, castling, etc info from the end
    // we're only interested in position information
    fen = fen.replace(/ .+$/, '')

    var rows = fen.split('/')
    var position = {}

    var currentRow = 8
    for (var i = 0; i < 8; i++) {
      var row = rows[i].split('')
      var colIdx = 0

      // loop through each character in the FEN section
      for (var j = 0; j < row.length; j++) {
        // number / empty squares
        if (row[j].search(/[1-8]/) !== -1) {
          var numEmptySquares = parseInt(row[j], 10)
          colIdx = colIdx + numEmptySquares
        } else {
          // piece
          var square = COLUMNS[colIdx] + currentRow
          position[square] = fenToPieceCode(row[j])
          colIdx = colIdx + 1
        }
      }

      currentRow = currentRow - 1
    }

    return position
  }

  // position object to FEN string
  // returns false if the obj is not a valid position object
  function objToFen (obj) {
    if (!validPositionObject(obj)) return false

    var fen = ''

    var currentRow = 8
    for (var i = 0; i < 8; i++) {
      for (var j = 0; j < 8; j++) {
        var square = COLUMNS[j] + currentRow

        // piece exists
        if (obj.hasOwnProperty(square)) {
          fen = fen + pieceCodeToFen(obj[square])
        } else {
          // empty space
          fen = fen + '1'
        }
      }

      if (i !== 7) {
        fen = fen + '/'
      }

      currentRow = currentRow - 1
    }

    // squeeze the empty numbers together
    fen = squeezeFenEmptySquares(fen)

    return fen
  }

  if (RUN_ASSERTS) {
    console.assert(objToFen(START_POSITION) === START_FEN)
    console.assert(objToFen({}) === '8/8/8/8/8/8/8/8')
    console.assert(objToFen({a2: 'wP', 'b2': 'bP'}) === '8/8/8/8/8/8/Pp6/8')
  }

  function squeezeFenEmptySquares (fen) {
    return fen.replace(/11111111/g, '8')
      .replace(/1111111/g, '7')
      .replace(/111111/g, '6')
      .replace(/11111/g, '5')
      .replace(/1111/g, '4')
      .replace(/111/g, '3')
      .replace(/11/g, '2')
  }

  function expandFenEmptySquares (fen) {
    return fen.replace(/8/g, '11111111')
      .replace(/7/g, '1111111')
      .replace(/6/g, '111111')
      .replace(/5/g, '11111')
      .replace(/4/g, '1111')
      .replace(/3/g, '111')
      .replace(/2/g, '11')
  }

  // returns the distance between two squares
  function squareDistance (squareA, squareB) {
    var squareAArray = squareA.split('')
    var squareAx = COLUMNS.indexOf(squareAArray[0]) + 1
    var squareAy = parseInt(squareAArray[1], 10)

    var squareBArray = squareB.split('')
    var squareBx = COLUMNS.indexOf(squareBArray[0]) + 1
    var squareBy = parseInt(squareBArray[1], 10)

    var xDelta = Math.abs(squareAx - squareBx)
    var yDelta = Math.abs(squareAy - squareBy)

    if (xDelta >= yDelta) return xDelta
    return yDelta
  }

  // returns the square of the closest instance of piece
  // returns false if no instance of piece is found in position
  function findClosestPiece (position, piece, square) {
    // create array of closest squares from square
    var closestSquares = createRadius(square)

    // search through the position in order of distance for the piece
    for (var i = 0; i < closestSquares.length; i++) {
      var s = closestSquares[i]

      if (position.hasOwnProperty(s) && position[s] === piece) {
        return s
      }
    }

    return false
  }

  // returns an array of closest squares from square
  function createRadius (square) {
    var squares = []

    // calculate distance of all squares
    for (var i = 0; i < 8; i++) {
      for (var j = 0; j < 8; j++) {
        var s = COLUMNS[i] + (j + 1)

        // skip the square we're starting from
        if (square === s) continue

        squares.push({
          square: s,
          distance: squareDistance(square, s)
        })
      }
    }

    // sort by distance
    squares.sort(function (a, b) {
      return a.distance - b.distance
    })

    // just return the square code
    var surroundingSquares = []
    for (i = 0; i < squares.length; i++) {
      surroundingSquares.push(squares[i].square)
    }

    return surroundingSquares
  }

  // given a position and a set of moves, return a new position
  // with the moves executed
  function calculatePositionFromMoves (position, moves) {
    var newPosition = deepCopy(position)

    for (var i in moves) {
      if (!moves.hasOwnProperty(i)) continue

      // skip the move if the position doesn't have a piece on the source square
      if (!newPosition.hasOwnProperty(i)) continue

      var piece = newPosition[i]
      delete newPosition[i]
      newPosition[moves[i]] = piece
    }

    return newPosition
  }

  // TODO: add some asserts here for calculatePositionFromMoves

  // ---------------------------------------------------------------------------
  // HTML
  // ---------------------------------------------------------------------------

  function buildContainerHTML (hasSparePieces) {
    var html = '<div class="{chessboard}">'

    if (hasSparePieces) {
      html += '<div class="{sparePieces} {sparePiecesTop}"></div>'
    }

    html += '<div class="{board}"></div>'

    if (hasSparePieces) {
      html += '<div class="{sparePieces} {sparePiecesBottom}"></div>'
    }

    html += '</div>'

    return interpolateTemplate(html, CSS)
  }

  // ---------------------------------------------------------------------------
  // Config
  // ---------------------------------------------------------------------------

  function expandConfigArgumentShorthand (config) {
    if (config === 'start') {
      config = {position: deepCopy(START_POSITION)}
    } else if (validFen(config)) {
      config = {position: fenToObj(config)}
    } else if (validPositionObject(config)) {
      config = {position: deepCopy(config)}
    }

    // config must be an object
    if (!$.isPlainObject(config)) config = {}

    return config
  }

  // validate config / set default options
  function expandConfig (config) {
    // default for orientation is white
    if (config.orientation !== 'black') config.orientation = 'white'

    // default for showNotation is true
    if (config.showNotation !== false) config.showNotation = true

    // default for draggable is false
    if (config.draggable !== true) config.draggable = false

    // default for dropOffBoard is 'snapback'
    if (config.dropOffBoard !== 'trash') config.dropOffBoard = 'snapback'

    // default for sparePieces is false
    if (config.sparePieces !== true) config.sparePieces = false

    // draggable must be true if sparePieces is enabled
    if (config.sparePieces) config.draggable = true

    // default piece theme is wikipedia
    if (!config.hasOwnProperty('pieceTheme') ||
        (!isString(config.pieceTheme) && !isFunction(config.pieceTheme))) {
      config.pieceTheme = 'img/chesspieces/wikipedia/{piece}.png'
    }

    // animation speeds
    if (!validAnimationSpeed(config.appearSpeed)) config.appearSpeed = DEFAULT_APPEAR_SPEED
    if (!validAnimationSpeed(config.moveSpeed)) config.moveSpeed = DEFAULT_MOVE_SPEED
    if (!validAnimationSpeed(config.snapbackSpeed)) config.snapbackSpeed = DEFAULT_SNAPBACK_SPEED
    if (!validAnimationSpeed(config.snapSpeed)) config.snapSpeed = DEFAULT_SNAP_SPEED
    if (!validAnimationSpeed(config.trashSpeed)) config.trashSpeed = DEFAULT_TRASH_SPEED

    // throttle rate
    if (!validThrottleRate(config.dragThrottleRate)) config.dragThrottleRate = DEFAULT_DRAG_THROTTLE_RATE

    return config
  }

  // ---------------------------------------------------------------------------
  // Dependencies
  // ---------------------------------------------------------------------------

  // check for a compatible version of jQuery
  function checkJQuery () {
    if (!validJQueryVersion()) {
      var errorMsg = 'Chessboard Error 1005: Unable to find a valid version of jQuery. ' +
        'Please include jQuery ' + MINIMUM_JQUERY_VERSION + ' or higher on the page' +
        '\n\n' +
        'Exiting' + ELLIPSIS
      console.error(errorMsg)
      return false
    }

    return true
  }

  // return either boolean false or the $container element
  function checkContainerArg (containerElOrString) {
    if (containerElOrString === '') {
      var errorMsg1 = 'Chessboard Error 1001: ' +
        'The first argument to Chessboard() cannot be an empty string.' +
        '\n\n' +
        'Exiting' + ELLIPSIS
      console.error(errorMsg1)
      return false
    }

    // convert containerEl to query selector if it is a string
    if (isString(containerElOrString) &&
        containerElOrString.charAt(0) !== '#') {
      containerElOrString = '#' + containerElOrString
    }

    // containerEl must be something that becomes a jQuery collection of size 1
    var $container = $(containerElOrString)
    if ($container.length !== 1) {
      var errorMsg2 = 'Chessboard Error 1003: ' +
        'The first argument to Chessboard() must be the ID of a DOM node, ' +
        'an ID query selector, or a single DOM node.' +
        '\n\n' +
        'Exiting' + ELLIPSIS
      console.error(errorMsg2)
      return false
    }

    return $container
  }

  // ---------------------------------------------------------------------------
  // Constructor
  // ---------------------------------------------------------------------------

  function constructor (containerElOrString, config) {
    // first things first: check basic dependencies
    if (!checkJQuery()) return null
    var $container = checkContainerArg(containerElOrString)
    if (!$container) return null

    // ensure the config object is what we expect
    config = expandConfigArgumentShorthand(config)
    config = expandConfig(config)

    // DOM elements
    var $board = null
    var $draggedPiece = null
    var $sparePiecesTop = null
    var $sparePiecesBottom = null

    // constructor return object
    var widget = {}

    // -------------------------------------------------------------------------
    // Stateful
    // -------------------------------------------------------------------------

    var boardBorderSize = 2
    var currentOrientation = 'white'
    var currentPosition = {}
    var draggedPiece = null
    var draggedPieceLocation = null
    var draggedPieceSource = null
    var isDragging = false
    var sparePiecesElsIds = {}
    var squareElsIds = {}
    var squareElsOffsets = {}
    var squareSize = 16
    
    // Click-to-move state
    var selectedSquare = null
    var selectedPiece = null

    // -------------------------------------------------------------------------
    // Validation / Errors
    // -------------------------------------------------------------------------

    function error (code, msg, obj) {
      // do nothing if showErrors is not set
      if (
        config.hasOwnProperty('showErrors') !== true ||
          config.showErrors === false
      ) {
        return
      }

      var errorText = 'Chessboard Error ' + code + ': ' + msg

      // print to console
      if (
        config.showErrors === 'console' &&
          typeof console === 'object' &&
          typeof console.log === 'function'
      ) {
        console.log(errorText)
        if (arguments.length >= 2) {
          console.log(obj)
        }
        return
      }

      // alert errors (replaced with console.error for better UX)
      if (config.showErrors === 'alert') {
        if (obj) {
          errorText += '\n\n' + JSON.stringify(obj)
        }
        console.error(errorText)
        return
      }

      // custom function
      if (isFunction(config.showErrors)) {
        config.showErrors(code, msg, obj)
      }
    }

    function setInitialState () {
      currentOrientation = config.orientation

      // make sure position is valid
      if (config.hasOwnProperty('position')) {
        if (config.position === 'start') {
          currentPosition = deepCopy(START_POSITION)
        } else if (validFen(config.position)) {
          currentPosition = fenToObj(config.position)
        } else if (validPositionObject(config.position)) {
          currentPosition = deepCopy(config.position)
        } else {
          error(
            7263,
            'Invalid value passed to config.position.',
            config.position
          )
        }
      }
    }

    // -------------------------------------------------------------------------
    // DOM Misc
    // -------------------------------------------------------------------------

    // calculates square size based on the width of the container
    // got a little CSS black magic here, so let me explain:
    // get the width of the container element (could be anything), reduce by 1 for
    // fudge factor, and then keep reducing until we find an exact mod 8 for
    // our square size
    function calculateSquareSize () {
      var containerWidth = parseInt($container.width(), 10)

      // defensive, prevent infinite loop
      if (!containerWidth || containerWidth <= 0) {
        return 0
      }

      // pad one pixel
      var boardWidth = containerWidth - 1

      while (boardWidth % 8 !== 0 && boardWidth > 0) {
        boardWidth = boardWidth - 1
      }

      return boardWidth / 8
    }

    // create random IDs for elements
    function createElIds () {
      // squares on the board
      for (var i = 0; i < COLUMNS.length; i++) {
        for (var j = 1; j <= 8; j++) {
          var square = COLUMNS[i] + j
          squareElsIds[square] = square + '-' + uuid()
        }
      }

      // spare pieces
      var pieces = 'KQRNBP'.split('')
      for (i = 0; i < pieces.length; i++) {
        var whitePiece = 'w' + pieces[i]
        var blackPiece = 'b' + pieces[i]
        sparePiecesElsIds[whitePiece] = whitePiece + '-' + uuid()
        sparePiecesElsIds[blackPiece] = blackPiece + '-' + uuid()
      }
    }

    // -------------------------------------------------------------------------
    // Markup Building
    // -------------------------------------------------------------------------

    function buildBoardHTML (orientation) {
      if (orientation !== 'black') {
        orientation = 'white'
      }

      var html = ''

      // algebraic notation / orientation
      var alpha = deepCopy(COLUMNS)
      var row = 8
      if (orientation === 'black') {
        alpha.reverse()
        row = 1
      }

      var squareColor = 'white'
      for (var i = 0; i < 8; i++) {
        html += '<div class="{row}">'
        for (var j = 0; j < 8; j++) {
          var square = alpha[j] + row

          html += '<div class="{square} ' + CSS[squareColor] + ' ' +
            'square-' + square + '" ' +
            'style="width:' + squareSize + 'px;height:' + squareSize + 'px;" ' +
            'id="' + squareElsIds[square] + '" ' +
            'data-square="' + square + '">'

          if (config.showNotation) {
            // alpha notation
            if ((orientation === 'white' && row === 1) ||
                (orientation === 'black' && row === 8)) {
              html += '<div class="{notation} {alpha}">' + alpha[j] + '</div>'
            }

            // numeric notation
            if (j === 0) {
              html += '<div class="{notation} {numeric}">' + row + '</div>'
            }
          }

          html += '</div>' // end .square

          squareColor = (squareColor === 'white') ? 'black' : 'white'
        }
        html += '<div class="{clearfix}"></div></div>'

        squareColor = (squareColor === 'white') ? 'black' : 'white'

        if (orientation === 'white') {
          row = row - 1
        } else {
          row = row + 1
        }
      }

      return interpolateTemplate(html, CSS)
    }

    function buildPieceImgSrc (piece) {
      if (isFunction(config.pieceTheme)) {
        return config.pieceTheme(piece)
      }

      if (isString(config.pieceTheme)) {
        return interpolateTemplate(config.pieceTheme, {piece: piece})
      }

      // NOTE: this should never happen
      error(8272, 'Unable to build image source for config.pieceTheme.')
      return ''
    }

    function buildPieceHTML (piece, hidden, id) {
      var html = '<img src="' + buildPieceImgSrc(piece) + '" '
      if (isString(id) && id !== '') {
        html += 'id="' + id + '" '
      }
      html += 'alt="" ' +
        'class="{piece}" ' +
        'data-piece="' + piece + '" ' +
        'style="width:' + squareSize + 'px;' + 'height:' + squareSize + 'px;'

      if (hidden) {
        html += 'display:none;'
      }

      html += '" />'

      return interpolateTemplate(html, CSS)
    }

    function buildSparePiecesHTML (color) {
      var pieces = ['wK', 'wQ', 'wR', 'wB', 'wN', 'wP']
      if (color === 'black') {
        pieces = ['bK', 'bQ', 'bR', 'bB', 'bN', 'bP']
      }

      var html = ''
      for (var i = 0; i < pieces.length; i++) {
        html += buildPieceHTML(pieces[i], false, sparePiecesElsIds[pieces[i]])
      }

      return html
    }

    // -------------------------------------------------------------------------
    // Animations
    // -------------------------------------------------------------------------

    function animateSquareToSquare (src, dest, piece, completeFn) {
      // get information about the source and destination squares
      var $srcSquare = $('#' + squareElsIds[src])
      var srcSquarePosition = $srcSquare.offset()
      var $destSquare = $('#' + squareElsIds[dest])
      var destSquarePosition = $destSquare.offset()

      // create the animated piece and absolutely position it
      // over the source square
      var animatedPieceId = uuid()
      $('body').append(buildPieceHTML(piece, true, animatedPieceId))
      var $animatedPiece = $('#' + animatedPieceId)
      $animatedPiece.css({
        display: '',
        position: 'absolute',
        top: srcSquarePosition.top,
        left: srcSquarePosition.left
      })

      // remove original piece from source square
      $srcSquare.find('.' + CSS.piece).remove()

      function onFinishAnimation1 () {
        // add the "real" piece to the destination square
        $destSquare.append(buildPieceHTML(piece))

        // remove the animated piece
        $animatedPiece.remove()

        // run complete function
        if (isFunction(completeFn)) {
          completeFn()
        }
      }

      // animate the piece to the destination square
      var opts = {
        duration: config.moveSpeed,
        complete: onFinishAnimation1
      }
      $animatedPiece.animate(destSquarePosition, opts)
    }

    function animateSparePieceToSquare (piece, dest, completeFn) {
      var srcOffset = $('#' + sparePiecesElsIds[piece]).offset()
      var $destSquare = $('#' + squareElsIds[dest])
      var destOffset = $destSquare.offset()

      // create the animate piece
      var pieceId = uuid()
      $('body').append(buildPieceHTML(piece, true, pieceId))
      var $animatedPiece = $('#' + pieceId)
      $animatedPiece.css({
        display: '',
        position: 'absolute',
        left: srcOffset.left,
        top: srcOffset.top
      })

      // on complete
      function onFinishAnimation2 () {
        // add the "real" piece to the destination square
        $destSquare.find('.' + CSS.piece).remove()
        $destSquare.append(buildPieceHTML(piece))

        // remove the animated piece
        $animatedPiece.remove()

        // run complete function
        if (isFunction(completeFn)) {
          completeFn()
        }
      }

      // animate the piece to the destination square
      var opts = {
        duration: config.moveSpeed,
        complete: onFinishAnimation2
      }
      $animatedPiece.animate(destOffset, opts)
    }

    // execute an array of animations
    function doAnimations (animations, oldPos, newPos) {
      if (animations.length === 0) return

      var numFinished = 0
      function onFinishAnimation3 () {
        // exit if all the animations aren't finished
        numFinished = numFinished + 1
        if (numFinished !== animations.length) return

        drawPositionInstant()

        // run their onMoveEnd function
        if (isFunction(config.onMoveEnd)) {
          config.onMoveEnd(deepCopy(oldPos), deepCopy(newPos))
        }
      }

      for (var i = 0; i < animations.length; i++) {
        var animation = animations[i]

        // clear a piece
        if (animation.type === 'clear') {
          $('#' + squareElsIds[animation.square] + ' .' + CSS.piece)
            .fadeOut(config.trashSpeed, onFinishAnimation3)

        // add a piece with no spare pieces - fade the piece onto the square
        } else if (animation.type === 'add' && !config.sparePieces) {
          $('#' + squareElsIds[animation.square])
            .append(buildPieceHTML(animation.piece, true))
            .find('.' + CSS.piece)
            .fadeIn(config.appearSpeed, onFinishAnimation3)

        // add a piece with spare pieces - animate from the spares
        } else if (animation.type === 'add' && config.sparePieces) {
          animateSparePieceToSquare(animation.piece, animation.square, onFinishAnimation3)

        // move a piece from squareA to squareB
        } else if (animation.type === 'move') {
          animateSquareToSquare(animation.source, animation.destination, animation.piece, onFinishAnimation3)
        }
      }
    }

    // calculate an array of animations that need to happen in order to get
    // from pos1 to pos2
    function calculateAnimations (pos1, pos2) {
      // make copies of both
      pos1 = deepCopy(pos1)
      pos2 = deepCopy(pos2)

      var animations = []
      var squaresMovedTo = {}

      // remove pieces that are the same in both positions
      for (var i in pos2) {
        if (!pos2.hasOwnProperty(i)) continue

        if (pos1.hasOwnProperty(i) && pos1[i] === pos2[i]) {
          delete pos1[i]
          delete pos2[i]
        }
      }

      // find all the "move" animations
      for (i in pos2) {
        if (!pos2.hasOwnProperty(i)) continue

        var closestPiece = findClosestPiece(pos1, pos2[i], i)
        if (closestPiece) {
          animations.push({
            type: 'move',
            source: closestPiece,
            destination: i,
            piece: pos2[i]
          })

          delete pos1[closestPiece]
          delete pos2[i]
          squaresMovedTo[i] = true
        }
      }

      // "add" animations
      for (i in pos2) {
        if (!pos2.hasOwnProperty(i)) continue

        animations.push({
          type: 'add',
          square: i,
          piece: pos2[i]
        })

        delete pos2[i]
      }

      // "clear" animations
      for (i in pos1) {
        if (!pos1.hasOwnProperty(i)) continue

        // do not clear a piece if it is on a square that is the result
        // of a "move", ie: a piece capture
        if (squaresMovedTo.hasOwnProperty(i)) continue

        animations.push({
          type: 'clear',
          square: i,
          piece: pos1[i]
        })

        delete pos1[i]
      }

      return animations
    }

    // -------------------------------------------------------------------------
    // Control Flow
    // -------------------------------------------------------------------------

    function drawPositionInstant () {
      // clear the board
      $board.find('.' + CSS.piece).remove()

      // add the pieces
      for (var i in currentPosition) {
        if (!currentPosition.hasOwnProperty(i)) continue

        $('#' + squareElsIds[i]).append(buildPieceHTML(currentPosition[i]))
      }
    }

    function drawBoard () {
      $board.html(buildBoardHTML(currentOrientation, squareSize, config.showNotation))
      drawPositionInstant()

      if (config.sparePieces) {
        if (currentOrientation === 'white') {
          $sparePiecesTop.html(buildSparePiecesHTML('black'))
          $sparePiecesBottom.html(buildSparePiecesHTML('white'))
        } else {
          $sparePiecesTop.html(buildSparePiecesHTML('white'))
          $sparePiecesBottom.html(buildSparePiecesHTML('black'))
        }
      }
    }

    function setCurrentPosition (position) {
      var oldPos = deepCopy(currentPosition)
      var newPos = deepCopy(position)
      var oldFen = objToFen(oldPos)
      var newFen = objToFen(newPos)

      // do nothing if no change in position
      if (oldFen === newFen) return

      // run their onChange function
      if (isFunction(config.onChange)) {
        config.onChange(oldPos, newPos)
      }

      // update state
      currentPosition = position
    }

    function isXYOnSquare (x, y) {
      for (var i in squareElsOffsets) {
        if (!squareElsOffsets.hasOwnProperty(i)) continue

        var s = squareElsOffsets[i]
        if (x >= s.left &&
            x < s.left + squareSize &&
            y >= s.top &&
            y < s.top + squareSize) {
          return i
        }
      }

      return 'offboard'
    }

    // records the XY coords of every square into memory
    function captureSquareOffsets () {
      squareElsOffsets = {}

      for (var i in squareElsIds) {
        if (!squareElsIds.hasOwnProperty(i)) continue

        squareElsOffsets[i] = $('#' + squareElsIds[i]).offset()
      }
    }

    function removeSquareHighlights () {
      $board
        .find('.' + CSS.square)
        .removeClass(CSS.highlight1 + ' ' + CSS.highlight2)
    }

    function snapbackDraggedPiece () {
      // there is no "snapback" for spare pieces
      if (draggedPieceSource === 'spare') {
        trashDraggedPiece()
        return
      }

      removeSquareHighlights()

      // animation complete
      function complete () {
        drawPositionInstant()
        $draggedPiece.css('display', 'none')

        // run their onSnapbackEnd function
        if (isFunction(config.onSnapbackEnd)) {
          config.onSnapbackEnd(
            draggedPiece,
            draggedPieceSource,
            deepCopy(currentPosition),
            currentOrientation
          )
        }
      }

      // get source square position
      var sourceSquarePosition = $('#' + squareElsIds[draggedPieceSource]).offset()

      // animate the piece to the target square
      var opts = {
        duration: config.snapbackSpeed,
        complete: complete
      }
      $draggedPiece.animate(sourceSquarePosition, opts)

      // set state
      isDragging = false
    }

    function trashDraggedPiece () {
      removeSquareHighlights()

      // remove the source piece
      var newPosition = deepCopy(currentPosition)
      delete newPosition[draggedPieceSource]
      setCurrentPosition(newPosition)

      // redraw the position
      drawPositionInstant()

      // hide the dragged piece
      $draggedPiece.fadeOut(config.trashSpeed)

      // set state
      isDragging = false
    }

    function dropDraggedPieceOnSquare (square) {
      removeSquareHighlights()

      // update position
      var newPosition = deepCopy(currentPosition)
      delete newPosition[draggedPieceSource]
      newPosition[square] = draggedPiece
      setCurrentPosition(newPosition)

      // get target square information
      var targetSquarePosition = $('#' + squareElsIds[square]).offset()

      // animation complete
      function onAnimationComplete () {
        drawPositionInstant()
        $draggedPiece.css('display', 'none')

        // execute their onSnapEnd function
        if (isFunction(config.onSnapEnd)) {
          config.onSnapEnd(draggedPieceSource, square, draggedPiece)
        }
      }

      // snap the piece to the target square
      var opts = {
        duration: config.snapSpeed,
        complete: onAnimationComplete
      }
      $draggedPiece.animate(targetSquarePosition, opts)

      // set state
      isDragging = false
    }

    function beginDraggingPiece (source, piece, x, y) {
      // run their custom onDragStart function
      // their custom onDragStart function can cancel drag start
      if (isFunction(config.onDragStart) &&
          config.onDragStart(source, piece, deepCopy(currentPosition), currentOrientation) === false) {
        return
      }

      // set state
      isDragging = true
      draggedPiece = piece
      draggedPieceSource = source

      // if the piece came from spare pieces, location is offboard
      if (source === 'spare') {
        draggedPieceLocation = 'offboard'
      } else {
        draggedPieceLocation = source
      }

      // capture the x, y coords of all squares in memory
      captureSquareOffsets()

      // create the dragged piece
      $draggedPiece.attr('src', buildPieceImgSrc(piece)).css({
        display: '',
        position: 'absolute',
        left: x - squareSize / 2,
        top: y - squareSize / 2
      })

      if (source !== 'spare') {
        // highlight the source square and hide the piece
        $('#' + squareElsIds[source])
          .addClass(CSS.highlight1)
          .find('.' + CSS.piece)
          .css('display', 'none')
      }
    }

    function updateDraggedPiece (x, y) {
      // put the dragged piece over the mouse cursor
      $draggedPiece.css({
        left: x - squareSize / 2,
        top: y - squareSize / 2
      })

      // get location
      var location = isXYOnSquare(x, y)

      // do nothing if the location has not changed
      if (location === draggedPieceLocation) return

      // remove highlight from previous square
      if (validSquare(draggedPieceLocation)) {
        $('#' + squareElsIds[draggedPieceLocation]).removeClass(CSS.highlight2)
      }

      // add highlight to new square
      if (validSquare(location)) {
        $('#' + squareElsIds[location]).addClass(CSS.highlight2)
      }

      // run onDragMove
      if (isFunction(config.onDragMove)) {
        config.onDragMove(
          location,
          draggedPieceLocation,
          draggedPieceSource,
          draggedPiece,
          deepCopy(currentPosition),
          currentOrientation
        )
      }

      // update state
      draggedPieceLocation = location
    }

    function stopDraggedPiece (location) {
      // determine what the action should be
      var action = 'drop'
      if (location === 'offboard' && config.dropOffBoard === 'snapback') {
        action = 'snapback'
      }
      if (location === 'offboard' && config.dropOffBoard === 'trash') {
        action = 'trash'
      }

      // run their onDrop function, which can potentially change the drop action
      if (isFunction(config.onDrop)) {
        var newPosition = deepCopy(currentPosition)

        // source piece is a spare piece and position is off the board
        // if (draggedPieceSource === 'spare' && location === 'offboard') {...}
        // position has not changed; do nothing

        // source piece is a spare piece and position is on the board
        if (draggedPieceSource === 'spare' && validSquare(location)) {
          // add the piece to the board
          newPosition[location] = draggedPiece
        }

        // source piece was on the board and position is off the board
        if (validSquare(draggedPieceSource) && location === 'offboard') {
          // remove the piece from the board
          delete newPosition[draggedPieceSource]
        }

        // source piece was on the board and position is on the board
        if (validSquare(draggedPieceSource) && validSquare(location)) {
          // move the piece
          delete newPosition[draggedPieceSource]
          newPosition[location] = draggedPiece
        }

        var oldPosition = deepCopy(currentPosition)

        var result = config.onDrop(
          draggedPieceSource,
          location,
          draggedPiece,
          newPosition,
          oldPosition,
          currentOrientation
        )
        if (result === 'snapback' || result === 'trash') {
          action = result
        }
      }

      // do it!
      if (action === 'snapback') {
        snapbackDraggedPiece()
      } else if (action === 'trash') {
        trashDraggedPiece()
      } else if (action === 'drop') {
        dropDraggedPieceOnSquare(location)
      }
    }

    // -------------------------------------------------------------------------
    // Public Methods
    // -------------------------------------------------------------------------

    // clear the board
    widget.clear = function (useAnimation) {
      widget.position({}, useAnimation)
    }

    // remove the widget from the page
    widget.destroy = function () {
      // remove markup
      $container.html('')
      $draggedPiece.remove()

      // remove event handlers
      $container.unbind()
    }

    // shorthand method to get the current FEN
    widget.fen = function () {
      return widget.position('fen')
    }

    // flip orientation
    widget.flip = function () {
      return widget.orientation('flip')
    }

    // move pieces
    // TODO: this method should be variadic as well as accept an array of moves
    widget.move = function () {
      // no need to throw an error here; just do nothing
      // TODO: this should return the current position
      if (arguments.length === 0) return

      var useAnimation = true

      // collect the moves into an object
      var moves = {}
      for (var i = 0; i < arguments.length; i++) {
        // any "false" to this function means no animations
        if (arguments[i] === false) {
          useAnimation = false
          continue
        }

        // skip invalid arguments
        if (!validMove(arguments[i])) {
          error(2826, 'Invalid move passed to the move method.', arguments[i])
          continue
        }

        var tmp = arguments[i].split('-')
        moves[tmp[0]] = tmp[1]
      }

      // calculate position from moves
      var newPos = calculatePositionFromMoves(currentPosition, moves)

      // update the board
      widget.position(newPos, useAnimation)

      // return the new position object
      return newPos
    }

    widget.orientation = function (arg) {
      // no arguments, return the current orientation
      if (arguments.length === 0) {
        return currentOrientation
      }

      // set to white or black
      if (arg === 'white' || arg === 'black') {
        currentOrientation = arg
        drawBoard()
        return currentOrientation
      }

      // flip orientation
      if (arg === 'flip') {
        currentOrientation = currentOrientation === 'white' ? 'black' : 'white'
        drawBoard()
        return currentOrientation
      }

      error(5482, 'Invalid value passed to the orientation method.', arg)
    }

    widget.position = function (position, useAnimation) {
      // no arguments, return the current position
      if (arguments.length === 0) {
        return deepCopy(currentPosition)
      }

      // get position as FEN
      if (isString(position) && position.toLowerCase() === 'fen') {
        return objToFen(currentPosition)
      }

      // start position
      if (isString(position) && position.toLowerCase() === 'start') {
        position = deepCopy(START_POSITION)
      }

      // convert FEN to position object
      if (validFen(position)) {
        position = fenToObj(position)
      }

      // validate position object
      if (!validPositionObject(position)) {
        error(6482, 'Invalid value passed to the position method.', position)
        return
      }

      // default for useAnimations is true
      if (useAnimation !== false) useAnimation = true

      if (useAnimation) {
        // start the animations
        var animations = calculateAnimations(currentPosition, position)
        doAnimations(animations, currentPosition, position)

        // set the new position
        setCurrentPosition(position)
      } else {
        // instant update
        setCurrentPosition(position)
        drawPositionInstant()
      }
    }

    widget.resize = function () {
      // calulate the new square size
      squareSize = calculateSquareSize()

      // set board width
      $board.css('width', squareSize * 8 + 'px')

      // set drag piece size
      $draggedPiece.css({
        height: squareSize,
        width: squareSize
      })

      // spare pieces
      if (config.sparePieces) {
        $container
          .find('.' + CSS.sparePieces)
          .css('paddingLeft', squareSize + boardBorderSize + 'px')
      }

      // redraw the board
      drawBoard()
    }

    // set the starting position
    widget.start = function (useAnimation) {
      widget.position('start', useAnimation)
    }

    // -------------------------------------------------------------------------
    // Arrow Drawing (v1.0.1 addition)
    // -------------------------------------------------------------------------

    var $arrowSvg = null
    var currentArrow = null

    function initArrowSvg () {
      if ($arrowSvg) return

      // Create SVG overlay
      var svgId = 'arrow-svg-' + uuid()
      var markerId = 'arrowhead-' + uuid()
      var svgHTML = '<svg id="' + svgId + '" style="position: absolute; top: 0; left: 0; width: 100%; height: 100%; pointer-events: none; z-index: 100 !important; overflow: visible;">' +
        '<defs>' +
        '<marker id="' + markerId + '" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">' +
        '<polygon points="0 0, 10 3, 0 6" fill="rgba(50, 150, 255, 0.8)" />' +
        '</marker>' +
        '</defs>' +
        '</svg>'
      
      // Ensure container has position relative
      if ($container.css('position') === 'static') {
        $container.css('position', 'relative')
      }
      
      // Get board position and size for SVG positioning
      var boardOffset = $board.position()
      var boardWidth = $board.width()
      var boardHeight = $board.height()
      
      // Create SVG using DOM methods with proper namespace
      var svgNS = 'http://www.w3.org/2000/svg'
      var svg = document.createElementNS(svgNS, 'svg')
      svg.setAttribute('id', svgId)
      svg.setAttribute('width', boardWidth)
      svg.setAttribute('height', boardHeight)
      svg.style.position = 'absolute'
      svg.style.top = boardOffset.top + 'px'
      svg.style.left = boardOffset.left + 'px'
      svg.style.pointerEvents = 'none'
      svg.style.zIndex = '10'  // Above pieces but below modals
      svg.style.overflow = 'visible'
      
      // Create defs element for markers (will be populated dynamically)
      var defs = document.createElementNS(svgNS, 'defs')
      svg.appendChild(defs)
      
      $container[0].appendChild(svg)
      $arrowSvg = $('#' + svgId)
      $arrowSvg.data('markers', {})  // Store marker IDs by color
    }

    function getSquareCenter (square) {
      var squareEl = $('#' + squareElsIds[square])
      if (!squareEl.length) return null

      var offset = squareEl.position()
      return {
        x: offset.left + squareSize / 2,
        y: offset.top + squareSize / 2
      }
    }

    function getTextStrokeColor (arrowColor) {
      // Remove # if present
      var hex = arrowColor.replace('#', '')
      
      // Handle 8-digit hex (with alpha channel)
      if (hex.length === 8) {
        hex = hex.substring(0, 6)
      }
      
      // Convert to RGB
      var r = parseInt(hex.substring(0, 2), 16)
      var g = parseInt(hex.substring(2, 4), 16)
      var b = parseInt(hex.substring(4, 6), 16)
      
      // Calculate relative luminance using WCAG formula
      // https://www.w3.org/TR/WCAG20-TECHS/G17.html
      var luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255
      
      // Return opposite color for stroke:
      // - Light arrows get black stroke
      // - Dark arrows get white stroke
      return luminance > 0.5 ? '#000000' : '#ffffff'
    }

    function getOrCreateMarker (color, opacity) {
      // Create a unique key for this color/opacity combination
      var markerKey = color + '_' + opacity
      var markers = $arrowSvg.data('markers')
      
      // Return existing marker if already created
      if (markers[markerKey]) {
        return markers[markerKey]
      }
      
      // Create new marker
      var svgNS = 'http://www.w3.org/2000/svg'
      var markerId = 'arrowhead-' + uuid()
      var marker = document.createElementNS(svgNS, 'marker')
      marker.setAttribute('id', markerId)
      marker.setAttribute('markerWidth', '4')
      marker.setAttribute('markerHeight', '4')
      marker.setAttribute('refX', '2')  // Position middle of arrowhead at line end for centered appearance
      marker.setAttribute('refY', '1.5')
      marker.setAttribute('orient', 'auto')
      marker.setAttribute('markerUnits', 'strokeWidth')
      
      var polygon = document.createElementNS(svgNS, 'polygon')
      polygon.setAttribute('points', '0 0, 4 1.5, 0 3')
      polygon.setAttribute('fill', color)
      polygon.setAttribute('fill-opacity', opacity)
      
      marker.appendChild(polygon)
      
      // Add marker to defs
      var defs = $arrowSvg.find('defs')[0]
      defs.appendChild(marker)
      
      // Cache the marker ID
      markers[markerKey] = markerId
      $arrowSvg.data('markers', markers)
      
      return markerId
    }

    widget.drawArrow = function (fromSquare, toSquare, color, label, opacity, clearPrevious, moveNumber) {
      if (!validSquare(fromSquare) || !validSquare(toSquare)) {
        return
      }

      initArrowSvg()

      var from = getSquareCenter(fromSquare)
      var to = getSquareCenter(toSquare)

      if (!from || !to) {
        return
      }

      // Calculate arrow direction - start from center, shorten only at end for arrowhead
      var dx = to.x - from.x
      var dy = to.y - from.y
      var length = Math.sqrt(dx * dx + dy * dy)
      
      var unitX = dx / length
      var unitY = dy / length
      
      // Start from center of source square
      var startX = from.x
      var startY = from.y
      
      // Calculate stroke width for arrowhead sizing
      var strokeWidth = squareSize * 0.12
      
      // Shorten arrow so arrowhead tip lands at center
      // With refX=2, the arrowhead extends 2 strokeWidths forward from line end
      var shortenBy = strokeWidth * 2
      var endX = to.x - unitX * shortenBy
      var endY = to.y - unitY * shortenBy

      // Clear previous arrows if requested (default true for backward compatibility)
      if (clearPrevious === undefined || clearPrevious === true) {
        $arrowSvg.find('line').remove()
        $arrowSvg.find('text').remove()
      }

      // Default color - solid blue
      if (!color) color = '#3296FF'
      
      // Default opacity
      if (!opacity) opacity = 0.8

      // Get or create marker for this color/opacity combination
      var markerId = getOrCreateMarker(color, opacity)

      // Draw new arrow using proper SVG namespace
      var svgNS = 'http://www.w3.org/2000/svg'
      var line = document.createElementNS(svgNS, 'line')
      line.setAttribute('x1', startX)
      line.setAttribute('y1', startY)
      line.setAttribute('x2', endX)
      line.setAttribute('y2', endY)
      line.setAttribute('stroke', color)
      line.setAttribute('stroke-width', strokeWidth)
      line.setAttribute('stroke-linecap', 'round')
      line.setAttribute('opacity', opacity)
      line.setAttribute('stroke-opacity', opacity)
      line.setAttribute('marker-end', 'url(#' + markerId + ')')
      
      $arrowSvg[0].appendChild(line)
      
      // Add text label if provided (score label at midpoint)
      if (label) {
        // Position label at the midpoint of the arrow, centered on it
        var midX = (startX + endX) / 2
        var midY = (startY + endY) / 2
        
        // Calculate contrasting stroke color based on arrow color
        var strokeColor = getTextStrokeColor(color)
        
        // Create text element
        var text = document.createElementNS(svgNS, 'text')
        text.setAttribute('x', midX)
        text.setAttribute('y', midY)
        text.setAttribute('text-anchor', 'middle')
        text.setAttribute('dominant-baseline', 'middle')
        text.setAttribute('font-family', 'Arial, sans-serif')
        text.setAttribute('font-size', squareSize * 0.28)
        text.setAttribute('font-weight', 'bold')
        text.setAttribute('fill', color)
        text.setAttribute('stroke', strokeColor)
        text.setAttribute('stroke-width', '2')
        text.setAttribute('paint-order', 'stroke')
        text.setAttribute('opacity', opacity)
        text.textContent = label
        
        $arrowSvg[0].appendChild(text)
      }
      
      // Add move number near arrowhead if provided
      if (moveNumber !== undefined && moveNumber !== null) {
        // Position near the arrowhead, offset perpendicular to arrow
        var perpX = -unitY
        var perpY = unitX
        var numberOffset = squareSize * 0.25
        var numberX = endX + perpX * numberOffset
        var numberY = endY + perpY * numberOffset
        
        // Calculate contrasting stroke color based on arrow color
        var strokeColor = getTextStrokeColor(color)
        
        // Create text element for move number
        var numberText = document.createElementNS(svgNS, 'text')
        numberText.setAttribute('x', numberX)
        numberText.setAttribute('y', numberY)
        numberText.setAttribute('text-anchor', 'middle')
        numberText.setAttribute('dominant-baseline', 'middle')
        numberText.setAttribute('font-family', 'Arial, sans-serif')
        numberText.setAttribute('font-size', squareSize * 0.22)
        numberText.setAttribute('font-weight', 'bold')
        numberText.setAttribute('fill', color)
        numberText.setAttribute('stroke', strokeColor)
        numberText.setAttribute('stroke-width', '2')
        numberText.setAttribute('paint-order', 'stroke')
        numberText.setAttribute('opacity', opacity)
        numberText.textContent = moveNumber
        
        $arrowSvg[0].appendChild(numberText)
      }
      
      currentArrow = { from: fromSquare, to: toSquare, color: color, label: label, opacity: opacity, moveNumber: moveNumber }
    }

    widget.clearArrow = function () {
      if ($arrowSvg) {
        $arrowSvg.find('line').remove()
        $arrowSvg.find('text').remove()
      }
      currentArrow = null
    }

    widget.getArrow = function () {
      return currentArrow
    }

    // -------------------------------------------------------------------------
    // Ghost Pieces and PV Animation
    // -------------------------------------------------------------------------

    var pvAnimationTimeouts = [] // Track animation timeouts for cancellation
    var currentPVSequence = null // Track current PV sequence for comparison
    var ghostPieces = [] // Track ghost pieces for cleanup
    var pendingPVData = null // Store pending PV to apply after current loop finishes
    var isAnimating = false // Track if animation is currently running
    var positionChanged = false // Track if position changed (move made) to force-start next PV

    // Cancel ongoing PV animation
    widget.cancelPVAnimation = function () {
      // Clear all pending timeouts
      for (var i = 0; i < pvAnimationTimeouts.length; i++) {
        clearTimeout(pvAnimationTimeouts[i])
      }
      pvAnimationTimeouts = []
      currentPVSequence = null
      pendingPVData = null
      isAnimating = false
      positionChanged = false
      // Also clear ghost pieces
      widget.clearGhostPieces()
    }

    // Set position changed flag (called when move is made)
    widget.setPositionChanged = function () {
      positionChanged = true
    }

    // Add a ghost piece to the board
    widget.addGhostPiece = function (fromSquare, toSquare, piece) {
      // Remove any existing ghost pieces from source and destination squares
      var $fromSquare = $container.find('.square-' + fromSquare)
      if ($fromSquare.length > 0) {
        $fromSquare.find('.ghost-piece').remove()
      }

      var $toSquare = $container.find('.square-' + toSquare)
      if ($toSquare.length === 0) return
      $toSquare.find('.ghost-piece').remove()

      // Hide the piece on the source square
      var $sourcePiece = null
      if ($fromSquare.length > 0) {
        $sourcePiece = $fromSquare.find('img.' + CSS['piece'])
        if ($sourcePiece.length > 0) {
          $sourcePiece.css('visibility', 'hidden')
        }
      }

      // Hide any piece on the destination square (for captures)
      var $destPiece = $toSquare.find('img.' + CSS['piece'])
      if ($destPiece.length > 0) {
        $destPiece.css('visibility', 'hidden')
      }

      // Create ghost piece image
      var pieceImage = config.pieceTheme.replace('{piece}', piece)
      var $ghostPiece = $('<img>')
        .attr('src', pieceImage)
        .addClass('ghost-piece')
        .addClass('ghost-fade-in')
        .css({
          width: '100%',
          height: '100%',
          position: 'absolute',
          top: 0,
          left: 0
        })

      // Add to destination square
      $toSquare.append($ghostPiece)

      // Track for cleanup (including both source and destination pieces for restoration)
      ghostPieces.push({
        fromSquare: fromSquare,
        toSquare: toSquare,
        element: $ghostPiece,
        sourcePiece: $sourcePiece,
        destPiece: $destPiece
      })
    }

    // Clear all ghost pieces
    widget.clearGhostPieces = function () {
      for (var i = 0; i < ghostPieces.length; i++) {
        ghostPieces[i].element.remove()
        // Restore source piece visibility if it was hidden
        if (ghostPieces[i].sourcePiece && ghostPieces[i].sourcePiece.length > 0) {
          ghostPieces[i].sourcePiece.css('visibility', 'visible')
        }
        // Restore destination piece visibility if it was hidden (for captures)
        if (ghostPieces[i].destPiece && ghostPieces[i].destPiece.length > 0) {
          ghostPieces[i].destPiece.css('visibility', 'visible')
        }
      }
      ghostPieces = []
    }

    // -------------------------------------------------------------------------
    // Case 1: Draw single best move arrow (no animation, no ghost pieces)
    // -------------------------------------------------------------------------
    
    widget.drawBestMove = function (pvData) {
      if (!pvData || !pvData.moves || pvData.moves.length === 0) {
        console.error('drawBestMove requires pre-computed pvData with at least one move')
        return
      }

      // Cancel any ongoing animation
      widget.cancelPVAnimation()
      
      // Draw just the first move without ghost pieces
      widget.drawPVArrowAtIndex(pvData, 0, true, false)
    }

    // -------------------------------------------------------------------------
    // Case 2: Draw PV animation (looping animation with ghost pieces)
    // -------------------------------------------------------------------------
    
    widget.drawPVAnimation = function (pvData, options) {
      if (!pvData || !pvData.moves || pvData.moves.length === 0) {
        console.error('drawPVAnimation requires pre-computed pvData with moves')
        return
      }

      // Default options
      options = options || {}
      var maxMoves = options.maxMoves !== undefined ? options.maxMoves : Math.min(6, pvData.moves.length)
      var firstMoveDelay = options.firstMoveDelay || 2000  // 2 seconds for first move
      var subsequentMoveDelay = options.subsequentMoveDelay || 1500  // 1.5 seconds for subsequent
      var pauseAfterLoop = options.pauseAfterLoop || 2000  // 2 seconds pause after last move

      // Check if this is a new PV sequence
      var newPVSequence = pvData.moves.length + ':' + pvData.moves[0].from + pvData.moves[0].to
      if (newPVSequence === currentPVSequence) {
        // Same PV, don't restart animation
        return
      }

      // If position changed (move made), cancel everything and start immediately
      if (positionChanged) {
        widget.cancelPVAnimation()
        currentPVSequence = newPVSequence
        isAnimating = true
        positionChanged = false
      } else {
        // Engine sent new PV during animation - queue it
        if (isAnimating) {
          pendingPVData = { pvData: pvData, options: options }
          return
        }
        // No animation running, start new one
        currentPVSequence = newPVSequence
        isAnimating = true
      }

      function scheduleAnimation (loopIteration) {
        var cumulativeDelay = 0

        for (var i = 0; i < maxMoves; i++) {
          (function (index) {
            var timeout = setTimeout(function () {
              // Check if this animation is still valid
              if (currentPVSequence === newPVSequence) {
                // Clear previous arrows but keep ghost pieces
                var clearPrevious = true
                // Only clear ghost pieces at the start of a new loop (first move)
                var clearGhosts = (index === 0)
                widget.drawPVArrowAtIndex(pvData, index, clearPrevious, true, clearGhosts)

                // If this is the last arrow, check for pending PV or schedule next loop
                if (index === maxMoves - 1) {
                  var finalDelay = index === 0 ? firstMoveDelay : subsequentMoveDelay
                  var loopTimeout = setTimeout(function () {
                    if (currentPVSequence === newPVSequence) {
                      // Check if there's a pending PV to start
                      if (pendingPVData) {
                        var pending = pendingPVData
                        pendingPVData = null
                        isAnimating = false
                        // Start the pending PV animation
                        widget.drawPVAnimation(pending.pvData, pending.options)
                      } else {
                        // Continue looping with current PV
                        scheduleAnimation(loopIteration + 1)
                      }
                    }
                  }, finalDelay + pauseAfterLoop)
                  pvAnimationTimeouts.push(loopTimeout)
                }
              }
            }, cumulativeDelay)
            pvAnimationTimeouts.push(timeout)

            // Calculate delay for next arrow
            cumulativeDelay += (index === 0) ? firstMoveDelay : subsequentMoveDelay
          })(i)
        }
      }

      // Start the first animation loop
      scheduleAnimation(0)
    }

    // -------------------------------------------------------------------------
    // Backward compatibility: Keep old method name as wrapper
    // -------------------------------------------------------------------------
    
    widget.drawPrincipalVariation = function (pvData, showPV, showBestMove) {
      // Deprecated: Use drawBestMove() or drawPVAnimation() instead
      if (!showPV) {
        widget.drawBestMove(pvData)
      } else {
        widget.drawPVAnimation(pvData)
      }
    }

    // Draw a single PV arrow at the specified index
    // Accepts pre-computed move data (no Chess.js dependency)
    widget.drawPVArrowAtIndex = function (pvData, index, clearPrevious, showGhostPieces, clearGhosts) {
      // Default parameters
      if (showGhostPieces === undefined) showGhostPieces = false
      if (clearGhosts === undefined) clearGhosts = true
      
      if (!pvData || !pvData.moves || index >= pvData.moves.length) {
        console.error('drawPVArrowAtIndex: invalid pvData or index')
        return
      }

      var move = pvData.moves[index]

      // Clear ghost pieces only if requested (at start of new loop)
      if (clearGhosts) {
        widget.clearGhostPieces()
      }

      // Add ghost piece at destination square (only in PV mode)
      if (showGhostPieces && move.piece) {
        var pieceCode = move.piece.color + move.piece.type.toUpperCase()
        widget.addGhostPiece(move.from, move.to, pieceCode)
      }

      // Use different colors based on whose turn it is
      // White to move = white arrow, Black to move = black arrow
      var arrowColor = move.isBlackMove ? '#000000' : '#FFFFFF'

      // Use 0.8 opacity for arrows
      var opacity = 0.8

      // Only show score label on the first arrow
      var scoreLabel = ''
      if (index === 0) {
        scoreLabel = formatScoreLabel(pvData.scoreType, pvData.score)
      }

      // Add move number to all arrows
      var moveNumberLabel = move.isBlackMove ? move.moveNumber + '...' : move.moveNumber.toString()

      // Draw arrow
      widget.drawArrow(move.from, move.to, arrowColor, scoreLabel, opacity, clearPrevious, moveNumberLabel)
    }

    // -------------------------------------------------------------------------
    // Score Formatting Helper
    // -------------------------------------------------------------------------

    function formatScoreLabel (scoreType, score) {
      if (scoreType === 'cp' && score !== undefined) {
        var scoreValue = (score / 100).toFixed(2)
        return (score >= 0 ? '+' : '') + scoreValue
      } else if (scoreType === 'mate' && score !== undefined) {
        // Show sign for mate: positive = White mates, negative = Black mates
        return (score >= 0 ? '+' : '-') + 'M' + Math.abs(score)
      }
      return ''
    }

    // Make it available externally if needed
    widget.formatScoreLabel = formatScoreLabel

    // -------------------------------------------------------------------------
    // Multiple Best Moves Visualization
    // -------------------------------------------------------------------------

    // Draw multiple best moves (alternative first moves from multi-PV analysis)
    // Accepts pre-computed move data (no Chess.js dependency)
    widget.drawMultipleBestMoves = function (multiPVLines, options) {
      if (!multiPVLines || !Array.isArray(multiPVLines)) {
        console.error('drawMultipleBestMoves requires pre-computed multiPVLines array')
        return
      }

      // Default options
      options = options || {}
      var colors = options.colors || ['#15781Bff', '#FFD700ff', '#DC3545ff'] // Green, Yellow, Red
      var maxLines = options.maxLines || 3
      var opacity = options.opacity !== undefined ? options.opacity : 1.0

      for (var i = 0; i < Math.min(maxLines, multiPVLines.length); i++) {
        var line = multiPVLines[i]

        if (!line || !line.from || !line.to) continue

        var arrowColor = colors[i % colors.length] // Support more than 3 lines if colors provided

        // Format score label
        var scoreLabel = formatScoreLabel(line.scoreType, line.score)

        // Draw arrow (clear previous only on first arrow)
        var clearPrevious = (i === 0)
        widget.drawArrow(line.from, line.to, arrowColor, scoreLabel, opacity, clearPrevious, null)
      }
    }

    // -------------------------------------------------------------------------
    // Browser Events
    // -------------------------------------------------------------------------

    function stopDefault (evt) {
      evt.preventDefault()
    }

    function mousedownSquare (evt) {
      // do nothing if we're not draggable
      if (!config.draggable) return

      var square = $(this).attr('data-square')
      if (!validSquare(square)) return

      var startX = evt.pageX
      var startY = evt.pageY
      var hasPiece = currentPosition.hasOwnProperty(square)
      var dragStarted = false
      
      // Mouse move handler - start drag if moved enough
      var mousemoveHandler = function(moveEvt) {
        var dx = Math.abs(moveEvt.pageX - startX)
        var dy = Math.abs(moveEvt.pageY - startY)
        
        // Start dragging if moved more than 5 pixels and has a piece
        if ((dx > 5 || dy > 5) && hasPiece && !dragStarted) {
          dragStarted = true
          $(document).off('mousemove', mousemoveHandler)
          $(document).off('mouseup', mouseupHandler)
          beginDraggingPiece(square, currentPosition[square], moveEvt.pageX, moveEvt.pageY)
        }
      }
      
      // Mouse up handler - treat as click if didn't drag
      var mouseupHandler = function(upEvt) {
        $(document).off('mousemove', mousemoveHandler)
        $(document).off('mouseup', mouseupHandler)
        
        if (!dragStarted) {
          handleSquareClick(square)
          evt.preventDefault()
        }
      }
      
      // Register temporary handlers
      $(document).on('mousemove', mousemoveHandler)
      $(document).one('mouseup', mouseupHandler)
    }
    
    function handleSquareClick (square) {
      console.log('Square clicked:', square, 'Selected:', selectedSquare)
      
      // If a piece is already selected
      if (selectedSquare !== null) {
        // If clicking the same square, deselect
        if (selectedSquare === square) {
          console.log('Deselecting piece')
          deselectPiece()
          return
        }
        
        // Try to move the selected piece to this square
        console.log('Attempting move from', selectedSquare, 'to', square)
        var move = attemptClickMove(selectedSquare, square)
        
        if (move === 'valid') {
          // Move was successful, deselect
          console.log('Move successful')
          deselectPiece()
        } else if (currentPosition.hasOwnProperty(square)) {
          // Clicked on another piece, select it instead
          console.log('Selecting different piece')
          deselectPiece()
          selectPiece(square, currentPosition[square])
        } else {
          // Invalid move, keep selection
          console.log('Invalid move, keeping selection')
        }
      } else {
        // No piece selected, select this one if it exists
        if (currentPosition.hasOwnProperty(square)) {
          console.log('Selecting piece at', square)
          selectPiece(square, currentPosition[square])
        }
      }
    }
    
    function selectPiece (square, piece) {
      selectedSquare = square
      selectedPiece = piece
      
      // Highlight the selected square
      $('#' + squareElsIds[square]).addClass(CSS.highlight1)
    }
    
    function deselectPiece () {
      if (selectedSquare !== null) {
        $('#' + squareElsIds[selectedSquare]).removeClass(CSS.highlight1)
      }
      selectedSquare = null
      selectedPiece = null
    }
    
    function attemptClickMove (fromSquare, toSquare) {
      // Build new position
      var newPosition = deepCopy(currentPosition)
      delete newPosition[fromSquare]
      newPosition[toSquare] = selectedPiece
      
      var oldPosition = deepCopy(currentPosition)
      
      // Call onDrop callback if it exists
      if (isFunction(config.onDrop)) {
        var result = config.onDrop(
          fromSquare,
          toSquare,
          selectedPiece,
          newPosition,
          oldPosition,
          currentOrientation
        )
        
        // If onDrop returns 'snapback', the move is invalid
        if (result === 'snapback' || result === 'trash') {
          return 'invalid'
        }
      }
      
      // Update the position
      setCurrentPosition(newPosition)
      
      // Animate the move
      var $piece = $('#' + squareElsIds[fromSquare]).find('.' + CSS.piece)
      var targetSquarePosition = $('#' + squareElsIds[toSquare]).offset()
      
      // Temporarily show piece at original location
      $piece.css('display', '')
      
      var opts = {
        duration: config.moveSpeed,
        complete: function() {
          drawPositionInstant()
          
          // Call onSnapEnd callback
          if (isFunction(config.onSnapEnd)) {
            config.onSnapEnd(fromSquare, toSquare, selectedPiece)
          }
        }
      }
      
      $piece.animate(targetSquarePosition, opts)
      
      return 'valid'
    }

    function touchstartSquare (e) {
      // do nothing if we're not draggable
      if (!config.draggable) return

      // do nothing if there is no piece on this square
      var square = $(this).attr('data-square')
      if (!validSquare(square)) return
      if (!currentPosition.hasOwnProperty(square)) return

      e = e.originalEvent
      beginDraggingPiece(
        square,
        currentPosition[square],
        e.changedTouches[0].pageX,
        e.changedTouches[0].pageY
      )
    }

    function mousedownSparePiece (evt) {
      // do nothing if sparePieces is not enabled
      if (!config.sparePieces) return

      var piece = $(this).attr('data-piece')

      beginDraggingPiece('spare', piece, evt.pageX, evt.pageY)
    }

    function touchstartSparePiece (e) {
      // do nothing if sparePieces is not enabled
      if (!config.sparePieces) return

      var piece = $(this).attr('data-piece')

      e = e.originalEvent
      beginDraggingPiece(
        'spare',
        piece,
        e.changedTouches[0].pageX,
        e.changedTouches[0].pageY
      )
    }

    function mousemoveWindow (evt) {
      if (isDragging) {
        updateDraggedPiece(evt.pageX, evt.pageY)
      }
    }

    var throttledMousemoveWindow = throttle(mousemoveWindow, config.dragThrottleRate)

    function touchmoveWindow (evt) {
      // do nothing if we are not dragging a piece
      if (!isDragging) return

      // prevent screen from scrolling
      evt.preventDefault()

      updateDraggedPiece(evt.originalEvent.changedTouches[0].pageX,
        evt.originalEvent.changedTouches[0].pageY)
    }

    var throttledTouchmoveWindow = throttle(touchmoveWindow, config.dragThrottleRate)

    function mouseupWindow (evt) {
      // do nothing if we are not dragging a piece
      if (!isDragging) return

      // get the location
      var location = isXYOnSquare(evt.pageX, evt.pageY)

      stopDraggedPiece(location)
    }

    function touchendWindow (evt) {
      // do nothing if we are not dragging a piece
      if (!isDragging) return

      // get the location
      var location = isXYOnSquare(evt.originalEvent.changedTouches[0].pageX,
        evt.originalEvent.changedTouches[0].pageY)

      stopDraggedPiece(location)
    }

    function mouseenterSquare (evt) {
      // do not fire this event if we are dragging a piece
      // NOTE: this should never happen, but it's a safeguard
      if (isDragging) return

      // exit if they did not provide a onMouseoverSquare function
      if (!isFunction(config.onMouseoverSquare)) return

      // get the square
      var square = $(evt.currentTarget).attr('data-square')

      // NOTE: this should never happen; defensive
      if (!validSquare(square)) return

      // get the piece on this square
      var piece = false
      if (currentPosition.hasOwnProperty(square)) {
        piece = currentPosition[square]
      }

      // execute their function
      config.onMouseoverSquare(square, piece, deepCopy(currentPosition), currentOrientation)
    }

    function mouseleaveSquare (evt) {
      // do not fire this event if we are dragging a piece
      // NOTE: this should never happen, but it's a safeguard
      if (isDragging) return

      // exit if they did not provide an onMouseoutSquare function
      if (!isFunction(config.onMouseoutSquare)) return

      // get the square
      var square = $(evt.currentTarget).attr('data-square')

      // NOTE: this should never happen; defensive
      if (!validSquare(square)) return

      // get the piece on this square
      var piece = false
      if (currentPosition.hasOwnProperty(square)) {
        piece = currentPosition[square]
      }

      // execute their function
      config.onMouseoutSquare(square, piece, deepCopy(currentPosition), currentOrientation)
    }

    // -------------------------------------------------------------------------
    // Initialization
    // -------------------------------------------------------------------------

    function addEvents () {
      // prevent "image drag"
      $('body').on('mousedown mousemove', '.' + CSS.piece, stopDefault)

      // mouse drag pieces
      $board.on('mousedown', '.' + CSS.square, mousedownSquare)
      $container.on('mousedown', '.' + CSS.sparePieces + ' .' + CSS.piece, mousedownSparePiece)

      // mouse enter / leave square
      $board
        .on('mouseenter', '.' + CSS.square, mouseenterSquare)
        .on('mouseleave', '.' + CSS.square, mouseleaveSquare)

      // piece drag
      var $window = $(window)
      $window
        .on('mousemove', throttledMousemoveWindow)
        .on('mouseup', mouseupWindow)

      // touch drag pieces
      if (isTouchDevice()) {
        $board.on('touchstart', '.' + CSS.square, touchstartSquare)
        $container.on('touchstart', '.' + CSS.sparePieces + ' .' + CSS.piece, touchstartSparePiece)
        $window
          .on('touchmove', throttledTouchmoveWindow)
          .on('touchend', touchendWindow)
      }
    }

    function initDOM () {
      // create unique IDs for all the elements we will create
      createElIds()

      // build board and save it in memory
      $container.html(buildContainerHTML(config.sparePieces))
      $board = $container.find('.' + CSS.board)

      if (config.sparePieces) {
        $sparePiecesTop = $container.find('.' + CSS.sparePiecesTop)
        $sparePiecesBottom = $container.find('.' + CSS.sparePiecesBottom)
      }

      // create the drag piece
      var draggedPieceId = uuid()
      $('body').append(buildPieceHTML('wP', true, draggedPieceId))
      $draggedPiece = $('#' + draggedPieceId)

      // TODO: need to remove this dragged piece element if the board is no
      // longer in the DOM

      // get the border size
      boardBorderSize = parseInt($board.css('borderLeftWidth'), 10)

      // set the size and draw the board
      widget.resize()
    }

    // -------------------------------------------------------------------------
    // Initialization
    // -------------------------------------------------------------------------

    setInitialState()
    initDOM()
    addEvents()

    // return the widget object
    return widget
  } // end constructor

  // TODO: do module exports here
  window['Chessboard'] = constructor

  // support legacy ChessBoard name
  window['ChessBoard'] = window['Chessboard']

  // expose util functions
  window['Chessboard']['fenToObj'] = fenToObj
  window['Chessboard']['objToFen'] = objToFen
})() // end anonymous wrapper
