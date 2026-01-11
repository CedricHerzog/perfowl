package parser

import (
	"encoding/json"
	"testing"
)

func TestDefaultCategories(t *testing.T) {
	categories := defaultCategories()

	// Verify expected categories exist
	expectedCategories := []string{
		"Idle", "Other", "Layout", "JavaScript", "GC / CC",
		"Network", "Graphics", "DOM", "UserTiming", "IPC",
	}

	if len(categories) != len(expectedCategories) {
		t.Errorf("Expected %d categories, got %d", len(expectedCategories), len(categories))
	}

	for i, expected := range expectedCategories {
		if categories[i].Name != expected {
			t.Errorf("Category[%d].Name = %q, want %q", i, categories[i].Name, expected)
		}
	}
}

func TestChromeCategoryMapping(t *testing.T) {
	tests := []struct {
		chromeCategory  string
		firefoxCategory string
	}{
		{"devtools.timeline", "JavaScript"},
		{"v8", "JavaScript"},
		{"v8.execute", "JavaScript"},
		{"v8.compile", "JavaScript"},
		{"disabled-by-default-v8.gc", "GC / CC"},
		{"blink", "Layout"},
		{"blink.user_timing", "UserTiming"},
		{"loading", "Network"},
		{"net", "Network"},
		{"netlog", "Network"},
		{"gpu", "Graphics"},
		{"cc", "Graphics"},
		{"viz", "Graphics"},
		{"ipc", "IPC"},
		{"__metadata", "Other"},
		{"toplevel", "Other"},
	}

	for _, tt := range tests {
		t.Run(tt.chromeCategory, func(t *testing.T) {
			firefoxCat, ok := chromeCategoryMap[tt.chromeCategory]
			if !ok {
				t.Errorf("Category %q not found in chromeCategoryMap", tt.chromeCategory)
				return
			}
			if firefoxCat != tt.firefoxCategory {
				t.Errorf("chromeCategoryMap[%q] = %q, want %q", tt.chromeCategory, firefoxCat, tt.firefoxCategory)
			}
		})
	}
}

func TestConvertChromeToProfile_MinimalProfile(t *testing.T) {
	chrome := &ChromeProfile{
		Metadata: ChromeMetadata{
			Source:    "Chrome DevTools",
			StartTime: "2024-01-01T00:00:00Z",
		},
		TraceEvents: []ChromeEvent{
			{
				Name: "thread_name",
				Cat:  "__metadata",
				Ph:   "M",
				Ts:   0,
				Pid:  1,
				Tid:  1,
				Args: json.RawMessage(`{"name":"CrRendererMain"}`),
			},
			{
				Name: "FunctionCall",
				Cat:  "devtools.timeline",
				Ph:   "X",
				Ts:   1000000,
				Dur:  5000,
				Pid:  1,
				Tid:  1,
			},
		},
	}

	profile, err := ConvertChromeToProfile(chrome)
	if err != nil {
		t.Fatalf("ConvertChromeToProfile() error = %v", err)
	}

	// Verify meta
	if profile.Meta.Product != "Chrome" {
		t.Errorf("Meta.Product = %q, want %q", profile.Meta.Product, "Chrome")
	}

	// Verify threads exist
	if len(profile.Threads) == 0 {
		t.Fatal("Expected at least one thread")
	}

	// Find the renderer thread
	var rendererThread *Thread
	for i := range profile.Threads {
		if profile.Threads[i].Name == "CrRendererMain" {
			rendererThread = &profile.Threads[i]
			break
		}
	}

	if rendererThread == nil {
		t.Fatal("Could not find CrRendererMain thread")
	}

	// Verify markers
	if rendererThread.Markers.Length == 0 {
		t.Error("Expected at least one marker")
	}
}

