package parser

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestChromeEventPhaseConstants(t *testing.T) {
	// Verify phase constants match expected values
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"PhaseBegin", PhaseBegin, "B"},
		{"PhaseEnd", PhaseEnd, "E"},
		{"PhaseDuration", PhaseDuration, "X"},
		{"PhaseMetadata", PhaseMetadata, "M"},
		{"PhaseInstant", PhaseInstant, "I"},
		{"PhaseCounter", PhaseCounter, "C"},
		{"PhaseAsyncStart", PhaseAsyncStart, "S"},
		{"PhaseAsyncEnd", PhaseAsyncEnd, "F"},
		{"PhaseAsyncBegin", PhaseAsyncBegin, "b"},
		{"PhaseAsyncEnd2", PhaseAsyncEnd2, "e"},
		{"PhaseAsyncStep", PhaseAsyncStep, "n"},
		{"PhaseFlowStart", PhaseFlowStart, "s"},
		{"PhaseFlowEnd", PhaseFlowEnd, "f"},
		{"PhaseSample", PhaseSample, "P"},
		{"PhaseObject", PhaseObject, "O"},
		{"PhaseCreate", PhaseCreate, "N"},
		{"PhaseDestroy", PhaseDestroy, "D"},
		{"PhaseMark", PhaseMark, "R"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestChromeEventUnmarshal(t *testing.T) {
	// Test that ChromeEvent correctly unmarshals various event types
	tests := []struct {
		name     string
		json     string
		expected ChromeEvent
	}{
		{
			name: "Duration event (X phase)",
			json: `{"name":"FunctionCall","cat":"devtools.timeline","ph":"X","ts":1000000,"dur":5000,"pid":1,"tid":2}`,
			expected: ChromeEvent{
				Name: "FunctionCall",
				Cat:  "devtools.timeline",
				Ph:   "X",
				Ts:   1000000,
				Dur:  5000,
				Pid:  1,
				Tid:  2,
			},
		},
		{
			name: "Metadata event (M phase)",
			json: `{"name":"thread_name","cat":"__metadata","ph":"M","ts":0,"pid":1,"tid":2,"args":{"name":"CrRendererMain"}}`,
			expected: ChromeEvent{
				Name: "thread_name",
				Cat:  "__metadata",
				Ph:   "M",
				Ts:   0,
				Pid:  1,
				Tid:  2,
			},
		},
		{
			name: "Instant event (I phase)",
			json: `{"name":"LayoutShift","cat":"loading","ph":"I","ts":500000,"pid":1,"tid":1}`,
			expected: ChromeEvent{
				Name: "LayoutShift",
				Cat:  "loading",
				Ph:   "I",
				Ts:   500000,
				Pid:  1,
				Tid:  1,
			},
		},
		{
			name: "Event with string ID",
			json: `{"name":"AsyncTask","ph":"b","ts":1000,"pid":1,"tid":1,"id":"abc123"}`,
			expected: ChromeEvent{
				Name: "AsyncTask",
				Ph:   "b",
				Ts:   1000,
				Pid:  1,
				Tid:  1,
				ID:   "abc123",
			},
		},
		{
			name: "Event with numeric ID",
			json: `{"name":"AsyncTask","ph":"b","ts":1000,"pid":1,"tid":1,"id":12345}`,
			expected: ChromeEvent{
				Name: "AsyncTask",
				Ph:   "b",
				Ts:   1000,
				Pid:  1,
				Tid:  1,
				ID:   float64(12345), // JSON numbers unmarshal as float64
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event ChromeEvent
			if err := json.Unmarshal([]byte(tt.json), &event); err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if event.Name != tt.expected.Name {
				t.Errorf("Name = %q, want %q", event.Name, tt.expected.Name)
			}
			if event.Ph != tt.expected.Ph {
				t.Errorf("Ph = %q, want %q", event.Ph, tt.expected.Ph)
			}
			if event.Ts != tt.expected.Ts {
				t.Errorf("Ts = %v, want %v", event.Ts, tt.expected.Ts)
			}
			if event.Dur != tt.expected.Dur {
				t.Errorf("Dur = %v, want %v", event.Dur, tt.expected.Dur)
			}
			if event.Pid != tt.expected.Pid {
				t.Errorf("Pid = %v, want %v", event.Pid, tt.expected.Pid)
			}
			if event.Tid != tt.expected.Tid {
				t.Errorf("Tid = %v, want %v", event.Tid, tt.expected.Tid)
			}
		})
	}
}

func TestV8CPUProfileUnmarshal(t *testing.T) {
	jsonData := `{
		"nodes": [
			{"id": 1, "callFrame": {"functionName": "(root)", "scriptId": 0, "url": "", "lineNumber": -1, "columnNumber": -1}},
			{"id": 2, "callFrame": {"functionName": "main", "scriptId": "1", "url": "file://test.js", "lineNumber": 10, "columnNumber": 5}, "parent": 1}
		],
		"samples": [1, 2, 2, 1],
		"timeDeltas": [1000, 1000, 1000, 1000],
		"startTime": 0,
		"endTime": 4000
	}`

	var profile V8CPUProfile
	if err := json.Unmarshal([]byte(jsonData), &profile); err != nil {
		t.Fatalf("Failed to unmarshal V8CPUProfile: %v", err)
	}

	if len(profile.Nodes) != 2 {
		t.Errorf("Nodes count = %d, want 2", len(profile.Nodes))
	}
	if len(profile.Samples) != 4 {
		t.Errorf("Samples count = %d, want 4", len(profile.Samples))
	}
	if len(profile.TimeDeltas) != 4 {
		t.Errorf("TimeDeltas count = %d, want 4", len(profile.TimeDeltas))
	}

	// Check first node
	if profile.Nodes[0].ID != 1 {
		t.Errorf("Node[0].ID = %d, want 1", profile.Nodes[0].ID)
	}
	if profile.Nodes[0].CallFrame.FunctionName != "(root)" {
		t.Errorf("Node[0].CallFrame.FunctionName = %q, want (root)", profile.Nodes[0].CallFrame.FunctionName)
	}

	// Check second node with parent
	if profile.Nodes[1].Parent != 1 {
		t.Errorf("Node[1].Parent = %d, want 1", profile.Nodes[1].Parent)
	}
}

func TestLoadChromeProfile(t *testing.T) {
	// Create a minimal Chrome profile
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "chrome.json")

	chromeProfile := map[string]any{
		"metadata": map[string]any{
			"enhancedTraceVersion": 1,
			"source":               "Chrome DevTools",
			"startTime":            "2024-01-01T00:00:00Z",
		},
		"traceEvents": []map[string]any{
			{
				"name": "thread_name",
				"cat":  "__metadata",
				"ph":   "M",
				"ts":   0,
				"pid":  1,
				"tid":  1,
				"args": map[string]string{"name": "CrRendererMain"},
			},
			{
				"name": "FunctionCall",
				"cat":  "devtools.timeline",
				"ph":   "X",
				"ts":   1000000,
				"dur":  5000,
				"pid":  1,
				"tid":  1,
			},
		},
	}

	data, err := json.Marshal(chromeProfile)
	if err != nil {
		t.Fatalf("Failed to marshal profile: %v", err)
	}

	if err := os.WriteFile(profilePath, data, 0644); err != nil {
		t.Fatalf("Failed to write profile: %v", err)
	}

	profile, err := LoadChromeProfile(profilePath)
	if err != nil {
		t.Fatalf("LoadChromeProfile() error = %v", err)
	}

	if profile.Metadata.Source != "Chrome DevTools" {
		t.Errorf("Metadata.Source = %q, want %q", profile.Metadata.Source, "Chrome DevTools")
	}
	if len(profile.TraceEvents) != 2 {
		t.Errorf("TraceEvents count = %d, want 2", len(profile.TraceEvents))
	}
}

