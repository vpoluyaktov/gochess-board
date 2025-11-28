#!/bin/bash
# UCI Protocol Test Script
#
# Tests the UCI protocol implementation of the built-in engine
# Run from project root: ./tests/engine/test_uci.sh

echo "=================================="
echo "GoChess UCI Protocol Test"
echo "=================================="
echo ""

# Get project root directory
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Build if needed
if [ ! -f "$PROJECT_ROOT/gochess-board" ]; then
    echo "Building gochess-board..."
    cd "$PROJECT_ROOT"
    go build -o gochess-board
    echo ""
fi

# Change to project root for relative paths
cd "$PROJECT_ROOT"

# Test 1: Basic UCI handshake
echo "Test 1: UCI Handshake"
echo "----------------------"
result=$(cat << 'EOF' | timeout 2s ./gochess-board --engine-only
uci
quit
EOF
)

if [[ $result == *"uciok"* ]]; then
    echo "✓ UCI handshake successful"
else
    echo "✗ UCI handshake failed"
    echo "Output: $result"
fi
echo ""

# Test 2: Ready check
echo "Test 2: Ready Check"
echo "-------------------"
result=$(cat << 'EOF' | timeout 2s ./gochess-board --engine-only
isready
quit
EOF
)

if [[ $result == *"readyok"* ]]; then
    echo "✓ Ready check successful"
else
    echo "✗ Ready check failed"
fi
echo ""

# Test 3: Position and search
echo "Test 3: Position and Search"
echo "---------------------------"
result=$(cat << 'EOF' | timeout 5s ./gochess-board --engine-only
uci
position startpos
go movetime 1000
quit
EOF
)

if [[ $result == *"bestmove"* ]]; then
    echo "✓ Search successful"
    # Extract move
    move=$(echo "$result" | grep "bestmove" | awk '{print $2}')
    echo "  Best move: $move"
else
    echo "✗ Search failed"
fi
echo ""

# Test 4: FEN position
echo "Test 4: FEN Position"
echo "--------------------"
result=$(cat << 'EOF' | timeout 5s ./gochess-board --engine-only
position fen r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4
go movetime 1000
quit
EOF
)

if [[ $result == *"bestmove"* ]]; then
    echo "✓ FEN position search successful"
    move=$(echo "$result" | grep "bestmove" | awk '{print $2}')
    echo "  Best move: $move"
else
    echo "✗ FEN position search failed"
fi
echo ""

# Test 5: Multiple positions
echo "Test 5: Multiple Positions"
echo "--------------------------"
result=$(cat << 'EOF' | timeout 10s ./gochess-board --engine-only
position startpos
go movetime 500
position fen rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1
go movetime 500
quit
EOF
)

move_count=$(echo "$result" | grep -c "bestmove")
if [ $move_count -eq 2 ]; then
    echo "✓ Multiple positions successful ($move_count moves)"
else
    echo "✗ Multiple positions failed (expected 2, got $move_count)"
fi
echo ""

echo "=================================="
echo "UCI Protocol Test Complete"
echo "=================================="
echo ""
echo "Next steps:"
echo "1. Test with cutechess-cli:"
echo "   cutechess-cli -engine cmd=./gochess-board args=\"--engine-only\" name=GoChess proto=uci \\"
echo "                 -engine cmd=stockfish name=Stockfish depth=1 proto=uci \\"
echo "                 -each tc=60+1 -rounds 10"
echo ""
echo "2. Test with a chess GUI (Arena, Cute Chess, etc.)"
echo ""
echo "3. Set up Lichess BOT for online play"
echo ""
