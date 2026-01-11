package analyzer

import (
	"testing"

	"github.com/CedricHerzog/perfowl/internal/testutil"
)

func TestAnalyzeBatch_EmptyProfiles(t *testing.T) {
	var profiles []ProfileEntry

	result, err := AnalyzeBatch(profiles)

	if err != nil {
		t.Fatalf("AnalyzeBatch error: %v", err)
	}

	if result.Summary.TotalProfiles != 0 {
		t.Errorf("TotalProfiles = %v, want 0", result.Summary.TotalProfiles)
	}
}

func TestAnalyzeBatch_SingleProfile(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)
	path := testutil.TempProfileFile(t, profile)

	profiles := []ProfileEntry{
		{Path: path, WorkerCount: 2, Label: "Test"},
	}

	result, err := AnalyzeBatch(profiles)

	if err != nil {
		t.Fatalf("AnalyzeBatch error: %v", err)
	}

	if result.Summary.TotalProfiles != 1 {
		t.Errorf("TotalProfiles = %v, want 1", result.Summary.TotalProfiles)
	}
}

func TestAnalyzeBatch_MultipleProfiles(t *testing.T) {
	profile1 := testutil.ProfileWithWorkers(2)
	profile2 := testutil.ProfileWithWorkers(4)
	path1 := testutil.TempProfileFile(t, profile1)
	path2 := testutil.TempProfileFile(t, profile2)

	profiles := []ProfileEntry{
		{Path: path1, WorkerCount: 2, Label: "Firefox"},
		{Path: path2, WorkerCount: 4, Label: "Firefox"},
	}

	result, err := AnalyzeBatch(profiles)

	if err != nil {
		t.Fatalf("AnalyzeBatch error: %v", err)
	}

	if result.Summary.TotalProfiles != 2 {
		t.Errorf("TotalProfiles = %v, want 2", result.Summary.TotalProfiles)
	}
}

func TestAnalyzeBatch_DifferentLabels(t *testing.T) {
	profile1 := testutil.ProfileWithWorkers(2)
	profile2 := testutil.ProfileWithWorkers(2)
	path1 := testutil.TempProfileFile(t, profile1)
	path2 := testutil.TempProfileFile(t, profile2)

	profiles := []ProfileEntry{
		{Path: path1, WorkerCount: 2, Label: "Firefox"},
		{Path: path2, WorkerCount: 2, Label: "Chrome"},
	}

	result, err := AnalyzeBatch(profiles)

	if err != nil {
		t.Fatalf("AnalyzeBatch error: %v", err)
	}

	if len(result.Summary.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(result.Summary.Labels))
	}
}

func TestAnalyzeBatch_SortedByWorkerCount(t *testing.T) {
	profile1 := testutil.ProfileWithWorkers(4)
	profile2 := testutil.ProfileWithWorkers(2)
	profile3 := testutil.ProfileWithWorkers(8)
	path1 := testutil.TempProfileFile(t, profile1)
	path2 := testutil.TempProfileFile(t, profile2)
	path3 := testutil.TempProfileFile(t, profile3)

	profiles := []ProfileEntry{
		{Path: path1, WorkerCount: 4, Label: "Test"},
		{Path: path2, WorkerCount: 2, Label: "Test"},
		{Path: path3, WorkerCount: 8, Label: "Test"},
	}

	result, err := AnalyzeBatch(profiles)

	if err != nil {
		t.Fatalf("AnalyzeBatch error: %v", err)
	}

	// Data points should be sorted by worker count
	series := result.Series["Test"]
	for i := 1; i < len(series); i++ {
		if series[i].WorkerCount < series[i-1].WorkerCount {
			t.Errorf("series not sorted by worker count")
		}
	}
}

func TestAnalyzeBatch_SummaryLabels(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)
	path := testutil.TempProfileFile(t, profile)

	profiles := []ProfileEntry{
		{Path: path, WorkerCount: 2, Label: "Firefox"},
	}

	result, err := AnalyzeBatch(profiles)

	if err != nil {
		t.Fatalf("AnalyzeBatch error: %v", err)
	}

	if len(result.Summary.Labels) == 0 {
		t.Error("expected labels in summary")
	}

	found := false
	for _, label := range result.Summary.Labels {
		if label == "Firefox" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'Firefox' in labels")
	}
}