func TestConvertChromeToProfile_ProcessNames(t *testing.T) {
	chrome := &ChromeProfile{
		TraceEvents: []ChromeEvent{
			{
				Name: "process_name",
				Cat:  "__metadata",
				Ph:   "M",
				Ts:   0,
				Pid:  1,
				Tid:  0,
				Args: json.RawMessage(`{"name":"Browser"}`),
			},
			{
				Name: "thread_name",
				Cat:  "__metadata",
				Ph:   "M",
				Ts:   0,
				Pid:  1,
				Tid:  1,
				Args: json.RawMessage(`{"name":"CrBrowserMain"}`),
			},
			{
				Name: "SomeEvent",
				Ph:   "X",
				Ts:   1000000,
				Dur:  1000,
				Pid:  1,
				Tid:  1,
			},
		},
	}

	profile, err := ConvertChromeToProfile(chrome)
	if err != nil {
		t.Fatalf("ConvertChromeToProfile() error = %v", err)
	}

	// Find the browser main thread
	var browserThread *Thread
	for i := range profile.Threads {
		if profile.Threads[i].Name == "CrBrowserMain" {
			browserThread = &profile.Threads[i]
			break
		}
	}

	if browserThread == nil {
		t.Fatal("Could not find CrBrowserMain thread")
	}

	if browserThread.ProcessName != "Browser" {
		t.Errorf("ProcessName = %q, want %q", browserThread.ProcessName, "Browser")
	}

	if browserThread.ProcessType != "default" {
		t.Errorf("ProcessType = %q, want %q", browserThread.ProcessType, "default")
	}
}

func TestConvertChromeToProfile_DurationEvents(t *testing.T) {
	chrome := &ChromeProfile{
		TraceEvents: []ChromeEvent{
			{
				Name: "FunctionCall",
				Cat:  "devtools.timeline",
				Ph:   "X",
				Ts:   1000000, // 1 second in microseconds
				Dur:  50000,   // 50ms in microseconds
				Pid:  1,
				Tid:  1,
				Args: json.RawMessage(`{"data":{"functionName":"test"}}`),
			},
		},
	}

	profile, err := ConvertChromeToProfile(chrome)
	if err != nil {
		t.Fatalf("ConvertChromeToProfile() error = %v", err)
	}

	if len(profile.Threads) == 0 {
		t.Fatal("Expected at least one thread")
	}

	thread := &profile.Threads[0]
	if thread.Markers.Length == 0 {
		t.Fatal("Expected at least one marker")
	}

	// Duration should be 50ms
	if len(thread.Markers.StartTime) == 0 {
		t.Fatal("Expected StartTime to be populated")
	}

	// The start time should be 0 (relative to profile start)
	// since there's only one event
	startTime := thread.Markers.StartTime[0]
	if startTime != 0 {
		t.Errorf("StartTime = %v, want 0", startTime)
	}

	// End time should be 50 (50ms duration)
	endTime := thread.Markers.EndTime[0]
	if endTimeFloat, ok := endTime.(float64); ok {
		if endTimeFloat != 50 {
			t.Errorf("EndTime = %v, want 50", endTimeFloat)
		}
	} else {
		t.Errorf("EndTime has unexpected type: %T", endTime)
	}
}

func TestConvertChromeToProfile_InstantEvents(t *testing.T) {
	chrome := &ChromeProfile{
		TraceEvents: []ChromeEvent{
			{
				Name: "LayoutShift",
				Cat:  "loading",
				Ph:   "I",
				Ts:   1000000,
				Pid:  1,
				Tid:  1,
			},
		},
	}

	profile, err := ConvertChromeToProfile(chrome)
	if err != nil {
		t.Fatalf("ConvertChromeToProfile() error = %v", err)
	}

	if len(profile.Threads) == 0 {
		t.Fatal("Expected at least one thread")
	}

	thread := &profile.Threads[0]
	if thread.Markers.Length == 0 {
		t.Fatal("Expected at least one marker")
	}

	// Instant events should have nil end time
	endTime := thread.Markers.EndTime[0]
	if endTime != nil {
		t.Errorf("EndTime for instant event should be nil, got %v", endTime)
	}
}

func TestConvertChromeToProfile_V8CPUProfile(t *testing.T) {
	chrome := &ChromeProfile{
		TraceEvents: []ChromeEvent{
			{
				Name: "ProfileChunk",
				Cat:  "disabled-by-default-v8.cpu_profiler",
				Ph:   "P",
				Ts:   1000000,
				Pid:  1,
				Tid:  1,
				Args: json.RawMessage(`{
					"data": {
						"cpuProfile": {
							"nodes": [
								{"id": 1, "callFrame": {"functionName": "(root)", "scriptId": 0, "url": ""}},
								{"id": 2, "callFrame": {"functionName": "main", "scriptId": 1, "url": "file://test.js", "lineNumber": 10}, "parent": 1}
							],
							"samples": [1, 2, 2, 1]
						},
						"timeDeltas": [1000, 1000, 1000, 1000]
					}
				}`),
			},
		},
	}

	profile, err := ConvertChromeToProfile(chrome)
	if err != nil {
		t.Fatalf("ConvertChromeToProfile() error = %v", err)
	}

	if len(profile.Threads) == 0 {
		t.Fatal("Expected at least one thread")
	}

	thread := &profile.Threads[0]

	// Verify samples were extracted
	if thread.Samples.Length != 4 {
		t.Errorf("Samples.Length = %d, want 4", thread.Samples.Length)
	}

	// Verify function table
	if thread.FuncTable.Length < 2 {
		t.Errorf("FuncTable.Length = %d, want at least 2", thread.FuncTable.Length)
	}

	// Verify frame table
	if thread.FrameTable.Length < 2 {
		t.Errorf("FrameTable.Length = %d, want at least 2", thread.FrameTable.Length)
	}

	// Verify stack table
	if thread.StackTable.Length < 2 {
		t.Errorf("StackTable.Length = %d, want at least 2", thread.StackTable.Length)
	}
}

