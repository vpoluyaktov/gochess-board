package builtin

// Bratko-Kopec Test Suite
// A standard chess test suite designed by Dr. Ivan Bratko and Dr. Danny Kopec in 1982
// to evaluate human or machine chess ability.
// This test has been a standard for nearly 40 years in computer chess.

import (
	"fmt"
	"time"
)

// BratkoKopecPosition represents a test position from the Bratko-Kopec suite
type BratkoKopecPosition struct {
	ID          string
	FEN         string
	BestMoves   []string // Best moves in UCI format
	Description string
}

// GetBratkoKopecSuite returns the complete Bratko-Kopec test suite (24 positions)
func GetBratkoKopecSuite() []BratkoKopecPosition {
	return []BratkoKopecPosition{
		{
			ID:          "BK.01",
			FEN:         "1k1r4/pp1b1R2/3q2pp/4p3/2B5/4Q3/PPP2B2/2K5 b - - 0 1",
			BestMoves:   []string{"d6d1"},
			Description: "Qd1+ wins",
		},
		{
			ID:          "BK.02",
			FEN:         "3r1k2/4npp1/1ppr3p/p6P/P2PPPP1/1NR5/5K2/2R5 w - - 0 1",
			BestMoves:   []string{"d4d5"},
			Description: "d5 breakthrough",
		},
		{
			ID:          "BK.03",
			FEN:         "2q1rr1k/3bbnnp/p2p1pp1/2pPp3/PpP1P1P1/1P2BNNP/2BQ1PRK/7R b - - 0 1",
			BestMoves:   []string{"f6f5"},
			Description: "f5 counterplay",
		},
		{
			ID:          "BK.04",
			FEN:         "rnbqkb1r/p3pppp/1p6/2ppP3/3N4/2P5/PPP1QPPP/R1B1KB1R w KQkq - 0 1",
			BestMoves:   []string{"e5e6"},
			Description: "e6 pawn advance",
		},
		{
			ID:          "BK.05",
			FEN:         "r1b2rk1/2q1b1pp/p2ppn2/1p6/3QP3/1BN1B3/PPP3PP/R4RK1 w - - 0 1",
			BestMoves:   []string{"c3d5", "a2a4"},
			Description: "Nd5 or a4",
		},
		{
			ID:          "BK.06",
			FEN:         "2r3k1/pppR1pp1/4p3/4P1P1/5P2/1P4K1/P1P5/8 w - - 0 1",
			BestMoves:   []string{"g5g6"},
			Description: "g6 breakthrough",
		},
		{
			ID:          "BK.07",
			FEN:         "1nk1r1r1/pp2n1pp/4p3/q2pPp1N/b1pP1P2/B1P2R2/2P1B1PP/R2Q2K1 w - - 0 1",
			BestMoves:   []string{"h5f6"},
			Description: "Nf6 attack",
		},
		{
			ID:          "BK.08",
			FEN:         "4b3/p3kp2/6p1/3pP2p/2pP1P2/4K1P1/P3N2P/8 w - - 0 1",
			BestMoves:   []string{"f4f5"},
			Description: "f5 pawn break",
		},
		{
			ID:          "BK.09",
			FEN:         "2kr1bnr/pbpq4/2n1pp2/3p3p/3P1P1B/2N2N1Q/PPP3PP/2KR1B1R w - - 0 1",
			BestMoves:   []string{"f4f5"},
			Description: "f5 attack",
		},
		{
			ID:          "BK.10",
			FEN:         "3rr1k1/pp3pp1/1qn2np1/8/3p4/PP1R1P2/2P1NQPP/R1B3K1 b - - 0 1",
			BestMoves:   []string{"c6e5"},
			Description: "Ne5 centralization",
		},
		{
			ID:          "BK.11",
			FEN:         "2r1nrk1/p2q1ppp/bp1p4/n1pPp3/P1P1P3/2PBB1N1/4QPPP/R4RK1 w - - 0 1",
			BestMoves:   []string{"f2f4"},
			Description: "f4 pawn break",
		},
		{
			ID:          "BK.12",
			FEN:         "r3r1k1/ppqb1ppp/8/4p1NQ/8/2P5/PP3PPP/R3R1K1 b - - 0 1",
			BestMoves:   []string{"d7f5"},
			Description: "Bf5 defense",
		},
		{
			ID:          "BK.13",
			FEN:         "r2q1rk1/4bppp/p2p4/2pP4/3pP3/3Q4/PP1B1PPP/R3R1K1 w - - 0 1",
			BestMoves:   []string{"b2b4"},
			Description: "b4 pawn break",
		},
		{
			ID:          "BK.14",
			FEN:         "rnb2r1k/pp2p2p/2pp2p1/q2P1p2/8/1Pb2NP1/PB2PPBP/R2Q1RK1 w - - 0 1",
			BestMoves:   []string{"d1d2", "d1e1"},
			Description: "Qd2 or Qe1",
		},
		{
			ID:          "BK.15",
			FEN:         "2r3k1/1p2q1pp/2b1pr2/p1pp4/6Q1/1P1PP1R1/P1PN2PP/5RK1 w - - 0 1",
			BestMoves:   []string{"g4g7"},
			Description: "Qxg7+ wins",
		},
		{
			ID:          "BK.16",
			FEN:         "r1bqkb1r/4npp1/p1p4p/1p1pP1B1/8/1B6/PPPN1PPP/R2Q1RK1 w kq - 0 1",
			BestMoves:   []string{"d2e4"},
			Description: "Ne4 attack",
		},
		{
			ID:          "BK.17",
			FEN:         "r2q1rk1/1ppnbppp/p2p1nb1/3Pp3/2P1P1P1/2N2N1P/PPB1QP2/R1B2RK1 b - - 0 1",
			BestMoves:   []string{"h7h5"},
			Description: "h5 counterplay",
		},
		{
			ID:          "BK.18",
			FEN:         "r1bq1rk1/pp2ppbp/2np2p1/2n5/P3PP2/N1P2N2/1PB3PP/R1B1QRK1 b - - 0 1",
			BestMoves:   []string{"c5b3"},
			Description: "Nb3 fork",
		},
		{
			ID:          "BK.19",
			FEN:         "3rr3/2pq2pk/p2p1pnp/8/2QBPP2/1P6/P5PP/4RRK1 b - - 0 1",
			BestMoves:   []string{"e8e4"},
			Description: "Rxe4 wins exchange",
		},
		{
			ID:          "BK.20",
			FEN:         "r4k2/pb2bp1r/1p1qp2p/3pNp2/3P1P2/2N3P1/PPP1Q2P/2KRR3 w - - 0 1",
			BestMoves:   []string{"g3g4"},
			Description: "g4 pawn break",
		},
		{
			ID:          "BK.21",
			FEN:         "3rn2k/ppb2rpp/2ppqp2/5N2/2P1P3/1P5Q/PB3PPP/3RR1K1 w - - 0 1",
			BestMoves:   []string{"f5h6"},
			Description: "Nh6 attack",
		},
		{
			ID:          "BK.22",
			FEN:         "2r2rk1/1bqnbpp1/1p1ppn1p/pP6/N1P1P3/P2B1N1P/1B2QPP1/R2R2K1 b - - 0 1",
			BestMoves:   []string{"b7e4"},
			Description: "Bxe4 wins pawn",
		},
		{
			ID:          "BK.23",
			FEN:         "r1bqk2r/pp2bppp/2p5/3pP3/P2Q1P2/2N1B3/1PP3PP/R4RK1 b kq - 0 1",
			BestMoves:   []string{"f7f6"},
			Description: "f6 challenge",
		},
		{
			ID:          "BK.24",
			FEN:         "r2qnrnk/p2b2b1/1p1p2pp/2pPpp2/1PP1P3/PRNBB3/3QNPPP/5RK1 w - - 0 1",
			BestMoves:   []string{"f2f4"},
			Description: "f4 pawn break",
		},
	}
}

