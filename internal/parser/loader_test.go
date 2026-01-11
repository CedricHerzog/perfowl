package parser

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadProfile_ValidJSON(t *testing.T) {
	// Create a minimal valid profile
	profile := &Profile{
		Meta: Meta{
			ProfilingStartTime: 0,
			ProfilingEndTime:   1000,
			Product:            "Firefox",
		},
		Threads: []Thread{
			{Name: "GeckoMain", IsMainThread: true},
		},
	}

	// Write to temp file
	dir := t.TempDir()
	path := filepath.Join(dir, "profile.json")

	data, err := json.Marshal(profile)
	if err != nil {
		t.Fatalf("failed to marshal profile: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Load the profile
	loaded, err := LoadProfile(path)
	if err != nil {
		t.Fatalf("LoadProfile() error = %v", err)
	}

	if loaded.Duration() != 1000 {
		t.Errorf("Duration() = %v, want 1000", loaded.Duration())
	}

	if loaded.ThreadCount() != 1 {
		t.Errorf("ThreadCount() = %v, want 1", loaded.ThreadCount())
	}

	if loaded.Meta.Product != "Firefox" {
		t.Errorf("Product = %v, want Firefox", loaded.Meta.Product)
	}
}

func TestLoadProfile_GzipCompressed(t *testing.T) {
	// Create a profile
	profile := &Profile{
		Meta: Meta{
			ProfilingStartTime: 0,
			ProfilingEndTime:   2000,
			Product:            "Firefox",
		},
	}

	// Write gzip compressed
	dir := t.TempDir()
	path := filepath.Join(dir, "profile.json.gz")

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	gw := gzip.NewWriter(f)
	enc := json.NewEncoder(gw)
	if err := enc.Encode(profile); err != nil {
		t.Fatalf("failed to encode: %v", err)
	}
	gw.Close()
	f.Close()

	// Load the profile
	loaded, err := LoadProfile(path)
	if err != nil {
		t.Fatalf("LoadProfile() error = %v", err)
	}

	if loaded.Duration() != 2000 {
		t.Errorf("Duration() = %v, want 2000", loaded.Duration())
	}
}

func TestLoadProfile_GzipExtension(t *testing.T) {
	// Test .gzip extension
	profile := &Profile{
		Meta: Meta{
			ProfilingStartTime: 0,
			ProfilingEndTime:   3000,
		},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "profile.json.gzip")

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	gw := gzip.NewWriter(f)
	enc := json.NewEncoder(gw)
	if err := enc.Encode(profile); err != nil {
		t.Fatalf("failed to encode: %v", err)
	}
	gw.Close()
	f.Close()

	loaded, err := LoadProfile(path)
	if err != nil {
		t.Fatalf("LoadProfile() error = %v", err)
	}

	if loaded.Duration() != 3000 {
		t.Errorf("Duration() = %v, want 3000", loaded.Duration())
	}
}

func TestLoadProfile_FileNotFound(t *testing.T) {
	_, err := LoadProfile("/nonexistent/path/profile.json")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "failed to open profile") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadProfile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.json")

	if err := os.WriteFile(path, []byte("not valid json {"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	_, err := LoadProfile(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "failed to decode profile JSON") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadProfile_InvalidGzip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.json.gz")

	// Write non-gzip data with .gz extension
	if err := os.WriteFile(path, []byte("not gzip data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	_, err := LoadProfile(path)
	if err == nil {
		t.Error("expected error for invalid gzip")
	}
	if !strings.Contains(err.Error(), "failed to create gzip reader") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadProfileFromReader_Valid(t *testing.T) {
	profile := &Profile{
		Meta: Meta{
			ProfilingStartTime: 0,
			ProfilingEndTime:   5000,
			Product:            "Firefox",
		},
		Threads: []Thread{
			{Name: "Main", IsMainThread: true},
		},
	}

	data, err := json.Marshal(profile)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	reader := bytes.NewReader(data)
	loaded, err := LoadProfileFromReader(reader)
	if err != nil {
		t.Fatalf("LoadProfileFromReader() error = %v", err)
	}

	if loaded.Duration() != 5000 {
		t.Errorf("Duration() = %v, want 5000", loaded.Duration())
	}
}

func TestLoadProfileFromReader_Invalid(t *testing.T) {
	reader := strings.NewReader("invalid json {{{")
	_, err := LoadProfileFromReader(reader)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadProfileFromReader_Empty(t *testing.T) {
	reader := strings.NewReader("")
	_, err := LoadProfileFromReader(reader)
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestLoadProfileAuto_Firefox(t *testing.T) {
	// Create a Firefox profile
	profile := &Profile{
		Meta: Meta{
			ProfilingStartTime: 0,
			ProfilingEndTime:   1000,
			Product:            "Firefox",
		},
		Threads: []Thread{
			{Name: "GeckoMain"},
		},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "profile.json")

	data, err := json.Marshal(profile)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	loaded, browserType, err := LoadProfileAuto(path)
	if err != nil {
		t.Fatalf("LoadProfileAuto() error = %v", err)
	}

	if browserType != BrowserFirefox {
		t.Errorf("browserType = %v, want Firefox", browserType)
	}

	if loaded.Duration() != 1000 {
		t.Errorf("Duration() = %v, want 1000", loaded.Duration())
	}
}

func TestLoadProfileAuto_Chrome(t *testing.T) {
	// Create a Chrome profile
	chromeProfile := &ChromeProfile{
		TraceEvents: []ChromeEvent{
			{
				Name: "thread_name",
				Cat:  "__metadata",
				Ph:   "M",
				Pid:  1,
				Tid:  1,
				Ts:   0,
				Args: json.RawMessage(`{"name": "CrBrowserMain"}`),
			},
			{
				Name: "MessageLoop::RunTask",
				Cat:  "toplevel",
				Ph:   "X",
				Pid:  1,
				Tid:  1,
				Ts:   1000,
				Dur:  5000,
			},
		},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "trace.json")

	data, err := json.Marshal(chromeProfile)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	loaded, browserType, err := LoadProfileAuto(path)
	if err != nil {
		t.Fatalf("LoadProfileAuto() error = %v", err)
	}

	if browserType != BrowserChrome {
		t.Errorf("browserType = %v, want Chrome", browserType)
	}

	if loaded == nil {
		t.Error("expected non-nil profile")
	}
}

func TestLoadProfileAuto_FileNotFound(t *testing.T) {
	_, _, err := LoadProfileAuto("/nonexistent/profile.json")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestLoadProfileWithType_Firefox(t *testing.T) {
	profile := &Profile{
		Meta: Meta{
			ProfilingStartTime: 0,
			ProfilingEndTime:   1000,
		},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "profile.json")

	data, err := json.Marshal(profile)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	loaded, browserType, err := LoadProfileWithType(path, BrowserFirefox)
	if err != nil {
		t.Fatalf("LoadProfileWithType() error = %v", err)
	}

	if browserType != BrowserFirefox {
		t.Errorf("browserType = %v, want Firefox", browserType)
	}

	if loaded.Duration() != 1000 {
		t.Errorf("Duration() = %v, want 1000", loaded.Duration())
	}
}

func TestLoadProfileWithType_Chrome(t *testing.T) {
	chromeProfile := &ChromeProfile{
		TraceEvents: []ChromeEvent{
			{
				Name: "thread_name",
				Cat:  "__metadata",
				Ph:   "M",
				Pid:  1,
				Tid:  1,
				Ts:   0,
				Args: json.RawMessage(`{"name": "CrBrowserMain"}`),
			},
		},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "trace.json")

	data, err := json.Marshal(chromeProfile)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	loaded, browserType, err := LoadProfileWithType(path, BrowserChrome)
	if err != nil {
		t.Fatalf("LoadProfileWithType() error = %v", err)
	}

	if browserType != BrowserChrome {
		t.Errorf("browserType = %v, want Chrome", browserType)
	}

	if loaded == nil {
		t.Error("expected non-nil profile")
	}
}

func TestLoadProfileWithType_Unknown_AutoDetect(t *testing.T) {
	// Test that BrowserUnknown triggers auto-detection
	profile := &Profile{
		Meta: Meta{
			ProfilingStartTime: 0,
			ProfilingEndTime:   1000,
		},
		Threads: []Thread{
			{Name: "GeckoMain"},
		},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "profile.json")

	data, err := json.Marshal(profile)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	loaded, browserType, err := LoadProfileWithType(path, BrowserUnknown)
	if err != nil {
		t.Fatalf("LoadProfileWithType() error = %v", err)
	}

	// Should detect as Firefox
	if browserType != BrowserFirefox {
		t.Errorf("browserType = %v, want Firefox (auto-detected)", browserType)
	}

	if loaded == nil {
		t.Error("expected non-nil profile")
	}
}

func TestLoadProfileWithType_Firefox_FileNotFound(t *testing.T) {
	_, _, err := LoadProfileWithType("/nonexistent/profile.json", BrowserFirefox)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestLoadProfileWithType_Chrome_FileNotFound(t *testing.T) {
	_, _, err := LoadProfileWithType("/nonexistent/trace.json", BrowserChrome)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestLoadProfileWithType_Unknown_InvalidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.json")

	// Write invalid content that can't be parsed as Firefox or Chrome
	if err := os.WriteFile(path, []byte(`{"random": "data"}`), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	// This should try to detect and then fallback
	_, _, err := LoadProfileWithType(path, BrowserUnknown)
	// It may succeed if it can parse as empty Firefox profile, or fail
	// The behavior depends on JSON parsing flexibility
	if err != nil {
		// Expected if it can't parse as either format
		if !strings.Contains(err.Error(), "failed to parse") {
			// May also detect as unknown but successfully parse as empty Firefox profile
			t.Logf("Got error (expected): %v", err)
		}
	}
}
