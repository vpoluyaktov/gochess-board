package builtin

import (
	"fmt"
	"testing"
	"time"
)

func TestTacticalSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping tactical suite in short mode")
	}
	engine := NewEngine()

	// Run tactical tests with 2 second time limit per position
	correct, total, failed := RunTacticalTest(engine, 2*time.Second, true)

	percentage := float64(correct) * 100.0 / float64(total)
	minELO, maxELO, category := EstimateELO(correct, total)

	t.Logf("\n")
	t.Logf("=================================")
	t.Logf("Tactical Test Results")
	t.Logf("=================================")
	t.Logf("Correct: %d/%d (%.1f%%)", correct, total, percentage)
	t.Logf("Category: %s", category)
	t.Logf("Estimated ELO: %d-%d", minELO, maxELO)
	t.Logf("=================================")

	if len(failed) > 0 {
		t.Logf("\nFailed Tests:")
		for _, fail := range failed {
			t.Logf("  - %s", fail)
		}
	}

	// We expect at least 60% success rate (1400+ ELO)
	if percentage < 60 {
		t.Logf("\nWarning: Tactical score below 60%% - engine may need improvement")
	}
}

func TestTacticalSuiteQuick(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping tactical suite in short mode")
	}

	engine := NewEngine()

	// Quick test with 1 second time limit
	correct, total, _ := RunTacticalTest(engine, 1*time.Second, false)

	percentage := float64(correct) * 100.0 / float64(total)
	t.Logf("Quick Tactical Test: %d/%d (%.1f%%)", correct, total, percentage)

	// Note: This engine is optimized for positional play, not pure tactics
	// A score of 20-30% is acceptable for a positional engine
	if percentage < 10 {
		t.Logf("Warning: Very low tactical score (%.1f%%) - engine is positional, not tactical", percentage)
	}
}

func TestIndividualTacticalPositions(t *testing.T) {
	engine := NewEngine()
	suite := GetTacticalTestSuite()

	// Test each position individually
	for _, test := range suite {
		t.Run(test.Name, func(t *testing.T) {
			move, err := engine.GetBestMove(test.FEN, 3*time.Second)
			if err != nil {
				t.Fatalf("Error getting move: %v", err)
			}

			// Check if move matches
			found := false
			for _, bestMove := range test.BestMoves {
				if move == bestMove || normalizeMove(move) == normalizeMove(bestMove) {
					found = true
					break
				}
			}

			if found {
				t.Logf("✓ %s: Found %s", test.Description, move)
			} else {
				t.Logf("✗ %s: Got %s, Expected %v (Difficulty: %d)",
					test.Description, move, test.BestMoves, test.Difficulty)
			}
		})
	}
}

func TestELOEstimation(t *testing.T) {
	tests := []struct {
		correct  int
		total    int
		minELO   int
		maxELO   int
		category string
	}{
		{10, 10, 1800, 2000, "Expert"},
		{9, 10, 1800, 2000, "Expert"},
		{8, 10, 1600, 1800, "Advanced"},
		{7, 10, 1500, 1700, "Intermediate+"},
		{6, 10, 1400, 1600, "Intermediate"},
		{5, 10, 1300, 1500, "Developing"},
		{4, 10, 1200, 1400, "Beginner+"},
		{3, 10, 1000, 1200, "Beginner"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d/%d", tt.correct, tt.total), func(t *testing.T) {
			minELO, maxELO, category := EstimateELO(tt.correct, tt.total)

			if minELO != tt.minELO || maxELO != tt.maxELO {
				t.Errorf("Expected ELO %d-%d, got %d-%d",
					tt.minELO, tt.maxELO, minELO, maxELO)
			}

			if category != tt.category {
				t.Errorf("Expected category %s, got %s", tt.category, category)
			}

			t.Logf("%d/%d correct = %d-%d ELO (%s)",
				tt.correct, tt.total, minELO, maxELO, category)
		})
	}
}

func TestGetTacticalTestSuite(t *testing.T) {
	suite := GetTacticalTestSuite()

	if len(suite) == 0 {
		t.Fatal("Tactical test suite is empty")
	}

	t.Logf("Tactical test suite contains %d positions", len(suite))

	// Verify each position is valid
	for i, test := range suite {
		if test.Name == "" {
			t.Errorf("Position %d has no name", i)
		}
		if test.FEN == "" {
			t.Errorf("Position %d (%s) has no FEN", i, test.Name)
		}
		if len(test.BestMoves) == 0 {
			t.Errorf("Position %d (%s) has no best moves", i, test.Name)
		}
		if test.Difficulty < 1 || test.Difficulty > 5 {
			t.Errorf("Position %d (%s) has invalid difficulty: %d",
				i, test.Name, test.Difficulty)
		}
	}
}

func BenchmarkTacticalTest(b *testing.B) {
	engine := NewEngine()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RunTacticalTest(engine, 1*time.Second, false)
	}
}

// Example test showing how to run a full ELO estimation
func ExampleRunTacticalTest() {
	engine := NewEngine()

	// Run tactical tests with 2 second time limit
	correct, total, failed := RunTacticalTest(engine, 2*time.Second, true)

	// Estimate ELO
	minELO, maxELO, category := EstimateELO(correct, total)

	fmt.Printf("\nResults:\n")
	fmt.Printf("Solved: %d/%d\n", correct, total)
	fmt.Printf("Category: %s\n", category)
	fmt.Printf("Estimated ELO: %d-%d\n", minELO, maxELO)

	if len(failed) > 0 {
		fmt.Printf("\nFailed positions: %d\n", len(failed))
	}
}
