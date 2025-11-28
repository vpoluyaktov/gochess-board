package analysis

import (
	"testing"
	"time"
)

func TestBuiltinAnalysisEngine(t *testing.T) {
	// Create the built-in analysis engine
	engine, err := NewBuiltinAnalysisEngine()
	if err != nil {
		t.Fatalf("Failed to create built-in analysis engine: %v", err)
	}
	defer engine.Close()

	// Create a channel to receive analysis info
	analysisChannel := make(chan AnalysisInfo, 10)

	// Start analysis on the starting position
	startFEN := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	err = engine.StartAnalysis(startFEN, analysisChannel)
	if err != nil {
		t.Fatalf("Failed to start analysis: %v", err)
	}

	// Collect analysis results for 2 seconds
	timeout := time.After(2 * time.Second)
	var lastInfo AnalysisInfo
	receivedCount := 0

	for {
		select {
		case info := <-analysisChannel:
			receivedCount++
			lastInfo = info
			t.Logf("Received analysis: depth=%d, score=%d (%s), move=%s, nodes=%d, nps=%d",
				info.Depth, info.Score, info.ScoreType, info.BestMove, info.Nodes, info.NPS)
		case <-timeout:
			goto done
		}
	}

done:
	// Stop analysis
	engine.StopAnalysis()

	// Verify we received some analysis
	if receivedCount == 0 {
		t.Fatal("Did not receive any analysis info")
	}

	// Verify the last analysis info has valid data
	if lastInfo.BestMove == "" {
		t.Error("Last analysis info has no best move")
	}
	if lastInfo.Depth == 0 {
		t.Error("Last analysis info has depth 0")
	}
	if lastInfo.Nodes == 0 {
		t.Error("Last analysis info has 0 nodes")
	}

	t.Logf("Test completed successfully. Received %d analysis updates", receivedCount)
	t.Logf("Final analysis: depth=%d, score=%d (%s), move=%s, nodes=%d",
		lastInfo.Depth, lastInfo.Score, lastInfo.ScoreType, lastInfo.BestMove, lastInfo.Nodes)
}

func TestBuiltinAnalysisEngineStop(t *testing.T) {
	// Create the built-in analysis engine
	engine, err := NewBuiltinAnalysisEngine()
	if err != nil {
		t.Fatalf("Failed to create built-in analysis engine: %v", err)
	}
	defer engine.Close()

	// Create a channel to receive analysis info
	analysisChannel := make(chan AnalysisInfo, 10)

	// Start analysis
	startFEN := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	err = engine.StartAnalysis(startFEN, analysisChannel)
	if err != nil {
		t.Fatalf("Failed to start analysis: %v", err)
	}

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Stop analysis
	engine.StopAnalysis()

	// Verify we can stop without errors
	t.Log("Analysis stopped successfully")
}

func TestBuiltinAnalysisEngineMultiplePositions(t *testing.T) {
	// Create the built-in analysis engine
	engine, err := NewBuiltinAnalysisEngine()
	if err != nil {
		t.Fatalf("Failed to create built-in analysis engine: %v", err)
	}
	defer engine.Close()

	positions := []string{
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",         // Starting position
		"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",      // After 1.e4
		"r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3", // After 1.e4 e5 2.Nf3 Nc6
	}

	for i, fen := range positions {
		t.Logf("Testing position %d: %s", i+1, fen)

		analysisChannel := make(chan AnalysisInfo, 10)
		err = engine.StartAnalysis(fen, analysisChannel)
		if err != nil {
			t.Fatalf("Failed to start analysis for position %d: %v", i+1, err)
		}

		// Wait for at least one result
		timeout := time.After(500 * time.Millisecond)
		select {
		case info := <-analysisChannel:
			t.Logf("Position %d analysis: depth=%d, move=%s, score=%d",
				i+1, info.Depth, info.BestMove, info.Score)
		case <-timeout:
			t.Errorf("Timeout waiting for analysis on position %d", i+1)
		}

		engine.StopAnalysis()
		time.Sleep(50 * time.Millisecond)
	}
}
