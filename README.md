# GoChess Board

A web-based chess application written in Go that lets you play against a **built-in chess engine** or any installed **UCI/CECP engine** (Stockfish, GNU Chess, Crafty, and more). The application is fully self-contained — all assets are embedded in a single binary that works offline with no external dependencies.

![Screenshot](docs/screenshot.png)
<!-- TODO: Add an actual screenshot of the application -->

## Live Demo

Try it now at **[https://demo.gochess-board.org](https://demo.gochess-board.org)**

The demo instance runs the Docker image with the following engines:

| Engine | Protocol | ELO Range |
|--------|----------|-----------|
| GoChess (Built-in) | Internal | ~1000-1200 |
| Stockfish 16 | UCI | 1320-3190 |
| Fairy-Stockfish 11.1 | UCI | 1350-2850 |
| Fruit 2.1 | UCI | ~2700 |
| Toga II 3.0 | UCI | ~2700 |
| Glaurung 2.2 | UCI | ~2700 |
| Crafty 23.4 | CECP | ~2400 |
| HoiChess | CECP | ~2200 |
| Fairy-Max | CECP | ~1800 |

## Features

- **Built-in Chess Engine** — native Go engine (~1000-1200 ELO), works immediately with zero setup
- **Multi-Engine Support** — automatic discovery of UCI and CECP/XBoard engines on the system PATH
- **Persistent Engine Pool** — optional engine pooling to eliminate startup overhead between moves
- **Opening Book** — native Polyglot `.bin` reader (pure Go, no external tools)
- **Opening Recognition** — identifies openings in real-time with ECO codes (3,594 openings from the Lichess database)
- **Real-time TUI** — terminal interface showing live game stats, engine analysis, and active engine status
- **Self-Contained Binary** — all HTML, JS, CSS, images, and opening database embedded via Go's `embed` package
- **WebSocket Analysis** — real-time position analysis streamed to the browser
- **Docker Support** — pre-built multi-platform images with 10+ chess engines included
- **CI/CD** — GitHub Actions workflows for Docker builds and releases

## Quick Start

### Option 1: Docker (Recommended)

```bash
docker run -p 35256:35256 vpoluyaktov/gochess-board:latest
```

Then open **http://localhost:35256** in your browser.

See [Docker documentation](docker/DOCKER.md) for more options (custom ports, volumes, docker-compose).

### Option 2: Build from Source

**Prerequisites:** Go 1.24+

```bash
git clone https://github.com/vpoluyaktov/gochess-board.git
cd gochess-board
go build -o gochess-board ./cmd/gochess-board
./gochess-board
```

The built-in engine works immediately. Install external engines for stronger play (see [Installing Engines](#installing-chess-engines)).

## Usage

```bash
# Default: opens browser, starts TUI
./gochess-board

# Headless mode
./gochess-board --no-browser --no-tui

# With opening book
./gochess-board --book-file /path/to/book.bin

# Custom port
./gochess-board --port 8080

# Persistent engine pool (engines stay alive between moves)
./gochess-board --persistent-engines

# Full featured
./gochess-board --persistent-engines --book-file /path/to/book.bin --port 8080
```

### Command Line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `35256` | Web server port |
| `--no-browser` | `false` | Don't auto-open browser |
| `--no-tui` | `false` | Disable terminal UI |
| `--book-file` | — | Path to Polyglot `.bin` opening book |
| `--persistent-engines` | `false` | Keep engines alive between moves |
| `--log-level` | `INFO` | Log level (`DEBUG`, `INFO`, `WARN`, `ERROR`) |
| `--restart` | `false` | Kill existing process on the same port before starting |

## Supported Chess Engines

The application automatically discovers engines on the system PATH.

### Built-in Engine
- **GoChess** — pure Go, minimax with alpha-beta pruning (~1000-1200 ELO)

### UCI Engines
Stockfish, Fruit, Toga II, Leela Chess Zero (lc0), Dragon by Komodo, Fairy-Stockfish, Glaurung, Ethereal, Berserk, Koivisto, RubiChess, and more.

### CECP/XBoard Engines
GNU Chess, Crafty, Sjeng, Phalanx, HoiChess, Fairy-Max.

### Opening Book
- Native Polyglot `.bin` reader (pure Go)
- Book moves checked before engine calculation (instant 0ms response)
- All engines (UCI and CECP) can use the book

## Installing Chess Engines

### Linux (Ubuntu/Debian)

```bash
# UCI engines
sudo apt-get install stockfish fruit toga2

# CECP engines
sudo apt-get install crafty sjeng phalanx hoichess fairymax

# GNU Chess 6.2.9 from source (recommended — fixes CVE-2021-30184)
sudo apt-get install build-essential
cd /tmp && wget https://ftp.gnu.org/gnu/chess/gnuchess-6.2.9.tar.gz
tar xzf gnuchess-6.2.9.tar.gz && cd gnuchess-6.2.9
./configure && make && sudo make install
```

### macOS (Homebrew)

```bash
brew install stockfish fruit gnu-chess crafty
```

### Windows

Download engines from their official sites and add to PATH:
- **Stockfish**: https://stockfishchess.org/download/
- **GNU Chess**: https://ftp.gnu.org/gnu/chess/

## Opening Books

### Sources

- **Included with Fruit**: `/usr/share/games/fruit/book_small.bin` (31K entries)
- **Included with GNU Chess**: `/usr/share/gnuchess/smallbook.bin`
- **Popular books**: Performance.bin, Gm2001.bin, ProDeo.bin
- **Create your own**: `polyglot make-book -pgn games.pgn -bin mybook.bin`

### Usage

```bash
./gochess-board --book-file /usr/share/games/fruit/book_small.bin
```

## Project Structure

```
gochess-board/
├── cmd/gochess-board/       # Application entry point
│   └── main.go
├── internal/
│   ├── analysis/            # Position analysis (WebSocket streaming)
│   ├── book/                # Polyglot opening book reader
│   ├── engines/             # Engine discovery, UCI, CECP, pool, monitor
│   │   └── builtin/         # Built-in Go chess engine
│   ├── logger/              # Structured logging
│   ├── opening/             # Opening name recognition (ECO codes)
│   ├── server/              # HTTP server, handlers, embedded assets
│   │   ├── assets/          # JS, CSS, images, opening database (embedded)
│   │   └── templates/       # HTML templates (embedded)
│   ├── tui/                 # Terminal UI (Bubble Tea)
│   └── utils/               # Shared utilities
├── docker/                  # Dockerfile, docker-compose, documentation
├── scripts/                 # Test and maintenance scripts
├── tests/                   # API, engine, integration, UI tests
├── .github/workflows/       # CI/CD (Docker build & release)
├── go.mod
└── README.md
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/` | Chess board UI |
| `GET` | `/api/engines` | List discovered engines |
| `POST` | `/api/computer-move` | Request engine move for a FEN position |
| `POST` | `/api/opening` | Look up opening name by move sequence |
| `WS` | `/ws/analysis` | WebSocket stream for real-time position analysis |

### Example: Request a Move

```bash
curl -X POST http://localhost:35256/api/computer-move \
  -H "Content-Type: application/json" \
  -d '{"fen": "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"}'
```

## TUI Interface

The terminal UI displays a three-column layout:

- **Game Stats** — move counts, game duration
- **Engine Info** — last move, think time, ELO rating
- **Position** — current FEN

Plus an **Active Engines** table showing running/idle engine instances when using `--persistent-engines`.

Press `q` or `Ctrl+C` to quit.

## Technologies

### Backend
- **Go 1.24** — server, engine communication, game logic
- **[notnil/chess](https://github.com/notnil/chess)** — move generation and validation
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** — terminal UI framework
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** — terminal styling
- **[gorilla/websocket](https://github.com/gorilla/websocket)** — WebSocket support

### Frontend
- **[chessboard.js](https://chessboardjs.com/) v1.0.0** — interactive chess board (embedded)
- **[chess.js](https://github.com/jhlywa/chess.js) v0.10.3** — game state management (embedded)
- **jQuery v3.5.1** — DOM manipulation (embedded)

### Infrastructure
- **Docker** — multi-platform images with pre-installed engines
- **GitHub Actions** — automated Docker builds and releases

## Building for Different Platforms

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o gochess-board-linux ./cmd/gochess-board

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o gochess-board-macos ./cmd/gochess-board

# Windows
GOOS=windows GOARCH=amd64 go build -o gochess-board.exe ./cmd/gochess-board
```

## License

MIT License. See [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
