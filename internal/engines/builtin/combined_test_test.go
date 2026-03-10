package builtin

import (
	"testing"
	"time"
)

// TestAllSuites runs both the custom tactical test suite and the Bratko-Kopec test suite
func TestAllSuites(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping combined test suites in short mode")
	}

	engine := NewEngine()

	t.Log("=================================================================")
	t.Log("                    COMBINED TEST SUITES")
	t.Log("=================================================================")
	t.Log("")

	// Run Custom Tactical Test Suite
	t.Log("-----------------------------------------------------------------")
	t.Log("                  CUSTOM TACTICAL TEST SUITE")
	t.Log("-----------------------------------------------------------------")
	tacticalCorrect, tacticalTotal, tacticalFailed := RunTacticalTest(engine, 2*time.Second, true)
	tacticalPct := float64(tacticalCorrect) * 100.0 / float64(tacticalTotal)
	_, _, tacticalCategory := EstimateELO(tacticalCorrect, tacticalTotal)

	t.Log("")

	// Run Bratko-Kopec Test Suite
	t.Log("-----------------------------------------------------------------")
	t.Log("                  BRATKO-KOPEC TEST SUITE")
	t.Log("-----------------------------------------------------------------")
	bkCorrect, bkTotal, bkFailed := RunBratkoKopecTest(engine, 3*time.Second, true)
	bkPct := float64(bkCorrect) * 100.0 / float64(bkTotal)
	bkRating, bkCategory := EstimateBKRating(bkCorrect)

	// Combined Summary
	totalCorrect := tacticalCorrect + bkCorrect
	totalPositions := tacticalTotal + bkTotal
	totalPct := float64(totalCorrect) * 100.0 / float64(totalPositions)

	t.Log("")
	t.Log("=================================================================")
	t.Log("                       COMBINED RESULTS")
	t.Log("=================================================================")
	t.Logf("")
	t.Logf("Custom Tactical Suite:")
	t.Logf("  Score: %d/%d (%.1f%%)", tacticalCorrect, tacticalTotal, tacticalPct)
	t.Logf("  Category: %s", tacticalCategory)
	t.Logf("")
	t.Logf("Bratko-Kopec Suite:")
	t.Logf("  Score: %d/%d (%.1f%%)", bkCorrect, bkTotal, bkPct)
	t.Logf("  Category: %s (Rating: %d)", bkCategory, bkRating)
	t.Logf("")
	t.Logf("TOTAL: %d/%d (%.1f%%)", totalCorrect, totalPositions, totalPct)
	t.Log("=================================================================")

	if len(tacticalFailed) > 0 || len(bkFailed) > 0 {
		t.Log("")
		t.Log("Failed Tests Summary:")
		if len(tacticalFailed) > 0 {
			t.Log("  Tactical:")
			for _, fail := range tacticalFailed {
				t.Logf("    - %s", fail)
			}
		}
		if len(bkFailed) > 0 {
			t.Log("  Bratko-Kopec:")
			for _, fail := range bkFailed {
				t.Logf("    - %s", fail)
			}
		}
	}
}

// TestAllSuitesQuick runs a quick version of both test suites with shorter time limits
func TestAllSuitesQuick(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping combined test suites in short mode")
	}

	engine := NewEngine()

	// Run Custom Tactical Test Suite (1 second per position)
	tacticalCorrect, tacticalTotal, _ := RunTacticalTest(engine, 1*time.Second, false)
	tacticalPct := float64(tacticalCorrect) * 100.0 / float64(tacticalTotal)

	// Run Bratko-Kopec Test Suite (2 seconds per position)
	bkCorrect, bkTotal, _ := RunBratkoKopecTest(engine, 2*time.Second, false)
	bkPct := float64(bkCorrect) * 100.0 / float64(bkTotal)
	bkRating, bkCategory := EstimateBKRating(bkCorrect)

	// Combined Summary
	totalCorrect := tacticalCorrect + bkCorrect
	totalPositions := tacticalTotal + bkTotal
	totalPct := float64(totalCorrect) * 100.0 / float64(totalPositions)

	t.Log("")
	t.Log("=== QUICK TEST RESULTS ===")
	t.Logf("Tactical:     %d/%d (%.1f%%)", tacticalCorrect, tacticalTotal, tacticalPct)
	t.Logf("Bratko-Kopec: %d/%d (%.1f%%) - %s (%d)", bkCorrect, bkTotal, bkPct, bkCategory, bkRating)
	t.Logf("TOTAL:        %d/%d (%.1f%%)", totalCorrect, totalPositions, totalPct)
}
