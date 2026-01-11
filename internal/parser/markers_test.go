package parser

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParsedMarker_IsDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration float64
		expected bool
	}{
		{"positive duration", 100.0, true},
		{"zero duration", 0.0, false},
		{"negative duration", -1.0, false},
		{"small positive", 0.001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ParsedMarker{Duration: tt.duration}
			if got := m.IsDuration(); got != tt.expected {
				t.Errorf("IsDuration() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParsedMarker_DurationMs(t *testing.T) {
	tests := []struct {
		name     string
		duration float64
	}{
		{"100ms", 100.0},
		{"zero", 0.0},
		{"1.5ms", 1.5},
		{"large", 10000.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ParsedMarker{Duration: tt.duration}
			if got := m.DurationMs(); got != tt.duration {
				t.Errorf("DurationMs() = %v, want %v", got, tt.duration)
			}
		})
	}
}

func TestParsedMarker_DurationTime(t *testing.T) {
	tests := []struct {
		name     string
		duration float64
		expected time.Duration
	}{
		{"100ms", 100.0, 100 * time.Millisecond},
		{"1ms", 1.0, 1 * time.Millisecond},
		{"zero", 0.0, 0},
		{"1.5ms", 1.5, 1500 * time.Microsecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ParsedMarker{Duration: tt.duration}
			if got := m.DurationTime(); got != tt.expected {
				t.Errorf("DurationTime() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExtractMarkers_EmptyThread(t *testing.T) {
	thread := &Thread{
		Name:        "TestThread",
		PID:         "123",
		StringArray: []string{},
		Markers:     Markers{Length: 0},
	}
	categories := []Category{{Name: "Other", Color: "grey"}}

	markers := ExtractMarkers(thread, categories)
	if len(markers) != 0 {
		t.Errorf("expected 0 markers, got %d", len(markers))
	}
}

func TestExtractMarkers_SingleMarker(t *testing.T) {
	data, _ := json.Marshal(map[string]interface{}{"type": "DOMEvent", "eventType": "click"})
	thread := &Thread{
		Name:        "GeckoMain",
		PID:         "1",
		StringArray: []string{"DOMEvent"},
		Markers: Markers{
			Length:    1,
			Name:      []int{0},
			Category:  []int{0},
			StartTime: []float64{100.0},
			EndTime:   []interface{}{150.0},
			Phase:     []int{1},
			Data:      []json.RawMessage{data},
		},
	}
	categories := []Category{{Name: "DOM", Color: "blue"}}

	markers := ExtractMarkers(thread, categories)
	if len(markers) != 1 {
		t.Fatalf("expected 1 marker, got %d", len(markers))
	}

	m := markers[0]
	if m.Name != "DOMEvent" {
		t.Errorf("Name = %v, want DOMEvent", m.Name)
	}
	if m.Type != MarkerTypeDOMEvent {
		t.Errorf("Type = %v, want DOMEvent", m.Type)
	}
	if m.Category != "DOM" {
		t.Errorf("Category = %v, want DOM", m.Category)
	}
	if m.StartTime != 100.0 {
		t.Errorf("StartTime = %v, want 100.0", m.StartTime)
	}
	if m.EndTime != 150.0 {
		t.Errorf("EndTime = %v, want 150.0", m.EndTime)
	}
	if m.Duration != 50.0 {
		t.Errorf("Duration = %v, want 50.0", m.Duration)
	}
	if m.ThreadName != "GeckoMain" {
		t.Errorf("ThreadName = %v, want GeckoMain", m.ThreadName)
	}
}

func TestExtractMarkers_MultipleMarkers(t *testing.T) {
	data1, _ := json.Marshal(map[string]interface{}{"type": "GCMajor"})
	data2, _ := json.Marshal(map[string]interface{}{"type": "GCMinor"})
	thread := &Thread{
		Name:        "GeckoMain",
		PID:         "1",
		StringArray: []string{"GCMajor", "GCMinor"},
		Markers: Markers{
			Length:    2,
			Name:      []int{0, 1},
			Category:  []int{0, 0},
			StartTime: []float64{100.0, 200.0},
			EndTime:   []interface{}{150.0, 220.0},
			Phase:     []int{1, 1},
			Data:      []json.RawMessage{data1, data2},
		},
	}
	categories := []Category{{Name: "GC / CC", Color: "orange"}}

	markers := ExtractMarkers(thread, categories)
	if len(markers) != 2 {
		t.Fatalf("expected 2 markers, got %d", len(markers))
	}

	if markers[0].Type != MarkerTypeGCMajor {
		t.Errorf("first marker Type = %v, want GCMajor", markers[0].Type)
	}
	if markers[1].Type != MarkerTypeGCMinor {
		t.Errorf("second marker Type = %v, want GCMinor", markers[1].Type)
	}
}

func TestExtractMarkers_NoEndTime(t *testing.T) {
	data, _ := json.Marshal(map[string]interface{}{"type": "Awake"})
	thread := &Thread{
		Name:        "GeckoMain",
		PID:         "1",
		StringArray: []string{"Awake"},
		Markers: Markers{
			Length:    1,
			Name:      []int{0},
			Category:  []int{0},
			StartTime: []float64{100.0},
			EndTime:   []interface{}{nil},
			Phase:     []int{0},
			Data:      []json.RawMessage{data},
		},
	}
	categories := []Category{{Name: "Other", Color: "grey"}}

	markers := ExtractMarkers(thread, categories)
	if len(markers) != 1 {
		t.Fatalf("expected 1 marker, got %d", len(markers))
	}

	if markers[0].Duration != 0 {
		t.Errorf("Duration = %v, want 0 (no end time)", markers[0].Duration)
	}
}

func TestExtractMarkers_InferTypeFromName(t *testing.T) {
	// No data field, type should be inferred from name
	thread := &Thread{
		Name:        "GeckoMain",
		PID:         "1",
		StringArray: []string{"GCMajor"},
		Markers: Markers{
			Length:    1,
			Name:      []int{0},
			Category:  []int{0},
			StartTime: []float64{100.0},
			EndTime:   []interface{}{150.0},
			Phase:     []int{1},
			Data:      []json.RawMessage{nil},
		},
	}
	categories := []Category{{Name: "GC / CC", Color: "orange"}}

	markers := ExtractMarkers(thread, categories)
	if len(markers) != 1 {
		t.Fatalf("expected 1 marker, got %d", len(markers))
	}

	if markers[0].Type != MarkerTypeGCMajor {
		t.Errorf("Type = %v, want GCMajor (inferred from name)", markers[0].Type)
	}
}

func TestExtractMarkers_NegativeDuration(t *testing.T) {
	// EndTime before StartTime - should result in 0 duration
	data, _ := json.Marshal(map[string]interface{}{"type": "test"})
	thread := &Thread{
		Name:        "GeckoMain",
		PID:         "1",
		StringArray: []string{"test"},
		Markers: Markers{
			Length:    1,
			Name:      []int{0},
			Category:  []int{0},
			StartTime: []float64{200.0},
			EndTime:   []interface{}{100.0}, // Before start
			Phase:     []int{1},
			Data:      []json.RawMessage{data},
		},
	}
	categories := []Category{{Name: "Other", Color: "grey"}}

	markers := ExtractMarkers(thread, categories)
	if len(markers) != 1 {
		t.Fatalf("expected 1 marker, got %d", len(markers))
	}

	if markers[0].Duration != 0 {
		t.Errorf("Duration = %v, want 0 (end before start)", markers[0].Duration)
	}
}

func TestExtractMarkers_InvalidData(t *testing.T) {
	// Invalid JSON in data field
	thread := &Thread{
		Name:        "GeckoMain",
		PID:         "1",
		StringArray: []string{"test"},
		Markers: Markers{
			Length:    1,
			Name:      []int{0},
			Category:  []int{0},
			StartTime: []float64{100.0},
			EndTime:   []interface{}{150.0},
			Phase:     []int{1},
			Data:      []json.RawMessage{[]byte("invalid json")},
		},
	}
	categories := []Category{{Name: "Other", Color: "grey"}}

	// Should not panic, just skip data parsing
	markers := ExtractMarkers(thread, categories)
	if len(markers) != 1 {
		t.Fatalf("expected 1 marker, got %d", len(markers))
	}

	if markers[0].Data != nil {
		t.Errorf("Data = %v, want nil (invalid JSON)", markers[0].Data)
	}
}

func TestExtractMarkers_OutOfBoundsIndices(t *testing.T) {
	// Name index out of bounds
	thread := &Thread{
		Name:        "GeckoMain",
		PID:         "1",
		StringArray: []string{"only_one"},
		Markers: Markers{
			Length:    1,
			Name:      []int{5}, // Out of bounds
			Category:  []int{0},
			StartTime: []float64{100.0},
			EndTime:   []interface{}{nil},
			Phase:     []int{1},
			Data:      []json.RawMessage{nil},
		},
	}
	categories := []Category{{Name: "Other", Color: "grey"}}

	markers := ExtractMarkers(thread, categories)
	if len(markers) != 1 {
		t.Fatalf("expected 1 marker, got %d", len(markers))
	}

	if markers[0].Name != "" {
		t.Errorf("Name = %v, want empty (index out of bounds)", markers[0].Name)
	}
}

func TestFilterMarkersByType_Match(t *testing.T) {
	markers := []ParsedMarker{
		{Type: MarkerTypeGCMajor, Name: "GCMajor"},
		{Type: MarkerTypeGCMinor, Name: "GCMinor"},
		{Type: MarkerTypeGCMajor, Name: "GCMajor"},
		{Type: MarkerTypeDOMEvent, Name: "DOMEvent"},
	}

	filtered := FilterMarkersByType(markers, MarkerTypeGCMajor)
	if len(filtered) != 2 {
		t.Errorf("expected 2 GCMajor markers, got %d", len(filtered))
	}
}

func TestFilterMarkersByType_NoMatch(t *testing.T) {
	markers := []ParsedMarker{
		{Type: MarkerTypeGCMajor, Name: "GCMajor"},
		{Type: MarkerTypeGCMinor, Name: "GCMinor"},
	}

	filtered := FilterMarkersByType(markers, MarkerTypeDOMEvent)
	if len(filtered) != 0 {
		t.Errorf("expected 0 markers, got %d", len(filtered))
	}
}

func TestFilterMarkersByType_MatchByName(t *testing.T) {
	// When Type doesn't match but Name does (as MarkerType)
	markers := []ParsedMarker{
		{Type: "", Name: "CustomMarker"},
		{Type: MarkerType("CustomMarker"), Name: "CustomMarker"},
	}

	filtered := FilterMarkersByType(markers, MarkerType("CustomMarker"))
	if len(filtered) != 2 {
		t.Errorf("expected 2 markers (by type or name), got %d", len(filtered))
	}
}

func TestFilterMarkersByType_Empty(t *testing.T) {
	var markers []ParsedMarker
	filtered := FilterMarkersByType(markers, MarkerTypeGCMajor)
	if len(filtered) != 0 {
		t.Errorf("expected 0 markers, got %d", len(filtered))
	}
}

func TestFilterMarkersByCategory_Match(t *testing.T) {
	markers := []ParsedMarker{
		{Category: "JavaScript"},
		{Category: "Layout"},
		{Category: "JavaScript"},
		{Category: "GC / CC"},
	}

	filtered := FilterMarkersByCategory(markers, "JavaScript")
	if len(filtered) != 2 {
		t.Errorf("expected 2 JavaScript markers, got %d", len(filtered))
	}
}

func TestFilterMarkersByCategory_NoMatch(t *testing.T) {
	markers := []ParsedMarker{
		{Category: "JavaScript"},
		{Category: "Layout"},
	}

	filtered := FilterMarkersByCategory(markers, "Network")
	if len(filtered) != 0 {
		t.Errorf("expected 0 markers, got %d", len(filtered))
	}
}

func TestFilterMarkersByCategory_Empty(t *testing.T) {
	var markers []ParsedMarker
	filtered := FilterMarkersByCategory(markers, "JavaScript")
	if len(filtered) != 0 {
		t.Errorf("expected 0 markers, got %d", len(filtered))
	}
}

func TestFilterMarkersByDuration_AboveThreshold(t *testing.T) {
	markers := []ParsedMarker{
		{Duration: 10.0},
		{Duration: 50.0},
		{Duration: 100.0},
		{Duration: 5.0},
	}

	filtered := FilterMarkersByDuration(markers, 50.0)
	if len(filtered) != 2 {
		t.Errorf("expected 2 markers >= 50ms, got %d", len(filtered))
	}
}

func TestFilterMarkersByDuration_BelowThreshold(t *testing.T) {
	markers := []ParsedMarker{
		{Duration: 10.0},
		{Duration: 20.0},
	}

	filtered := FilterMarkersByDuration(markers, 50.0)
	if len(filtered) != 0 {
		t.Errorf("expected 0 markers >= 50ms, got %d", len(filtered))
	}
}

func TestFilterMarkersByDuration_ZeroDuration(t *testing.T) {
	markers := []ParsedMarker{
		{Duration: 0.0},
		{Duration: 50.0},
	}

	filtered := FilterMarkersByDuration(markers, 0.0)
	if len(filtered) != 2 {
		t.Errorf("expected 2 markers >= 0ms, got %d", len(filtered))
	}
}

func TestFilterMarkersByDuration_Empty(t *testing.T) {
	var markers []ParsedMarker
	filtered := FilterMarkersByDuration(markers, 50.0)
	if len(filtered) != 0 {
		t.Errorf("expected 0 markers, got %d", len(filtered))
	}
}

func TestGetMarkerStats_Empty(t *testing.T) {
	var markers []ParsedMarker
	stats := GetMarkerStats(markers)

	if stats.TotalCount != 0 {
		t.Errorf("TotalCount = %d, want 0", stats.TotalCount)
	}
	if stats.TotalDuration != 0 {
		t.Errorf("TotalDuration = %f, want 0", stats.TotalDuration)
	}
	if stats.AvgDuration != 0 {
		t.Errorf("AvgDuration = %f, want 0", stats.AvgDuration)
	}
}

func TestGetMarkerStats_Populated(t *testing.T) {
	markers := []ParsedMarker{
		{Type: MarkerTypeGCMajor, Category: "GC / CC", Duration: 50.0},
		{Type: MarkerTypeGCMinor, Category: "GC / CC", Duration: 10.0},
		{Type: MarkerTypeDOMEvent, Category: "DOM", Duration: 20.0},
	}

	stats := GetMarkerStats(markers)

	if stats.TotalCount != 3 {
		t.Errorf("TotalCount = %d, want 3", stats.TotalCount)
	}
	if stats.TotalDuration != 80.0 {
		t.Errorf("TotalDuration = %f, want 80.0", stats.TotalDuration)
	}
	if stats.MaxDuration != 50.0 {
		t.Errorf("MaxDuration = %f, want 50.0", stats.MaxDuration)
	}
	if stats.MinDuration != 10.0 {
		t.Errorf("MinDuration = %f, want 10.0", stats.MinDuration)
	}
}

func TestGetMarkerStats_ByType(t *testing.T) {
	markers := []ParsedMarker{
		{Type: MarkerTypeGCMajor},
		{Type: MarkerTypeGCMajor},
		{Type: MarkerTypeGCMinor},
	}

	stats := GetMarkerStats(markers)

	if stats.ByType["GCMajor"] != 2 {
		t.Errorf("ByType[GCMajor] = %d, want 2", stats.ByType["GCMajor"])
	}
	if stats.ByType["GCMinor"] != 1 {
		t.Errorf("ByType[GCMinor] = %d, want 1", stats.ByType["GCMinor"])
	}
}

func TestGetMarkerStats_ByCategory(t *testing.T) {
	markers := []ParsedMarker{
		{Category: "JavaScript"},
		{Category: "JavaScript"},
		{Category: "Layout"},
	}

	stats := GetMarkerStats(markers)

	if stats.ByCategory["JavaScript"] != 2 {
		t.Errorf("ByCategory[JavaScript] = %d, want 2", stats.ByCategory["JavaScript"])
	}
	if stats.ByCategory["Layout"] != 1 {
		t.Errorf("ByCategory[Layout] = %d, want 1", stats.ByCategory["Layout"])
	}
}

func TestGetMarkerStats_ZeroDurationMarkers(t *testing.T) {
	markers := []ParsedMarker{
		{Duration: 0.0}, // Should not affect min/max
		{Duration: 50.0},
	}

	stats := GetMarkerStats(markers)

	if stats.MaxDuration != 50.0 {
		t.Errorf("MaxDuration = %f, want 50.0", stats.MaxDuration)
	}
	if stats.MinDuration != 50.0 {
		t.Errorf("MinDuration = %f, want 50.0 (zero durations excluded)", stats.MinDuration)
	}
}

func TestInferMarkerType(t *testing.T) {
	tests := []struct {
		name     string
		expected MarkerType
	}{
		{"GCMajor", MarkerTypeGCMajor},
		{"GCMinor", MarkerTypeGCMinor},
		{"GCSlice", MarkerTypeGCSlice},
		{"CC", MarkerTypeCC},
		{"CCSlice", MarkerTypeCCSlice},
		{"DOMEvent", MarkerTypeDOMEvent},
		{"Styles", MarkerTypeStyles},
		{"UserTiming", MarkerTypeUserTiming},
		{"MainThreadLongTask", MarkerTypeMainThreadLongTask},
		{"tracing", MarkerTypeTracing},
		{"ChannelMarker", MarkerTypeChannelMarker},
		{"HostResolver", MarkerTypeHostResolver},
		{"JSActorMessage", MarkerTypeJSActorMessage},
		{"FrameMessage", MarkerTypeFrameMessage},
		{"Awake", MarkerTypeAwake},
		{"UnknownMarker", MarkerType("UnknownMarker")}, // Falls back to name
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inferMarkerType(tt.name); got != tt.expected {
				t.Errorf("inferMarkerType(%q) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestGetInt(t *testing.T) {
	slice := []int{10, 20, 30}

	tests := []struct {
		name     string
		index    int
		expected int
	}{
		{"first", 0, 10},
		{"middle", 1, 20},
		{"last", 2, 30},
		{"negative", -1, -1},
		{"out of bounds", 5, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getInt(slice, tt.index); got != tt.expected {
				t.Errorf("getInt() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetFloat(t *testing.T) {
	slice := []float64{1.5, 2.5, 3.5}

	tests := []struct {
		name     string
		index    int
		expected float64
	}{
		{"first", 0, 1.5},
		{"middle", 1, 2.5},
		{"last", 2, 3.5},
		{"negative", -1, 0},
		{"out of bounds", 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFloat(slice, tt.index); got != tt.expected {
				t.Errorf("getFloat() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetPhase(t *testing.T) {
	slice := []int{0, 1, 2}

	tests := []struct {
		name     string
		index    int
		expected int
	}{
		{"first", 0, 0},
		{"middle", 1, 1},
		{"last", 2, 2},
		{"negative", -1, 0},
		{"out of bounds", 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPhase(slice, tt.index); got != tt.expected {
				t.Errorf("getPhase() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetEndTime(t *testing.T) {
	slice := []interface{}{100.0, nil, 200.0, "invalid"}

	t.Run("valid float", func(t *testing.T) {
		result := getEndTime(slice, 0)
		if result == nil || *result != 100.0 {
			t.Errorf("expected 100.0, got %v", result)
		}
	})

	t.Run("nil value", func(t *testing.T) {
		result := getEndTime(slice, 1)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("another float", func(t *testing.T) {
		result := getEndTime(slice, 2)
		if result == nil || *result != 200.0 {
			t.Errorf("expected 200.0, got %v", result)
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		result := getEndTime(slice, 3)
		if result != nil {
			t.Errorf("expected nil for non-float, got %v", result)
		}
	})

	t.Run("out of bounds", func(t *testing.T) {
		result := getEndTime(slice, 10)
		if result != nil {
			t.Errorf("expected nil for out of bounds, got %v", result)
		}
	})

	t.Run("negative index", func(t *testing.T) {
		result := getEndTime(slice, -1)
		if result != nil {
			t.Errorf("expected nil for negative index, got %v", result)
		}
	})
}
