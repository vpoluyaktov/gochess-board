# Polyglot Opening Book Implementation Status

## Overview

Implementation of native Polyglot opening book support in pure Go (no third-party libraries) to enable chess opening book functionality for both UCI and CECP engines.

**Status**: ✅ **COMPLETE** - All bugs fixed, tests passing, ready for integration

---

## Motivation

Previously, the application used Polyglot as an external wrapper to add opening book support to engines. However:
- Polyglot works as a UCI-to-CECP adapter (not CECP-to-UCI)
- After implementing native CECP support, we no longer need Polyglot for engine communication
- We still want opening book support for all engines (UCI and CECP)

**Solution**: Implement native Polyglot book file (.bin) reading in Go.

---

## Implementation Details

### Files Created

1. **`server/polyglot_book.go`** (~320 lines)
   - `PolyglotBook` struct with book entry management
   - `LoadFromFile()` - Reads .bin files (16 bytes per entry: key, move, weight, learn)
   - `Probe()` - Returns all book moves for a position
   - `ProbeWeighted()` - Returns a random weighted book move
   - `zobristHash()` - Calculates Polyglot Zobrist hash for positions
   - `polyglotMoveToUCI()` - Converts Polyglot move format to UCI notation

2. **`server/polyglot_zobrist.go`** (~786 lines)
   - Contains the official 781 Zobrist random numbers from Polyglot specification
   - Extracted from: https://github.com/ddugovic/polyglot/blob/master/book_format.html

3. **`server/polyglot_book_test.go`**
   - Unit tests for book loading and move probing

### Key Features Implemented

✅ **Binary file parsing** - Reads Polyglot .bin format (big-endian)
✅ **Zobrist hashing** - Full implementation with official hash values
✅ **Move format conversion** - Polyglot format → UCI notation
✅ **Weighted move selection** - Respects book move frequencies
✅ **Binary search** - Efficient position lookup in sorted book
✅ **Piece type mapping** - Handles differences between Go chess library and Polyglot piece numbering

### Architecture

```
┌─────────────┐         ┌──────────────────────────────────────┐         ┌────────────┐
│   Browser   │ ◄─────► │         Go Server                    │ ◄─────► │   Engine   │
│     UI      │         │                                      │         │            │
└─────────────┘         │  1. Receive position                 │         └────────────┘
                        │  2. Check book for opening moves     │
                        │  3. If book hit: return book move    │
                        │  4. If no book: ask engine for move  │
                        └──────────────────────────────────────┘
```

---

## Bugs Fixed

### Bug #1: Zobrist Hash Turn Logic (FIXED ✅)

**Problem**: The hash calculation XORed the turn value when it was **black's turn**, but the Polyglot specification requires XORing when it's **white's turn**.

**Root Cause**: The Polyglot Zobrist hash uses 0 as the base for black-to-move positions, and XORs a specific value when white is to move. This is counterintuitive but verified against python-chess library.

**Symptoms**:
- Expected hash for starting position: `0x463B96181691FC9C`
- Calculated hash: `0xBEEDB0B2B9B67995`
- The calculated hash was actually correct for the position with **black to move**!

**Fix**: Changed line 264 in `polyglot_book.go`:
```go
// Before (WRONG):
if position.Turn() == chess.Black {
    hash ^= polyglotRandom64[polyglotBlackToMoveOffset]
}

// After (CORRECT):
if position.Turn() == chess.White {
    hash ^= polyglotRandom64[polyglotBlackToMoveOffset]
}
```

### Bug #2: Move Format Bit Order (FIXED ✅)

**Problem**: The move format conversion had the FROM and TO squares swapped.

**Root Cause**: The Polyglot move format stores the **TO square** in bits 0-5 and the **FROM square** in bits 6-11, which is opposite of what some documentation suggests.

**Symptoms**:
- Expected moves: `e2e4`, `d2d4`, `g1f3`
- Calculated moves: `e4e2`, `d4d2`, `f3g1` (reversed)

**Fix**: Swapped the bit extraction in `polyglotMoveToUCI()`:
```go
// Before (WRONG):
from := int(move & 0x3F)
to := int((move >> 6) & 0x3F)

// After (CORRECT):
to := int(move & 0x3F)
from := int((move >> 6) & 0x3F)
```

