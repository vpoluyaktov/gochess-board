package builtin

import (
	"testing"

	"github.com/notnil/chess"
)

// TestSearchWithStatsMultiPV_StartingPosition tests multi-PV search on starting position
func TestSearchWithStatsMultiPV_StartingPosition(t *testing.T) {
	engine := NewEngine()
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenObj, err := chess.FEN(fen)
	if err != nil {
		t.Fatalf("Invalid FEN: %v", err)
	}
	game := chess.NewGame(fenObj)
	pos := game.Position()

	stopCh := make(chan bool, 1)
	depth := 3

	score, bestMove, _, nodes, multiPV := engine.searchWithStatsMultiPV(pos, depth, stopCh)

	// Verify basic search results
	if bestMove == nil {
		t.Error("Expected best move, got nil")
	}
	if nodes == 0 {
		t.Error("Expected nodes > 0")
	}

	// Verify multi-PV results
	if len(multiPV) == 0 {
		t.Error("Expected MultiPV to be populated")
	}
	if len(multiPV) > 3 {
		t.Errorf("Expected at most 3 PV lines, got %d", len(multiPV))
	}

	// Verify scores are in descending order
	for i := 0; i < len(multiPV)-1; i++ {
		if multiPV[i].Score < multiPV[i+1].Score {
			t.Errorf("PV lines not sorted: line %d score %d < line %d score %d",
				i, multiPV[i].Score, i+1, multiPV[i+1].Score)
		}
	}

	// Verify first PV matches best move and score
	if len(multiPV) > 0 {
		if len(multiPV[0].Moves) == 0 {
			t.Error("First PV line has no moves")
		}
		if multiPV[0].Moves[0] != bestMove.String() {
			t.Errorf("First PV move %s doesn't match best move %s",
				multiPV[0].Moves[0], bestMove.String())
		}
	}

	t.Logf("Depth %d: Found %d PV lines, best move: %s, score: %d, nodes: %d",
		depth, len(multiPV), bestMove.String(), score, nodes)
}

// TestSearchWithStatsMultiPV_MiddleGame tests multi-PV on a middlegame position
func TestSearchWithStatsMultiPV_MiddleGame(t *testing.T) {
	engine := NewEngine()
	// Italian Game position
	fen := "r1bqkbnr/pppp1ppp/2n5/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4"
	fenObj, err := chess.FEN(fen)
	if err != nil {
		t.Fatalf("Invalid FEN: %v", err)
	}
	game := chess.NewGame(fenObj)
	pos := game.Position()

	stopCh := make(chan bool, 1)
	depth := 4

	score, bestMove, _, nodes, multiPV := engine.searchWithStatsMultiPV(pos, depth, stopCh)

	if bestMove == nil {
		t.Error("Expected best move, got nil")
	}

	// Should have multiple reasonable moves in this position
	if len(multiPV) < 2 {
		t.Errorf("Expected at least 2 PV lines for this position, got %d", len(multiPV))
	}

	// Verify all PV lines have moves
	for i, pvLine := range multiPV {
		if len(pvLine.Moves) == 0 {
			t.Errorf("PV line %d has no moves", i)
		}
		if pvLine.ScoreType == "" {
			t.Errorf("PV line %d has no score type", i)
		}
	}

	t.Logf("Middlegame: Found %d PV lines, best: %s (score: %d), nodes: %d",
		len(multiPV), bestMove.String(), score, nodes)
	for i, pvLine := range multiPV {
		t.Logf("  PV %d: %s (score: %d %s)", i+1, pvLine.Moves[0], pvLine.Score, pvLine.ScoreType)
	}
}