func TestAnalyzeBatch_BestWorkers(t *testing.T) {
	profile := testutil.ProfileWithWorkers(4)
	path := testutil.TempProfileFile(t, profile)

	profiles := []ProfileEntry{
		{Path: path, WorkerCount: 4, Label: "Test"},
	}

	result, err := AnalyzeBatch(profiles)

	if err != nil {
		t.Fatalf("AnalyzeBatch error: %v", err)
	}

	if result.Summary.BestWorkers == nil {
		t.Error("expected BestWorkers map")
	}
}

func TestAnalyzeBatch_MinWallClock(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)
	path := testutil.TempProfileFile(t, profile)

	profiles := []ProfileEntry{
		{Path: path, WorkerCount: 2, Label: "Test"},
	}

	result, err := AnalyzeBatch(profiles)

	if err != nil {
		t.Fatalf("AnalyzeBatch error: %v", err)
	}

	if result.Summary.MinWallClock == nil {
		t.Error("expected MinWallClock map")
	}
}

func TestAnalyzeBatch_LoadError(t *testing.T) {
	profiles := []ProfileEntry{
		{Path: "/nonexistent/profile.json", WorkerCount: 2, Label: "Test"},
	}

	result, err := AnalyzeBatch(profiles)

	// Should not fail completely, but profile won't be included
	if err != nil {
		t.Logf("AnalyzeBatch returned error (may be expected): %v", err)
	}

	if result != nil && result.Summary.TotalProfiles != 0 {
		t.Errorf("TotalProfiles = %v, want 0 for failed load", result.Summary.TotalProfiles)
	}
}

func TestAnalyzeBatch_OperationTimeMeasurement(t *testing.T) {
	profile := testutil.ProfileWithDelimiters()
	path := testutil.TempProfileFile(t, profile)

	profiles := []ProfileEntry{
		{
			Path:         path,
			WorkerCount:  2,
			Label:        "Test",
			StartPattern: "DOMEvent",
			EndPattern:   "Paint",
		},
	}

	result, err := AnalyzeBatch(profiles)

	if err != nil {
		t.Fatalf("AnalyzeBatch error: %v", err)
	}

	// Operation time measurement is optional
	if result.Summary.TotalProfiles != 1 {
		t.Errorf("TotalProfiles = %v, want 1", result.Summary.TotalProfiles)
	}
}

func TestProfileDataPoint_Fields(t *testing.T) {
	point := ProfileDataPoint{
		WorkerCount:     4,
		Label:           "Firefox",
		FilePath:        "/path/to/profile.json",
		WallClockMs:     1000,
		OperationTimeMs: 800,
		TotalWorkMs:     3500,
		Efficiency:      87.5,
		Speedup:         3.5,
		CryptoTimeMs:    200,
	}

	if point.WorkerCount != 4 {
		t.Errorf("WorkerCount = %v, want 4", point.WorkerCount)
	}
	if point.Label != "Firefox" {
		t.Errorf("Label = %v, want Firefox", point.Label)
	}
	if point.Efficiency < 0 || point.Efficiency > 100 {
		t.Errorf("Efficiency out of range: %f", point.Efficiency)
	}
}

func TestBatchSummary_Fields(t *testing.T) {
	summary := BatchSummary{
		TotalProfiles:    5,
		Labels:           []string{"Firefox", "Chrome"},
		BestWorkers:      map[string]int{"Firefox": 4, "Chrome": 8},
		MinWallClock:     map[string]float64{"Firefox": 1000, "Chrome": 900},
		MinOperationTime: map[string]float64{"Firefox": 800, "Chrome": 700},
		MaxSpeedup:       map[string]float64{"Firefox": 3.5, "Chrome": 4.0},
		PeakEfficiency:   map[string]float64{"Firefox": 87.5, "Chrome": 90.0},
	}

	if summary.TotalProfiles != 5 {
		t.Errorf("TotalProfiles = %v, want 5", summary.TotalProfiles)
	}
	if len(summary.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(summary.Labels))
	}
}
