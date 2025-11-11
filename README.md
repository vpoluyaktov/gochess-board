# Go Chess Board Application

A Go web server that lets you play chess against **Stockfish**, one of the strongest chess engines in the world. The chess game logic runs on the backend using Go and Stockfish via UCI protocol, while the frontend provides an interactive chess board. The application automatically opens your default browser when started.

## Features

- **Play vs Stockfish**: Play chess against Stockfish chess engine (world-class strength!)
- **Backend AI**: Stockfish runs on the Go server via UCI protocol
- **Real-time TUI**: Beautiful terminal interface showing live game stats and Stockfish analysis
- **Move Validation**: Only legal moves are allowed
- **Auto-Browser Opening**: Automatically opens the chess board in your default browser
- **Fully Self-Contained**: All assets embedded in the binary - works completely offline!
  - Chess piece images
  - ChessboardJS library (ready for your modifications!)
  - Chess.js library (frontend game state)
  - jQuery library
  - CSS styles
- Interactive chess board with drag-and-drop functionality
- Beautiful, modern UI with gradient background
- Responsive design
- New Game and Flip Board controls

## Prerequisites

- Go 1.21 or higher
- **Stockfish chess engine** installed on your system
  ```bash
  # Ubuntu/Debian
  sudo apt-get install stockfish
  
  # macOS (using Homebrew)
  brew install stockfish
  
  # Windows
  # Download from https://stockfishchess.org/download/
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
1. Start a web server on `http://localhost:8080`
2. Automatically open the chess board in your default browser

### Command Line Options

- `-port`: Port for the web server (default: `8080`)
  ```bash
  ./go-chess -port=3000
  ```

- `-no-browser`: Don't automatically open the browser
  ```bash
  ./go-chess -no-browser
  ```
  Then manually open `http://localhost:8080` in your browser

- `-no-tui`: Disable the TUI interface (run in simple mode)
  ```bash
  ./go-chess -no-tui
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

1. **You play as White** - Drag and drop pieces to make your move
2. **Stockfish plays as Black** - The backend uses Stockfish engine to calculate the best move
3. **Move validation** - Illegal moves are rejected (pieces snap back)
4. **Game state** - Managed by chess.js on the frontend, Stockfish on the backend via UCI protocol

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
- **Stockfish**: World-class chess engine (via UCI protocol)
- **notnil/chess**: Go chess library for move generation and validation
- **Bubble Tea**: TUI framework for the terminal interface
- **Lipgloss**: Styling library for beautiful terminal output
- **Embedded Templates & Assets**: Everything embedded in the binary using Go's `embed` package

### Frontend
- **chessboardjs v1.0.0**: Interactive chess board library (local copy - modify as needed!)
- **chess.js v0.10.3**: Chess game state management (local copy)
- **jQuery v3.5.1**: JavaScript library (local copy, required by chessboardjs)

## Notes

- The application uses Go's `embed` package to include all templates, libraries, and assets in the binary
- **100% self-contained** - the compiled binary works completely offline with zero external dependencies
- All JavaScript libraries (chessboardjs, jQuery) are embedded and served locally
- **Easy to modify**: Edit `server/assets/js/chessboard-1.0.0.min.js` and rebuild to customize the chess board behavior
- Browser auto-opening works on Linux, macOS, and Windows
- Use the `-no-browser` flag if you prefer to open the browser manually
- All assets are served from the embedded `/assets/` directory structure

## License

This is a template application. Add your own license as needed.

## Contributing

This is a template project. Feel free to fork and modify for your needs.
