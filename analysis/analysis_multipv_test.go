package analysis

import (
	"testing"
	"time"

	"github.com/notnil/chess"
)

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

// TestPVLine_MateScore tests PVLine with mate score
func TestPVLine_MateScore(t *testing.T) {
	pv := PVLine{
		Score:     3,
		ScoreType: "mate",
		Moves:     []string{"f7f8q"},
	}

	if pv.ScoreType != "mate" {
		t.Errorf("Expected scoreType 'mate', got %s", pv.ScoreType)
	}
	if pv.Score != 3 {
		t.Errorf("Expected mate in 3, got %d", pv.Score)
	}
}

// TestAnalysisInfo_MultiPV tests AnalysisInfo with multiple PV lines
func TestAnalysisInfo_MultiPV(t *testing.T) {
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

	// Verify scores are in descending order
	if info.MultiPV[0].Score < info.MultiPV[1].Score {
		t.Error("PV lines should be sorted by score (descending)")
	}
	if info.MultiPV[1].Score < info.MultiPV[2].Score {
		t.Error("PV lines should be sorted by score (descending)")
	}

	// Verify best move matches first PV line
	if info.BestMove != info.MultiPV[0].Moves[0] {
		t.Errorf("Best move %s should match first PV line move %s", info.BestMove, info.MultiPV[0].Moves[0])
	}
}

// TestAnalysisInfo_SinglePV tests backward compatibility with single PV
func TestAnalysisInfo_SinglePV(t *testing.T) {
	info := AnalysisInfo{
		Depth:     10,
		Score:     50,
		BestMove:  "e2e4",
		PV:        []string{"e2e4", "e7e5", "g1f3"},
		ScoreType: "cp",
		MultiPV: []PVLine{
			{Score: 50, ScoreType: "cp", Moves: []string{"e2e4", "e7e5", "g1f3"}},
		},
	}

	// Verify backward compatibility fields
	if info.BestMove != "e2e4" {
		t.Errorf("Expected best move e2e4, got %s", info.BestMove)
	}
	if len(info.PV) != 3 {
		t.Errorf("Expected PV length 3, got %d", len(info.PV))
	}

	// Verify MultiPV consistency
	if len(info.MultiPV) != 1 {
		t.Errorf("Expected 1 PV line, got %d", len(info.MultiPV))
	}
	if info.MultiPV[0].Moves[0] != info.BestMove {
		t.Error("MultiPV first move should match BestMove")
	}
}

// TestAnalysisInfo_EmptyMultiPV tests AnalysisInfo with no MultiPV
func TestAnalysisInfo_EmptyMultiPV(t *testing.T) {
	info := AnalysisInfo{
		Depth:     5,
		Score:     25,
		BestMove:  "e2e4",
		PV:        []string{"e2e4"},
		ScoreType: "cp",
		MultiPV:   []PVLine{}, // Empty MultiPV
	}

	if len(info.MultiPV) != 0 {
		t.Errorf("Expected empty MultiPV, got %d lines", len(info.MultiPV))
	}

	// Should still have valid single PV data
	if info.BestMove == "" {
		t.Error("BestMove should not be empty")
	}
	if len(info.PV) == 0 {
		t.Error("PV should not be empty")
	}
}

// TestBuiltinAnalysisEngine_MultiPV tests builtin engine multi-PV generation
func TestBuiltinAnalysisEngine_MultiPV(t *testing.T) {
	engine, err := NewBuiltinAnalysisEngine()
	if err != nil {
		t.Fatalf("Failed to create builtin analysis engine: %v", err)
	}
	defer engine.Close()

	analysisChannel := make(chan AnalysisInfo, 10)
	stopCh := make(chan bool, 1)

	// Start analysis on starting position
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	err = engine.StartAnalysis(fen, analysisChannel)
	if err != nil {
		t.Fatalf("Failed to start analysis: %v", err)
	}

	// Wait for at least one analysis result
	var lastInfo AnalysisInfo
	timeout := time.After(3 * time.Second)
	gotResult := false

	for !gotResult {
		select {
		case info := <-analysisChannel:
			lastInfo = info
			// Wait for depth >= 2 to ensure we have meaningful multi-PV
			if info.Depth >= 2 {
				gotResult = true
			}
		case <-timeout:
			t.Fatal("Timeout waiting for analysis result")
		}
	}

	// Stop analysis
	stopCh <- true
	engine.StopAnalysis()

	// Verify multi-PV results
	if len(lastInfo.MultiPV) == 0 {
		t.Error("Expected MultiPV to be populated")
	}

	if len(lastInfo.MultiPV) > 3 {
		t.Errorf("Expected at most 3 PV lines, got %d", len(lastInfo.MultiPV))
	}

	// Verify each PV line has moves
	for i, pv := range lastInfo.MultiPV {
		if len(pv.Moves) == 0 {
			t.Errorf("PV line %d has no moves", i)
		}
		if pv.ScoreType == "" {
			t.Errorf("PV line %d has no score type", i)
		}
	}

	// Verify scores are in descending order
	for i := 0; i < len(lastInfo.MultiPV)-1; i++ {
		if lastInfo.MultiPV[i].Score < lastInfo.MultiPV[i+1].Score {
			t.Errorf("PV lines not sorted: line %d score %d < line %d score %d",
				i, lastInfo.MultiPV[i].Score, i+1, lastInfo.MultiPV[i+1].Score)
		}
	}

	// Verify backward compatibility
	if lastInfo.BestMove == "" {
		t.Error("BestMove should be populated")
	}
	if len(lastInfo.PV) == 0 {
		t.Error("PV should be populated for backward compatibility")
	}
	if len(lastInfo.MultiPV) > 0 && lastInfo.BestMove != lastInfo.MultiPV[0].Moves[0] {
		t.Error("BestMove should match first move of first PV line")
	}
}

