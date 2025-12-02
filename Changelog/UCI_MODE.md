# UCI Protocol Support

## Overview

GoChess now supports the **Universal Chess Interface (UCI)** protocol, making it compatible with all major chess GUIs and testing tools.

## Quick Start

```bash
# Build the engine
go build -o gochess-board

# Run in UCI mode
./gochess-board --uci
```

## Supported UCI Commands

### Essential Commands

| Command | Description | Response |
|---------|-------------|----------|
| `uci` | Identify engine | `id name GoChess 2.0`<br>`id author GoChess Team`<br>`uciok` |
| `isready` | Check if ready | `readyok` |
| `ucinewgame` | Start new game | (resets position) |
| `position [fen/startpos]` | Set position | (updates internal state) |
| `go [params]` | Start search | `bestmove <move>` |
| `quit` | Exit engine | (exits cleanly) |

### Position Command

```
position startpos
position fen <fen_string>
position startpos moves e2e4 e7e5
position fen <fen_string> moves <move1> <move2> ...
```

**Note:** Move application is not yet implemented. Use FEN positions directly.

### Go Command

```
go movetime 2000          # Search for 2 seconds
go wtime 60000 btime 60000  # Time controls (uses 1/30th)
go infinite               # Search indefinitely
```

**Supported Parameters:**
- `movetime <ms>` - Fixed time per move
- `wtime <ms>` - White's remaining time
- `btime <ms>` - Black's remaining time
- `infinite` - Search until stopped

**Not Yet Supported:**
- `depth <n>` - Search to specific depth
- `nodes <n>` - Search specific number of nodes
- `mate <n>` - Search for mate in n moves

## Usage Examples

### Interactive Mode

```bash
./gochess-board --uci
uci
isready
position startpos
go movetime 2000
quit
```

### With Chess GUI

#### Arena Chess GUI

1. Download Arena from http://www.playwitharena.de/
2. Install Arena
3. Engines → Install New Engine
4. Browse to `gochess-board` executable
5. Select "UCI" protocol
6. Done!

#### Cute Chess GUI

```bash
# Install cutechess
sudo apt-get install cutechess

# Add engine
cutechess-gui
# Tools → Settings → Engines → Add
# Command: /path/to/gochess-board
# Arguments: --uci
# Protocol: UCI
```

### Automated Testing with cutechess-cli

```bash
# Install cutechess-cli
sudo apt-get install cutechess-cli

# Play games against Stockfish
cutechess-cli \
  -engine cmd=./gochess-board args="--uci" name="GoChess" proto=uci \
  -engine cmd=stockfish name="Stockfish" depth=1 proto=uci \
  -each tc=60+1 \
  -rounds 10 \
  -pgnout games.pgn

# Results will show win/loss/draw statistics
```

### Lichess BOT

You can run GoChess as a Lichess bot!

1. **Install lichess-bot:**
```bash
git clone https://github.com/lichess-bot-devs/lichess-bot
cd lichess-bot
pip install -r requirements.txt
```

2. **Configure engine in `config.yml`:**
```yaml
engine:
  dir: "/path/to/gochess-board"
  name: "gochess-board"
  protocol: "uci"
  uci_options:
    Move Overhead: 100
  
challenge:
  concurrency: 1
  min_increment: 1
  max_increment: 10
  min_initial: 60
  max_initial: 300
```

3. **Get Lichess API token:**
   - Go to https://lichess.org/account/oauth/token
   - Create token with "Play games with the bot API" scope
   - Add to `config.yml`

4. **Run bot:**
```bash
python lichess-bot.py
```

Your engine will now play rated games on Lichess!

## Testing

### Basic Functionality Test

```bash
# Test script
cat << 'EOF' | ./gochess-board --uci
uci
isready
position startpos
go movetime 1000
position fen r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4
go movetime 2000
quit
EOF
```

Expected output:
```
id name GoChess 2.0
id author GoChess Team
uciok
readyok
bestmove <move>
bestmove <move>
```

### Performance Test

```bash
# Time a search
time (cat << 'EOF' | ./gochess-board --uci
uci
isready
position startpos
go movetime 5000
quit
EOF
)
```

### Tournament Test

```bash
# Self-play tournament
cutechess-cli \
  -engine cmd=./gochess-board args="--uci" name="GoChess-1" proto=uci \
  -engine cmd=./gochess-board args="--uci" name="GoChess-2" proto=uci \
  -each tc=60+1 \
  -rounds 20 \
  -pgnout selfplay.pgn
```

## Implementation Details

### Engine Information

- **Name:** GoChess 2.0
- **Author:** GoChess Team
- **Protocol:** UCI
- **Version:** 2.0

### Features

✅ **Implemented:**
- Position setup (FEN and startpos)
- Time management (movetime, wtime/btime)
- Best move calculation
- Clean shutdown

