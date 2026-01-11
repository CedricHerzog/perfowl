package mcp

import (
	"strings"
	"testing"

	"github.com/CedricHerzog/perfowl/internal/format/toon"
	"github.com/CedricHerzog/perfowl/internal/parser"
	"github.com/CedricHerzog/perfowl/internal/testutil"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestTOONOutputFormat(t *testing.T) {
	// Load a profile
	profile, _, err := parser.LoadProfileAuto("../../profiles/firefox/new-4-core.json.gz")
	if err != nil {
		t.Skipf("Skipping test: could not load profile: %v", err)
	}

	// Build summary using the same function as the MCP handler
	summary := buildSummary(profile)

	// Encode to TOON
	output, err := toon.Encode(summary)
	if err != nil {
		t.Fatalf("Failed to encode summary: %v", err)
	}

	// Verify it's TOON format, not JSON
	if strings.HasPrefix(strings.TrimSpace(output), "{") {
		t.Errorf("Output looks like JSON, expected TOON format:\n%s", output)
	}

	// Verify expected TOON fields are present
	if !strings.Contains(output, "duration_seconds:") {
		t.Errorf("Missing 'duration_seconds:' field in output:\n%s", output)
	}
	if !strings.Contains(output, "platform:") {
		t.Errorf("Missing 'platform:' field in output:\n%s", output)
	}
	if !strings.Contains(output, "features[") {
		t.Errorf("Missing 'features[' array in output:\n%s", output)
	}

	t.Logf("TOON Output:\n%s", output)
}

func TestTOONTabularFormat(t *testing.T) {
	// Test that arrays of structs use tabular format
	type Item struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	type Report struct {
		Items []Item `json:"items"`
	}

	report := Report{
		Items: []Item{
			{Name: "first", Value: 1},
			{Name: "second", Value: 2},
		},
	}

	output, err := toon.Encode(report)
	if err != nil {
		t.Fatalf("Failed to encode report: %v", err)
	}

	// Should use tabular format
	if !strings.Contains(output, "items[2]{name,value}:") {
		t.Errorf("Expected tabular header 'items[2]{name,value}:', got:\n%s", output)
	}

	t.Logf("TOON Tabular Output:\n%s", output)
}

func TestNewServer(t *testing.T) {
	server := NewServer()

	if server == nil {
		t.Error("expected non-nil server")
	}
	if server.server == nil {
		t.Error("expected non-nil internal MCP server")
	}
}

func TestBuildSummary_BasicFields(t *testing.T) {
	// Note: WithMeta must come before WithDuration since WithMeta replaces the entire Meta
	profile := testutil.NewProfileBuilder().
		WithMeta(parser.Meta{
			Platform:     "Linux",
			OSCPU:        "x86_64",
			Product:      "Firefox",
			AppBuildID:   "20240101",
			CPUName:      "Intel",
			PhysicalCPUs: 4,
			LogicalCPUs:  8,
		}).
		WithDuration(5000).
		WithThread(testutil.NewThreadBuilder("GeckoMain").AsMainThread().Build()).
		WithThread(testutil.NewThreadBuilder("DOM Worker").Build()).
		Build()

	summary := buildSummary(profile)

	if summary.DurationSeconds != 5 {
		t.Errorf("DurationSeconds = %v, want 5", summary.DurationSeconds)
	}
	if summary.Platform != "Linux" {
		t.Errorf("Platform = %v, want Linux", summary.Platform)
	}
	if summary.Product != "Firefox" {
		t.Errorf("Product = %v, want Firefox", summary.Product)
	}
	if summary.ThreadCount != 2 {
		t.Errorf("ThreadCount = %v, want 2", summary.ThreadCount)
	}
	if summary.MainThreadCount != 1 {
		t.Errorf("MainThreadCount = %v, want 1", summary.MainThreadCount)
	}
	if summary.PhysicalCPUs != 4 {
		t.Errorf("PhysicalCPUs = %v, want 4", summary.PhysicalCPUs)
	}
	if summary.LogicalCPUs != 8 {
		t.Errorf("LogicalCPUs = %v, want 8", summary.LogicalCPUs)
	}
}

func TestBuildSummary_WithExtensions(t *testing.T) {
	profile := testutil.ProfileWithExtensions()

	summary := buildSummary(profile)

	if summary.ExtensionCount != 2 {
		t.Errorf("ExtensionCount = %v, want 2", summary.ExtensionCount)
	}
	if len(summary.Extensions) != 2 {
		t.Errorf("Extensions length = %v, want 2", len(summary.Extensions))
	}
}

func TestBuildSummary_Markers(t *testing.T) {
	mb := testutil.NewMarkerBuilder()
	mb.AddGCMajor(0, 10)
	mb.AddGCMajor(100, 20)
	markers, strings := mb.Build()

	profile := testutil.NewProfileBuilder().
		WithDuration(1000).
		WithThread(testutil.NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithMarkers(markers).
			WithStringArray(strings).
			Build()).
		Build()

	summary := buildSummary(profile)

	if summary.TotalMarkers != 2 {
		t.Errorf("TotalMarkers = %v, want 2", summary.TotalMarkers)
	}
}

func TestBuildSummary_Samples(t *testing.T) {
	sb := testutil.NewSamplesBuilder()
	for i := 0; i < 100; i++ {
		sb.AddSample(0, float64(i))
	}

	profile := testutil.NewProfileBuilder().
		WithDuration(1000).
		WithThread(testutil.NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithSamples(sb.Build()).
			Build()).
		Build()

	summary := buildSummary(profile)

	if summary.TotalSamples != 100 {
		t.Errorf("TotalSamples = %v, want 100", summary.TotalSamples)
	}
}

func TestSplitAndTrim_Basic(t *testing.T) {
	result := splitAndTrim("a, b, c")

	if len(result) != 3 {
		t.Errorf("expected 3 items, got %d", len(result))
	}
	if result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestSplitAndTrim_WithWhitespace(t *testing.T) {
	result := splitAndTrim("  DOM  ,  Layout  ,  JavaScript  ")

	if len(result) != 3 {
		t.Errorf("expected 3 items, got %d", len(result))
	}
	if result[0] != "DOM" {
		t.Errorf("expected 'DOM', got '%s'", result[0])
	}
	if result[1] != "Layout" {
		t.Errorf("expected 'Layout', got '%s'", result[1])
	}
	if result[2] != "JavaScript" {
		t.Errorf("expected 'JavaScript', got '%s'", result[2])
	}
}

func TestSplitAndTrim_Empty(t *testing.T) {
	result := splitAndTrim("")

	if len(result) != 0 {
		t.Errorf("expected empty result, got %d items", len(result))
	}
}

func TestSplitAndTrim_SingleItem(t *testing.T) {
	result := splitAndTrim("single")

	if len(result) != 1 {
		t.Errorf("expected 1 item, got %d", len(result))
	}
	if result[0] != "single" {
		t.Errorf("expected 'single', got '%s'", result[0])
	}
}

func TestSplitAndTrim_EmptyParts(t *testing.T) {
	result := splitAndTrim("a,,b,  ,c")

	if len(result) != 3 {
		t.Errorf("expected 3 items (empty parts filtered), got %d", len(result))
	}
}

func TestProfileSummary_Fields(t *testing.T) {
	summary := ProfileSummary{
		DurationSeconds: 10.5,
		Platform:        "Linux",
		OSCPU:           "x86_64",
		Product:         "Firefox",
		BuildID:         "20240101",
		CPUName:         "Intel",
		PhysicalCPUs:    4,
		LogicalCPUs:     8,
		ThreadCount:     10,
		MainThreadCount: 1,
		ExtensionCount:  2,
		Extensions:      map[string]string{"ext1": "Extension 1"},
		Features:        []string{"feature1", "feature2"},
		TotalMarkers:    100,
		TotalSamples:    5000,
	}

	// Verify struct can be encoded
	output, err := toon.Encode(summary)
	if err != nil {
		t.Fatalf("failed to encode ProfileSummary: %v", err)
	}

	if !strings.Contains(output, "duration_seconds:") {
		t.Error("missing duration_seconds in output")
	}
	if !strings.Contains(output, "platform:") {
		t.Error("missing platform in output")
	}
}

func TestBuildSummary_EmptyProfile(t *testing.T) {
	profile := testutil.MinimalProfile()

	summary := buildSummary(profile)

	if summary.ThreadCount != 0 {
		t.Errorf("ThreadCount = %v, want 0", summary.ThreadCount)
	}
	if summary.MainThreadCount != 0 {
		t.Errorf("MainThreadCount = %v, want 0", summary.MainThreadCount)
	}
	if summary.TotalMarkers != 0 {
		t.Errorf("TotalMarkers = %v, want 0", summary.TotalMarkers)
	}
	if summary.TotalSamples != 0 {
		t.Errorf("TotalSamples = %v, want 0", summary.TotalSamples)
	}
}

// Handler Tests - using temp profile files

func TestHandleGetSummary_Success(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{"path": path})

	result, err := server.handleGetSummary(nil, req)

	if err != nil {
		t.Fatalf("handleGetSummary error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleGetSummary_MissingPath(t *testing.T) {
	server := NewServer()
	req := mockRequest(map[string]any{})

	_, err := server.handleGetSummary(nil, req)

	if err == nil {
		t.Error("expected error for missing path")
	}
	if !strings.Contains(err.Error(), "path is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHandleGetSummary_FileNotFound(t *testing.T) {
	server := NewServer()
	req := mockRequest(map[string]any{"path": "/nonexistent/profile.json"})

	_, err := server.handleGetSummary(nil, req)

	if err == nil {
		t.Error("expected error for file not found")
	}
	if !strings.Contains(err.Error(), "failed to load profile") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHandleGetBottlenecks_Success(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{"path": path})

	result, err := server.handleGetBottlenecks(nil, req)

	if err != nil {
		t.Fatalf("handleGetBottlenecks error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleGetBottlenecks_WithSeverityFilter(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{
		"path":         path,
		"min_severity": "high",
	})

	result, err := server.handleGetBottlenecks(nil, req)

	if err != nil {
		t.Fatalf("handleGetBottlenecks error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleGetMarkers_Success(t *testing.T) {
	mb := testutil.NewMarkerBuilder()
	mb.AddGCMajor(0, 10)
	mb.AddDOMEvent("click", 100)
	markers, strings := mb.Build()

	profile := testutil.NewProfileBuilder().
		WithDuration(1000).
		WithThread(testutil.NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithMarkers(markers).
			WithStringArray(strings).
			Build()).
		Build()
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{"path": path})

	result, err := server.handleGetMarkers(nil, req)

	if err != nil {
		t.Fatalf("handleGetMarkers error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleGetMarkers_WithFilters(t *testing.T) {
	mb := testutil.NewMarkerBuilder()
	mb.AddGCMajor(0, 50)
	mb.AddGCMajor(100, 30)
	markers, strings := mb.Build()

	profile := testutil.NewProfileBuilder().
		WithDuration(1000).
		WithCategories(testutil.DefaultCategories()).
		WithThread(testutil.NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithMarkers(markers).
			WithStringArray(strings).
			Build()).
		Build()
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{
		"path":         path,
		"type":         "GCMajor",
		"category":     "GC",
		"min_duration": float64(20),
		"limit":        float64(10),
	})

	result, err := server.handleGetMarkers(nil, req)

	if err != nil {
		t.Fatalf("handleGetMarkers error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleAnalyzeExtension_Success(t *testing.T) {
	profile := testutil.ProfileWithExtensions()
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{"path": path})

	result, err := server.handleAnalyzeExtension(nil, req)

	if err != nil {
		t.Fatalf("handleAnalyzeExtension error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleAnalyzeExtension_WithFilter(t *testing.T) {
	profile := testutil.ProfileWithExtensions()
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{
		"path":         path,
		"extension_id": "ext1@test.com",
	})

	result, err := server.handleAnalyzeExtension(nil, req)

	if err != nil {
		t.Fatalf("handleAnalyzeExtension error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleAnalyzeProfile_Success(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{"path": path})

	result, err := server.handleAnalyzeProfile(nil, req)

	if err != nil {
		t.Fatalf("handleAnalyzeProfile error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleGetCallTree_Success(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{"path": path})

	result, err := server.handleGetCallTree(nil, req)

	if err != nil {
		t.Fatalf("handleGetCallTree error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleGetCallTree_WithThreadFilter(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{
		"path":   path,
		"thread": "GeckoMain",
		"limit":  float64(10),
	})

	result, err := server.handleGetCallTree(nil, req)

	if err != nil {
		t.Fatalf("handleGetCallTree error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleGetCategoryBreakdown_Success(t *testing.T) {
	profile := testutil.ProfileWithCategories()
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{"path": path})

	result, err := server.handleGetCategoryBreakdown(nil, req)

	if err != nil {
		t.Fatalf("handleGetCategoryBreakdown error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleGetCategoryBreakdown_WithThreadFilter(t *testing.T) {
	profile := testutil.ProfileWithCategories()
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{
		"path":   path,
		"thread": "GeckoMain",
	})

	result, err := server.handleGetCategoryBreakdown(nil, req)

	if err != nil {
		t.Fatalf("handleGetCategoryBreakdown error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleGetThreadAnalysis_Success(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{"path": path})

	result, err := server.handleGetThreadAnalysis(nil, req)

	if err != nil {
		t.Fatalf("handleGetThreadAnalysis error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleCompareProfiles_Success(t *testing.T) {
	profile1 := testutil.ProfileWithWorkers(2)
	profile2 := testutil.ProfileWithWorkers(4)
	path1 := testutil.TempProfileFile(t, profile1)
	path2 := testutil.TempProfileFile(t, profile2)

	server := NewServer()
	req := mockRequest(map[string]any{
		"baseline":   path1,
		"comparison": path2,
	})

	result, err := server.handleCompareProfiles(nil, req)

	if err != nil {
		t.Fatalf("handleCompareProfiles error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleCompareProfiles_MissingBaseline(t *testing.T) {
	server := NewServer()
	req := mockRequest(map[string]any{
		"comparison": "/some/path.json",
	})

	_, err := server.handleCompareProfiles(nil, req)

	if err == nil {
		t.Error("expected error for missing baseline")
	}
	if !strings.Contains(err.Error(), "baseline path is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHandleCompareProfiles_MissingComparison(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{
		"baseline": path,
	})

	_, err := server.handleCompareProfiles(nil, req)

	if err == nil {
		t.Error("expected error for missing comparison")
	}
	if !strings.Contains(err.Error(), "comparison path is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHandleAnalyzeWorkers_Success(t *testing.T) {
	profile := testutil.ProfileWithWorkers(4)
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{"path": path})

	result, err := server.handleAnalyzeWorkers(nil, req)

	if err != nil {
		t.Fatalf("handleAnalyzeWorkers error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleAnalyzeCrypto_Success(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{"path": path})

	result, err := server.handleAnalyzeCrypto(nil, req)

	if err != nil {
		t.Fatalf("handleAnalyzeCrypto error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleAnalyzeJSCrypto_Success(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{"path": path})

	result, err := server.handleAnalyzeJSCrypto(nil, req)

	if err != nil {
		t.Fatalf("handleAnalyzeJSCrypto error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleAnalyzeContention_Success(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{"path": path})

	result, err := server.handleAnalyzeContention(nil, req)

	if err != nil {
		t.Fatalf("handleAnalyzeContention error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleAnalyzeScaling_Success(t *testing.T) {
	profile := testutil.ProfileWithWorkers(4)
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{"path": path})

	result, err := server.handleAnalyzeScaling(nil, req)

	if err != nil {
		t.Fatalf("handleAnalyzeScaling error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleCompareScaling_Success(t *testing.T) {
	profile1 := testutil.ProfileWithWorkers(2)
	profile2 := testutil.ProfileWithWorkers(4)
	path1 := testutil.TempProfileFile(t, profile1)
	path2 := testutil.TempProfileFile(t, profile2)

	server := NewServer()
	req := mockRequest(map[string]any{
		"baseline":   path1,
		"comparison": path2,
	})

	result, err := server.handleCompareScaling(nil, req)

	if err != nil {
		t.Fatalf("handleCompareScaling error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleCompareScaling_MissingBaseline(t *testing.T) {
	server := NewServer()
	req := mockRequest(map[string]any{
		"comparison": "/some/path.json",
	})

	_, err := server.handleCompareScaling(nil, req)

	if err == nil {
		t.Error("expected error for missing baseline")
	}
}

func TestHandleBatchAnalyze_Success(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	profilesJSON := `[{"path": "` + path + `", "workers": 2, "label": "Test"}]`
	req := mockRequest(map[string]any{"profiles": profilesJSON})

	result, err := server.handleBatchAnalyze(nil, req)

	if err != nil {
		t.Fatalf("handleBatchAnalyze error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleBatchAnalyze_InvalidJSON(t *testing.T) {
	server := NewServer()
	req := mockRequest(map[string]any{"profiles": "invalid json"})

	_, err := server.handleBatchAnalyze(nil, req)

	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "failed to parse profiles JSON") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHandleBatchAnalyze_EmptyProfiles(t *testing.T) {
	server := NewServer()
	req := mockRequest(map[string]any{"profiles": "[]"})

	_, err := server.handleBatchAnalyze(nil, req)

	if err == nil {
		t.Error("expected error for empty profiles")
	}
	if !strings.Contains(err.Error(), "no profiles provided") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHandleGenerateChart_InlineMode(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	profilesJSON := `[{"path": "` + path + `", "workers": 2, "label": "Test"}]`
	req := mockRequest(map[string]any{
		"profiles":   profilesJSON,
		"chart_type": "wall_clock",
		"output":     "inline",
	})

	result, err := server.handleGenerateChart(nil, req)

	if err != nil {
		t.Fatalf("handleGenerateChart error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleGenerateChart_FileMode(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)
	path := testutil.TempProfileFile(t, profile)

	// Create temp output path
	tmpDir := t.TempDir()
	outputPath := tmpDir + "/test_chart.svg"

	server := NewServer()
	profilesJSON := `[{"path": "` + path + `", "workers": 2, "label": "Test"}]`
	req := mockRequest(map[string]any{
		"profiles":    profilesJSON,
		"chart_type":  "efficiency",
		"output":      "file",
		"output_path": outputPath,
	})

	result, err := server.handleGenerateChart(nil, req)

	if err != nil {
		t.Fatalf("handleGenerateChart error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleGetDelimiterMarkers_Success(t *testing.T) {
	profile := testutil.ProfileWithDelimiters()
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{"path": path})

	result, err := server.handleGetDelimiterMarkers(nil, req)

	if err != nil {
		t.Fatalf("handleGetDelimiterMarkers error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleGetDelimiterMarkers_WithFilters(t *testing.T) {
	profile := testutil.ProfileWithDelimiters()
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{
		"path":       path,
		"categories": "DOM,Layout",
		"limit":      float64(10),
	})

	result, err := server.handleGetDelimiterMarkers(nil, req)

	if err != nil {
		t.Fatalf("handleGetDelimiterMarkers error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleMeasureOperation_PatternBased(t *testing.T) {
	profile := testutil.ProfileWithDelimiters()
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{
		"path":          path,
		"start_pattern": "DOMEvent",
		"end_pattern":   "Paint",
	})

	result, err := server.handleMeasureOperation(nil, req)

	if err != nil {
		t.Fatalf("handleMeasureOperation error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleMeasureOperation_IndexBased(t *testing.T) {
	profile := testutil.ProfileWithDelimiters()
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{
		"path":        path,
		"start_index": float64(0),
		"end_index":   float64(1),
	})

	result, err := server.handleMeasureOperation(nil, req)

	if err != nil {
		t.Fatalf("handleMeasureOperation error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleMeasureOperation_WithOptions(t *testing.T) {
	profile := testutil.ProfileWithDelimiters()
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{
		"path":               path,
		"start_pattern":      "DOMEvent",
		"end_pattern":        "Paint",
		"start_after_ms":     float64(0),
		"end_before_ms":      float64(1000),
		"find_last":          true,
		"start_min_duration": float64(0),
		"end_min_duration":   float64(0),
	})

	result, err := server.handleMeasureOperation(nil, req)

	if err != nil {
		t.Fatalf("handleMeasureOperation error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleMeasureOperation_MissingPattern(t *testing.T) {
	profile := testutil.ProfileWithDelimiters()
	path := testutil.TempProfileFile(t, profile)

	server := NewServer()
	req := mockRequest(map[string]any{
		"path":          path,
		"start_pattern": "DOMEvent",
		// missing end_pattern
	})

	_, err := server.handleMeasureOperation(nil, req)

	if err == nil {
		t.Error("expected error for missing end_pattern")
	}
}

// mockRequest creates a mock CallToolRequest for testing
func mockRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}