// TestBuiltinAnalysisEngine_MultiPV_TacticalPosition tests multi-PV on a tactical position
func TestBuiltinAnalysisEngine_MultiPV_TacticalPosition(t *testing.T) {
	engine, err := NewBuiltinAnalysisEngine()
	if err != nil {
		t.Fatalf("Failed to create builtin analysis engine: %v", err)
	}
	defer engine.Close()

	analysisChannel := make(chan AnalysisInfo, 10)
	stopCh := make(chan bool, 1)

	// Position with clear best move (scholar's mate threat)
	fen := "r1bqkb1r/pppp1ppp/2n2n2/4p2Q/2B1P3/8/PPPP1PPP/RNB1K1NR w KQkq - 4 4"
	err = engine.StartAnalysis(fen, analysisChannel)
	if err != nil {
		t.Fatalf("Failed to start analysis: %v", err)
	}

	// Wait for depth 3 analysis
	var lastInfo AnalysisInfo
	timeout := time.After(5 * time.Second)
	gotResult := false

	for !gotResult {
		select {
		case info := <-analysisChannel:
			lastInfo = info
			if info.Depth >= 3 {
				gotResult = true
			}
		case <-timeout:
			t.Fatal("Timeout waiting for analysis result")
		}
	}

	stopCh <- true
	engine.StopAnalysis()

	// Should have multiple PV lines
	if len(lastInfo.MultiPV) < 2 {
		t.Errorf("Expected at least 2 PV lines for this position, got %d", len(lastInfo.MultiPV))
	}

	// Log the PV lines for debugging
	for i, pv := range lastInfo.MultiPV {
		t.Logf("PV %d: %s (score: %d %s)", i+1, pv.Moves[0], pv.Score, pv.ScoreType)
	}
}

// TestParseCECPAnalysis_MultiPV tests CECP analysis parsing with MultiPV
func TestParseCECPAnalysis_MultiPV(t *testing.T) {
	// This test verifies that CECP engines create single-entry MultiPV
	// We can't easily test the full CECP engine without an actual engine binary,
	// but we can test the parsing function

	// Mock position (starting position)
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fenObj, _ := chess.FEN(fen)
	game := chess.NewGame(fenObj)
	pos := game.Position()

	// Sample CECP output
	line := "5 +25 123 45678 Nf3 Nc6 Nc3 Nf6 e3"

	info := parseCECPAnalysis(line, pos)

	// Verify basic parsing
	if info.Depth != 5 {
		t.Errorf("Expected depth 5, got %d", info.Depth)
	}
	if info.Score != 25 {
		t.Errorf("Expected score 25, got %d", info.Score)
	}

	// Verify MultiPV is created
	if len(info.MultiPV) != 1 {
		t.Errorf("Expected single-entry MultiPV for CECP, got %d entries", len(info.MultiPV))
	}

	// Verify MultiPV content matches main PV
	if len(info.MultiPV) > 0 {
		if info.MultiPV[0].Score != info.Score {
			t.Error("MultiPV score should match main score")
		}
		if info.MultiPV[0].ScoreType != info.ScoreType {
			t.Error("MultiPV scoreType should match main scoreType")
		}
		if len(info.MultiPV[0].Moves) != len(info.PV) {
			t.Error("MultiPV moves should match main PV")
		}
	}
}