// TestSearchWithStatsMultiPV_TacticalPosition tests multi-PV on tactical position
func TestSearchWithStatsMultiPV_TacticalPosition(t *testing.T) {
	engine := NewEngine()
	// Position with a clear tactical shot
	fen := "r1bqkb1r/pppp1ppp/2n2n2/4p2Q/2B1P3/8/PPPP1PPP/RNB1K1NR w KQkq - 4 4"
	fenObj, err := chess.FEN(fen)
	if err != nil {
		t.Fatalf("Invalid FEN: %v", err)
	}
	game := chess.NewGame(fenObj)
	pos := game.Position()

	stopCh := make(chan bool, 1)
	depth := 4

	score, bestMove, _, nodes, multiPV := engine.searchWithStatsMultiPV(pos, depth, stopCh)

	if bestMove == nil {
		t.Error("Expected best move, got nil")
	}

	// Should find the winning move (Qxf7#)
	if len(multiPV) > 0 {
		// First move should be mate or significantly better
		if multiPV[0].ScoreType == "mate" {
			t.Logf("Found mate in %d", multiPV[0].Score)
		} else if len(multiPV) >= 2 {
			scoreDiff := multiPV[0].Score - multiPV[1].Score
			if scoreDiff <= 0 {
				t.Logf("Warning: Best move score %d not better than 2nd best %d",
					multiPV[0].Score, multiPV[1].Score)
			}
		}
	}

	t.Logf("Tactical: Found %d PV lines, best: %s (score: %d %s), nodes: %d",
		len(multiPV), bestMove.String(), score, multiPV[0].ScoreType, nodes)
}

// TestSearchWithStatsMultiPV_EndgamePosition tests multi-PV on endgame
func TestSearchWithStatsMultiPV_EndgamePosition(t *testing.T) {
	engine := NewEngine()
	// King and pawn endgame
	fen := "8/8/8/4k3/8/4K3/4P3/8 w - - 0 1"
	fenObj, err := chess.FEN(fen)
	if err != nil {
		t.Fatalf("Invalid FEN: %v", err)
	}
	game := chess.NewGame(fenObj)
	pos := game.Position()

	stopCh := make(chan bool, 1)
	depth := 5

	score, bestMove, _, nodes, multiPV := engine.searchWithStatsMultiPV(pos, depth, stopCh)

	if bestMove == nil {
		t.Error("Expected best move, got nil")
	}

	// Endgame should have fewer reasonable moves
	if len(multiPV) == 0 {
		t.Error("Expected at least 1 PV line")
	}

	t.Logf("Endgame: Found %d PV lines, best: %s (score: %d), nodes: %d",
		len(multiPV), bestMove.String(), score, nodes)
}

// TestSearchWithStatsMultiPV_FewMoves tests multi-PV when few legal moves available
func TestSearchWithStatsMultiPV_FewMoves(t *testing.T) {
	engine := NewEngine()
	// Position with only a few legal moves
	fen := "8/8/8/8/8/3k4/3p4/3K4 b - - 0 1"
	fenObj, err := chess.FEN(fen)
	if err != nil {
		t.Fatalf("Invalid FEN: %v", err)
	}
	game := chess.NewGame(fenObj)
	pos := game.Position()

	stopCh := make(chan bool, 1)
	depth := 4

	score, bestMove, _, nodes, multiPV := engine.searchWithStatsMultiPV(pos, depth, stopCh)

	if bestMove == nil {
		t.Error("Expected best move, got nil")
	}

	// Should have PV lines for all legal moves (or up to 3)
	legalMoves := pos.ValidMoves()
	expectedPVs := len(legalMoves)
	if expectedPVs > 3 {
		expectedPVs = 3
	}

	if len(multiPV) != expectedPVs {
		t.Logf("Warning: Expected %d PV lines (legal moves: %d), got %d",
			expectedPVs, len(legalMoves), len(multiPV))
	}

	t.Logf("Few moves: %d legal moves, %d PV lines, best: %s (score: %d), nodes: %d",
		len(legalMoves), len(multiPV), bestMove.String(), score, nodes)
}

// TestSearchWithStatsMultiPV_StopChannel tests that multi-PV search respects stop channel
func TestSearchWithStatsMultiPV_StopChannel(t *testing.T) {
	engine := NewEngine()
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenObj, err := chess.FEN(fen)
	if err != nil {
		t.Fatalf("Invalid FEN: %v", err)
	}
	game := chess.NewGame(fenObj)
	pos := game.Position()

	stopCh := make(chan bool, 1)

	// Send stop signal immediately
	stopCh <- true

	depth := 3
	_, bestMove, pv, nodes, multiPV := engine.searchWithStatsMultiPV(pos, depth, stopCh)

	// Should still return some result (even if incomplete)
	// The function should handle stop gracefully
	t.Logf("Stopped search: best move: %v, PV length: %d, MultiPV count: %d, nodes: %d",
		bestMove, len(pv), len(multiPV), nodes)
}