// RunBratkoKopecTest runs the Bratko-Kopec test suite
func RunBratkoKopecTest(engine *InternalEngine, timeLimit time.Duration, verbose bool) (int, int, []string) {
	suite := GetBratkoKopecSuite()
	correct := 0
	total := len(suite)
	var failedTests []string

	if verbose {
		fmt.Println("Running Bratko-Kopec Test Suite")
		fmt.Println("================================")
		fmt.Println()
	}

	for i, test := range suite {
		if verbose {
			fmt.Printf("Test %d/%d: %s\n", i+1, total, test.ID)
			fmt.Printf("  Position: %s\n", test.Description)
		}

		// Get engine's best move
		move, err := engine.GetBestMove(test.FEN, timeLimit)
		if err != nil {
			if verbose {
				fmt.Printf("  ✗ Error: %v\n\n", err)
			}
			failedTests = append(failedTests, fmt.Sprintf("%s: Error - %v", test.ID, err))
			continue
		}

		// Check if move matches any of the best moves
		found := false
		for _, bestMove := range test.BestMoves {
			if move == bestMove {
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
			failedTests = append(failedTests, fmt.Sprintf("%s: Got %s, Expected %v", test.ID, move, test.BestMoves))
		}
	}

	return correct, total, failedTests
}

// EstimateBKRating estimates rating based on Bratko-Kopec score
// Based on the original Bratko-Kopec rating scale
func EstimateBKRating(correct int) (int, string) {
	switch {
	case correct >= 22:
		return 2500, "Grandmaster"
	case correct >= 19:
		return 2400, "International Master"
	case correct >= 16:
		return 2200, "Master"
	case correct >= 13:
		return 2000, "Expert"
	case correct >= 10:
		return 1800, "Class A"
	case correct >= 7:
		return 1600, "Class B"
	case correct >= 4:
		return 1400, "Class C"
	default:
		return 1200, "Class D"
	}
}
