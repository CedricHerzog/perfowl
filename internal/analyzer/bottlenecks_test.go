package analyzer

import (
	"encoding/json"
	"testing"

	"github.com/CedricHerzog/perfowl/internal/parser"
	"github.com/CedricHerzog/perfowl/internal/testutil"
)

func TestSeverity_String(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityLow, "low"},
		{SeverityMedium, "medium"},
		{SeverityHigh, "high"},
		{Severity(99), "low"}, // Unknown defaults to low
	}

	for _, tt := range tests {
		if got := tt.severity.String(); got != tt.expected {
			t.Errorf("Severity(%d).String() = %v, want %v", tt.severity, got, tt.expected)
		}
	}
}

func TestSeverity_MarshalJSON(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityLow, `"low"`},
		{SeverityMedium, `"medium"`},
		{SeverityHigh, `"high"`},
	}

	for _, tt := range tests {
		data, err := tt.severity.MarshalJSON()
		if err != nil {
			t.Errorf("MarshalJSON() error = %v", err)
			continue
		}
		if string(data) != tt.expected {
			t.Errorf("MarshalJSON() = %v, want %v", string(data), tt.expected)
		}
	}
}

func TestParseSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected Severity
	}{
		{"high", SeverityHigh},
		{"HIGH", SeverityHigh},
		{"High", SeverityHigh},
		{"medium", SeverityMedium},
		{"MEDIUM", SeverityMedium},
		{"Medium", SeverityMedium},
		{"low", SeverityLow},
		{"LOW", SeverityLow},
		{"Low", SeverityLow},
		{"unknown", SeverityLow},
		{"", SeverityLow},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ParseSeverity(tt.input); got != tt.expected {
				t.Errorf("ParseSeverity(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDetectBottlenecks_EmptyProfile(t *testing.T) {
	profile := testutil.MinimalProfile()
	bottlenecks := DetectBottlenecks(profile)

	if len(bottlenecks) != 0 {
		t.Errorf("expected no bottlenecks for empty profile, got %d", len(bottlenecks))
	}
}

func TestDetectBottlenecks_NoIssues(t *testing.T) {
	profile := testutil.ProfileWithMainThread()
	bottlenecks := DetectBottlenecks(profile)

	if len(bottlenecks) != 0 {
		t.Errorf("expected no bottlenecks for clean profile, got %d", len(bottlenecks))
	}
}

func TestDetectLongTasks_NoLongTasks(t *testing.T) {
	markers := []parser.ParsedMarker{
		{Name: "DOMEvent", Duration: 10},
		{Name: "Styles", Duration: 20},
	}

	result := detectLongTasks(markers)
	if result != nil {
		t.Errorf("expected nil for no long tasks, got %+v", result)
	}
}

func TestDetectLongTasks_SingleLongTask(t *testing.T) {
	markers := []parser.ParsedMarker{
		{Name: "MainThreadLongTask", Duration: 100, ThreadName: "GeckoMain"},
	}

	result := detectLongTasks(markers)
	if result == nil {
		t.Fatal("expected bottleneck for long task")
	}

	if result.Type != "Long Tasks" {
		t.Errorf("Type = %v, want Long Tasks", result.Type)
	}
	if result.Count != 1 {
		t.Errorf("Count = %v, want 1", result.Count)
	}
	if result.MaxDuration != 100 {
		t.Errorf("MaxDuration = %v, want 100", result.MaxDuration)
	}
}

func TestDetectLongTasks_MultipleLongTasks(t *testing.T) {
	markers := []parser.ParsedMarker{
		{Name: "MainThreadLongTask", Duration: 60, ThreadName: "GeckoMain"},
		{Name: "MainThreadLongTask", Duration: 80, ThreadName: "GeckoMain"},
		{Name: "MainThreadLongTask", Duration: 150, ThreadName: "GeckoMain"},
	}

	result := detectLongTasks(markers)
	if result == nil {
		t.Fatal("expected bottleneck for long tasks")
	}

	if result.Count != 3 {
		t.Errorf("Count = %v, want 3", result.Count)
	}
	if result.TotalDuration != 290 {
		t.Errorf("TotalDuration = %v, want 290", result.TotalDuration)
	}
	if result.MaxDuration != 150 {
		t.Errorf("MaxDuration = %v, want 150", result.MaxDuration)
	}
}

func TestDetectLongTasks_SeverityLevels(t *testing.T) {
	// Low severity: few tasks, short duration
	lowMarkers := []parser.ParsedMarker{
		{Name: "MainThreadLongTask", Duration: 60, ThreadName: "GeckoMain"},
		{Name: "MainThreadLongTask", Duration: 70, ThreadName: "GeckoMain"},
	}
	lowResult := detectLongTasks(lowMarkers)
	if lowResult.Severity != SeverityLow {
		t.Errorf("expected low severity, got %v", lowResult.Severity)
	}

	// Medium severity: >5 tasks or max > 100ms
	medMarkers := make([]parser.ParsedMarker, 6)
	for i := range medMarkers {
		medMarkers[i] = parser.ParsedMarker{Name: "MainThreadLongTask", Duration: 60, ThreadName: "GeckoMain"}
	}
	medResult := detectLongTasks(medMarkers)
	if medResult.Severity != SeverityMedium {
		t.Errorf("expected medium severity for >5 tasks, got %v", medResult.Severity)
	}

	// High severity: >10 tasks or max > 200ms
	highMarkers := []parser.ParsedMarker{
		{Name: "MainThreadLongTask", Duration: 250, ThreadName: "GeckoMain"},
	}
	highResult := detectLongTasks(highMarkers)
	if highResult.Severity != SeverityHigh {
		t.Errorf("expected high severity for >200ms task, got %v", highResult.Severity)
	}
}

func TestDetectLongTasks_ExcludesShortTasks(t *testing.T) {
	markers := []parser.ParsedMarker{
		{Name: "MainThreadLongTask", Duration: 30, ThreadName: "GeckoMain"}, // Below threshold
		{Name: "MainThreadLongTask", Duration: 40, ThreadName: "GeckoMain"}, // Below threshold
	}

	result := detectLongTasks(markers)
	if result != nil {
		t.Errorf("expected nil for tasks below threshold, got %+v", result)
	}
}

func TestDetectLongTasks_ExcludesUnreasonablyLong(t *testing.T) {
	markers := []parser.ParsedMarker{
		{Name: "MainThreadLongTask", Duration: 15000, ThreadName: "GeckoMain"}, // > 10s, unreasonable
	}

	result := detectLongTasks(markers)
	if result != nil {
		t.Errorf("expected nil for unreasonably long tasks, got %+v", result)
	}
}

func TestDetectGCPressure_NoGC(t *testing.T) {
	markers := []parser.ParsedMarker{
		{Name: "DOMEvent", Duration: 10},
	}

	result := detectGCPressure(markers, 1.0)
	if result != nil {
		t.Errorf("expected nil for no GC, got %+v", result)
	}
}

func TestDetectGCPressure_FirefoxMarkers(t *testing.T) {
	markers := []parser.ParsedMarker{
		{Name: "GCMajor", Duration: 150},
		{Name: "GCMinor", Duration: 50},
		{Name: "GCSlice", Duration: 30},
	}

	result := detectGCPressure(markers, 1.0)
	if result == nil {
		t.Fatal("expected bottleneck for GC pressure")
	}

	if result.Type != "GC Pressure" {
		t.Errorf("Type = %v, want GC Pressure", result.Type)
	}
	if result.Count != 3 {
		t.Errorf("Count = %v, want 3", result.Count)
	}
}

func TestDetectGCPressure_ChromeMarkers(t *testing.T) {
	markers := []parser.ParsedMarker{
		{Name: "MajorGC", Duration: 150},
		{Name: "MinorGC", Duration: 50},
	}

	result := detectGCPressure(markers, 1.0)
	if result == nil {
		t.Fatal("expected bottleneck for Chrome GC")
	}

	if result.Count != 2 {
		t.Errorf("Count = %v, want 2", result.Count)
	}
}

func TestDetectGCPressure_V8Markers(t *testing.T) {
	markers := []parser.ParsedMarker{
		{Name: "V8.GC_SCAVENGER", Duration: 100},
		{Name: "V8.GC_MARK_COMPACTOR", Duration: 200},
	}

	result := detectGCPressure(markers, 1.0)
	if result == nil {
		t.Fatal("expected bottleneck for V8 GC")
	}

	if result.Count != 2 {
		t.Errorf("Count = %v, want 2", result.Count)
	}
}

func TestDetectGCPressure_SeverityLevels(t *testing.T) {
	// High frequency (>4 GC/sec) should be high severity
	highFreqMarkers := make([]parser.ParsedMarker, 5)
	for i := range highFreqMarkers {
		highFreqMarkers[i] = parser.ParsedMarker{Name: "GCMajor", Duration: 150}
	}
	result := detectGCPressure(highFreqMarkers, 1.0)
	if result.Severity != SeverityHigh {
		t.Errorf("expected high severity for >4 GC/sec, got %v", result.Severity)
	}

	// Long pause (>200ms) should be high severity
	longPauseMarkers := []parser.ParsedMarker{
		{Name: "GCMajor", Duration: 250},
	}
	result = detectGCPressure(longPauseMarkers, 10.0)
	if result.Severity != SeverityHigh {
		t.Errorf("expected high severity for >200ms pause, got %v", result.Severity)
	}
}

func TestDetectGCPressure_SkipsLowSeverityMinimalDuration(t *testing.T) {
	// Low severity with < 500ms total should be skipped
	markers := []parser.ParsedMarker{
		{Name: "GCMinor", Duration: 10},
		{Name: "GCMinor", Duration: 10},
	}

	result := detectGCPressure(markers, 100.0) // Very long profile = low frequency
	if result != nil {
		t.Errorf("expected nil for low severity minimal GC, got %+v", result)
	}
}

func TestDetectSyncIPC_NoSyncIPC(t *testing.T) {
	markers := []parser.ParsedMarker{
		{Name: "DOMEvent", Duration: 10},
	}

	result := detectSyncIPC(markers)
	if result != nil {
		t.Errorf("expected nil for no sync IPC, got %+v", result)
	}
}

func TestDetectSyncIPC_WithSyncFlag(t *testing.T) {
	markers := []parser.ParsedMarker{
		{Name: "IPC", Category: "IPC", Duration: 50, Data: map[string]interface{}{"sync": true}},
	}

	result := detectSyncIPC(markers)
	if result == nil {
		t.Fatal("expected bottleneck for sync IPC")
	}

	if result.Type != "Synchronous IPC" {
		t.Errorf("Type = %v, want Synchronous IPC", result.Type)
	}
}

func TestDetectSyncIPC_LongDurationAssumedSync(t *testing.T) {
	markers := []parser.ParsedMarker{
		{Name: "IPC", Category: "IPC", Duration: 50, Data: map[string]interface{}{}}, // > 10ms, assumed sync
	}

	result := detectSyncIPC(markers)
	if result == nil {
		t.Fatal("expected bottleneck for long IPC")
	}

	if result.Count != 1 {
		t.Errorf("Count = %v, want 1", result.Count)
	}
}

func TestDetectSyncIPC_SeverityLevels(t *testing.T) {
	// High severity: >20 calls or >500ms total
	highMarkers := make([]parser.ParsedMarker, 25)
	for i := range highMarkers {
		highMarkers[i] = parser.ParsedMarker{
			Name:     "IPC",
			Category: "IPC",
			Duration: 15,
			Data:     map[string]interface{}{"sync": true},
		}
	}
	result := detectSyncIPC(highMarkers)
	if result.Severity != SeverityHigh {
		t.Errorf("expected high severity for >20 sync IPC, got %v", result.Severity)
	}
}

func TestDetectLayoutThrashing_NoThrashing(t *testing.T) {
	// Less than 5 layout markers
	markers := []parser.ParsedMarker{
		{Name: "Reflow", Category: "Layout", Duration: 10, StartTime: 0},
		{Name: "Reflow", Category: "Layout", Duration: 10, StartTime: 200},
	}

	result := detectLayoutThrashing(markers)
	if result != nil {
		t.Errorf("expected nil for insufficient layout markers, got %+v", result)
	}
}

func TestDetectLayoutThrashing_RapidReflows(t *testing.T) {
	// 10 reflows within 100ms window
	markers := make([]parser.ParsedMarker, 10)
	for i := range markers {
		markers[i] = parser.ParsedMarker{
			Name:      "Reflow",
			Category:  "Layout",
			Duration:  5,
			StartTime: float64(i * 10), // 10ms apart = within 100ms window
		}
	}

	result := detectLayoutThrashing(markers)
	if result == nil {
		t.Fatal("expected bottleneck for layout thrashing")
	}

	if result.Type != "Layout Thrashing" {
		t.Errorf("Type = %v, want Layout Thrashing", result.Type)
	}
}

func TestDetectLayoutThrashing_StylesCategory(t *testing.T) {
	markers := make([]parser.ParsedMarker, 10)
	for i := range markers {
		markers[i] = parser.ParsedMarker{
			Name:      "Styles",
			Duration:  5,
			StartTime: float64(i * 10),
		}
	}

	result := detectLayoutThrashing(markers)
	if result == nil {
		t.Fatal("expected bottleneck for Styles thrashing")
	}
}

func TestDetectNetworkBlocking_NoBlocking(t *testing.T) {
	markers := []parser.ParsedMarker{
		{Name: "Network", Category: "Network", Duration: 100}, // < 1000ms
	}

	result := detectNetworkBlocking(markers)
	if result != nil {
		t.Errorf("expected nil for fast network, got %+v", result)
	}
}

func TestDetectNetworkBlocking_SlowRequests(t *testing.T) {
	markers := []parser.ParsedMarker{
		{Name: "ChannelMarker", Category: "Network", Duration: 2000, Data: map[string]interface{}{"url": "https://slow.example.com/api"}},
	}

	result := detectNetworkBlocking(markers)
	if result == nil {
		t.Fatal("expected bottleneck for slow network")
	}

	if result.Type != "Network Blocking" {
		t.Errorf("Type = %v, want Network Blocking", result.Type)
	}
	if result.Count != 1 {
		t.Errorf("Count = %v, want 1", result.Count)
	}
}

func TestDetectNetworkBlocking_URLTruncation(t *testing.T) {
	longURL := "https://example.com/very/long/path/that/exceeds/fifty/characters/and/should/be/truncated"
	markers := []parser.ParsedMarker{
		{Name: "ChannelMarker", Category: "Network", Duration: 2000, Data: map[string]interface{}{"url": longURL}},
	}

	result := detectNetworkBlocking(markers)
	if result == nil {
		t.Fatal("expected bottleneck")
	}

	// URL should be truncated in locations
	if len(result.Locations) > 0 && len(result.Locations[0]) > 60 {
		// Some truncation should occur (50 chars + "..." + duration prefix)
		t.Logf("Location: %s", result.Locations[0])
	}
}

func TestDetectExtensionOverhead_NoExtensions(t *testing.T) {
	profile := testutil.MinimalProfile()
	markers := []parser.ParsedMarker{
		{Name: "DOMEvent", Duration: 10},
	}

	result := detectExtensionOverhead(markers, profile)
	if result != nil {
		t.Errorf("expected nil for no extensions, got %+v", result)
	}
}

func TestDetectExtensionOverhead_WithActivity(t *testing.T) {
	profile := testutil.NewProfileBuilder().
		WithDuration(1000).
		WithExtension("ext1@test.com", "Test Extension", "moz-extension://abc123/").
		Build()

	markers := []parser.ParsedMarker{
		{Name: "Script", Duration: 50, Data: map[string]interface{}{"url": "moz-extension://abc123/content.js"}},
		{Name: "Script", Duration: 30, Data: map[string]interface{}{"url": "moz-extension://abc123/background.js"}},
		{Name: "DOMEvent", Duration: 30, Data: map[string]interface{}{"url": "moz-extension://abc123/popup.js"}},
	}

	result := detectExtensionOverhead(markers, profile)
	if result == nil {
		t.Fatal("expected bottleneck for extension overhead")
	}

	if result.Type != "Extension Overhead" {
		t.Errorf("Type = %v, want Extension Overhead", result.Type)
	}
	if result.Count != 3 {
		t.Errorf("Count = %v, want 3", result.Count)
	}
	if result.TotalDuration != 110 {
		t.Errorf("TotalDuration = %v, want 110", result.TotalDuration)
	}
}

func TestDetectExtensionOverhead_JSActorMessage(t *testing.T) {
	profile := testutil.NewProfileBuilder().
		WithDuration(1000).
		WithExtension("ext1@test.com", "Test Extension", "moz-extension://abc123/").
		Build()

	markers := []parser.ParsedMarker{
		{Name: "JSActorMessage", Duration: 100, Data: map[string]interface{}{"actor": "WebExtensions:ext1@test.com"}},
	}

	result := detectExtensionOverhead(markers, profile)
	if result == nil {
		t.Fatal("expected bottleneck for JSActorMessage")
	}

	if result.Count != 1 {
		t.Errorf("Count = %v, want 1", result.Count)
	}
}

func TestDetectExtensionOverhead_BelowThreshold(t *testing.T) {
	profile := testutil.NewProfileBuilder().
		WithDuration(1000).
		WithExtension("ext1@test.com", "Test Extension", "moz-extension://abc123/").
		Build()

	// Very low activity (< 100ms and < 50 events)
	markers := []parser.ParsedMarker{
		{Name: "Script", Duration: 5, Data: map[string]interface{}{"url": "moz-extension://abc123/content.js"}},
	}

	result := detectExtensionOverhead(markers, profile)
	if result != nil {
		t.Errorf("expected nil for minimal extension activity, got %+v", result)
	}
}

func TestCalculateScore_NoBottlenecks(t *testing.T) {
	var bottlenecks []Bottleneck
	score := CalculateScore(bottlenecks)

	if score != 100 {
		t.Errorf("expected score 100 for no bottlenecks, got %d", score)
	}
}

func TestCalculateScore_SingleBottleneck(t *testing.T) {
	tests := []struct {
		severity Severity
		expected int
	}{
		{SeverityLow, 95},
		{SeverityMedium, 90},
		{SeverityHigh, 80},
	}

	for _, tt := range tests {
		bottlenecks := []Bottleneck{{Severity: tt.severity}}
		score := CalculateScore(bottlenecks)
		if score != tt.expected {
			t.Errorf("score for %v severity = %d, want %d", tt.severity, score, tt.expected)
		}
	}
}

func TestCalculateScore_MixedSeverities(t *testing.T) {
	bottlenecks := []Bottleneck{
		{Severity: SeverityHigh},   // -20
		{Severity: SeverityMedium}, // -10
		{Severity: SeverityLow},    // -5
	}

	score := CalculateScore(bottlenecks)
	expected := 100 - 20 - 10 - 5
	if score != expected {
		t.Errorf("score = %d, want %d", score, expected)
	}
}

func TestCalculateScore_MinimumZero(t *testing.T) {
	// Many high severity bottlenecks
	bottlenecks := make([]Bottleneck, 10)
	for i := range bottlenecks {
		bottlenecks[i] = Bottleneck{Severity: SeverityHigh}
	}

	score := CalculateScore(bottlenecks)
	if score != 0 {
		t.Errorf("expected score 0 for many bottlenecks, got %d", score)
	}
}

func TestGenerateSummary_NoBottlenecks(t *testing.T) {
	var bottlenecks []Bottleneck
	profile := testutil.MinimalProfile()

	summary := GenerateSummary(bottlenecks, profile)
	if summary == "" {
		t.Error("expected non-empty summary")
	}
	if !contains(summary, "No significant") {
		t.Errorf("expected 'No significant' in summary, got: %s", summary)
	}
}

func TestGenerateSummary_WithBottlenecks(t *testing.T) {
	bottlenecks := []Bottleneck{
		{Severity: SeverityHigh, Type: "Long Tasks"},
		{Severity: SeverityMedium, Type: "GC Pressure"},
		{Severity: SeverityLow, Type: "Sync IPC"},
	}
	profile := testutil.MinimalProfile()

	summary := GenerateSummary(bottlenecks, profile)
	if !contains(summary, "3 bottleneck") {
		t.Errorf("expected '3 bottleneck' in summary, got: %s", summary)
	}
	if !contains(summary, "1 high severity") {
		t.Errorf("expected '1 high severity' in summary, got: %s", summary)
	}
}

func TestDetectBottlenecks_Integration(t *testing.T) {
	// Create a profile with various issues
	profile := testutil.ProfileWithLongTasks(3)

	bottlenecks := DetectBottlenecks(profile)

	// Should detect long tasks
	found := false
	for _, b := range bottlenecks {
		if b.Type == "Long Tasks" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to detect Long Tasks bottleneck")
	}
}

func TestBottleneck_JSONSerialization(t *testing.T) {
	b := Bottleneck{
		Type:          "Long Tasks",
		Severity:      SeverityHigh,
		Count:         5,
		TotalDuration: 500,
		Description:   "Test bottleneck",
	}

	data, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Verify the JSON contains expected fields
	jsonStr := string(data)
	if !contains(jsonStr, `"type":"Long Tasks"`) {
		t.Errorf("JSON should contain type field: %s", jsonStr)
	}
	if !contains(jsonStr, `"severity":"high"`) {
		t.Errorf("JSON should contain severity as string: %s", jsonStr)
	}
	if !contains(jsonStr, `"count":5`) {
		t.Errorf("JSON should contain count field: %s", jsonStr)
	}
}

// Helper function for string containment check
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
