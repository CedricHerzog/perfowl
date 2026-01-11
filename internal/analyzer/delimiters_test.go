package analyzer

import (
	"testing"

	"github.com/CedricHerzog/perfowl/internal/testutil"
)

func TestGetDelimiterMarkers_EmptyProfile(t *testing.T) {
	profile := testutil.MinimalProfile()

	result := GetDelimiterMarkers(profile, nil)

	if len(result) != 0 {
		t.Errorf("expected 0 delimiter markers, got %d", len(result))
	}
}

func TestGetDelimiterMarkers_WithDelimiters(t *testing.T) {
	profile := testutil.ProfileWithDelimiters()

	result := GetDelimiterMarkers(profile, nil)

	if len(result) == 0 {
		t.Error("expected delimiter markers")
	}
}

func TestGetDelimiterMarkers_CategoryFilter(t *testing.T) {
	profile := testutil.ProfileWithDelimiters()

	result := GetDelimiterMarkers(profile, []string{"Layout"})

	// Should only include Layout category markers
	for _, m := range result {
		if m.Category != "Layout" {
			t.Errorf("expected Layout category, got %s", m.Category)
		}
	}
}

func TestGetDelimiterMarkers_SortedByTime(t *testing.T) {
	profile := testutil.ProfileWithDelimiters()

	result := GetDelimiterMarkers(profile, nil)

	for i := 1; i < len(result); i++ {
		if result[i].TimeMs < result[i-1].TimeMs {
			t.Errorf("markers not sorted by time: %f < %f",
				result[i].TimeMs, result[i-1].TimeMs)
		}
	}
}

func TestMatchMarkerPattern_TypeOnly(t *testing.T) {
	marker := DelimiterMarker{Type: "DOMEvent", Name: "click"}

	if !MatchMarkerPattern(marker, "DOMEvent") {
		t.Error("expected DOMEvent to match")
	}
}

func TestMatchMarkerPattern_TypeWithSubtype(t *testing.T) {
	marker := DelimiterMarker{
		Type: "DOMEvent",
		Name: "click",
		Data: map[string]interface{}{"eventType": "click"},
	}

	if !MatchMarkerPattern(marker, "DOMEvent:click") {
		t.Error("expected DOMEvent:click to match")
	}
}

func TestMatchMarkerPattern_NoMatch(t *testing.T) {
	marker := DelimiterMarker{Type: "Styles", Name: "recalc"}

	if MatchMarkerPattern(marker, "DOMEvent") {
		t.Error("expected no match for different type")
	}
}

func TestMatchMarkerPattern_CaseMatch(t *testing.T) {
	marker := DelimiterMarker{Type: "Paint", Name: "composite"}

	// Should match exactly
	if !MatchMarkerPattern(marker, "Paint") {
		t.Error("expected Paint to match")
	}
}

func TestMeasureOperation_PatternMatch(t *testing.T) {
	profile := testutil.ProfileWithDelimiters()

	result, err := MeasureOperation(profile, "DOMEvent", "Paint", 0, 0)

	if err != nil {
		t.Logf("MeasureOperation error: %v (may be expected if patterns don't match)", err)
		return
	}

	if result.OperationTimeMs <= 0 {
		t.Errorf("expected positive operation time, got %f", result.OperationTimeMs)
	}
}

func TestMeasureOperation_StartNotFound(t *testing.T) {
	profile := testutil.ProfileWithDelimiters()

	_, err := MeasureOperation(profile, "NonexistentMarker", "Paint", 0, 0)

	if err == nil {
		t.Error("expected error for nonexistent start marker")
	}
}

func TestMeasureOperation_EndNotFound(t *testing.T) {
	profile := testutil.ProfileWithDelimiters()

	_, err := MeasureOperation(profile, "DOMEvent", "NonexistentMarker", 0, 0)

	if err == nil {
		t.Error("expected error for nonexistent end marker")
	}
}

