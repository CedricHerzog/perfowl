package analyzer

import (
	"testing"

	"github.com/CedricHerzog/perfowl/internal/parser"
	"github.com/CedricHerzog/perfowl/internal/testutil"
)

func TestIsWorkerThread_DOMWorker(t *testing.T) {
	thread := &parser.Thread{Name: "DOM Worker"}
	if !isWorkerThread(thread) {
		t.Error("expected DOM Worker to be identified as worker thread")
	}
}

func TestIsWorkerThread_DedicatedWorker(t *testing.T) {
	thread := &parser.Thread{Name: "DedicatedWorker"}
	if !isWorkerThread(thread) {
		t.Error("expected DedicatedWorker to be identified as worker thread")
	}
}

func TestIsWorkerThread_SharedWorker(t *testing.T) {
	thread := &parser.Thread{Name: "SharedWorker"}
	if !isWorkerThread(thread) {
		t.Error("expected SharedWorker to be identified as worker thread")
	}
}

func TestIsWorkerThread_ServiceWorker(t *testing.T) {
	thread := &parser.Thread{Name: "ServiceWorker"}
	if !isWorkerThread(thread) {
		t.Error("expected ServiceWorker to be identified as worker thread")
	}
}

func TestIsWorkerThread_ExcludesThreadPoolForegroundWorker(t *testing.T) {
	// Chrome's internal thread pool patterns are excluded
	thread := &parser.Thread{Name: "ThreadPoolForegroundWorker"}
	if isWorkerThread(thread) {
		t.Error("expected ThreadPoolForegroundWorker to be excluded")
	}
}

func TestIsWorkerThread_ExcludesThreadPoolBackgroundWorker(t *testing.T) {
	thread := &parser.Thread{Name: "ThreadPoolBackgroundWorker"}
	if isWorkerThread(thread) {
		t.Error("expected ThreadPoolBackgroundWorker to be excluded")
	}
}

func TestIsWorkerThread_ExcludesCompositorTileWorker(t *testing.T) {
	thread := &parser.Thread{Name: "CompositorTileWorker"}
	if isWorkerThread(thread) {
		t.Error("expected CompositorTileWorker to be excluded")
	}
}

func TestIsWorkerThread_ExcludesAudioWorklet(t *testing.T) {
	thread := &parser.Thread{Name: "AudioWorklet"}
	if isWorkerThread(thread) {
		t.Error("expected AudioWorklet to be excluded")
	}
}

func TestIsWorkerThread_ExcludesV8Profiler(t *testing.T) {
	// The actual exclusion pattern is "v8:profevntproc"
	thread := &parser.Thread{Name: "V8:ProfEvntProc"}
	if isWorkerThread(thread) {
		t.Error("expected V8:ProfEvntProc to be excluded")
	}
}

func TestIsWorkerThread_NonWorkerThreadWithoutWorkerInName(t *testing.T) {
	// Compositor doesn't contain "worker" so it's not matched
	thread := &parser.Thread{Name: "Compositor"}
	if isWorkerThread(thread) {
		t.Error("expected Compositor to not be a worker thread")
	}
}

func TestIsWorkerThread_MainThreadExcluded(t *testing.T) {
	thread := &parser.Thread{Name: "GeckoMain"}
	if isWorkerThread(thread) {
		t.Error("expected main thread to be excluded")
	}
}

func TestAnalyzeWorkers_NoWorkers(t *testing.T) {
	profile := testutil.ProfileWithMainThread()

	result := AnalyzeWorkers(profile)

	if result.TotalWorkers != 0 {
		t.Errorf("TotalWorkers = %v, want 0", result.TotalWorkers)
	}
}

func TestAnalyzeWorkers_SingleWorker(t *testing.T) {
	profile := testutil.ProfileWithWorkers(1)

	result := AnalyzeWorkers(profile)

	if result.TotalWorkers != 1 {
		t.Errorf("TotalWorkers = %v, want 1", result.TotalWorkers)
	}
}

func TestAnalyzeWorkers_MultipleWorkers(t *testing.T) {
	profile := testutil.ProfileWithWorkers(3)

	result := AnalyzeWorkers(profile)

	if result.TotalWorkers != 3 {
		t.Errorf("TotalWorkers = %v, want 3", result.TotalWorkers)
	}
}

func TestAnalyzeWorkers_CPUTimeCalculation(t *testing.T) {
	profile := testutil.ProfileWithWorkers(1)

	result := AnalyzeWorkers(profile)

	if result.TotalCPUTimeMs <= 0 {
		t.Errorf("expected positive TotalCPUTimeMs, got %f", result.TotalCPUTimeMs)
	}
}

func TestAnalyzeWorkers_IdleTimeCalculation(t *testing.T) {
	profile := testutil.ProfileWithWorkers(1)

	result := AnalyzeWorkers(profile)

	// Idle time should be calculated
	if result.TotalIdleTimeMs < 0 {
		t.Errorf("expected non-negative TotalIdleTimeMs, got %f", result.TotalIdleTimeMs)
	}
}

func TestAnalyzeWorkers_ActivePercent(t *testing.T) {
	profile := testutil.ProfileWithWorkers(1)

	result := AnalyzeWorkers(profile)

	// Active percent should be between 0 and 100
	for _, w := range result.Workers {
		if w.ActivePercent < 0 || w.ActivePercent > 100 {
			t.Errorf("ActivePercent out of range: %f", w.ActivePercent)
		}
	}
}

func TestAnalyzeWorkers_OverallEfficiency(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)

	result := AnalyzeWorkers(profile)

	// Overall efficiency should be between 0 and 100
	if result.OverallEfficiency < 0 || result.OverallEfficiency > 100 {
		t.Errorf("OverallEfficiency out of range: %f", result.OverallEfficiency)
	}
}

func TestAnalyzeWorkers_EmptyProfile(t *testing.T) {
	profile := testutil.MinimalProfile()

	result := AnalyzeWorkers(profile)

	if result.TotalWorkers != 0 {
		t.Errorf("TotalWorkers = %v, want 0 for empty profile", result.TotalWorkers)
	}
}

func TestFormatWorkerAnalysis_Output(t *testing.T) {
	analysis := WorkerAnalysis{
		TotalWorkers:      2,
		ActiveWorkers:     2,
		TotalCPUTimeMs:    1000,
		TotalIdleTimeMs:   500,
		OverallEfficiency: 66.7,
		Workers: []WorkerStats{
			{ThreadName: "DOM Worker", CPUTimeMs: 500, IdleTimeMs: 250, ActivePercent: 66.7},
		},
	}

	output := FormatWorkerAnalysis(analysis)

	if output == "" {
		t.Error("expected non-empty output")
	}
}
