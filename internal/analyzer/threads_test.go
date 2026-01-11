package analyzer

import (
	"testing"

	"github.com/CedricHerzog/perfowl/internal/testutil"
)

func TestAnalyzeThreads_EmptyProfile(t *testing.T) {
	profile := testutil.MinimalProfile()

	result := AnalyzeThreads(profile)

	if result.TotalThreads != 0 {
		t.Errorf("TotalThreads = %v, want 0", result.TotalThreads)
	}
	if result.MainThreadCount != 0 {
		t.Errorf("MainThreadCount = %v, want 0", result.MainThreadCount)
	}
}

func TestAnalyzeThreads_SingleMainThread(t *testing.T) {
	profile := testutil.ProfileWithMainThread()

	result := AnalyzeThreads(profile)

	if result.TotalThreads != 1 {
		t.Errorf("TotalThreads = %v, want 1", result.TotalThreads)
	}
	if result.MainThreadCount != 1 {
		t.Errorf("MainThreadCount = %v, want 1", result.MainThreadCount)
	}
}

func TestAnalyzeThreads_MultipleThreads(t *testing.T) {
	profile := testutil.ProfileWithWorkers(3)

	result := AnalyzeThreads(profile)

	// 1 main thread + 3 workers
	if result.TotalThreads != 4 {
		t.Errorf("TotalThreads = %v, want 4", result.TotalThreads)
	}
}

func TestAnalyzeThreads_MainThreadCounting(t *testing.T) {
	profile := testutil.NewProfileBuilder().
		WithDuration(1000).
		WithThread(testutil.NewThreadBuilder("GeckoMain").AsMainThread().Build()).
		WithThread(testutil.NewThreadBuilder("ContentMain").AsMainThread().Build()).
		WithThread(testutil.NewThreadBuilder("Worker").Build()).
		Build()

	result := AnalyzeThreads(profile)

	if result.MainThreadCount != 2 {
		t.Errorf("MainThreadCount = %v, want 2", result.MainThreadCount)
	}
}

func TestAnalyzeThreads_ProcessTypeCounting(t *testing.T) {
	profile := testutil.NewProfileBuilder().
		WithDuration(1000).
		WithThread(testutil.NewThreadBuilder("GeckoMain").AsMainThread().WithProcessType("default").Build()).
		WithThread(testutil.NewThreadBuilder("ContentMain").WithProcessType("tab").Build()).
		WithThread(testutil.NewThreadBuilder("WebMain").WithProcessType("web").Build()).
		Build()

	result := AnalyzeThreads(profile)

	if result.ParentProcessThreads != 1 {
		t.Errorf("ParentProcessThreads = %v, want 1", result.ParentProcessThreads)
	}
	if result.ContentProcessThreads != 2 {
		t.Errorf("ContentProcessThreads = %v, want 2", result.ContentProcessThreads)
	}
}

func TestAnalyzeThreads_CPUTimeCalculation(t *testing.T) {
	profile := testutil.ProfileWithWorkers(1)

	result := AnalyzeThreads(profile)

	// Workers should have CPU time from samples
	foundWorker := false
	for _, thread := range result.Threads {
		if thread.Name == "DOM Worker" {
			foundWorker = true
			if thread.CPUTimeMs <= 0 {
				t.Errorf("expected positive CPU time for worker, got %f", thread.CPUTimeMs)
			}
		}
	}
	if !foundWorker {
		t.Error("expected to find DOM Worker thread")
	}
}

func TestAnalyzeThreads_SortedByCPUTime(t *testing.T) {
	profile := testutil.ProfileWithWorkers(3)

	result := AnalyzeThreads(profile)

	// Threads should be sorted by CPU time descending
	for i := 1; i < len(result.Threads); i++ {
		if result.Threads[i].CPUTimeMs > result.Threads[i-1].CPUTimeMs {
			t.Errorf("threads not sorted by CPU time: %f > %f",
				result.Threads[i].CPUTimeMs, result.Threads[i-1].CPUTimeMs)
		}
	}
}

func TestAnalyzeThreads_WakePattern(t *testing.T) {
	profile := testutil.ProfileWithWakePatterns()

	result := AnalyzeThreads(profile)

	// Should detect wake patterns
	if len(result.Threads) == 0 {
		t.Fatal("expected threads")
	}

	mainThread := result.Threads[0]
	if mainThread.WakeCount == 0 {
		t.Error("expected wake count > 0")
	}
}

func TestAnalyzeThreads_TopCategories(t *testing.T) {
	profile := testutil.ProfileWithCategories()

	result := AnalyzeThreads(profile)

	if len(result.Threads) == 0 {
		t.Fatal("expected threads")
	}

	// Main thread should have top categories
	mainThread := result.Threads[0]
	if len(mainThread.TopCategories) == 0 {
		t.Error("expected top categories for main thread")
	}
}
