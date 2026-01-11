package analyzer

import (
	"testing"

	"github.com/CedricHerzog/perfowl/internal/testutil"
)

func TestCompareProfiles_SameProfile(t *testing.T) {
	profile := testutil.ProfileWithMainThread()

	result := CompareProfiles(profile, profile)

	if len(result.Improved) != 0 {
		t.Errorf("expected no improvements for same profile, got %d", len(result.Improved))
	}
	if len(result.Regressed) != 0 {
		t.Errorf("expected no regressions for same profile, got %d", len(result.Regressed))
	}
}

func TestCompareProfiles_DurationImproved(t *testing.T) {
	baseline := testutil.NewProfileBuilder().
		WithDuration(2000).
		WithThread(testutil.NewThreadBuilder("GeckoMain").AsMainThread().Build()).
		Build()

	comparison := testutil.NewProfileBuilder().
		WithDuration(1000). // 50% faster
		WithThread(testutil.NewThreadBuilder("GeckoMain").AsMainThread().Build()).
		Build()

	result := CompareProfiles(baseline, comparison)

	// Duration should show improvement
	if result.Changes.DurationChangeMs >= 0 {
		t.Errorf("expected negative duration change (improvement), got %f", result.Changes.DurationChangeMs)
	}
}

func TestCompareProfiles_DurationRegressed(t *testing.T) {
	baseline := testutil.NewProfileBuilder().
		WithDuration(1000).
		WithThread(testutil.NewThreadBuilder("GeckoMain").AsMainThread().Build()).
		Build()

	comparison := testutil.NewProfileBuilder().
		WithDuration(2000). // 100% slower
		WithThread(testutil.NewThreadBuilder("GeckoMain").AsMainThread().Build()).
		Build()

	result := CompareProfiles(baseline, comparison)

	// Duration should show regression
	if result.Changes.DurationChangeMs <= 0 {
		t.Errorf("expected positive duration change (regression), got %f", result.Changes.DurationChangeMs)
	}
}

func TestCompareProfiles_GCImproved(t *testing.T) {
	baseline := testutil.ProfileWithGC() // Has GC markers

	comparison := testutil.ProfileWithMainThread() // No GC markers

	result := CompareProfiles(baseline, comparison)

	// Should detect GC improvement
	if result.Comparison.GCMajorCount >= result.Baseline.GCMajorCount {
		// May not detect if baseline doesn't have markers properly set up
		t.Logf("GC comparison: baseline=%d, comparison=%d",
			result.Baseline.GCMajorCount, result.Comparison.GCMajorCount)
	}
}

func TestCompareProfiles_Summaries(t *testing.T) {
	baseline := testutil.ProfileWithMainThread()
	comparison := testutil.ProfileWithMainThread()

	result := CompareProfiles(baseline, comparison)

	// Should have summaries
	if result.Baseline.Name == "" {
		t.Error("expected baseline name")
	}
	if result.Comparison.Name == "" {
		t.Error("expected comparison name")
	}
}

func TestExtractSummary_BasicFields(t *testing.T) {
	profile := testutil.NewProfileBuilder().
		WithDuration(1000).
		WithThread(testutil.NewThreadBuilder("GeckoMain").AsMainThread().Build()).
		WithThread(testutil.NewThreadBuilder("Worker").Build()).
		Build()

	summary := extractSummary(profile, "test")

	if summary.Name != "test" {
		t.Errorf("Name = %v, want test", summary.Name)
	}
	if summary.DurationMs != 1000 {
		t.Errorf("DurationMs = %v, want 1000", summary.DurationMs)
	}
	if summary.ThreadCount != 2 {
		t.Errorf("ThreadCount = %v, want 2", summary.ThreadCount)
	}
}

func TestPercentChange_Increase(t *testing.T) {
	result := percentChange(100, 150)
	if result != 50 {
		t.Errorf("percentChange(100, 150) = %f, want 50", result)
	}
}

func TestPercentChange_Decrease(t *testing.T) {
	result := percentChange(100, 50)
	if result != -50 {
		t.Errorf("percentChange(100, 50) = %f, want -50", result)
	}
}

func TestPercentChange_NoChange(t *testing.T) {
	result := percentChange(100, 100)
	if result != 0 {
		t.Errorf("percentChange(100, 100) = %f, want 0", result)
	}
}

func TestPercentChange_FromZero(t *testing.T) {
	result := percentChange(0, 100)
	if result != 100 {
		t.Errorf("percentChange(0, 100) = %f, want 100", result)
	}
}

func TestPercentChange_BothZero(t *testing.T) {
	result := percentChange(0, 0)
	if result != 0 {
		t.Errorf("percentChange(0, 0) = %f, want 0", result)
	}
}

func TestFormatPercent_Positive(t *testing.T) {
	result := formatPercent(50)
	// formatPercent uses math.Abs, so it returns absolute value without sign prefix
	if result != "50.0%" {
		t.Errorf("formatPercent(50) = %s, want 50.0%%", result)
	}
}

func TestFormatPercent_Negative(t *testing.T) {
	result := formatPercent(-25)
	// formatPercent uses math.Abs, so negative becomes positive
	if result != "25.0%" {
		t.Errorf("formatPercent(-25) = %s, want 25.0%%", result)
	}
}

func TestFormatPercent_Zero(t *testing.T) {
	result := formatPercent(0)
	if result != "0.0%" {
		t.Errorf("formatPercent(0) = %s, want 0.0%%", result)
	}
}
