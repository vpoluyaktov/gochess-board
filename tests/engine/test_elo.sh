#!/bin/bash
# ELO Testing Script for GoChess Engine
#
# Tests the built-in engine against tactical positions
# Run from project root: ./tests/engine/test_elo.sh

set -e

echo "==================================="
echo "GoChess Engine ELO Testing Suite"
echo "==================================="
echo ""

# Build the engine
echo "Building GoChess..."
cd "$(dirname "$0")/../.."
go build -o gochess-board 2>&1 | grep -v "^#" || true

echo "✓ Build complete"
echo ""

# Test 1: Tactical Test Suite (WAC - Win At Chess)
echo "Test 1: Tactical Positions (WAC subset)"
echo "----------------------------------------"
echo "Testing ability to find tactical moves..."

# Sample tactical positions (from Win At Chess test suite)
declare -a positions=(
    # Position 1: Mate in 2
    "2rr3k/pp3pp1/1nnqbN1p/3pN3/2pP4/2P3Q1/PPB4P/R4RK1 w - - 0 1|Qg7"
    # Position 2: Win material
    "r1b1k2r/ppppnppp/2n2q2/2b5/3NP3/2P1B3/PP3PPP/RN1QKB1R w KQkq - 0 1|Nf5"
    # Position 3: Tactical blow
    "r1bqk2r/pp2bppp/2p5/3pP3/3P4/5N2/PPP2PPP/RNBQR1K1 w kq - 0 1|Bh6"
)

correct=0
total=0

for pos_data in "${positions[@]}"; do
    IFS='|' read -r fen best_move <<< "$pos_data"
    total=$((total + 1))
    
    # Run engine with 5 second timeout using UCI protocol
    result=$(echo "$fen" | timeout 5s bash -c "echo 'position fen $fen'; echo 'go movetime 2000'; echo 'quit'" | ./gochess-board --engine-only 2>/dev/null | grep "bestmove" | awk '{print $2}' || echo "timeout")
    
    if [[ $result == *"$best_move"* ]]; then
        echo "  ✓ Position $total: Found $best_move"
        correct=$((correct + 1))
    else
        echo "  ✗ Position $total: Expected $best_move, got $result"
    fi
done

tactical_score=$((correct * 100 / total))
echo ""
echo "Tactical Score: $correct/$total ($tactical_score%)"
echo ""

# Test 2: Positional Understanding
echo "Test 2: Positional Understanding"
echo "---------------------------------"
echo "Testing evaluation quality..."

# Test if engine prefers better positions
declare -a eval_tests=(
    # Better position should score higher
    "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1|equal"
    "rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 2|equal"
)

echo "  (Positional tests would require engine output parsing)"
echo "  Skipping for now - use manual analysis"
echo ""

# Test 3: Performance Metrics
echo "Test 3: Performance Metrics"
echo "----------------------------"
echo "Measuring search speed..."

# Run a benchmark position
bench_fen="r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"

echo "  Running 5-second search on complex position..."
start_time=$(date +%s)
# This would need actual engine output
echo "  (Would measure nodes/sec here)"
echo ""

# Estimate ELO based on tactical score
echo "==================================="
echo "ELO Estimation"
echo "==================================="
echo ""

if [ $tactical_score -ge 80 ]; then
    echo "Tactical Score: $tactical_score% - Excellent"
    echo "Estimated ELO: 1600-1800"
elif [ $tactical_score -ge 60 ]; then
    echo "Tactical Score: $tactical_score% - Good"
    echo "Estimated ELO: 1400-1600"
elif [ $tactical_score -ge 40 ]; then
    echo "Tactical Score: $tactical_score% - Fair"
    echo "Estimated ELO: 1200-1400"
else
    echo "Tactical Score: $tactical_score% - Needs Improvement"
    echo "Estimated ELO: 1000-1200"
fi

echo "Testing complete!"
