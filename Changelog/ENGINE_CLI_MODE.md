# Engine Command-Line Mode

## Overview

The `gochess-board` application now supports a special `--engine-only` mode for testing the built-in chess engine from the command line.

## Usage

```bash
# Build the application first
go build -o gochess-board

# Basic usage
echo "FEN_STRING" | ./gochess-board --engine-only

# With custom think time
echo "FEN_STRING" | ./gochess-board --engine-only --think-time 5s

# Example
echo "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1" | ./gochess-board --engine-only
# Output: b1c3
```

## Command-Line Flags

### `--engine-only`
Run the built-in engine in command-line mode for testing. The engine:
- Reads a FEN position from stdin
- Calculates the best move
- Outputs the move in algebraic notation
- Exits

### `--think-time duration`
Time to think per move in engine-only mode (default: 2s)

Examples:
- `--think-time 1s` - 1 second
- `--think-time 500ms` - 500 milliseconds
- `--think-time 5s` - 5 seconds

## Use Cases

### 1. **Tactical Testing**

Test the engine against tactical positions:

```bash
# Mate in 1
echo "6k1/5ppp/8/8/8/8/5PPP/R5K1 w - - 0 1" | ./gochess-board --engine-only
# Expected: a1a8 (Ra8#)
```

### 2. **Automated Testing**

Use in shell scripts for ELO testing:

```bash
#!/bin/bash
positions=(
    "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1|e2e4"
    "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4|f3g5"
)

for pos_data in "${positions[@]}"; do
    IFS='|' read -r fen expected <<< "$pos_data"
    result=$(echo "$fen" | ./gochess-board --engine-only --think-time 2s)
    
    if [[ $result == $expected ]]; then
        echo "✓ Found $expected"
    else
        echo "✗ Expected $expected, got $result"
    fi
done
```

### 3. **Integration with Chess Tools**

Use with cutechess-cli or other chess testing frameworks:

```bash
# Example wrapper script for cutechess-cli
#!/bin/bash
while IFS= read -r line; do
    if [[ $line == position* ]]; then
        fen=$(echo "$line" | cut -d' ' -f2-)
        move=$(echo "$fen" | ./gochess-board --engine-only --think-time 1s)
        echo "bestmove $move"
    fi
done
```

### 4. **Quick Testing**

Test engine improvements quickly:

```bash
# Test starting position
echo "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1" | \
  ./gochess-board --engine-only --think-time 3s

# Test tactical position
echo "2rr3k/pp3pp1/1nnqbN1p/3pN3/2pP4/2P3Q1/PPB4P/R4RK1 w - - 0 1" | \
  ./gochess-board --engine-only --think-time 5s
```

## Implementation Details

### Code Location
- **File:** `main.go`
- **Function:** `runEngineOnly(thinkTime time.Duration)`

### How It Works

1. **Flag Parsing:** Checks for `--engine-only` flag
2. **Engine Creation:** Creates `builtin.NewEngine()`
3. **Input Reading:** Reads FEN from stdin using `bufio.Scanner`
4. **Move Calculation:** Calls `engine.GetBestMove(fen, thinkTime)`
5. **Output:** Prints move to stdout
6. **Exit:** Exits immediately

### Error Handling

Errors are written to stderr:
```bash
# Invalid FEN
echo "invalid" | ./gochess-board --engine-only
# Output to stderr: Error getting best move: invalid FEN: ...

# No input
./gochess-board --engine-only < /dev/null
# Output to stderr: Error: No input provided
```

## Testing Script

The `engines/builtin/test_elo.sh` script uses this mode:

```bash
cd engines/builtin
./test_elo.sh
```

This script:
- Builds the engine
- Tests tactical positions
- Estimates ELO based on results
- Provides recommendations

## Performance

Engine-only mode has minimal overhead:
- **Startup time:** ~10-50ms
- **Think time:** As specified (default 2s)
- **Output time:** <1ms
- **Total:** ~2-5 seconds per position

## Comparison with Server Mode

| Feature | Server Mode | Engine-Only Mode |
|---------|-------------|------------------|
| Web UI | ✅ | ❌ |
| TUI | ✅ | ❌ |
| Browser | ✅ | ❌ |
| Command-line | ❌ | ✅ |
| Scripting | ❌ | ✅ |
| Testing | ❌ | ✅ |
| Startup time | ~500ms | ~10ms |
| Memory | ~100MB | ~70MB |

## Examples

### Test Multiple Positions

```bash
# Create test file
cat > positions.txt << EOF
rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1
r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4
EOF

# Test each position
while IFS= read -r fen; do
    echo "Position: $fen"
    echo "Best move: $(echo "$fen" | ./gochess-board --engine-only)"
    echo ""
done < positions.txt
```

### Benchmark Engine Speed

```bash
#!/bin/bash
fen="r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"

for time in 1s 2s 3s 5s; do
    echo "Think time: $time"
    /usr/bin/time -f "Real: %E" \
        bash -c "echo '$fen' | ./gochess-board --engine-only --think-time $time" 2>&1
    echo ""
done
```

### Compare with Previous Version

```bash
# Build current version
go build -o gochess-new

# Build old version (from git)
git stash
git checkout HEAD~5
go build -o gochess-old
git stash pop

# Compare
fen="rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

echo "Old engine:"
echo "$fen" | ./gochess-old --engine-only --think-time 2s

echo "New engine:"
echo "$fen" | ./gochess-new --engine-only --think-time 2s
```

## Troubleshooting

### Engine Takes Too Long

Reduce think time:
```bash
echo "FEN" | ./gochess-board --engine-only --think-time 500ms
```

### No Output

Check stderr for errors:
```bash
echo "FEN" | ./gochess-board --engine-only 2>&1
```

### Timeout in Scripts

Use `timeout` command:
```bash
echo "FEN" | timeout 5s ./gochess-board --engine-only --think-time 3s
```

## Future Enhancements

Potential improvements:

1. **UCI Protocol Support**
   - Full UCI command support
   - Multi-move analysis
   - Position setup commands

2. **Batch Mode**
   - Process multiple FENs from file
   - Output to file
   - Progress reporting

3. **Analysis Output**
   - Show evaluation score
   - Display principal variation
   - Node count and NPS

4. **JSON Output**
   - Machine-readable format
   - Include metadata
   - Multiple move suggestions

## Conclusion

The `--engine-only` mode makes the GoChess engine:
- ✅ Easy to test
- ✅ Scriptable
- ✅ Fast to run
- ✅ Simple to integrate

Perfect for automated testing, ELO estimation, and development!