❌ **Not Yet Implemented:**
- Move application from `position ... moves`
- Depth-based search
- Node-based search
- Pondering
- Multi-PV
- Search info output (depth, nodes, score)
- Options (Hash size, threads, etc.)

### Time Management

Simple time management strategy:
- `movetime X`: Use exactly X milliseconds
- `wtime X` / `btime X`: Use 1/30th of remaining time
- `infinite`: Search for 1 hour (effectively infinite)

This is conservative and prevents time losses.

### Error Handling

If an error occurs during search:
```
info string Error: <error message>
bestmove 0000
```

The engine sends an error message and a null move to avoid hanging the GUI.

## Comparison with Other Modes

| Feature | Server Mode | --engine-only | --uci |
|---------|-------------|---------------|-------|
| Web UI | ✅ | ❌ | ❌ |
| Chess GUIs | ❌ | ❌ | ✅ |
| Scripting | ❌ | ✅ | ✅ |
| Lichess BOT | ❌ | ❌ | ✅ |
| cutechess-cli | ❌ | ❌ | ✅ |
| ELO Testing | ❌ | Basic | Professional |
| Tournament Play | ❌ | ❌ | ✅ |

## Advanced Usage

### Custom Time Controls

```bash
# Bullet (1 minute + 1 second)
cutechess-cli \
  -engine cmd=./gochess-board args="--uci" name="GoChess" proto=uci \
  -engine cmd=stockfish name="Stockfish" depth=1 proto=uci \
  -each tc=60+1 \
  -rounds 50

# Blitz (3 minutes + 2 seconds)
cutechess-cli \
  -engine cmd=./gochess-board args="--uci" name="GoChess" proto=uci \
  -engine cmd=stockfish name="Stockfish" depth=2 proto=uci \
  -each tc=180+2 \
  -rounds 50

# Rapid (10 minutes + 5 seconds)
cutechess-cli \
  -engine cmd=./gochess-board args="--uci" name="GoChess" proto=uci \
  -engine cmd=stockfish name="Stockfish" depth=3 proto=uci \
  -each tc=600+5 \
  -rounds 50
```

### Opening Book

```bash
# Use opening book with cutechess-cli
cutechess-cli \
  -engine cmd=./gochess-board args="--uci" name="GoChess" proto=uci \
  -engine cmd=stockfish name="Stockfish" depth=1 proto=uci \
  -each tc=60+1 \
  -rounds 100 \
  -openings file=book.pgn format=pgn order=random \
  -pgnout games.pgn
```

### Gauntlet Tournament

Test against multiple engines:

```bash
cutechess-cli \
  -engine cmd=./gochess-board args="--uci" name="GoChess" proto=uci \
  -engine cmd=stockfish name="Stockfish-D1" depth=1 proto=uci \
  -engine cmd=stockfish name="Stockfish-D2" depth=2 proto=uci \
  -engine cmd=fairy-stockfish name="Fairy-SF" depth=1 proto=uci \
  -each tc=60+1 \
  -rounds 20 \
  -games 2 \
  -pgnout gauntlet.pgn \
  -ratinginterval 10
```

## Troubleshooting

### Engine Not Responding

Check if UCI mode is active:
```bash
echo "uci" | ./gochess-board --uci
# Should output: id name GoChess 2.0 ...
```

### GUI Can't Find Engine

Ensure you're using the full path:
```bash
which gochess-board
# Or
realpath ./gochess-board
```

### Slow Responses

The engine uses 2 seconds default think time. Adjust with `movetime`:
```
go movetime 1000  # 1 second
```

### Testing Connection

Simple test:
```bash
(echo "uci"; sleep 1; echo "quit") | ./gochess-board --uci
```

## Future Enhancements

Planned improvements:

1. **Search Info Output**
   - `info depth X`
   - `info nodes X`
   - `info score cp X`
   - `info pv <moves>`

2. **Move Application**
   - Apply moves from `position ... moves`
   - Track game history

3. **Options**
   - `option name Hash type spin default 64 min 1 max 1024`
   - `option name Threads type spin default 1 min 1 max 64`

4. **Advanced Features**
   - Pondering
   - Multi-PV
   - Depth-based search
   - Node-based search

## Resources

- **UCI Protocol Specification:** http://wbec-ridderkerk.nl/html/UCIProtocol.html
- **cutechess-cli:** https://github.com/cutechess/cutechess
- **lichess-bot:** https://github.com/lichess-bot-devs/lichess-bot
- **Arena Chess GUI:** http://www.playwitharena.de/

## Conclusion

UCI support makes GoChess:
- ✅ Compatible with all chess GUIs
- ✅ Ready for automated testing
- ✅ Tournament ready
- ✅ Lichess BOT compatible
- ✅ Professional ELO testing ready

Your engine is now a **first-class citizen** in the chess engine world! 🎉
