# Go Chess Board Application

A Go web server that lets you play chess against multiple chess engines including **Stockfish**, **Fruit**, **Toga**, **Crafty**, and **GNU Chess**. The application supports both UCI and CECP/XBoard protocol engines with optional opening book support via Polyglot. The chess game logic runs on the backend, while the frontend provides an interactive chess board. The application automatically opens your default browser when started.

## Features

- **Multiple Chess Engines**: Play against various engines (Stockfish, Fruit, Toga, Crafty, GNU Chess, and more)
- **UCI & CECP Support**: Native UCI protocol support + CECP engines via Polyglot wrapper
- **Opening Book Support**: Optional opening book integration for stronger play
- **Opening Name Recognition**: Identifies chess openings in real-time with ECO codes (3,594 openings from Lichess database)
- **Automatic Engine Discovery**: Detects installed engines on your system
- **Real-time TUI**: Beautiful terminal interface showing live game stats and engine analysis
- **Move Validation**: Only legal moves are allowed
- **Auto-Browser Opening**: Automatically opens the chess board in your default browser
- **Fully Self-Contained**: All assets embedded in the binary - works completely offline!
  - Chess piece images
  - ChessboardJS library (ready for your modifications!)
  - Chess.js library (frontend game state)
  - jQuery library
  - CSS styles
  - Chess opening database (3,594 openings)
- Interactive chess board with drag-and-drop functionality
- Beautiful, modern UI with gradient background
- Responsive design
- New Game and Flip Board controls

## Prerequisites