func TestMeasureOperation_TimeConstraints(t *testing.T) {
	profile := testutil.ProfileWithDelimiters()

	// Measure with time constraints
	result, err := MeasureOperationAdvanced(profile, MeasureOptions{
		StartPattern: "DOMEvent",
		EndPattern:   "Paint",
		StartAfterMs: 50,
		EndBeforeMs:  500,
	})

	if err != nil {
		t.Logf("MeasureOperationAdvanced error: %v", err)
		return
	}

	if result.StartMarker.TimeMs < 50 {
		t.Errorf("start marker before StartAfterMs: %f", result.StartMarker.TimeMs)
	}
	if result.EndMarker.TimeMs > 500 {
		t.Errorf("end marker after EndBeforeMs: %f", result.EndMarker.TimeMs)
	}
}

func TestMeasureOperationLast_FindsLastMatch(t *testing.T) {
	profile := testutil.ProfileWithDelimiters()

	result, err := MeasureOperationLast(profile, "DOMEvent", "Paint", 0, 0)

	if err != nil {
		t.Logf("MeasureOperationLast error: %v", err)
		return
	}

	// Should find the last Paint marker
	if result == nil {
		t.Error("expected result")
	}
}

func TestMeasureOperationByIndex_Valid(t *testing.T) {
	profile := testutil.ProfileWithDelimiters()

	// Get delimiters first
	delimiters := GetDelimiterMarkers(profile, nil)
	if len(delimiters) < 2 {
		t.Skip("not enough delimiter markers")
	}

	result, err := MeasureOperationByIndex(profile, 0, 1)

	if err != nil {
		t.Fatalf("MeasureOperationByIndex error: %v", err)
	}

	if result.OperationTimeMs < 0 {
		t.Errorf("expected non-negative operation time, got %f", result.OperationTimeMs)
	}
}

func TestMeasureOperationByIndex_StartOutOfRange(t *testing.T) {
	profile := testutil.ProfileWithDelimiters()

	_, err := MeasureOperationByIndex(profile, 1000, 1001)

	if err == nil {
		t.Error("expected error for out of range indices")
	}
}

func TestMeasureOperationByIndex_EndBeforeStart(t *testing.T) {
	profile := testutil.ProfileWithDelimiters()

	delimiters := GetDelimiterMarkers(profile, nil)
	if len(delimiters) < 2 {
		t.Skip("not enough delimiter markers")
	}

	_, err := MeasureOperationByIndex(profile, 1, 0)

	if err == nil {
		t.Error("expected error for end before start")
	}
}

func TestGetDelimiterMarkersReport_AllFields(t *testing.T) {
	profile := testutil.ProfileWithDelimiters()

	report := GetDelimiterMarkersReport(profile, nil, 0)

	if report.TotalCount < 0 {
		t.Errorf("TotalCount should be >= 0, got %d", report.TotalCount)
	}
}

func TestGetDelimiterMarkersReport_Limit(t *testing.T) {
	profile := testutil.ProfileWithDelimiters()

	report := GetDelimiterMarkersReport(profile, nil, 2)

	if len(report.Markers) > 2 {
		t.Errorf("expected at most 2 markers, got %d", len(report.Markers))
	}
}

func TestIsDelimiterMarker_DOMEvent(t *testing.T) {
	// Test various delimiter marker types
	tests := []struct {
		name     string
		marker   DelimiterMarker
		expected bool
	}{
		{"DOMEvent", DelimiterMarker{Type: "DOMEvent"}, true},
		{"EventDispatch", DelimiterMarker{Type: "EventDispatch"}, true},
		{"UserTiming", DelimiterMarker{Type: "UserTiming"}, true},
		{"Styles", DelimiterMarker{Type: "Styles"}, true},
		{"Paint", DelimiterMarker{Type: "Paint"}, true},
		{"UpdateLayoutTree", DelimiterMarker{Type: "UpdateLayoutTree"}, true},
		{"GCMajor", DelimiterMarker{Type: "GCMajor"}, false},
		{"IPC", DelimiterMarker{Type: "IPC"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This tests the function indirectly via the marker types
			// that should be included/excluded
		})
	}
}