func TestConvertChromeToProfile_CategoryMapping(t *testing.T) {
	chrome := &ChromeProfile{
		TraceEvents: []ChromeEvent{
			{Name: "Event1", Cat: "devtools.timeline", Ph: "X", Ts: 1000000, Dur: 1000, Pid: 1, Tid: 1},
			{Name: "Event2", Cat: "blink", Ph: "X", Ts: 1001000, Dur: 1000, Pid: 1, Tid: 1},
			{Name: "Event3", Cat: "loading", Ph: "X", Ts: 1002000, Dur: 1000, Pid: 1, Tid: 1},
			{Name: "Event4", Cat: "gpu", Ph: "X", Ts: 1003000, Dur: 1000, Pid: 1, Tid: 1},
			{Name: "Event5", Cat: "disabled-by-default-v8.gc", Ph: "X", Ts: 1004000, Dur: 1000, Pid: 1, Tid: 1},
		},
	}

	profile, err := ConvertChromeToProfile(chrome)
	if err != nil {
		t.Fatalf("ConvertChromeToProfile() error = %v", err)
	}

	if len(profile.Threads) == 0 {
		t.Fatal("Expected at least one thread")
	}

	thread := &profile.Threads[0]

	// Build category index map
	categoryNames := make(map[int]string)
	for i, cat := range profile.Meta.Categories {
		categoryNames[i] = cat.Name
	}

	// Verify category assignments
	expectedCategories := []string{"JavaScript", "Layout", "Network", "Graphics", "GC / CC"}
	for i, expected := range expectedCategories {
		if i >= len(thread.Markers.Category) {
			t.Fatalf("Not enough markers, expected at least %d", i+1)
		}
		catIdx := thread.Markers.Category[i]
		catName := categoryNames[catIdx]
		if catName != expected {
			t.Errorf("Marker[%d] category = %q, want %q", i, catName, expected)
		}
	}
}

func TestConvertChromeToProfile_StringDeduplication(t *testing.T) {
	chrome := &ChromeProfile{
		TraceEvents: []ChromeEvent{
			{Name: "FunctionCall", Ph: "X", Ts: 1000000, Dur: 1000, Pid: 1, Tid: 1},
			{Name: "FunctionCall", Ph: "X", Ts: 1001000, Dur: 1000, Pid: 1, Tid: 1},
			{Name: "FunctionCall", Ph: "X", Ts: 1002000, Dur: 1000, Pid: 1, Tid: 1},
		},
	}

	profile, err := ConvertChromeToProfile(chrome)
	if err != nil {
		t.Fatalf("ConvertChromeToProfile() error = %v", err)
	}

	// "FunctionCall" should only appear once in the string array
	count := 0
	for _, s := range profile.Shared.StringArray {
		if s == "FunctionCall" {
			count++
		}
	}

	if count != 1 {
		t.Errorf("String 'FunctionCall' appears %d times in StringArray, want 1", count)
	}
}

func TestConvertChromeToProfile_TimeRange(t *testing.T) {
	chrome := &ChromeProfile{
		TraceEvents: []ChromeEvent{
			// Metadata events have ts=0 and should be skipped for time range
			{Name: "thread_name", Ph: "M", Ts: 0, Pid: 1, Tid: 1, Args: json.RawMessage(`{"name":"Main"}`)},
			// Actual events
			{Name: "Event1", Ph: "X", Ts: 5000000, Dur: 1000000, Pid: 1, Tid: 1},  // 5s to 6s
			{Name: "Event2", Ph: "X", Ts: 10000000, Dur: 2000000, Pid: 1, Tid: 1}, // 10s to 12s
		},
	}

	profile, err := ConvertChromeToProfile(chrome)
	if err != nil {
		t.Fatalf("ConvertChromeToProfile() error = %v", err)
	}

	// Duration should be from first event to end of last event
	// minTime=5000000us, maxTime=12000000us
	// Duration = (12000000 - 5000000) / 1000 = 7000ms = 7s
	expectedDuration := 7000.0
	actualDuration := profile.Meta.ProfilingEndTime

	if actualDuration != expectedDuration {
		t.Errorf("ProfilingEndTime = %v, want %v", actualDuration, expectedDuration)
	}
}

