package builtin

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// RunUCI runs the engine in UCI protocol mode
func RunUCI() {
	engine := NewEngine()
	scanner := bufio.NewScanner(os.Stdin)

	// Default to starting position
	currentFEN := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]

		switch command {
		case "uci":
			// Identify engine
			fmt.Println("id name GoChess Board Engine")
			fmt.Println("id author Vladimir Poluyaktov")
			// Options can be added here
			// fmt.Println("option name Hash type spin default 64 min 1 max 1024")
			fmt.Println("uciok")

		case "isready":
			// Engine is always ready
			fmt.Println("readyok")

		case "ucinewgame":
			// Reset to starting position
			currentFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

		case "position":
			// Parse position command
			currentFEN = parsePositionCommand(parts)

		case "go":
			// Parse go command and search
			thinkTime := parseGoCommand(parts)

			// Get best move
			move, err := engine.GetBestMove(currentFEN, thinkTime)
			if err != nil {
				// Send error as comment (UCI allows this)
				fmt.Printf("info string Error: %v\n", err)
				// Send a dummy move to avoid hanging
				fmt.Println("bestmove 0000")
				continue
			}

			// Send best move
			fmt.Printf("bestmove %s\n", move)

		case "stop":
			// Stop searching (our engine doesn't support async search yet)
			// Just acknowledge

		case "quit":
			// Exit cleanly
			return

		case "setoption":
			// Parse setoption command (currently ignored)
			// Format: setoption name <name> value <value>
			// We could implement Hash size, threads, etc.

		default:
			// Unknown command - UCI spec says to ignore
		}
	}

	// Handle scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
}

// parsePositionCommand parses UCI position command
// Format: position [fen <fenstring> | startpos] moves <move1> ... <moveN>
func parsePositionCommand(parts []string) string {
	if len(parts) < 2 {
		return "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	}

	fen := ""

	if parts[1] == "startpos" {
		fen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	} else if parts[1] == "fen" {
		// Extract FEN string (parts 2 onwards until "moves" or end)
		fenParts := []string{}
		for i := 2; i < len(parts); i++ {
			if parts[i] == "moves" {
				break
			}
			fenParts = append(fenParts, parts[i])
		}
		fen = strings.Join(fenParts, " ")
	}

	// TODO: Apply moves if present
	// For now, we just return the FEN position
	// A full implementation would apply the moves to get the final position

	return fen
}

// parseGoCommand parses UCI go command and returns think time
// Format: go [searchmoves <move1> ... <moveN>] [ponder] [wtime <x>] [btime <x>]
//
//	[winc <x>] [binc <x>] [movestogo <x>] [depth <x>] [nodes <x>] [mate <x>]
//	[movetime <x>] [infinite]
func parseGoCommand(parts []string) time.Duration {
	// Default think time
	defaultTime := 2 * time.Second

	for i := 1; i < len(parts); i++ {
		switch parts[i] {
		case "movetime":
			// Fixed time per move in milliseconds
			if i+1 < len(parts) {
				if ms, err := time.ParseDuration(parts[i+1] + "ms"); err == nil {
					return ms
				}
			}

		case "wtime":
			// White time remaining in milliseconds
			// Simple time management: use 1/30th of remaining time
			if i+1 < len(parts) {
				if ms, err := time.ParseDuration(parts[i+1] + "ms"); err == nil {
					return ms / 30
				}
			}

		case "btime":
			// Black time remaining in milliseconds
			// Simple time management: use 1/30th of remaining time
			if i+1 < len(parts) {
				if ms, err := time.ParseDuration(parts[i+1] + "ms"); err == nil {
					return ms / 30
				}
			}

		case "infinite":
			// Search indefinitely (use a very long time)
			return 1 * time.Hour

		case "depth":
			// Search to specific depth (we don't support this yet)
			// Just use default time
		}
	}

	return defaultTime
}
