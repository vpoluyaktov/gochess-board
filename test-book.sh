#!/bin/bash
# Test if polyglot is using the opening book

echo "Testing Stockfish with opening book via Polyglot..."
echo ""

# Create a simple polyglot config
cat > /tmp/test-polyglot.ini << EOF
[PolyGlot]
EngineDir = .
EngineCommand = stockfish
Book = true
BookFile = /usr/share/games/gnuchess/book.bin
LogFile = /tmp/test-polyglot.log
BookDepth = 255
EOF

echo "Starting position, asking for move..."
echo ""

# Send commands to polyglot
(
  echo "uci"
  sleep 0.5
  echo "ucinewgame"
  echo "position startpos"
  echo "go movetime 100"
  sleep 1
  echo "quit"
) | polyglot /tmp/test-polyglot.ini 2>&1 | grep -E "(bestmove|book)"

echo ""
echo "Check the log file for book hits:"
if [ -f /tmp/test-polyglot.log ]; then
  grep -i "book" /tmp/test-polyglot.log | head -5
else
  echo "No log file created yet"
fi