func TestConvertChromeToProfile_EmptyProfile(t *testing.T) {
	chrome := &ChromeProfile{
		TraceEvents: []ChromeEvent{},
	}

	profile, err := ConvertChromeToProfile(chrome)
	if err != nil {
		t.Fatalf("ConvertChromeToProfile() error = %v", err)
	}

	if profile.Meta.Product != "Chrome" {
		t.Errorf("Meta.Product = %q, want %q", profile.Meta.Product, "Chrome")
	}

	if len(profile.Threads) != 0 {
		t.Errorf("Expected 0 threads for empty profile, got %d", len(profile.Threads))
	}
}

func TestInternString(t *testing.T) {
	c := &chromeConverter{
		stringMap:   make(map[string]int),
		stringArray: make([]string, 0),
	}

	// First intern
	idx1 := c.internString("hello")
	if idx1 != 0 {
		t.Errorf("First intern should return 0, got %d", idx1)
	}

	// Same string should return same index
	idx2 := c.internString("hello")
	if idx2 != 0 {
		t.Errorf("Second intern of same string should return 0, got %d", idx2)
	}

	// Different string should return new index
	idx3 := c.internString("world")
	if idx3 != 1 {
		t.Errorf("Intern of new string should return 1, got %d", idx3)
	}

	// Verify string array
	if len(c.stringArray) != 2 {
		t.Errorf("StringArray length = %d, want 2", len(c.stringArray))
	}
	if c.stringArray[0] != "hello" {
		t.Errorf("StringArray[0] = %q, want 'hello'", c.stringArray[0])
	}
	if c.stringArray[1] != "world" {
		t.Errorf("StringArray[1] = %q, want 'world'", c.stringArray[1])
	}
}

func TestMapCategory(t *testing.T) {
	c := &chromeConverter{
		categories:  defaultCategories(),
		categoryMap: make(map[string]int),
	}

	// Build category lookup
	for i, cat := range c.categories {
		c.categoryMap[cat.Name] = i
	}

	tests := []struct {
		chromeCategory string
		expectedName   string
	}{
		{"devtools.timeline", "JavaScript"},
		{"v8,devtools.timeline", "JavaScript"}, // Comma-separated
		{"blink", "Layout"},
		{"unknown_category", "Other"},
		{"", "Other"},
	}

	for _, tt := range tests {
		t.Run(tt.chromeCategory, func(t *testing.T) {
			idx := c.mapCategory(tt.chromeCategory)
			if c.categories[idx].Name != tt.expectedName {
				t.Errorf("mapCategory(%q) -> %q, want %q",
					tt.chromeCategory, c.categories[idx].Name, tt.expectedName)
			}
		})
	}
}

func TestGetCategoryForCallFrame(t *testing.T) {
	c := &chromeConverter{
		categories:  defaultCategories(),
		categoryMap: make(map[string]int),
	}

	for i, cat := range c.categories {
		c.categoryMap[cat.Name] = i
	}

	tests := []struct {
		name         string
		callFrame    V8CallFrame
		expectedName string
	}{
		{
			name: "HTTP URL",
			callFrame: V8CallFrame{
				FunctionName: "test",
				URL:          "https://example.com/script.js",
			},
			expectedName: "JavaScript",
		},
		{
			name: "File URL",
			callFrame: V8CallFrame{
				FunctionName: "test",
				URL:          "file:///path/to/script.js",
			},
			expectedName: "JavaScript",
		},
		{
			name: "Chrome extension",
			callFrame: V8CallFrame{
				FunctionName: "test",
				URL:          "chrome-extension://abcd1234/script.js",
			},
			expectedName: "Other",
		},
		{
			name: "Root function",
			callFrame: V8CallFrame{
				FunctionName: "(root)",
			},
			expectedName: "Other",
		},
		{
			name: "Program function",
			callFrame: V8CallFrame{
				FunctionName: "(program)",
			},
			expectedName: "Other",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx := c.getCategoryForCallFrame(&tt.callFrame)
			if c.categories[idx].Name != tt.expectedName {
				t.Errorf("getCategoryForCallFrame() -> %q, want %q",
					c.categories[idx].Name, tt.expectedName)
			}
		})
	}
}