func TestLoadChromeProfile_Gzip(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "chrome.json.gz")

	chromeProfile := map[string]any{
		"metadata": map[string]any{
			"source": "Chrome DevTools",
		},
		"traceEvents": []map[string]any{
			{
				"name": "FunctionCall",
				"ph":   "X",
				"ts":   1000000,
				"dur":  5000,
				"pid":  1,
				"tid":  1,
			},
		},
	}

	data, err := json.Marshal(chromeProfile)
	if err != nil {
		t.Fatalf("Failed to marshal profile: %v", err)
	}

	f, err := os.Create(profilePath)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	gzWriter := gzip.NewWriter(f)
	if _, err := gzWriter.Write(data); err != nil {
		f.Close()
		t.Fatalf("Failed to write gzip data: %v", err)
	}
	gzWriter.Close()
	f.Close()

	profile, err := LoadChromeProfile(profilePath)
	if err != nil {
		t.Fatalf("LoadChromeProfile() error = %v", err)
	}

	if len(profile.TraceEvents) != 1 {
		t.Errorf("TraceEvents count = %d, want 1", len(profile.TraceEvents))
	}
}

func TestLoadChromeProfile_NonexistentFile(t *testing.T) {
	_, err := LoadChromeProfile("/nonexistent/path/file.json")
	if err == nil {
		t.Error("LoadChromeProfile() expected error for nonexistent file")
	}
}

func TestLoadChromeProfile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "invalid.json")

	if err := os.WriteFile(invalidPath, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	_, err := LoadChromeProfile(invalidPath)
	if err == nil {
		t.Error("LoadChromeProfile() expected error for invalid JSON")
	}
}

func TestThreadNameArgsUnmarshal(t *testing.T) {
	jsonData := `{"name": "CrRendererMain"}`

	var args ThreadNameArgs
	if err := json.Unmarshal([]byte(jsonData), &args); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if args.Name != "CrRendererMain" {
		t.Errorf("Name = %q, want %q", args.Name, "CrRendererMain")
	}
}

func TestProcessNameArgsUnmarshal(t *testing.T) {
	jsonData := `{"name": "Browser"}`

	var args ProcessNameArgs
	if err := json.Unmarshal([]byte(jsonData), &args); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if args.Name != "Browser" {
		t.Errorf("Name = %q, want %q", args.Name, "Browser")
	}
}
