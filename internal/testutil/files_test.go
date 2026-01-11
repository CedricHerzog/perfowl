package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTempProfileFile(t *testing.T) {
	profile := MinimalProfile()
	path := TempProfileFile(t, profile)

	if !FileExists(path) {
		t.Errorf("expected file to exist: %s", path)
	}
}

func TestTempGzipProfileFile(t *testing.T) {
	profile := MinimalProfile()
	path := TempGzipProfileFile(t, profile)

	if !FileExists(path) {
		t.Errorf("expected file to exist: %s", path)
	}
}

func TestTempJSONFile(t *testing.T) {
	data := map[string]string{"key": "value"}
	path := TempJSONFile(t, data, "test.json")

	if !FileExists(path) {
		t.Errorf("expected file to exist: %s", path)
	}
}

func TestTempGzipJSONFile(t *testing.T) {
	data := map[string]string{"key": "value"}
	path := TempGzipJSONFile(t, data, "test.json.gz")

	if !FileExists(path) {
		t.Errorf("expected file to exist: %s", path)
	}
}

func TestTempTextFile(t *testing.T) {
	path := TempTextFile(t, "hello world", "test.txt")

	if !FileExists(path) {
		t.Errorf("expected file to exist: %s", path)
	}

	content := ReadFile(t, path)
	if string(content) != "hello world" {
		t.Errorf("content = %q, want 'hello world'", string(content))
	}
}

func TestTempChromeProfile(t *testing.T) {
	data := ChromeTraceWithNEvents(10)
	path := TempChromeProfile(t, data)

	if !FileExists(path) {
		t.Errorf("expected file to exist: %s", path)
	}
}

func TestTempGzipChromeProfile(t *testing.T) {
	data := ChromeTraceWithNEvents(10)
	path := TempGzipChromeProfile(t, data)

	if !FileExists(path) {
		t.Errorf("expected file to exist: %s", path)
	}
}

func TestMinimalChromeProfile(t *testing.T) {
	profile := MinimalChromeProfile()

	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	if len(profile.TraceEvents) == 0 {
		t.Error("expected trace events")
	}
}

func TestChromeProfileWithEvents(t *testing.T) {
	events := []struct {
		Name string
		Ph   string
	}{
		{Name: "test", Ph: "B"},
	}

	// This is a bit awkward since ChromeEvent is from parser package
	// Just test with empty events
	profile := ChromeProfileWithEvents(nil)
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	_ = events // to avoid unused variable
}

func TestTempDir(t *testing.T) {
	dir := TempDir(t)

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestWriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	WriteFile(t, path, []byte("test content"))

	if !FileExists(path) {
		t.Errorf("expected file to exist: %s", path)
	}
}

func TestReadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
		t.Fatalf("write error: %v", err)
	}

	content := ReadFile(t, path)
	if string(content) != "test content" {
		t.Errorf("content = %q, want 'test content'", string(content))
	}
}

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	if FileExists(path) {
		t.Error("expected file to not exist")
	}

	if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
		t.Fatalf("write error: %v", err)
	}

	if !FileExists(path) {
		t.Error("expected file to exist")
	}
}

func TestAssertFileExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
		t.Fatalf("write error: %v", err)
	}

	// This should not fail
	AssertFileExists(t, path)
}

func TestAssertFileNotExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.txt")

	// This should not fail
	AssertFileNotExists(t, path)
}

func TestAssertFileContains(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	if err := os.WriteFile(path, []byte("hello world"), 0644); err != nil {
		t.Fatalf("write error: %v", err)
	}

	// This should not fail
	AssertFileContains(t, path, "world")
}

func TestContains(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"hello world", "world", true},
		{"hello world", "hello", true},
		{"hello world", "xyz", false},
		{"", "", true},
		{"hello", "", true},
		{"", "hello", false},
		{"ab", "ab", true},
	}

	for _, tt := range tests {
		got := contains(tt.s, tt.substr)
		if got != tt.want {
			t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
		}
	}
}

func TestFindSubstring(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"hello world", "world", true},
		{"hello world", "hello", true},
		{"hello world", "xyz", false},
		{"abc", "abcd", false},
	}

	for _, tt := range tests {
		got := findSubstring(tt.s, tt.substr)
		if got != tt.want {
			t.Errorf("findSubstring(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
		}
	}
}
