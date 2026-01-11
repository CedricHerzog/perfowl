package analyzer

import (
	"testing"

	"github.com/CedricHerzog/perfowl/internal/testutil"
)

func TestAnalyzeCallTree_EmptyProfile(t *testing.T) {
	profile := testutil.MinimalProfile()

	result := AnalyzeCallTree(profile, "", 20)

	if result.TotalTimeMs != 0 {
		t.Errorf("TotalTimeMs = %v, want 0", result.TotalTimeMs)
	}
	if result.TotalSamples != 0 {
		t.Errorf("TotalSamples = %v, want 0", result.TotalSamples)
	}
}

func TestAnalyzeCallTree_WithSamples(t *testing.T) {
	profile := testutil.ProfileWithCallTree()

	result := AnalyzeCallTree(profile, "", 20)

	if result.TotalSamples == 0 {
		t.Error("expected samples")
	}
	if result.TotalTimeMs <= 0 {
		t.Errorf("expected positive TotalTimeMs, got %f", result.TotalTimeMs)
	}
}

func TestAnalyzeCallTree_TopFunctions(t *testing.T) {
	profile := testutil.ProfileWithCallTree()

	result := AnalyzeCallTree(profile, "", 20)

	if len(result.TopFunctions) == 0 {
		t.Error("expected top functions")
	}

	// First function should have highest self time
	for i := 1; i < len(result.TopFunctions); i++ {
		if result.TopFunctions[i].SelfTimeMs > result.TopFunctions[i-1].SelfTimeMs {
			t.Errorf("functions not sorted by self time")
		}
	}
}

func TestAnalyzeCallTree_HotPaths(t *testing.T) {
	profile := testutil.ProfileWithCallTree()

	result := AnalyzeCallTree(profile, "", 20)

	if len(result.HotPaths) == 0 {
		t.Error("expected hot paths")
	}
}

func TestAnalyzeCallTree_ThreadFilter(t *testing.T) {
	profile := testutil.ProfileWithCallTree()

	// Filter by specific thread
	result := AnalyzeCallTree(profile, "GeckoMain", 20)

	if result.ThreadName != "GeckoMain" {
		t.Errorf("ThreadName = %v, want GeckoMain", result.ThreadName)
	}
}

func TestAnalyzeCallTree_ThreadFilterNotFound(t *testing.T) {
	profile := testutil.ProfileWithCallTree()

	result := AnalyzeCallTree(profile, "NonexistentThread", 20)

	if result.TotalSamples != 0 {
		t.Errorf("expected 0 samples for nonexistent thread, got %d", result.TotalSamples)
	}
}

func TestAnalyzeCallTree_LimitParameter(t *testing.T) {
	profile := testutil.ProfileWithCallTree()

	result := AnalyzeCallTree(profile, "", 3)

	if len(result.TopFunctions) > 3 {
		t.Errorf("expected at most 3 functions, got %d", len(result.TopFunctions))
	}
}

func TestAnalyzeCallTree_DefaultLimit(t *testing.T) {
	profile := testutil.ProfileWithCallTree()

	result := AnalyzeCallTree(profile, "", 0)

	// Should use default limit (20)
	if len(result.TopFunctions) > 20 {
		t.Errorf("expected at most 20 functions with default limit, got %d", len(result.TopFunctions))
	}
}

func TestAnalyzeCallTree_SelfTimeCalculation(t *testing.T) {
	profile := testutil.ProfileWithCallTree()

	result := AnalyzeCallTree(profile, "", 20)

	// Self time should be <= running time for all functions
	for _, fn := range result.TopFunctions {
		if fn.SelfTimeMs > fn.RunningTimeMs {
			t.Errorf("SelfTimeMs (%f) > RunningTimeMs (%f) for %s",
				fn.SelfTimeMs, fn.RunningTimeMs, fn.Name)
		}
	}
}

func TestAnalyzeCallTree_PercentCalculation(t *testing.T) {
	profile := testutil.ProfileWithCallTree()

	result := AnalyzeCallTree(profile, "", 20)

	for _, fn := range result.TopFunctions {
		if fn.SelfPercent < 0 || fn.SelfPercent > 100 {
			t.Errorf("SelfPercent out of range: %f", fn.SelfPercent)
		}
		if fn.TotalPercent < 0 || fn.TotalPercent > 100 {
			t.Errorf("TotalPercent out of range: %f", fn.TotalPercent)
		}
	}
}

func TestFormatCallTree_Output(t *testing.T) {
	analysis := CallTreeAnalysis{
		TotalTimeMs:  1000,
		TotalSamples: 100,
		ThreadName:   "GeckoMain",
		TopFunctions: []FunctionStats{
			{Name: "main", SelfTimeMs: 500, RunningTimeMs: 1000, SelfPercent: 50, TotalPercent: 100},
		},
		HotPaths: []HotPath{
			{Path: "main -> child", SelfTimeMs: 500, Percent: 50},
		},
	}

	output := FormatCallTree(analysis)

	if output == "" {
		t.Error("expected non-empty output")
	}
}
