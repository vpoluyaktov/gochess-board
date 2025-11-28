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
