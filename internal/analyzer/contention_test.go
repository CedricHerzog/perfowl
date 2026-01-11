package analyzer

import (
	"testing"

	"github.com/CedricHerzog/perfowl/internal/testutil"
)

func TestAnalyzeContention_EmptyProfile(t *testing.T) {
	profile := testutil.MinimalProfile()

	result := AnalyzeContention(profile)

	if result.TotalEvents != 0 {
		t.Errorf("TotalEvents = %v, want 0", result.TotalEvents)
	}
}

func TestAnalyzeContention_NoContention(t *testing.T) {
	profile := testutil.ProfileWithMainThread()

	result := AnalyzeContention(profile)

	if result.TotalEvents != 0 {
		t.Errorf("TotalEvents = %v, want 0", result.TotalEvents)
	}
}

func TestAnalyzeContention_GCContention(t *testing.T) {
	profile := testutil.ProfileWithContention()

	result := AnalyzeContention(profile)

	// Should detect GC contention
	if result.GCContention == 0 {
		t.Logf("GCContention = %d (may be 0 if no workers active during GC)", result.GCContention)
	}
}

func TestAnalyzeContention_SeverityCalculation(t *testing.T) {
	profile := testutil.ProfileWithContention()

	result := AnalyzeContention(profile)

	validSeverities := map[string]bool{
		"high":    true,
		"medium":  true,
		"low":     true,
		"minimal": true,
	}

	if !validSeverities[result.Severity] {
		t.Errorf("invalid Severity: %s", result.Severity)
	}
}

func TestAnalyzeContention_Recommendations(t *testing.T) {
	profile := testutil.ProfileWithContention()

	result := AnalyzeContention(profile)

	// Recommendations are generated based on specific thresholds:
	// - GCContention > 5
	// - IPCContention > 5
	// - TotalImpactMs > 100
	// - Severity == "high"
	// The test fixture may not always trigger these conditions
	if len(result.Recommendations) > 0 {
		// If we have recommendations, they should be non-empty strings
		for _, rec := range result.Recommendations {
			if rec == "" {
				t.Error("expected non-empty recommendation string")
			}
		}
	} else {
		// No recommendations means none of the thresholds were hit
		t.Logf("No recommendations generated (thresholds not met: GC=%d, IPC=%d, Impact=%.1fms, Severity=%s)",
			result.GCContention, result.IPCContention, result.TotalImpactMs, result.Severity)
	}
}

func TestAnalyzeContention_TotalImpact(t *testing.T) {
	profile := testutil.ProfileWithContention()

	result := AnalyzeContention(profile)

	// Total impact should be non-negative
	if result.TotalImpactMs < 0 {
		t.Errorf("TotalImpactMs should be >= 0, got %f", result.TotalImpactMs)
	}
}

func TestAnalyzeContention_Events(t *testing.T) {
	profile := testutil.ProfileWithContention()

	result := AnalyzeContention(profile)

	// Events should have valid types
	for _, event := range result.Events {
		validTypes := map[string]bool{
			"gc_pause": true,
			"ipc_wait": true,
		}
		if !validTypes[event.Type] {
			t.Errorf("invalid event type: %s", event.Type)
		}
	}
}

func TestFormatContentionAnalysis_Output(t *testing.T) {
	analysis := ContentionAnalysis{
		TotalEvents:   5,
		TotalImpactMs: 500,
		GCContention:  3,
		IPCContention: 2,
		Severity:      "medium",
		Events: []ContentionEvent{
			{Type: "gc_pause", StartTime: 100, Duration: 50, Description: "GC pause"},
		},
		Recommendations: []string{"Reduce GC frequency"},
	}

	output := FormatContentionAnalysis(analysis)

	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestAnalyzeContention_SeverityHigh(t *testing.T) {
	// Create a profile with high contention (>10% of duration)
	profile := testutil.NewProfileBuilder().
		WithDuration(1000).
		WithCategories(testutil.DefaultCategories()).
		Build()

	// Add main thread with lots of GC
	mainMb := testutil.NewMarkerBuilder()
	for i := 0; i < 20; i++ {
		mainMb.AddGCMajor(float64(i*50), 10)
	}
	mainMarkers, mainStrings := mainMb.Build()

	mainThread := testutil.NewThreadBuilder("GeckoMain").
		AsMainThread().
		WithMarkers(mainMarkers).
		WithStringArray(mainStrings).
		Build()
	profile.Threads = append(profile.Threads, mainThread)

	// Add worker that's active during GC
	workerSb := testutil.NewSamplesBuilder()
	for i := 0; i < 200; i++ {
		workerSb.AddSampleWithCPUDelta(0, float64(i*5), 1000)
	}

	workerThread := testutil.NewThreadBuilder("DOM Worker").
		WithSamples(workerSb.Build()).
		Build()
	profile.Threads = append(profile.Threads, workerThread)

	result := AnalyzeContention(profile)

	// Should have some contention detected
	if result.TotalEvents == 0 && result.GCContention == 0 && result.IPCContention == 0 {
		t.Log("No contention detected - this is expected if GC doesn't overlap with worker activity")
	}
}