- Go 1.21 or higher
- **At least one chess engine** installed (see [Installing Chess Engines](#installing-chess-engines) below)
- **Polyglot** (optional, required for CECP engines and opening books)

## Supported Chess Engines

The application automatically discovers and supports engines using:

### UCI Protocol Engines (Native Support)
These engines work directly without any wrapper:
- **Stockfish** - World's strongest open-source engine
- **Fruit** - Strong tactical engine
- **Toga II** - Fruit derivative
- **Leela Chess Zero (lc0)** - Neural network engine
- **Komodo**, **Rybka**, **Houdini** (commercial)
- And many more UCI engines...

### CECP/XBoard Protocol Engines (via Polyglot)
These engines require Polyglot wrapper:
- **GNU Chess** - Classic free chess engine
- **Crafty** - Strong traditional engine
- **Sjeng** - Multi-variant engine
- **Phalanx** - Lightweight engine

### Opening Book Support
When using the `--book-file` flag:
- UCI engines get optional book variants (e.g., "Stockfish 16 + Book")
- CECP engines always use the book when available
- Books must be in Polyglot format (.bin files)

## Installing Chess Engines

### Linux (Ubuntu/Debian)

**UCI Engines:**
```bash
# Stockfish (strongest)
sudo apt-get install stockfish

# Fruit
sudo apt-get install fruit

# Toga II
sudo apt-get install toga2
```

**CECP Engines + Polyglot:**
```bash
# Polyglot wrapper (required for CECP engines and books)
sudo apt-get install polyglot

# GNU Chess
sudo apt-get install gnuchess

# Crafty
sudo apt-get install crafty

# Other CECP engines
sudo apt-get install sjeng phalanx
```

**Building GNU Chess from source (recommended for latest version):**
```bash
sudo apt-get install build-essential libreadline-dev
cd /tmp
wget https://ftp.gnu.org/gnu/chess/gnuchess-6.2.9.tar.gz
tar xzf gnuchess-6.2.9.tar.gz
cd gnuchess-6.2.9
./configure
make
sudo make install
```

### macOS

**Using Homebrew:**
```bash
# Install Homebrew if not already installed
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# UCI Engines
brew install stockfish
brew install fruit

# Polyglot wrapper
brew install polyglot

# CECP Engines
brew install gnu-chess
brew install crafty
```

### Windows

**Stockfish (UCI):**
1. Download from https://stockfishchess.org/download/
2. Extract to `C:\Program Files\Stockfish\`
3. Add to PATH: `C:\Program Files\Stockfish\`

**Polyglot:**
1. Download from http://hgm.nubati.net/cgi-bin/gitweb.cgi?p=polyglot.git
2. Or use pre-compiled binary from Arena Chess GUI
3. Extract to `C:\Program Files\Polyglot\`
4. Add to PATH

**GNU Chess (CECP):**
1. Download from https://ftp.gnu.org/gnu/chess/
2. Extract and add to PATH

**Alternative: Use Arena Chess GUI**
Arena includes many engines and Polyglot:
1. Download from http://www.playwitharena.de
2. Engines are in `Arena\Engines\` directory
3. Add the engines directory to your PATH

### Verifying Installation

Check which engines are detected:
```bash
./go-chess --no-browser
# Check the terminal output or chess-debug.log for discovered engines
```

Or check manually:
```bash
# Test UCI engine
echo "uci" | stockfish

# Test CECP engine (via polyglot)
echo "xboard" | gnuchess

# Check polyglot
polyglot --help
```

## Opening Books

### Where to Get Opening Books

**Included with GNU Chess:**
```bash
# Linux
/usr/share/games/gnuchess/book.bin (2.8MB)
/usr/local/share/gnuchess/smallbook.bin (38KB)
```

**Download Popular Books:**
- **Performance.bin** - Small, good for testing
- **Gm2001.bin** - Grandmaster games
- **ProDeo.bin** - Strong opening book
- Available from chess programming forums and Arena Chess GUI

**Create Your Own:**
```bash
# Install polyglot tools
sudo apt-get install polyglot

# Create book from PGN file
polyglot make-book -pgn games.pgn -bin mybook.bin -min-game 1
```

### Using Opening Books

```bash
# Run with opening book
./go-chess --book-file /usr/share/games/gnuchess/book.bin

# Without book (UCI engines only, no CECP)
./go-chess
```

## Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd go-chess
   ```

2. Download dependencies:
   ```bash
   go mod tidy
   ```

3. Build the application:
   ```bash
   go build -o go-chess
   ```

## Usage

### Basic Usage

Simply run the application:

```bash
./go-chess
```

This will:
1. Discover installed chess engines
2. Start a web server on `http://localhost:35256`
3. Automatically open the chess board in your default browser
4. Display a TUI showing discovered engines and game stats

### Command Line Options

All flags support both single dash (`-flag`) and double dash (`--flag`) formats:

- `--port`: Port for the web server (default: `35256`)
  ```bash
  ./go-chess --port 3000
  ```

- `--no-browser`: Don't automatically open the browser
  ```bash
  ./go-chess --no-browser
  ```
  Then manually open `http://localhost:35256` in your browser

- `--no-tui`: Disable the TUI interface (run in simple mode)
  ```bash
  ./go-chess --no-tui
  ```

- `--book-file`: Path to opening book file (Polyglot .bin format)
  ```bash
  ./go-chess --book-file /usr/share/games/gnuchess/book.bin
  ```
  When specified:
  - UCI engines get "+ Book" variants
  - CECP engines use the book
  - Without this flag, only UCI engines are available (no CECP)

### Examples

```bash
# Run with opening book and custom port
./go-chess --book-file /usr/share/games/gnuchess/book.bin --port 8080

# Run without browser, with book
./go-chess --no-browser --book-file /path/to/book.bin

# Simple mode (no TUI)
./go-chess --no-tui --book-file /usr/share/games/gnuchess/book.bin
```

### TUI Interface

By default, the application displays a beautiful terminal UI with a **horizontal layout optimized for 16:9 displays**:

**Three-column layout:**
- 📊 **Game Stats** (left): Total moves, white/black move counts, game duration
- 🤖 **Stockfish** (center): Last move, think time, time since last move, ELO rating
- 📍 **Position** (right): Current FEN notation

**Features:**
- 🔴 **Live Updates**: Real-time updates every second
- 🎨 **Color-coded**: Purple labels, green values, styled boxes
- ⚡ **Compact**: Fits nicely on standard terminal windows

Press `q` or `Ctrl+C` in the TUI to quit.

## Development

### Project Structure

```
go-chess/
├── main.go              # Main application entry point
├── go.mod               # Go module definition
├── server/
│   ├── server.go        # HTTP server implementation
│   ├── templates/
│   │   └── index.html   # Chess board HTML template
│   └── assets/          # All assets embedded in binary
│       ├── chess/
│       │   └── pieces/  # Chess piece images (12 PNG files)
│       ├── css/
│       │   └── chessboard-1.0.0.min.css
│       └── js/
│           ├── chessboard-1.0.0.min.js  # Modify this!
│           ├── chess.min.js             # Chess game logic
│           └── jquery-3.5.1.min.js
└── README.md            # This file
```

### Running in Development

```bash
# Run with auto-browser opening
go run main.go

# Run without opening browser
go run main.go -no-browser

# Run on custom port
go run main.go -port=3000
```

### Building for Different Platforms

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o go-chess-linux

# Windows
GOOS=windows GOARCH=amd64 go build -o go-chess.exe

# macOS
GOOS=darwin GOARCH=amd64 go build -o go-chess-macos
```

## How It Works

### Engine Discovery
On startup, the application:
1. Checks if Polyglot is installed
2. Discovers UCI engines (Stockfish, Fruit, Toga, etc.)
3. Discovers CECP engines if Polyglot is available (GNU Chess, Crafty, etc.)
4. Creates Polyglot-wrapped variants for opening book support
5. Generates unique configuration files in `/tmp/go-chess-polyglot/`

### Game Flow
1. **You play as White** - Drag and drop pieces to make your move
2. **Engine plays as Black** - The backend uses the selected engine to calculate the best move
3. **Move validation** - Illegal moves are rejected (pieces snap back)
4. **Game state** - Managed by chess.js on the frontend, engine on the backend via UCI protocol

### Engine Types
- **UCI engines** - Direct communication, no wrapper needed
- **CECP engines** - Wrapped by Polyglot to provide UCI interface
- **+ Book variants** - UCI engines with opening book support via Polyglot
- **CECP via Polyglot** - CECP engines accessible through Polyglot wrapper

## API Endpoints

### POST `/api/computer-move`

Request a computer move for the current position.

**Request:**
```json
{
  "fen": "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"
}
```

**Response:**
```json
{
  "move": "e7e5",
  "fen": "rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 2"
}
```

## Game Features

- **Drag and Drop**: Move pieces by dragging them
- **Move Validation**: Only legal chess moves are allowed
- **New Game**: Start a fresh game at any time
- **Flip Board**: Rotate the board 180 degrees
- **Stockfish Engine**: World-class chess AI (configurable strength via move time)
- **Adjustable Difficulty**: Modify `moveTime` in `server/chess.go` to change Stockfish's thinking time

## Technologies Used

### Backend
- **Go**: Backend server and application logic
- **Chess Engines**: Multiple engine support
  - **UCI Protocol**: Stockfish, Fruit, Toga, Leela, etc.
  - **CECP Protocol**: GNU Chess, Crafty, Sjeng (via Polyglot)
- **Polyglot**: Chess engine wrapper for CECP engines and opening books
- **notnil/chess**: Go chess library for move generation and validation
- **Bubble Tea**: TUI framework for the terminal interface
- **Lipgloss**: Styling library for beautiful terminal output
- **Embedded Templates & Assets**: Everything embedded in the binary using Go's `embed` package

### Frontend
- **chessboardjs v1.0.0**: Interactive chess board library (local copy - modify as needed!)
- **chess.js v0.10.3**: Chess game state management (local copy)
- **jQuery v3.5.1**: JavaScript library (local copy, required by chessboardjs)

## Notes

### Application
- The application uses Go's `embed` package to include all templates, libraries, and assets in the binary
- **Self-contained binary** - works offline, only requires external chess engines
- All JavaScript libraries (chessboardjs, jQuery) are embedded and served locally
- **Easy to modify**: Edit `server/assets/js/chessboard-1.0.0.min.js` and rebuild to customize the chess board behavior
- Browser auto-opening works on Linux, macOS, and Windows
- Use the `--no-browser` flag if you prefer to open the browser manually
- All assets are served from the embedded `/assets/` directory structure

### Engine Discovery
- Engines are discovered automatically from system PATH
- UCI engines work without any additional dependencies
- CECP engines require Polyglot to be installed
- Polyglot configs are generated in `/tmp/go-chess-polyglot/`
- Each engine gets a unique config file based on MD5 hash
- Check `chess-debug.log` for detailed discovery information

### Opening Books (Polyglot)
- Books must be in Polyglot binary format (.bin)
- GNU Chess includes books at `/usr/share/games/gnuchess/book.bin`
- Without `--book-file`, only UCI engines are available
- With `--book-file`, UCI engines get optional "+ Book" variants
- CECP engines always use the book when available
- Polyglot logs book usage in `/tmp/go-chess-polyglot/polyglot-*.log`

### Opening Name Recognition
- Opening database loaded from `server/assets/openings/*.tsv`
- Contains 3,594 chess openings from Lichess database
- Loaded at server startup (~7-8 seconds, one-time cost)
- Fast in-memory trie structure for microsecond lookups
- API endpoint: `POST /api/opening` with move array
- Update database: `./update-openings.sh [commit-hash]`
- Source: https://github.com/lichess-org/chess-openings

## License

This is a template application. Add your own license as needed.

## Contributing

This is a template project. Feel free to fork and modify for your needs.
