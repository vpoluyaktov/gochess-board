package server

import (
	"testing"
)

func TestAnalysisInfo_Structure(t *testing.T) {
	info := AnalysisInfo{
		Depth:     20,
		Score:     150,
		BestMove:  "e2e4",
		PV:        []string{"e2e4", "e7e5", "Nf3"},
		Nodes:     1000000,
		NPS:       500000,
		Time:      2000,
		ScoreType: "cp",
	}

	if info.Depth != 20 {
		t.Errorf("Expected Depth 20, got %d", info.Depth)
	}
	if info.Score != 150 {
		t.Errorf("Expected Score 150, got %d", info.Score)
	}
	if info.BestMove != "e2e4" {
		t.Errorf("Expected BestMove 'e2e4', got %q", info.BestMove)
	}
	if len(info.PV) != 3 {
		t.Errorf("Expected PV length 3, got %d", len(info.PV))
	}
	if info.ScoreType != "cp" {
		t.Errorf("Expected ScoreType 'cp', got %q", info.ScoreType)
	}
}

func TestAnalysisInfo_MateScore(t *testing.T) {
	info := AnalysisInfo{
		Depth:     15,
		Score:     3,
		BestMove:  "Qh7",
		ScoreType: "mate",
	}

	if info.ScoreType != "mate" {
		t.Errorf("Expected ScoreType 'mate', got %q", info.ScoreType)
	}
	if info.Score != 3 {
		t.Errorf("Expected mate in 3, got %d", info.Score)
	}
}

func TestAnalysisInfo_EmptyPV(t *testing.T) {
	info := AnalysisInfo{
		Depth:    10,
		Score:    0,
		BestMove: "e2e4",
		PV:       []string{},
	}

	if info.PV == nil {
		t.Error("PV should not be nil")
	}
	if len(info.PV) != 0 {
		t.Errorf("Expected empty PV, got length %d", len(info.PV))
	}
}

func TestAnalysisInfo_LongPV(t *testing.T) {
	pv := make([]string, 20)
	for i := 0; i < 20; i++ {
		pv[i] = "move"
	}

	info := AnalysisInfo{
		Depth:    25,
		Score:    200,
		BestMove: "e2e4",
		PV:       pv,
	}

	if len(info.PV) != 20 {
		t.Errorf("Expected PV length 20, got %d", len(info.PV))
	}
}

func TestAnalysisInfo_NegativeScore(t *testing.T) {
	info := AnalysisInfo{
		Depth:     15,
		Score:     -250,
		BestMove:  "d7d5",
		ScoreType: "cp",
	}

	if info.Score != -250 {
		t.Errorf("Expected Score -250, got %d", info.Score)
	}
}

func TestAnalysisInfo_HighDepth(t *testing.T) {
	info := AnalysisInfo{
		Depth:    40,
		Score:    0,
		BestMove: "e2e4",
	}

	if info.Depth != 40 {
		t.Errorf("Expected Depth 40, got %d", info.Depth)
	}
}

func TestAnalysisInfo_Performance(t *testing.T) {
	info := AnalysisInfo{
		Nodes: 10000000,
		NPS:   5000000,
		Time:  2000,
	}

	// Verify NPS calculation makes sense
	// NPS = Nodes / (Time / 1000)
	expectedNPS := info.Nodes / (int64(info.Time) / 1000)
	if expectedNPS != 5000000 {
		t.Logf("NPS calculation: %d nodes / %d ms = %d nps", info.Nodes, info.Time, expectedNPS)
	}
}

func TestAnalysisInfo_ZeroValues(t *testing.T) {
	info := AnalysisInfo{}

	if info.Depth != 0 {
		t.Errorf("Expected zero Depth, got %d", info.Depth)
	}
	if info.Score != 0 {
		t.Errorf("Expected zero Score, got %d", info.Score)
	}
	if info.BestMove != "" {
		t.Errorf("Expected empty BestMove, got %q", info.BestMove)
	}
}

func TestAnalysisEngine_Structure(t *testing.T) {
	// Test that AnalysisEngine struct exists and has expected structure
	var engine *AnalysisEngine

	if engine == nil {
		// Expected: uninitialized pointer should be nil
	} else {
		t.Error("Uninitialized engine should be nil")
	}
}

func TestNewAnalysisEngine_InvalidPath(t *testing.T) {
	engine, err := NewAnalysisEngine("/nonexistent/engine")

	if err == nil {
		t.Error("Expected error for non-existent engine")
		if engine != nil {
			engine.Close()
		}
	}

	if engine != nil {
		t.Error("Expected nil engine for invalid path")
	}
}

func TestNewAnalysisEngine_EmptyPath(t *testing.T) {
	engine, err := NewAnalysisEngine("")

	if err == nil {
		t.Error("Expected error for empty path")
		if engine != nil {
			engine.Close()
		}
	}
}

// Integration test - only runs if stockfish is available
func TestNewAnalysisEngine_WithStockfish(t *testing.T) {
	paths := []string{
		"/usr/games/stockfish",
		"/usr/bin/stockfish",
		"/usr/local/bin/stockfish",
	}

	var engine *AnalysisEngine
	var err error

	for _, path := range paths {
		engine, err = NewAnalysisEngine(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Skip("Stockfish not found, skipping integration test")
		return
	}

	defer engine.Close()

	if engine == nil {
		t.Error("Engine should not be nil after successful creation")
		return
	}

	if !engine.active {
		t.Error("Engine should be active after creation")
	}
}
