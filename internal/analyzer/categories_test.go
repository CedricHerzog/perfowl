package analyzer

import (
	"testing"

	"github.com/CedricHerzog/perfowl/internal/testutil"
)

func TestAnalyzeCategories_EmptyProfile(t *testing.T) {
	profile := testutil.MinimalProfile()

	result := AnalyzeCategories(profile, "")

	if result.TotalTimeMs != 0 {
		t.Errorf("TotalTimeMs = %v, want 0", result.TotalTimeMs)
	}
	if len(result.Categories) != 0 {
		t.Errorf("expected no categories, got %d", len(result.Categories))
	}
}

func TestAnalyzeCategories_SingleThread(t *testing.T) {
	profile := testutil.ProfileWithCategories()

	result := AnalyzeCategories(profile, "")

	if result.TotalTimeMs <= 0 {
		t.Errorf("expected positive TotalTimeMs, got %f", result.TotalTimeMs)
	}
	if len(result.Categories) == 0 {
		t.Error("expected categories")
	}
}

func TestAnalyzeCategories_FilterByThread(t *testing.T) {
	profile := testutil.ProfileWithCategories()

	// Filter by main thread
	result := AnalyzeCategories(profile, "GeckoMain")

	if result.TotalTimeMs <= 0 {
		t.Errorf("expected positive TotalTimeMs when filtering, got %f", result.TotalTimeMs)
	}
}

func TestAnalyzeCategories_ThreadNotFound(t *testing.T) {
	profile := testutil.ProfileWithCategories()

	result := AnalyzeCategories(profile, "NonexistentThread")

	if result.TotalTimeMs != 0 {
		t.Errorf("expected 0 TotalTimeMs for nonexistent thread, got %f", result.TotalTimeMs)
	}
}

func TestAnalyzeCategories_PercentCalculation(t *testing.T) {
	profile := testutil.ProfileWithCategories()

	result := AnalyzeCategories(profile, "")

	// Percentages should sum to approximately 100
	var totalPercent float64
	for _, cat := range result.Categories {
		totalPercent += cat.Percent
	}

	if totalPercent < 99 || totalPercent > 101 {
		t.Errorf("category percentages sum to %f, expected ~100", totalPercent)
	}
}

func TestAnalyzeCategories_SortedByTime(t *testing.T) {
	profile := testutil.ProfileWithCategories()

	result := AnalyzeCategories(profile, "")

	// Categories should be sorted by time descending
	for i := 1; i < len(result.Categories); i++ {
		if result.Categories[i].TimeMs > result.Categories[i-1].TimeMs {
			t.Errorf("categories not sorted by time: %f > %f",
				result.Categories[i].TimeMs, result.Categories[i-1].TimeMs)
		}
	}
}

func TestAnalyzeCategories_ByThreadBreakdown(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)

	result := AnalyzeCategories(profile, "")

	// Should have breakdown by thread
	if len(result.ByThread) == 0 {
		t.Error("expected ByThread breakdown")
	}
}
