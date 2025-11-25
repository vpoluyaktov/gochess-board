#!/bin/bash
# Verify if opening book is being used by comparing moves

echo "=== Testing Opening Book Usage ==="
echo ""

BOOK="/usr/share/games/gnuchess/book.bin"

echo "1. Checking if book file exists and is readable:"
if [ -f "$BOOK" ]; then
  echo "   ✓ Book file found: $BOOK ($(du -h "$BOOK" | cut -f1))"
else
  echo "   ✗ Book file not found!"
  exit 1
fi
echo ""

echo "2. Testing Stockfish WITHOUT book (direct):"
echo "   Asking for move from starting position with 100ms think time..."
MOVE_NO_BOOK=$(echo -e "uci\nucinewgame\nposition startpos\ngo movetime 100\nquit" | stockfish 2>/dev/null | grep "^bestmove" | head -1 | awk '{print $2}')
echo "   Move: $MOVE_NO_BOOK"
echo ""

echo "3. Testing Stockfish WITH book (via polyglot):"
cat > /tmp/test-book.ini << EOF
[PolyGlot]
EngineDir = .
EngineCommand = stockfish  
Book = true
BookFile = $BOOK
BookDepth = 255
EOF

MOVE_WITH_BOOK=$(echo -e "uci\nucinewgame\nposition startpos\ngo movetime 100\nquit" | polyglot /tmp/test-book.ini 2>/dev/null | grep "^bestmove" | head -1 | awk '{print $2}')
echo "   Move: $MOVE_WITH_BOOK"
echo ""

echo "4. Analysis:"
if [ "$MOVE_NO_BOOK" = "$MOVE_WITH_BOOK" ]; then
  echo "   ⚠ Moves are the same - book might not have this position"
  echo "   (This is normal if the book contains the same move the engine would choose)"
else
  echo "   ✓ Moves are different - book is likely being used!"
  echo "   Without book: $MOVE_NO_BOOK"
  echo "   With book:    $MOVE_WITH_BOOK"
fi
echo ""

echo "5. Common opening book moves from starting position:"
echo "   e2e4 (King's Pawn), d2d4 (Queen's Pawn), c2c4 (English), g1f3 (Reti)"
echo ""

# Test multiple times to see variety
echo "6. Testing 3 times to see if book provides variety:"
for i in 1 2 3; do
  MOVE=$(echo -e "uci\nucinewgame\nposition startpos\ngo movetime 50\nquit" | polyglot /tmp/test-book.ini 2>/dev/null | grep "^bestmove" | head -1 | awk '{print $2}')
  echo "   Attempt $i: $MOVE"
done
