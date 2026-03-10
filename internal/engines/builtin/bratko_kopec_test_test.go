package builtin

import (
	"testing"
	"time"
)

func TestBratkoKopec(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Bratko-Kopec test in short mode")
	}

	engine := NewEngine()

	// Run Bratko-Kopec tests with 5 second time limit per position
	correct, total, failed := RunBratkoKopecTest(engine, 5*time.Second, true)

	percentage := float64(correct) * 100.0 / float64(total)
	rating, category := EstimateBKRating(correct)

	t.Logf("\n")
	t.Logf("=================================")
	t.Logf("Bratko-Kopec Test Results")
	t.Logf("=================================")
	t.Logf("Correct: %d/%d (%.1f%%)", correct, total, percentage)
	t.Logf("Category: %s", category)
	t.Logf("Estimated Rating: %d", rating)
	t.Logf("=================================")

	if len(failed) > 0 {
		t.Logf("\nFailed Tests:")
		for _, fail := range failed {
			t.Logf("  - %s", fail)
		}
	}
}

func TestBratkoKopecQuick(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Bratko-Kopec test in short mode")
	}

	engine := NewEngine()

	// Quick test with 2 second time limit
	correct, total, _ := RunBratkoKopecTest(engine, 2*time.Second, false)

	percentage := float64(correct) * 100.0 / float64(total)
	rating, category := EstimateBKRating(correct)

	t.Logf("Bratko-Kopec Quick: %d/%d (%.1f%%) - %s (%d)", correct, total, percentage, category, rating)
}
