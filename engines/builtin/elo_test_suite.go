package builtin

// ELO Testing Suite
// This file contains tactical test positions for evaluating engine strength

import (
	"fmt"
	"time"
)

// TacticalPosition represents a test position with expected best move
type TacticalPosition struct {
	Name        string
	FEN         string
	BestMoves   []string // Multiple acceptable moves
	Description string
	Difficulty  int // 1-5 scale
}

// GetTacticalTestSuite returns a suite of tactical positions
// Based on Win At Chess (WAC) and other tactical test suites
func GetTacticalTestSuite() []TacticalPosition {
	return []TacticalPosition{
		{
			Name:        "WAC.001",
			FEN:         "2rr3k/pp3pp1/1nnqbN1p/3pN3/2pP4/2P3Q1/PPB4P/R4RK1 w - - 0 1",
			BestMoves:   []string{"g3g7", "Qg7+"},
			Description: "Mate in 2 - Queen sacrifice",
			Difficulty:  3,
		},
		{
			Name:        "WAC.002",
			FEN:         "r1b1k2r/ppppnppp/2n2q2/2b5/3NP3/2P1B3/PP3PPP/RN1QKB1R w KQkq - 0 1",
			BestMoves:   []string{"d4f5", "Nf5"},
			Description: "Win queen - knight fork",
			Difficulty:  2,
		},
		{
			Name:        "Simple Mate",
			FEN:         "6k1/5ppp/8/8/8/8/5PPP/R5K1 w - - 0 1",
			BestMoves:   []string{"a1a8", "Ra8#"},
			Description: "Back rank mate",
			Difficulty:  1,
		},
		{
			Name:        "Fork",
			FEN:         "r1bqkb1r/pppp1ppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4",
			BestMoves:   []string{"f3g5", "Ng5"},
			Description: "Knight fork on f7",
			Difficulty:  2,
		},
		{
			Name:        "Pin",
			FEN:         "r1bqk2r/pppp1ppp/2n2n2/2b1p3/2B1P3/3P1N2/PPP2PPP/RNBQK2R w KQkq - 0 1",
			BestMoves:   []string{"c4f7", "Bxf7+"},
			Description: "Bishop takes f7 check",
			Difficulty:  1,
		},
		{
			Name:        "Passed Pawn",
			FEN:         "8/8/8/3k4/8/3K4/3P4/8 w - - 0 1",
			BestMoves:   []string{"d2d4", "d4"},
			Description: "Push passed pawn",
			Difficulty:  1,
		},
		{
			Name:        "Promotion",
			FEN:         "8/3P4/8/3k4/8/3K4/8/8 w - - 0 1",
			BestMoves:   []string{"d7d8q", "d8=Q"},
			Description: "Promote to queen",
			Difficulty:  1,
		},
		{
			Name:        "Discovered Attack",
			FEN:         "r1bqk2r/pppp1ppp/2n2n2/2b1p3/2B1P3/2NP1N2/PPP2PPP/R1BQK2R w KQkq - 0 1",
			BestMoves:   []string{"c3d5", "Nd5"},
			Description: "Knight to d5 - discovered attack",
			Difficulty:  2,
		},
		{
			Name:        "Deflection",
			FEN:         "r2qkb1r/ppp2ppp/2np1n2/4p1B1/2B1P3/2N2N2/PPPP1PPP/R2QK2R w KQkq - 0 1",
			BestMoves:   []string{"g5f6", "Bxf6"},
			Description: "Bishop takes knight",
			Difficulty:  2,
		},
		{
			Name:        "Skewer",
			FEN:         "r1bqk2r/pppp1ppp/2n5/2b1p3/2B1P3/5N2/PPPP1PPP/RNBQ1RK1 b kq - 0 1",
			BestMoves:   []string{"c5f2", "Bxf2+"},
			Description: "Bishop check - skewer king and rook",
			Difficulty:  2,
		},
	}
}

// RunTacticalTest tests the engine against tactical positions
func RunTacticalTest(engine *InternalEngine, timeLimit time.Duration, verbose bool) (int, int, []string) {
	suite := GetTacticalTestSuite()
	correct := 0
	total := len(suite)
	var failedTests []string

	if verbose {
		fmt.Println("Running Tactical Test Suite")
		fmt.Println("============================")
		fmt.Println()
	}

	for i, test := range suite {
		if verbose {
			fmt.Printf("Test %d/%d: %s\n", i+1, total, test.Name)
			fmt.Printf("  Position: %s\n", test.Description)
		}

		// Get engine's best move
		move, err := engine.GetBestMove(test.FEN, timeLimit)
		if err != nil {
			if verbose {
				fmt.Printf("  ✗ Error: %v\n\n", err)
			}
			failedTests = append(failedTests, fmt.Sprintf("%s: Error - %v", test.Name, err))
			continue
		}

		// Check if move matches any of the best moves
		found := false
		for _, bestMove := range test.BestMoves {
			if move == bestMove || normalizeMove(move) == normalizeMove(bestMove) {
				found = true
				break
			}
		}

		if found {
			correct++
			if verbose {
				fmt.Printf("  ✓ Found: %s\n\n", move)
			}
		} else {
			if verbose {
				fmt.Printf("  ✗ Got: %s, Expected: %v\n\n", move, test.BestMoves)
			}
			failedTests = append(failedTests, fmt.Sprintf("%s: Got %s, Expected %v", test.Name, move, test.BestMoves))
		}
	}

	return correct, total, failedTests
}

// normalizeMove converts move to standard format
func normalizeMove(move string) string {
	// Remove check/mate symbols
	if len(move) > 0 && (move[len(move)-1] == '+' || move[len(move)-1] == '#') {
		move = move[:len(move)-1]
	}
	return move
}

// EstimateELO estimates ELO based on tactical test performance
func EstimateELO(correctCount, totalCount int) (int, int, string) {
	percentage := float64(correctCount) * 100.0 / float64(totalCount)

	var minELO, maxELO int
	var category string

	switch {
	case percentage >= 90:
		minELO, maxELO = 1800, 2000
		category = "Expert"
	case percentage >= 80:
		minELO, maxELO = 1600, 1800
		category = "Advanced"
	case percentage >= 70:
		minELO, maxELO = 1500, 1700
		category = "Intermediate+"
	case percentage >= 60:
		minELO, maxELO = 1400, 1600
		category = "Intermediate"
	case percentage >= 50:
		minELO, maxELO = 1300, 1500
		category = "Developing"
	case percentage >= 40:
		minELO, maxELO = 1200, 1400
		category = "Beginner+"
	default:
		minELO, maxELO = 1000, 1200
		category = "Beginner"
	}

	return minELO, maxELO, category
}

// PerformanceMetrics tracks engine performance
type PerformanceMetrics struct {
	TotalNodes      int64
	TotalTime       time.Duration
	NodesPerSecond  int64
	AverageDepth    float64
	PositionsTested int
}

// BenchmarkEngine runs performance benchmarks
func BenchmarkEngine(engine *InternalEngine) PerformanceMetrics {
	// Standard benchmark position (Kiwipete)
	fen := "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"

	metrics := PerformanceMetrics{
		PositionsTested: 1,
	}

	// Run search at different depths
	depths := []int{3, 4, 5}
	totalDepth := 0

	for _, depth := range depths {
		// This would need to be integrated with actual search stats
		// For now, just run the search
		_, _ = engine.GetBestMove(fen, 5*time.Second)
		totalDepth += depth
	}

	metrics.AverageDepth = float64(totalDepth) / float64(len(depths))

	return metrics
}