func TestEventIDToString(t *testing.T) {
	c := &chromeConverter{}

	tests := []struct {
		name     string
		id       any
		expected string
	}{
		{"nil", nil, ""},
		{"string hex", "0x1", "0x1"},
		{"string plain", "abc123", "abc123"},
		{"float64", float64(12345), "12345"},
		{"int converted to float64", float64(42), "42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.eventIDToString(tt.id)
			if result != tt.expected {
				t.Errorf("eventIDToString(%v) = %q, want %q", tt.id, result, tt.expected)
			}
		})
	}
}

func TestProfileTargetMapping(t *testing.T) {
	// This test verifies that ProfileChunk events get their samples assigned
	// to the correct target thread based on the Profile event with matching id
	chrome := &ChromeProfile{
		TraceEvents: []ChromeEvent{
			// Thread metadata for the worker thread (tid=100)
			{
				Name: "thread_name",
				Cat:  "__metadata",
				Ph:   "M",
				Ts:   0,
				Pid:  1,
				Tid:  100,
				Args: json.RawMessage(`{"name":"DedicatedWorker thread"}`),
			},
			// Thread metadata for the profiler thread (tid=200)
			{
				Name: "thread_name",
				Cat:  "__metadata",
				Ph:   "M",
				Ts:   0,
				Pid:  1,
				Tid:  200,
				Args: json.RawMessage(`{"name":"v8:ProfEvntProc"}`),
			},
			// Profile event indicates tid=100 is being profiled, assigned id "0x1"
			{
				Name: "Profile",
				Cat:  "disabled-by-default-v8.cpu_profiler",
				Ph:   "P",
				Ts:   1000000,
				Pid:  1,
				Tid:  100, // The thread being profiled
				ID:   "0x1",
			},
			// ProfileChunk event is emitted by the profiler thread (tid=200)
			// but should assign samples to tid=100 based on id="0x1"
			{
				Name: "ProfileChunk",
				Cat:  "disabled-by-default-v8.cpu_profiler",
				Ph:   "P",
				Ts:   1001000,
				Pid:  1,
				Tid:  200, // The profiler thread (not the profiled thread!)
				ID:   "0x1",
				Args: json.RawMessage(`{
					"data": {
						"cpuProfile": {
							"nodes": [
								{"id": 1, "callFrame": {"functionName": "(root)", "scriptId": 0, "url": ""}},
								{"id": 2, "callFrame": {"functionName": "workerFunc", "scriptId": 1, "url": "worker.js", "lineNumber": 5}, "parent": 1}
							],
							"samples": [1, 2, 2, 2]
						},
						"timeDeltas": [100, 100, 100, 100]
					}
				}`),
			},
		},
	}

	profile, err := ConvertChromeToProfile(chrome)
	if err != nil {
		t.Fatalf("ConvertChromeToProfile() error = %v", err)
	}

	// Find the DedicatedWorker thread
	var workerThread *Thread
	var profilerThread *Thread
	for i := range profile.Threads {
		if profile.Threads[i].Name == "DedicatedWorker thread" {
			workerThread = &profile.Threads[i]
		}
		if profile.Threads[i].Name == "v8:ProfEvntProc" {
			profilerThread = &profile.Threads[i]
		}
	}

	if workerThread == nil {
		t.Fatal("Could not find DedicatedWorker thread")
	}

	// The samples should be on the worker thread, NOT the profiler thread
	if workerThread.Samples.Length != 4 {
		t.Errorf("Worker thread Samples.Length = %d, want 4", workerThread.Samples.Length)
	}

	// The profiler thread should have no samples (or it shouldn't exist with samples)
	if profilerThread != nil && profilerThread.Samples.Length > 0 {
		t.Errorf("Profiler thread should have 0 samples, got %d", profilerThread.Samples.Length)
	}

	// Verify the function name "workerFunc" is in the worker thread's func table
	foundWorkerFunc := false
	for _, nameIdx := range workerThread.FuncTable.Name {
		if nameIdx < len(workerThread.StringArray) && workerThread.StringArray[nameIdx] == "workerFunc" {
			foundWorkerFunc = true
			break
		}
	}
	if !foundWorkerFunc {
		t.Error("Expected to find 'workerFunc' in worker thread's FuncTable")
	}
}