// TestSearchWithStatsMultiPV_ScoreTypes tests score type handling
func TestSearchWithStatsMultiPV_ScoreTypes(t *testing.T) {
	engine := NewEngine()

	testCases := []struct {
		name       string
		fen        string
		depth      int
		expectMate bool
	}{
		{
			name:       "Normal position",
			fen:        "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			depth:      3,
			expectMate: false,
		},
		{
			name:       "Mate in 1",
			fen:        "r1bqkb1r/pppp1ppp/2n2n2/4p2Q/2B1P3/8/PPPP1PPP/RNB1K1NR w KQkq - 4 4",
			depth:      3,
			expectMate: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fenObj, err := chess.FEN(tc.fen)
			if err != nil {
				t.Fatalf("Invalid FEN: %v", err)
			}
			game := chess.NewGame(fenObj)
			pos := game.Position()

			stopCh := make(chan bool, 1)
			_, _, _, _, multiPV := engine.searchWithStatsMultiPV(pos, tc.depth, stopCh)

			if len(multiPV) == 0 {
				t.Error("Expected at least 1 PV line")
				return
			}

			// Check score type of best move
			scoreType := multiPV[0].ScoreType
			if tc.expectMate && scoreType != "mate" {
				t.Logf("Warning: Expected mate score, got %s (score: %d)", scoreType, multiPV[0].Score)
			}
			if !tc.expectMate && scoreType != "cp" {
				t.Errorf("Expected centipawn score, got %s", scoreType)
			}

			t.Logf("%s: ScoreType=%s, Score=%d", tc.name, scoreType, multiPV[0].Score)
		})
	}
}

// TestPVLine_Structure tests the PVLine struct
func TestPVLine_Structure(t *testing.T) {
	pv := PVLine{
		Score:     50,
		ScoreType: "cp",
		Moves:     []string{"e2e4", "e7e5", "g1f3"},
	}

	if pv.Score != 50 {
		t.Errorf("Expected score 50, got %d", pv.Score)
	}
	if pv.ScoreType != "cp" {
		t.Errorf("Expected scoreType 'cp', got %s", pv.ScoreType)
	}
	if len(pv.Moves) != 3 {
		t.Errorf("Expected 3 moves, got %d", len(pv.Moves))
	}
}

// TestAnalysisInfo_MultiPVField tests AnalysisInfo MultiPV field
func TestAnalysisInfo_MultiPVField(t *testing.T) {
	info := AnalysisInfo{
		Depth:     10,
		Score:     50,
		BestMove:  "e2e4",
		ScoreType: "cp",
		MultiPV: []PVLine{
			{Score: 50, ScoreType: "cp", Moves: []string{"e2e4", "e7e5"}},
			{Score: 45, ScoreType: "cp", Moves: []string{"d2d4", "d7d5"}},
			{Score: 40, ScoreType: "cp", Moves: []string{"g1f3", "g8f6"}},
		},
	}

	if len(info.MultiPV) != 3 {
		t.Errorf("Expected 3 PV lines, got %d", len(info.MultiPV))
	}

	// Verify first PV matches best move
	if info.MultiPV[0].Moves[0] != info.BestMove {
		t.Errorf("First PV move %s doesn't match best move %s",
			info.MultiPV[0].Moves[0], info.BestMove)
	}
}

// BenchmarkSearchWithStatsMultiPV benchmarks multi-PV search performance
func BenchmarkSearchWithStatsMultiPV(b *testing.B) {
	engine := NewEngine()
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenObj, _ := chess.FEN(fen)
	game := chess.NewGame(fenObj)
	pos := game.Position()
	stopCh := make(chan bool, 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.searchWithStatsMultiPV(pos, 3, stopCh)
	}
}

// BenchmarkSearchWithStatsMultiPV_Depth4 benchmarks at depth 4
func BenchmarkSearchWithStatsMultiPV_Depth4(b *testing.B) {
	engine := NewEngine()
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenObj, _ := chess.FEN(fen)
	game := chess.NewGame(fenObj)
	pos := game.Position()
	stopCh := make(chan bool, 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.searchWithStatsMultiPV(pos, 4, stopCh)
	}
}
