package builtin

import (
	"testing"
	"time"
)

func TestInternalEngineAnalyze(t *testing.T) {
	engine := NewEngine()

	// Create channels
	stopCh := make(chan bool)
	resultCh := make(chan AnalysisInfo, 10)

	// Start analysis in a goroutine
	go func() {
		startFEN := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
		err := engine.Analyze(startFEN, 4, stopCh, resultCh)
		if err != nil {
			t.Errorf("Analyze failed: %v", err)
		}
		close(resultCh)
	}()

	// Collect results
	var results []AnalysisInfo
	timeout := time.After(3 * time.Second)

	for {
		select {
		case info, ok := <-resultCh:
			if !ok {
				goto done
			}
			results = append(results, info)
			t.Logf("Depth %d: move=%s, score=%d (%s), nodes=%d, nps=%d",
				info.Depth, info.BestMove, info.Score, info.ScoreType, info.Nodes, info.NPS)
		case <-timeout:
			t.Fatal("Timeout waiting for analysis")
		}
	}

done:
	// Verify we got results for all depths
	if len(results) != 4 {
		t.Errorf("Expected 4 depth results, got %d", len(results))
	}

	// Verify depths are sequential
	for i, info := range results {
		expectedDepth := i + 1
		if info.Depth != expectedDepth {
			t.Errorf("Result %d: expected depth %d, got %d", i, expectedDepth, info.Depth)
		}
		if info.BestMove == "" {
			t.Errorf("Result %d: no best move", i)
		}
		if info.Nodes == 0 {
			t.Errorf("Result %d: zero nodes", i)
		}
	}
}

func TestInternalEngineAnalyzeStop(t *testing.T) {
	engine := NewEngine()

	stopCh := make(chan bool)
	resultCh := make(chan AnalysisInfo, 10)

	// Start analysis
	go func() {
		startFEN := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
		err := engine.Analyze(startFEN, 10, stopCh, resultCh)
		if err != nil {
			t.Errorf("Analyze failed: %v", err)
		}
		close(resultCh)
	}()

	// Wait a bit then stop
	time.Sleep(100 * time.Millisecond)
	close(stopCh)

	// Drain results
	count := 0
	for range resultCh {
		count++
	}

	// Should have stopped before reaching depth 10
	if count >= 10 {
		t.Errorf("Expected analysis to stop early, but got %d results", count)
	}

	t.Logf("Analysis stopped after %d depths (expected < 10)", count)
}

func TestInternalEngineAnalyzeInvalidFEN(t *testing.T) {
	engine := NewEngine()

	stopCh := make(chan bool)
	resultCh := make(chan AnalysisInfo, 10)

	invalidFEN := "invalid fen string"
	err := engine.Analyze(invalidFEN, 4, stopCh, resultCh)

	if err == nil {
		t.Error("Expected error for invalid FEN, got nil")
	}
}

func TestInternalEngineAnalyzePV(t *testing.T) {
	engine := NewEngine()

	stopCh := make(chan bool)
	resultCh := make(chan AnalysisInfo, 10)

	// Start analysis
	go func() {
		startFEN := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
		err := engine.Analyze(startFEN, 4, stopCh, resultCh)
		if err != nil {
			t.Errorf("Analyze failed: %v", err)
		}
		close(resultCh)
	}()

	// Collect results and verify PV
	var lastInfo AnalysisInfo
	for info := range resultCh {
		lastInfo = info

		// PV should not be empty
		if len(info.PV) == 0 {
			t.Errorf("Depth %d: PV is empty", info.Depth)
		}

		// PV should start with the best move
		if len(info.PV) > 0 && info.PV[0] != info.BestMove {
			t.Errorf("Depth %d: PV[0]=%s doesn't match BestMove=%s",
				info.Depth, info.PV[0], info.BestMove)
		}

		// PV length should be reasonable (at most depth moves)
		if len(info.PV) > info.Depth {
			t.Errorf("Depth %d: PV has %d moves, expected at most %d",
				info.Depth, len(info.PV), info.Depth)
		}

		t.Logf("Depth %d: PV = %v (length: %d)", info.Depth, info.PV, len(info.PV))
	}

	// Verify we got at least depth 4
	if lastInfo.Depth < 4 {
		t.Errorf("Expected to reach depth 4, got %d", lastInfo.Depth)
	}

	// At depth 4, we should have a PV with multiple moves
	if len(lastInfo.PV) < 2 {
		t.Errorf("Depth %d: Expected PV with at least 2 moves, got %d",
			lastInfo.Depth, len(lastInfo.PV))
	}
}