// TestMultiPV_ScoreOrdering tests that PV lines are properly ordered by score
func TestMultiPV_ScoreOrdering(t *testing.T) {
	testCases := []struct {
		name     string
		multiPV  []PVLine
		expected bool // true if properly ordered
	}{
		{
			name: "Properly ordered",
			multiPV: []PVLine{
				{Score: 100, ScoreType: "cp", Moves: []string{"e2e4"}},
				{Score: 90, ScoreType: "cp", Moves: []string{"d2d4"}},
				{Score: 80, ScoreType: "cp", Moves: []string{"g1f3"}},
			},
			expected: true,
		},
		{
			name: "Improperly ordered",
			multiPV: []PVLine{
				{Score: 80, ScoreType: "cp", Moves: []string{"e2e4"}},
				{Score: 100, ScoreType: "cp", Moves: []string{"d2d4"}},
				{Score: 90, ScoreType: "cp", Moves: []string{"g1f3"}},
			},
			expected: false,
		},
		{
			name: "Single PV",
			multiPV: []PVLine{
				{Score: 50, ScoreType: "cp", Moves: []string{"e2e4"}},
			},
			expected: true,
		},
		{
			name:     "Empty MultiPV",
			multiPV:  []PVLine{},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isOrdered := true
			for i := 0; i < len(tc.multiPV)-1; i++ {
				if tc.multiPV[i].Score < tc.multiPV[i+1].Score {
					isOrdered = false
					break
				}
			}

			if isOrdered != tc.expected {
				t.Errorf("Expected ordering=%v, got %v", tc.expected, isOrdered)
			}
		})
	}
}

// TestMultiPV_MixedScoreTypes tests MultiPV with mixed score types
func TestMultiPV_MixedScoreTypes(t *testing.T) {
	// In practice, all PV lines should have the same score type,
	// but we test that the structure supports it
	multiPV := []PVLine{
		{Score: 3, ScoreType: "mate", Moves: []string{"f7f8q"}},
		{Score: 500, ScoreType: "cp", Moves: []string{"e2e4"}},
		{Score: 400, ScoreType: "cp", Moves: []string{"d2d4"}},
	}

	if len(multiPV) != 3 {
		t.Errorf("Expected 3 PV lines, got %d", len(multiPV))
	}

	// Verify each line has correct score type
	if multiPV[0].ScoreType != "mate" {
		t.Error("First line should be mate score")
	}
	if multiPV[1].ScoreType != "cp" {
		t.Error("Second line should be centipawn score")
	}
}

// TestBuiltinAnalysisEngine_MultiPV_BlackToMove tests that scores are correctly
// normalized to White's perspective when it's Black's turn.
// All scores should be from White's perspective: positive = White winning, negative = Black winning.
func TestBuiltinAnalysisEngine_MultiPV_BlackToMove(t *testing.T) {
	engine, err := NewBuiltinAnalysisEngine()
	if err != nil {
		t.Fatalf("Failed to create builtin analysis engine: %v", err)
	}
	defer engine.Close()

	analysisChannel := make(chan AnalysisInfo, 10)

	// Position where Black is to move
	// After 1.e4 e5 2.Nf3 - Black to move
	fen := "rnbqkbnr/pppp1ppp/8/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2"
	err = engine.StartAnalysis(fen, analysisChannel)
	if err != nil {
		t.Fatalf("Failed to start analysis: %v", err)
	}

	// Wait for analysis results
	var lastInfo AnalysisInfo
	timeout := time.After(3 * time.Second)
	gotResult := false

	for !gotResult {
		select {
		case info := <-analysisChannel:
			lastInfo = info
			if info.Depth >= 3 && len(info.MultiPV) >= 2 {
				gotResult = true
			}
		case <-timeout:
			gotResult = true
		}
	}

	engine.StopAnalysis()

	if len(lastInfo.MultiPV) < 2 {
		t.Skipf("Not enough PV lines to test (got %d)", len(lastInfo.MultiPV))
	}

	// Scores should be normalized to White's perspective and sorted descending
	// (best move for the side to move should still be first, but scores are from White's view)
	// For this roughly equal position, scores should be close to 0 or slightly negative
	// (since Black has a slight initiative after e5)
	for i := 0; i < len(lastInfo.MultiPV)-1; i++ {
		// Scores should be in descending order (best for current player first)
		// Since we normalize to White's perspective, for Black's turn the best move
		// (most negative from White's view) should be first
		if lastInfo.MultiPV[i].Score > lastInfo.MultiPV[i+1].Score {
			t.Errorf("MultiPV not sorted correctly: PV %d score %d > PV %d score %d (should be <=)",
				i+1, lastInfo.MultiPV[i].Score, i+2, lastInfo.MultiPV[i+1].Score)
		}
	}

	t.Logf("Black to move: Found %d PV lines (scores normalized to White's perspective)", len(lastInfo.MultiPV))
	for i, pv := range lastInfo.MultiPV {
		t.Logf("  PV %d: %s (score: %d %s)", i+1, pv.Moves[0], pv.Score, pv.ScoreType)
	}
}
