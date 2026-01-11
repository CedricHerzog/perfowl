package testutil

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

// TempProfileFile creates a temporary JSON profile file and returns its path.
// The file is automatically cleaned up when the test finishes.
func TempProfileFile(t *testing.T, profile *parser.Profile) string {
	t.Helper()
	return TempJSONFile(t, profile, "profile.json")
}

// TempGzipProfileFile creates a temporary gzip-compressed JSON profile file.
// The file is automatically cleaned up when the test finishes.
func TempGzipProfileFile(t *testing.T, profile *parser.Profile) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "profile.json.gz")

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	enc := json.NewEncoder(gw)
	if err := enc.Encode(profile); err != nil {
		t.Fatalf("failed to encode profile: %v", err)
	}

	return path
}

// TempJSONFile creates a temporary JSON file with the given content.
// The file is automatically cleaned up when the test finishes.
func TempJSONFile(t *testing.T, content interface{}, filename string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, filename)

	data, err := json.Marshal(content)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	return path
}

// TempGzipJSONFile creates a temporary gzip-compressed JSON file.
// The file is automatically cleaned up when the test finishes.
func TempGzipJSONFile(t *testing.T, content interface{}, filename string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, filename)

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	enc := json.NewEncoder(gw)
	if err := enc.Encode(content); err != nil {
		t.Fatalf("failed to encode JSON: %v", err)
	}

	return path
}

// TempTextFile creates a temporary text file with the given content.
// The file is automatically cleaned up when the test finishes.
func TempTextFile(t *testing.T, content string, filename string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, filename)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	return path
}

// TempChromeProfile creates a temporary Chrome profile file.
// The file is automatically cleaned up when the test finishes.
func TempChromeProfile(t *testing.T, profile *parser.ChromeProfile) string {
	t.Helper()
	return TempJSONFile(t, profile, "chrome-profile.json")
}

// TempGzipChromeProfile creates a temporary gzip-compressed Chrome profile file.
// The file is automatically cleaned up when the test finishes.
func TempGzipChromeProfile(t *testing.T, profile *parser.ChromeProfile) string {
	t.Helper()
	return TempGzipJSONFile(t, profile, "chrome-profile.json.gz")
}

// MinimalChromeProfile returns a minimal valid Chrome profile.
func MinimalChromeProfile() *parser.ChromeProfile {
	return &parser.ChromeProfile{
		TraceEvents: []parser.ChromeEvent{
			{
				Name:  "thread_name",
				Cat:   "__metadata",
				Ph:    "M",
				Pid:   1,
				Tid:   1,
				Ts:    0,
				Args:  json.RawMessage(`{"name": "CrBrowserMain"}`),
				Scope: "",
			},
		},
		Metadata: parser.ChromeMetadata{
			Source: "devtools",
		},
	}
}

// ChromeProfileWithEvents returns a Chrome profile with the given events.
func ChromeProfileWithEvents(events []parser.ChromeEvent) *parser.ChromeProfile {
	return &parser.ChromeProfile{
		TraceEvents: events,
		Metadata: parser.ChromeMetadata{
			Source: "devtools",
		},
	}
}

// TempDir creates a temporary directory and returns its path.
// The directory is automatically cleaned up when the test finishes.
func TempDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// WriteFile writes content to a file at the given path.
func WriteFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

// ReadFile reads and returns the content of a file.
func ReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}
	return data
}

// FileExists returns true if the file exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// AssertFileExists fails the test if the file doesn't exist.
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	if !FileExists(path) {
		t.Errorf("expected file to exist: %s", path)
	}
}

// AssertFileNotExists fails the test if the file exists.
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if FileExists(path) {
		t.Errorf("expected file to not exist: %s", path)
	}
}

// AssertFileContains fails the test if the file doesn't contain the substring.
func AssertFileContains(t *testing.T, path, substring string) {
	t.Helper()
	content := string(ReadFile(t, path))
	if !contains(content, substring) {
		t.Errorf("expected file %s to contain %q", path, substring)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
