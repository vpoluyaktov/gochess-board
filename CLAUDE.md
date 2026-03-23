# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Project Is

A web-based chess **analysis tool** written in Go. The primary use case is loading games from PGN libraries (single or multi-game files), navigating move by move, and getting real-time engine evaluation with an evaluation graph and variation explorer. Playing games against engines is a secondary feature.

Key analysis capabilities: PGN load/paste/save, game picker modal for multi-game PGN files, move navigation, live WebSocket engine analysis (best move / top-3 / principal variation arrows), evaluation graph, opening recognition (ECO codes), and a variation explorer that opens alternative lines in a separate window and can merge them back.

## Build & Run

```bash
# Build
go build -o gochess-board ./cmd/gochess-board

# Run (opens browser, starts TUI)
./gochess-board

# Common flags
./gochess-board --no-browser --no-tui --port 35256 --book-file /path/to/book.bin --log-level DEBUG
```

## Testing

```bash
# Unit tests
go test ./...
go test -v ./...                  # verbose
go test -run TestName ./...       # single test
go test -race ./...               # race detection

# Code quality
go fmt ./...
go vet ./...

# Shell-based integration/API tests
bash tests/quality/test_code.sh
bash tests/api/test_api.sh        # requires running server
bash tests/engine/test_uci.sh

# E2E (Playwright, requires running server)
cd tests/ui/playwright && npm install && npx playwright test
```

## Architecture

### Entry Point & Server

`cmd/gochess-board/main.go` — parses flags, discovers engines, starts the TUI and HTTP server.

`internal/server/` — HTTP server with embedded assets (HTML/JS/CSS/images via `go:embed`). Key handlers:
- `GET /` — serves the chess UI (`templates/index.html`)
- `POST /api/computer-move` — receives FEN + engine name, returns best move
- `GET /api/engines` — lists discovered engines
- `GET /api/opening` — looks up opening name by FEN
- `WS /ws/analysis` — streams real-time analysis lines

### Engine System

`internal/engines/` — engine abstraction layer:
- `engines.go` — `Engine` interface used by all engine types
- `discovery.go` — auto-discovers UCI/CECP engines on system PATH at startup
- `uci.go` / `cecp.go` — protocol implementations (spawn subprocess, communicate via stdin/stdout)
- `pool.go` — persistent engine process pool (avoids restart overhead per move)
- `monitor.go` — health monitoring for engine processes

`internal/engines/builtin/` — pure Go engine (~1400-1600 ELO):
- `engine.go` — top-level UCI-compatible interface
- `search.go` — minimax with alpha-beta pruning
- `quiescence.go` — tactical search extension
- `evaluation.go` — material + piece-square tables + king safety + mobility
- `transposition.go` — 64MB transposition table
- `move_ordering.go` + `killer_moves.go` — search optimizations

### Supporting Packages

- `internal/book/` — Polyglot `.bin` opening book reader (pure Go, no C deps); uses Zobrist hashing
- `internal/opening/` — ECO opening name lookup from embedded Lichess database (3,594 openings)
- `internal/analysis/` — WebSocket streaming of engine analysis lines (multi-PV)
- `internal/tui/` — Bubble Tea TUI showing live game/engine stats
- `internal/logger/` — structured logging

### Frontend

Embedded in `internal/server/assets/` — chessboard.js, chess.js, jQuery, CodeMirror, and custom JS modules (`gochess-main.js`, `gochess-board.js`, `gochess-engine.js`, `gochess-analysis.js`). All assets are embedded at compile time via `go:embed`.

## Docker

```bash
docker build -f docker/Dockerfile -t gochess-board .
docker-compose -f docker/docker-compose.yml up -d
```

The Docker image pre-installs ~10 engines (Stockfish, GNU Chess, Crafty, Fruit, etc.) and runs on port 35256 as a non-root `chess` user.

## Key Conventions

- The `Engine` interface in `internal/engines/engines.go` must be satisfied by all engine implementations.
- Engine discovery runs once at startup; the list is cached and served via `/api/engines`.
- The built-in engine is always available regardless of what external engines are installed.
- FEN strings are the primary position representation passed between frontend and backend.
