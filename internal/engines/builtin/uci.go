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
			// Parse go command and search (pass FEN for proper time management)
			thinkTime := parseGoCommandWithFEN(parts, currentFEN)

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

// TimeControl holds parsed time control parameters
type TimeControl struct {
	WhiteTime   time.Duration // White's remaining time
	BlackTime   time.Duration // Black's remaining time
	WhiteInc    time.Duration // White's increment per move
	BlackInc    time.Duration // Black's increment per move
	MovesToGo   int           // Moves until next time control (0 = sudden death)
	MoveTime    time.Duration // Fixed time per move (overrides other settings)
	Depth       int           // Search to specific depth (0 = no limit)
	Infinite    bool          // Search indefinitely
	IsWhiteTurn bool          // Whose turn it is
}

// parseGoCommand parses UCI go command and returns think time
// Format: go [searchmoves <move1> ... <moveN>] [ponder] [wtime <x>] [btime <x>]
//
//	[winc <x>] [binc <x>] [movestogo <x>] [depth <x>] [nodes <x>] [mate <x>]
//	[movetime <x>] [infinite]
func parseGoCommand(parts []string) time.Duration {
	return parseGoCommandWithFEN(parts, "")
}

// parseGoCommandWithFEN parses UCI go command with FEN context for proper time management
func parseGoCommandWithFEN(parts []string, fen string) time.Duration {
	tc := TimeControl{
		IsWhiteTurn: true, // Default to white
	}

	// Determine whose turn from FEN
	if fen != "" {
		fenParts := strings.Fields(fen)
		if len(fenParts) >= 2 && fenParts[1] == "b" {
			tc.IsWhiteTurn = false
		}
	}

	// Parse all time control parameters
	for i := 1; i < len(parts); i++ {
		switch parts[i] {
		case "movetime":
			if i+1 < len(parts) {
				if ms, err := parseDurationMs(parts[i+1]); err == nil {
					tc.MoveTime = ms
				}
				i++
			}

		case "wtime":
			if i+1 < len(parts) {
				if ms, err := parseDurationMs(parts[i+1]); err == nil {
					tc.WhiteTime = ms
				}
				i++
			}

		case "btime":
			if i+1 < len(parts) {
				if ms, err := parseDurationMs(parts[i+1]); err == nil {
					tc.BlackTime = ms
				}
				i++
			}

		case "winc":
			if i+1 < len(parts) {
				if ms, err := parseDurationMs(parts[i+1]); err == nil {
					tc.WhiteInc = ms
				}
				i++
			}

		case "binc":
			if i+1 < len(parts) {
				if ms, err := parseDurationMs(parts[i+1]); err == nil {
					tc.BlackInc = ms
				}
				i++
			}

		case "movestogo":
			if i+1 < len(parts) {
				if n, err := parseInt(parts[i+1]); err == nil {
					tc.MovesToGo = n
				}
				i++
			}

		case "depth":
			if i+1 < len(parts) {
				if n, err := parseInt(parts[i+1]); err == nil {
					tc.Depth = n
				}
				i++
			}

		case "infinite":
			tc.Infinite = true
		}
	}

	return calculateThinkTime(tc)
}

// parseDurationMs parses a string as milliseconds and returns a Duration
func parseDurationMs(s string) (time.Duration, error) {
	return time.ParseDuration(s + "ms")
}

// parseInt parses a string as an integer
func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// calculateThinkTime calculates optimal think time based on time control
func calculateThinkTime(tc TimeControl) time.Duration {
	// Default think time
	defaultTime := 2 * time.Second

	// Infinite search
	if tc.Infinite {
		return 1 * time.Hour
	}

	// Fixed time per move (highest priority)
	if tc.MoveTime > 0 {
		// Use 95% of movetime to have safety margin
		return tc.MoveTime * 95 / 100
	}

	// Get our time and increment based on whose turn it is
	var ourTime, ourInc time.Duration
	if tc.IsWhiteTurn {
		ourTime = tc.WhiteTime
		ourInc = tc.WhiteInc
	} else {
		ourTime = tc.BlackTime
		ourInc = tc.BlackInc
	}

	// If no time specified, use default
	if ourTime == 0 {
		return defaultTime
	}

	// Calculate think time based on game situation
	var thinkTime time.Duration

	if tc.MovesToGo > 0 {
		// Time control with moves to go
		// Divide remaining time by moves to go, plus some increment
		thinkTime = ourTime / time.Duration(tc.MovesToGo)
		// Add most of the increment (save a bit for safety)
		thinkTime += ourInc * 90 / 100
	} else {
		// Sudden death or increment-based time control
		// Estimate ~40 moves remaining in the game
		estimatedMovesLeft := 40

		// Base time: divide remaining time by estimated moves
		thinkTime = ourTime / time.Duration(estimatedMovesLeft)

		// Add increment (if any)
		if ourInc > 0 {
			// We can use most of the increment each move
			thinkTime += ourInc * 90 / 100
		}
	}

	// Safety margins
	minTime := 100 * time.Millisecond
	maxTime := ourTime / 4 // Never use more than 25% of remaining time

	// Apply minimum
	if thinkTime < minTime {
		thinkTime = minTime
	}

	// Apply maximum
	if thinkTime > maxTime {
		thinkTime = maxTime
	}

	// Emergency: if we have very little time, think fast
	if ourTime < 5*time.Second {
		// Use 5% of remaining time + half increment
		thinkTime = ourTime/20 + ourInc/2
		if thinkTime < minTime {
			thinkTime = minTime
		}
	}

	return thinkTime
}
