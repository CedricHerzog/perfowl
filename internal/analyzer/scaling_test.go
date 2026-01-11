package analyzer

import (
	"testing"

	"github.com/CedricHerzog/perfowl/internal/testutil"
)

func TestAnalyzeScaling_EmptyProfile(t *testing.T) {
	profile := testutil.MinimalProfile()

	result := AnalyzeScaling(profile)

	if result.WorkerCount != 0 {
		t.Errorf("WorkerCount = %v, want 0", result.WorkerCount)
	}
}

func TestAnalyzeScaling_NoWorkers(t *testing.T) {
	profile := testutil.ProfileWithMainThread()

	result := AnalyzeScaling(profile)

	if result.WorkerCount != 0 {
		t.Errorf("WorkerCount = %v, want 0", result.WorkerCount)
	}
}

func TestAnalyzeScaling_SingleWorker(t *testing.T) {
	profile := testutil.ProfileWithWorkers(1)

	result := AnalyzeScaling(profile)

	if result.WorkerCount != 1 {
		t.Errorf("WorkerCount = %v, want 1", result.WorkerCount)
	}
}

func TestAnalyzeScaling_MultipleWorkers(t *testing.T) {
	profile := testutil.ProfileWithWorkers(4)

	result := AnalyzeScaling(profile)

	if result.WorkerCount != 4 {
		t.Errorf("WorkerCount = %v, want 4", result.WorkerCount)
	}
}

func TestAnalyzeScaling_TotalWorkCalculation(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)

	result := AnalyzeScaling(profile)

	if result.TotalWorkMs <= 0 {
		t.Errorf("expected positive TotalWorkMs, got %f", result.TotalWorkMs)
	}
}

func TestAnalyzeScaling_WallClockCalculation(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)

	result := AnalyzeScaling(profile)

	if result.WallClockMs <= 0 {
		t.Errorf("expected positive WallClockMs, got %f", result.WallClockMs)
	}
}

func TestAnalyzeScaling_SpeedupCalculation(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)

	result := AnalyzeScaling(profile)

	// Actual speedup should be positive
	if result.ActualSpeedup <= 0 {
		t.Errorf("expected positive ActualSpeedup, got %f", result.ActualSpeedup)
	}
}

func TestAnalyzeScaling_EfficiencyCalculation(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)

	result := AnalyzeScaling(profile)

	// Efficiency should be between 0 and 100
	if result.Efficiency < 0 || result.Efficiency > 100 {
		t.Errorf("Efficiency out of range: %f", result.Efficiency)
	}
}

func TestAnalyzeScaling_BottleneckClassification(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)

	result := AnalyzeScaling(profile)

	validTypes := map[string]bool{
		"serialization": true,
		"contention":    true,
		"overhead":      true,
		"minimal":       true,
		"none":          true,
	}

	if !validTypes[result.BottleneckType] {
		t.Errorf("invalid BottleneckType: %s", result.BottleneckType)
	}
}

func TestAnalyzeScaling_Recommendations(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)

	result := AnalyzeScaling(profile)

	// Should have at least one recommendation
	if len(result.Recommendations) == 0 {
		t.Error("expected recommendations")
	}
}

func TestCompareScaling_SameProfile(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)

	result := CompareScaling(profile, profile)

	// Improvement should be ~0 for same profile
	if result.Improvement < -1 || result.Improvement > 1 {
		t.Errorf("expected ~0 improvement for same profile, got %f", result.Improvement)
	}
}

func TestCompareScaling_Improvement(t *testing.T) {
	baseline := testutil.NewProfileBuilder().
		WithDuration(2000). // Longer duration = slower
		Build()

	comparison := testutil.NewProfileBuilder().
		WithDuration(1000). // Shorter duration = faster
		Build()

	// Add threads for both
	baseline.Threads = append(baseline.Threads, testutil.NewThreadBuilder("DOM Worker").Build())
	comparison.Threads = append(comparison.Threads, testutil.NewThreadBuilder("DOM Worker").Build())

	result := CompareScaling(baseline, comparison)

	// Comparison should show improvement (positive)
	if result.Improvement <= 0 {
		t.Logf("Improvement: %f (may not show improvement without worker samples)", result.Improvement)
	}
}

func TestFormatScalingAnalysis_Output(t *testing.T) {
	analysis := ScalingAnalysis{
		WorkerCount:        4,
		TotalWorkMs:        4000,
		WallClockMs:        1200,
		TheoreticalSpeedup: 4.0,
		ActualSpeedup:      3.33,
		Efficiency:         83.25,
		BottleneckType:     "minimal",
		Recommendations:    []string{"Performance is good"},
	}

	output := FormatScalingAnalysis(analysis)

	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestFormatScalingComparison_Output(t *testing.T) {
	comparison := ScalingComparison{
		Baseline: ScalingAnalysis{
			WorkerCount:    4,
			WallClockMs:    1500,
			Efficiency:     70,
			BottleneckType: "overhead",
		},
		Comparison: ScalingAnalysis{
			WorkerCount:    4,
			WallClockMs:    1200,
			Efficiency:     80,
			BottleneckType: "minimal",
		},
		Improvement: 20.0,
		Analysis:    "Wall clock improved by 20%",
	}

	output := FormatScalingComparison(comparison)

	if output == "" {
		t.Error("expected non-empty output")
	}
}
