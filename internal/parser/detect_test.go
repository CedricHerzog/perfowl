package parser

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParseBrowserType(t *testing.T) {
	tests := []struct {
		input    string
		expected BrowserType
	}{
		{"firefox", BrowserFirefox},
		{"Firefox", BrowserFirefox},
		{"FIREFOX", BrowserFirefox},
		{"chrome", BrowserChrome},
		{"Chrome", BrowserChrome},
		{"CHROME", BrowserChrome},
		{"auto", BrowserUnknown},
		{"", BrowserUnknown},
		{"safari", BrowserUnknown},
		{"edge", BrowserUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseBrowserType(tt.input)
			if result != tt.expected {
				t.Errorf("ParseBrowserType(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDetectFromPeek(t *testing.T) {
	tests := []struct {
		name     string
		peek     *profilePeek
		expected BrowserType
	}{
		{
			name: "Firefox profile with meta and threads",
			peek: &profilePeek{
				Meta: &struct {
					Product string `json:"product"`
				}{Product: "Firefox"},
				Threads: []json.RawMessage{[]byte(`{}`)},
			},
			expected: BrowserFirefox,
		},
		{
			name: "Chrome profile with traceEvents",
			peek: &profilePeek{
				TraceEvents: []json.RawMessage{[]byte(`{}`)},
			},
			expected: BrowserChrome,
		},
		{
			name:     "Empty profile",
			peek:     &profilePeek{},
			expected: BrowserUnknown,
		},
		{
			name: "Firefox meta without threads",
			peek: &profilePeek{
				Meta: &struct {
					Product string `json:"product"`
				}{Product: "Firefox"},
			},
			expected: BrowserUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectFromPeek(tt.peek)
			if result != tt.expected {
				t.Errorf("detectFromPeek() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDetectBrowserType_FirefoxFile(t *testing.T) {
	// Create a temp Firefox profile
	tmpDir := t.TempDir()
	firefoxPath := filepath.Join(tmpDir, "firefox.json")

	firefoxProfile := map[string]any{
		"meta": map[string]any{
			"product": "Firefox",
			"version": 1,
		},
		"threads": []map[string]any{
			{"name": "GeckoMain"},
		},
	}

	data, err := json.Marshal(firefoxProfile)
	if err != nil {
		t.Fatalf("Failed to marshal Firefox profile: %v", err)
	}

	if err := os.WriteFile(firefoxPath, data, 0644); err != nil {
		t.Fatalf("Failed to write Firefox profile: %v", err)
	}

	browserType, err := DetectBrowserType(firefoxPath)
	if err != nil {
		t.Fatalf("DetectBrowserType() error = %v", err)
	}

	if browserType != BrowserFirefox {
		t.Errorf("DetectBrowserType() = %q, want %q", browserType, BrowserFirefox)
	}
}

func TestDetectBrowserType_ChromeFile(t *testing.T) {
	// Create a temp Chrome profile
	tmpDir := t.TempDir()
	chromePath := filepath.Join(tmpDir, "chrome.json")

	chromeProfile := map[string]any{
		"metadata": map[string]any{
			"source": "Chrome DevTools",
		},
		"traceEvents": []map[string]any{
			{"name": "thread_name", "ph": "M", "pid": 1, "tid": 1},
		},
	}

	data, err := json.Marshal(chromeProfile)
	if err != nil {
		t.Fatalf("Failed to marshal Chrome profile: %v", err)
	}

	if err := os.WriteFile(chromePath, data, 0644); err != nil {
		t.Fatalf("Failed to write Chrome profile: %v", err)
	}

	browserType, err := DetectBrowserType(chromePath)
	if err != nil {
		t.Fatalf("DetectBrowserType() error = %v", err)
	}

	if browserType != BrowserChrome {
		t.Errorf("DetectBrowserType() = %q, want %q", browserType, BrowserChrome)
	}
}

func TestDetectBrowserType_GzipFile(t *testing.T) {
	// Create a gzip-compressed Firefox profile
	tmpDir := t.TempDir()
	gzPath := filepath.Join(tmpDir, "firefox.json.gz")

	firefoxProfile := map[string]any{
		"meta": map[string]any{
			"product": "Firefox",
		},
		"threads": []map[string]any{
			{"name": "GeckoMain"},
		},
	}

	data, err := json.Marshal(firefoxProfile)
	if err != nil {
		t.Fatalf("Failed to marshal profile: %v", err)
	}

	f, err := os.Create(gzPath)
	if err != nil {
		t.Fatalf("Failed to create gzip file: %v", err)
	}

	gzWriter := gzip.NewWriter(f)
	if _, err := gzWriter.Write(data); err != nil {
		f.Close()
		t.Fatalf("Failed to write gzip data: %v", err)
	}
	gzWriter.Close()
	f.Close()

	browserType, err := DetectBrowserType(gzPath)
	if err != nil {
		t.Fatalf("DetectBrowserType() error = %v", err)
	}

	if browserType != BrowserFirefox {
		t.Errorf("DetectBrowserType() = %q, want %q", browserType, BrowserFirefox)
	}
}

func TestDetectBrowserType_NonexistentFile(t *testing.T) {
	_, err := DetectBrowserType("/nonexistent/path/file.json")
	if err == nil {
		t.Error("DetectBrowserType() expected error for nonexistent file")
	}
}

func TestDetectBrowserType_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "invalid.json")

	if err := os.WriteFile(invalidPath, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	_, err := DetectBrowserType(invalidPath)
	if err == nil {
		t.Error("DetectBrowserType() expected error for invalid JSON")
	}
}
