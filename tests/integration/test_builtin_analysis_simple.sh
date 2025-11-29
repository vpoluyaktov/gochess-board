#!/bin/bash

# Integration Test: Built-in Engine Analysis (Simple)
# 
# This test requires:
# 1. The gochess-board server to be running (from project root: ./gochess-board)
# 2. websocat installed (https://github.com/vi/websocat)
#
# Run from project root: ./tests/integration/test_builtin_analysis_simple.sh

PORT=35256
WS_URL="ws://localhost:$PORT/ws/analysis"

echo "Testing built-in engine analysis..."
echo "Connecting to: $WS_URL"
echo ""

# Start analysis and capture output for 3 seconds
(
  echo '{"action":"start","fen":"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1","enginePath":"internal"}'
  sleep 3
  echo '{"action":"stop"}'
) | websocat "$WS_URL" --text

echo ""
echo "Test complete!"