### Verification

Both bugs were verified by comparing against the python-chess library implementation:
- Hash values now match exactly: `0x463B96181691FC9C` ✓
- Move lists match: `[e2e4, d2d4, g1f3, c2c4, ...]` ✓
- All unit tests pass ✓

---

## Testing

### Test Book File Location

```bash
/usr/share/games/gnuchess/book.bin
```

This is the GNU Chess opening book (180,358 entries).

### Running Tests

```bash
# Unit tests
go test -v -run TestPolyglotBook ./server

# Integration test with server
go run . --no-browser --log-level DEBUG --book-file /usr/share/games/gnuchess/book.bin
```

### Test Results

✅ **All tests passing**:
```
=== RUN   TestPolyglotBook
Loaded 180358 entries from /usr/share/games/gnuchess/book.bin
Calculated hash for starting position: 0x463B96181691FC9C
Expected hash for starting position:  0x463B96181691FC9C
Starting position has 13 book moves: [e2e4 d2d4 g1f3 c2c4 g2g3 b2b3 f2f4 b1c3 b2b4 e2e3 d2d3 g2g4 a2a3]
Weighted book move: d2d4
--- PASS: TestPolyglotBook (0.32s)
```

### Expected Behavior

1. Server loads book file at startup
2. For UCI engines: creates "+ Book" variants that consult book before engine
3. For CECP engines: native book lookup before engine call
4. Book moves returned in UCI format for UI compatibility

---

## Polyglot Book Format Reference

### Entry Structure (16 bytes)
```
key    uint64  - Zobrist hash of position
move   uint16  - Move in Polyglot format
weight uint16  - Frequency/priority of move
learn  uint32  - Learning data (unused)
```

### Move Format (16 bits)
```
bits 0-5:   TO square (0=a1, 63=h8)
bits 6-11:  FROM square
bits 12-14: promotion piece (0=none, 1=knight, 2=bishop, 3=rook, 4=queen)
```
**Note**: Some documentation incorrectly lists FROM in low bits. Verified: TO is in bits 0-5.

### Zobrist Hash Components

The hash is XOR of:
1. **Pieces** (768 values): `Random64[64*pieceIdx + square]`
   - Piece ordering: black pawn=0, white pawn=1, black knight=2, white knight=3, ...
2. **Castle rights** (4 values): White K-side, White Q-side, Black K-side, Black Q-side
3. **En passant** (8 values): One per file (a-h)
4. **Side to move** (1 value): XOR if **white** to move (counterintuitive but verified)

---

## Next Steps for Integration

### Ready for Production Use

The Polyglot book implementation is complete and tested. To integrate into the application:

1. **Add book loading to server startup** - Load book file if `--book-file` flag is provided
2. **Integrate with move selection** - Check book before querying engine
3. **Update engine discovery** - Remove Polyglot wrapper, use native book for all engines
4. **Add UI indicators** - Show when a move comes from the book vs. engine

### Integration Points

```go
// In chess.go - computer move handler
if bookFile != nil {
    bookMoves := bookFile.Probe(game.Position())
    if len(bookMoves) > 0 {
        // Pick a random weighted move from book
        bookMove := bookFile.ProbeWeighted(game.Position())
        return bookMove
    }
}
// No book move, ask engine
engine.GetBestMove(...)
```

---

## Code Statistics

- **Total lines**: ~1,400 lines
- **New files**: 3
- **Dependencies**: Only `github.com/notnil/chess` (already in use)
- **External tools**: None (pure Go implementation)

---

## References

- [Polyglot Book Format Specification](http://hgm.nubati.net/book_format.html)
- [Polyglot Source Code](https://github.com/ddugovic/polyglot)
- [Python Chess Polyglot Implementation](https://python-chess.readthedocs.io/en/latest/polyglot.html)

---

## Conclusion

✅ **Implementation Complete!**

The Polyglot opening book feature is fully implemented and tested. All bugs have been fixed:
- ✅ Zobrist hash calculation corrected (turn logic)
- ✅ Move format conversion fixed (bit order)
- ✅ All unit tests passing
- ✅ Verified against python-chess library

The pure Go implementation eliminates external dependencies and provides full control over book handling for both UCI and CECP engines. Ready for integration into the main application.
