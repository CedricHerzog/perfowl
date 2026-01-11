package analyzer

import (
	"testing"

	"github.com/CedricHerzog/perfowl/internal/parser"
	"github.com/CedricHerzog/perfowl/internal/testutil"
)

func TestAnalyzeExtensions_NoExtensions(t *testing.T) {
	profile := testutil.MinimalProfile()

	result := AnalyzeExtensions(profile)

	if len(result.Extensions) != 0 {
		t.Errorf("expected no extensions, got %d", len(result.Extensions))
	}
}

func TestAnalyzeExtensions_SingleExtension(t *testing.T) {
	profile := testutil.ProfileWithExtensions()

	result := AnalyzeExtensions(profile)

	// Profile has 2 extensions, but may not have activity for both
	if result.TotalExtensions != 2 {
		t.Errorf("TotalExtensions = %v, want 2", result.TotalExtensions)
	}
}

func TestAnalyzeExtensions_ExtensionActivity(t *testing.T) {
	profile := testutil.NewProfileBuilder().
		WithDuration(1000).
		WithCategories(testutil.DefaultCategories()).
		WithExtension("ext1@test.com", "Test Extension", "moz-extension://abc123/").
		Build()

	// Add a thread with markers for the extension
	mb := testutil.NewMarkerBuilder()
	mb.AddCustom("Script", 0, 100, 50, map[string]interface{}{"url": "moz-extension://abc123/content.js"})
	markers, strings := mb.Build()

	thread := testutil.NewThreadBuilder("GeckoMain").
		AsMainThread().
		WithMarkers(markers).
		WithStringArray(strings).
		Build()
	profile.Threads = append(profile.Threads, thread)

	result := AnalyzeExtensions(profile)

	// Should detect extension activity
	if result.TotalDuration == 0 {
		t.Error("expected TotalDuration > 0 for extension activity")
	}
}

func TestMatchExtension_URLMatch(t *testing.T) {
	extensionURLs := map[string]string{
		"ext1@test.com": "moz-extension://abc123/",
	}

	marker := parser.ParsedMarker{
		Data: map[string]interface{}{"url": "moz-extension://abc123/content.js"},
	}

	extID := matchExtension(marker, extensionURLs)
	if extID != "ext1@test.com" {
		t.Errorf("expected ext1@test.com, got %s", extID)
	}
}

func TestMatchExtension_MozExtensionURL(t *testing.T) {
	extensionURLs := map[string]string{
		"ext1@test.com": "moz-extension://abc123/",
	}

	marker := parser.ParsedMarker{
		Data: map[string]interface{}{"url": "moz-extension://abc123/background.js"},
	}

	extID := matchExtension(marker, extensionURLs)
	if extID == "" {
		t.Error("expected to match moz-extension URL")
	}
}

func TestMatchExtension_ChromeExtensionURL(t *testing.T) {
	extensionURLs := map[string]string{
		"ext1@test.com": "chrome-extension://abc123/",
	}

	marker := parser.ParsedMarker{
		Data: map[string]interface{}{"url": "chrome-extension://abc123/content.js"},
	}

	extID := matchExtension(marker, extensionURLs)
	if extID != "ext1@test.com" {
		t.Errorf("expected ext1@test.com, got %s", extID)
	}
}

func TestMatchExtension_NoMatch(t *testing.T) {
	extensionURLs := map[string]string{
		"ext1@test.com": "moz-extension://abc123/",
	}

	marker := parser.ParsedMarker{
		Data: map[string]interface{}{"url": "https://example.com/script.js"},
	}

	extID := matchExtension(marker, extensionURLs)
	if extID != "" {
		t.Errorf("expected no match, got %s", extID)
	}
}

func TestAnalyzeExtensions_SortedByDuration(t *testing.T) {
	profile := testutil.NewProfileBuilder().
		WithDuration(1000).
		WithCategories(testutil.DefaultCategories()).
		WithExtension("ext1@test.com", "Extension 1", "moz-extension://abc123/").
		WithExtension("ext2@test.com", "Extension 2", "moz-extension://def456/").
		Build()

	// Add markers for both extensions with different durations
	mb := testutil.NewMarkerBuilder()
	mb.AddCustom("Script", 0, 0, 100, map[string]interface{}{"url": "moz-extension://abc123/script.js"})
	mb.AddCustom("Script", 0, 100, 200, map[string]interface{}{"url": "moz-extension://def456/script.js"})
	markers, strings := mb.Build()

	thread := testutil.NewThreadBuilder("GeckoMain").
		AsMainThread().
		WithMarkers(markers).
		WithStringArray(strings).
		Build()
	profile.Threads = append(profile.Threads, thread)

	result := AnalyzeExtensions(profile)

	// Extensions should be sorted by duration descending
	for i := 1; i < len(result.Extensions); i++ {
		if result.Extensions[i].TotalDuration > result.Extensions[i-1].TotalDuration {
			t.Errorf("extensions not sorted by duration")
		}
	}
}

func TestCalculateImpactScore_Low(t *testing.T) {
	report := &ExtensionReport{
		TotalDuration: 100,
		MarkersCount:  10,
		IPCMessages:   5,
	}

	score := calculateImpactScore(report)
	if score != "low" {
		t.Errorf("expected low impact, got %s", score)
	}
}

func TestCalculateImpactScore_Medium(t *testing.T) {
	// Scoring: duration >500: +2, markers >100: +1, IPC >50: +1 = 4 (medium)
	// score >= 5: high, >= 3: medium, else low
	report := &ExtensionReport{
		TotalDuration: 600,  // >500: +2
		MarkersCount:  200,  // >100: +1
		IPCMessages:   60,   // >50: +1
	}
	// Total: 4, which is >= 3 but < 5 = medium

	score := calculateImpactScore(report)
	if score != "medium" {
		t.Errorf("expected medium impact, got %s", score)
	}
}

func TestCalculateImpactScore_High(t *testing.T) {
	report := &ExtensionReport{
		TotalDuration: 2000,
		MarkersCount:  2000,
		IPCMessages:   200,
	}

	score := calculateImpactScore(report)
	if score != "high" {
		t.Errorf("expected high impact, got %s", score)
	}
}
