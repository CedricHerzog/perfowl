package testutil

import (
	"testing"
)

func TestMinimalProfile(t *testing.T) {
	profile := MinimalProfile()
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	if len(profile.Threads) != 0 {
		t.Errorf("expected 0 threads, got %d", len(profile.Threads))
	}
}

func TestProfileWithMainThread(t *testing.T) {
	profile := ProfileWithMainThread()
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	if len(profile.Threads) != 1 {
		t.Errorf("expected 1 thread, got %d", len(profile.Threads))
	}
	if !profile.Threads[0].IsMainThread {
		t.Error("expected main thread")
	}
}

func TestProfileWithGC(t *testing.T) {
	profile := ProfileWithGC()
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	if profile.Threads[0].Markers.Length == 0 {
		t.Error("expected markers")
	}
}

func TestProfileWithLongTasks(t *testing.T) {
	profile := ProfileWithLongTasks(5)
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	if profile.Threads[0].Markers.Length != 5 {
		t.Errorf("expected 5 markers, got %d", profile.Threads[0].Markers.Length)
	}
}

func TestProfileWithWorkers(t *testing.T) {
	profile := ProfileWithWorkers(3)
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	// 1 main thread + 3 workers
	if len(profile.Threads) != 4 {
		t.Errorf("expected 4 threads, got %d", len(profile.Threads))
	}
}

func TestProfileWithExtensions(t *testing.T) {
	profile := ProfileWithExtensions()
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	if profile.Meta.Extensions.Length != 2 {
		t.Errorf("expected 2 extensions, got %d", profile.Meta.Extensions.Length)
	}
}

func TestProfileWithLayoutThrashing(t *testing.T) {
	profile := ProfileWithLayoutThrashing()
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	// 10 reflows
	if profile.Threads[0].Markers.Length != 10 {
		t.Errorf("expected 10 markers, got %d", profile.Threads[0].Markers.Length)
	}
}

func TestProfileWithSyncIPC(t *testing.T) {
	profile := ProfileWithSyncIPC()
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	// 3 IPC events
	if profile.Threads[0].Markers.Length != 3 {
		t.Errorf("expected 3 markers, got %d", profile.Threads[0].Markers.Length)
	}
}

func TestProfileWithNetworkBlocking(t *testing.T) {
	profile := ProfileWithNetworkBlocking()
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	// 2 network events
	if profile.Threads[0].Markers.Length != 2 {
		t.Errorf("expected 2 markers, got %d", profile.Threads[0].Markers.Length)
	}
}

func TestProfileWithCrypto(t *testing.T) {
	profile := ProfileWithCrypto()
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	if profile.Threads[0].Samples.Length == 0 {
		t.Error("expected samples")
	}
}

func TestProfileWithJSCrypto(t *testing.T) {
	profile := ProfileWithJSCrypto()
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	// Main thread + 2 crypto workers
	if len(profile.Threads) != 3 {
		t.Errorf("expected 3 threads, got %d", len(profile.Threads))
	}
}

func TestProfileWithContention(t *testing.T) {
	profile := ProfileWithContention()
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	if len(profile.Threads) < 2 {
		t.Error("expected at least 2 threads")
	}
}

func TestProfileWithDelimiters(t *testing.T) {
	profile := ProfileWithDelimiters()
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	if profile.Threads[0].Markers.Length == 0 {
		t.Error("expected markers")
	}
}

func TestProfileWithCallTree(t *testing.T) {
	profile := ProfileWithCallTree()
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	if profile.Threads[0].Samples.Length == 0 {
		t.Error("expected samples")
	}
	if profile.Threads[0].StackTable.Length == 0 {
		t.Error("expected stack table")
	}
}

func TestProfileWithCategories(t *testing.T) {
	profile := ProfileWithCategories()
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	if profile.Threads[0].Samples.Length == 0 {
		t.Error("expected samples")
	}
}

func TestProfileWithWakePatterns(t *testing.T) {
	profile := ProfileWithWakePatterns()
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	// 10 awake markers
	if profile.Threads[0].Markers.Length != 10 {
		t.Errorf("expected 10 markers, got %d", profile.Threads[0].Markers.Length)
	}
}

func TestProfileWithBottlenecks(t *testing.T) {
	profile := ProfileWithBottlenecks()
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	if profile.Threads[0].Markers.Length == 0 {
		t.Error("expected markers")
	}
}

func TestProfileWithExtensionActivity(t *testing.T) {
	profile := ProfileWithExtensionActivity()
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	if profile.Meta.Extensions.Length != 2 {
		t.Errorf("expected 2 extensions, got %d", profile.Meta.Extensions.Length)
	}
	if profile.Threads[0].Samples.Length == 0 {
		t.Error("expected samples")
	}
}

func TestProfileWithContentionData(t *testing.T) {
	profile := ProfileWithContentionData()
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	if len(profile.Threads) < 3 {
		t.Errorf("expected at least 3 threads, got %d", len(profile.Threads))
	}
}

func TestProfileWithWorkersData(t *testing.T) {
	profile := ProfileWithWorkersData()
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	// 1 main + 4 workers
	if len(profile.Threads) != 5 {
		t.Errorf("expected 5 threads, got %d", len(profile.Threads))
	}
}
