#!/bin/bash

# Integration Test: Built-in Engine Analysis (Background Process)
# 
# This test requires:
# 1. The gochess-board server to be running (from project root: ./gochess-board)
# 2. websocat installed (https://github.com/vi/websocat)
#
# Run from project root: ./tests/integration/test_builtin_analysis.sh

PORT=35256
WS_URL="ws://localhost:$PORT/ws/analysis"

echo "Testing built-in engine analysis..."
echo "Connecting to: $WS_URL"
echo ""

# Test message to start analysis with built-in engine
# Using the starting position FEN
cat << 'EOF' | websocat "$WS_URL" -n1 --ping-interval 30 --ping-timeout 60 &
{"action":"start","fen":"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1","enginePath":"internal"}
EOF

WS_PID=$!

# Wait for analysis output (5 seconds)
echo "Waiting for analysis output (5 seconds)..."
sleep 5

# Stop the analysis
echo '{"action":"stop"}' | websocat "$WS_URL" -n1 --ping-interval 30 --ping-timeout 60

# Kill the background process if still running
kill $WS_PID 2>/dev/null

echo ""
echo "Test complete!"
