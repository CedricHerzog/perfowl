package toon

import (
	"strings"
	"testing"
)

func TestEncodePrimitives(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"string", "hello", "hello"},
		{"int", 42, "42"},
		{"float", 3.14, "3.14"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"float no trailing zeros", 10.0, "10"},
		{"negative int", -5, "-5"},
		{"large float", 1234567.89, "1234567.89"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Encode(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// For primitives at root level, there's no field name
			if strings.TrimSpace(result) != tt.expected {
				t.Errorf("got %q, want %q", strings.TrimSpace(result), tt.expected)
			}
		})
	}
}

func TestEncodeStruct(t *testing.T) {
	type Simple struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	input := Simple{Name: "test", Count: 42}
	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "name: test\ncount: 42\n"
	if result != expected {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}

func TestEncodeStructWithOmitEmpty(t *testing.T) {
	type WithOptional struct {
		Name     string  `json:"name"`
		Optional string  `json:"optional,omitempty"`
		Value    float64 `json:"value,omitempty"`
	}

	input := WithOptional{Name: "test"}
	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Optional and Value should be omitted
	if strings.Contains(result, "optional") {
		t.Errorf("expected optional to be omitted, got:\n%s", result)
	}
	if strings.Contains(result, "value") {
		t.Errorf("expected value to be omitted, got:\n%s", result)
	}
}

func TestEncodeNestedStruct(t *testing.T) {
	type Inner struct {
		Value int `json:"value"`
	}
	type Outer struct {
		Name  string `json:"name"`
		Inner Inner  `json:"inner"`
	}

	input := Outer{Name: "test", Inner: Inner{Value: 42}}
	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have nested indentation
	if !strings.Contains(result, "inner:") {
		t.Errorf("expected 'inner:' header, got:\n%s", result)
	}
	if !strings.Contains(result, "value: 42") {
		t.Errorf("expected 'value: 42', got:\n%s", result)
	}
}

func TestEncodeSimpleArray(t *testing.T) {
	input := struct {
		Tags []string `json:"tags"`
	}{
		Tags: []string{"a", "b", "c"},
	}

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "tags[3]: a,b,c\n"
	if result != expected {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}

func TestEncodeTabularArray(t *testing.T) {
	type Item struct {
		Name  string  `json:"name"`
		Value int     `json:"value"`
		Score float64 `json:"score"`
	}

	input := struct {
		Items []Item `json:"items"`
	}{
		Items: []Item{
			{Name: "first", Value: 1, Score: 1.5},
			{Name: "second", Value: 2, Score: 2.5},
		},
	}

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have tabular format header
	if !strings.Contains(result, "items[2]{name,value,score}:") {
		t.Errorf("expected tabular header, got:\n%s", result)
	}

	// Should have CSV-like rows
	if !strings.Contains(result, "first,1,1.5") {
		t.Errorf("expected first row, got:\n%s", result)
	}
	if !strings.Contains(result, "second,2,2.5") {
		t.Errorf("expected second row, got:\n%s", result)
	}
}

func TestEncodeMap(t *testing.T) {
	input := struct {
		Data map[string]string `json:"data"`
	}{
		Data: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have map entries
	if !strings.Contains(result, "data:") {
		t.Errorf("expected 'data:' header, got:\n%s", result)
	}
	if !strings.Contains(result, "key1: value1") {
		t.Errorf("expected 'key1: value1', got:\n%s", result)
	}
}

func TestEncodeEmptyArray(t *testing.T) {
	input := struct {
		Items []string `json:"items"`
	}{
		Items: []string{},
	}

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "items[0]:") {
		t.Errorf("expected 'items[0]:', got:\n%s", result)
	}
}

func TestEncodeQuotedStrings(t *testing.T) {
	input := struct {
		WithColon   string `json:"with_colon"`
		WithNewline string `json:"with_newline"`
		WithQuote   string `json:"with_quote"`
	}{
		WithColon:   "hello: world",
		WithNewline: "line1\nline2",
		WithQuote:   `say "hello"`,
	}

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be quoted
	if !strings.Contains(result, `"hello: world"`) {
		t.Errorf("expected quoted string with colon, got:\n%s", result)
	}
	if !strings.Contains(result, `"line1\nline2"`) {
		t.Errorf("expected quoted string with newline, got:\n%s", result)
	}
	if !strings.Contains(result, `"say \"hello\""`) {
		t.Errorf("expected quoted string with escaped quotes, got:\n%s", result)
	}
}

func TestEncodePointer(t *testing.T) {
	type Item struct {
		Name string `json:"name"`
	}

	item := &Item{Name: "test"}
	result, err := Encode(item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "name: test") {
		t.Errorf("expected 'name: test', got:\n%s", result)
	}
}

func TestEncodeNilPointer(t *testing.T) {
	var item *struct {
		Name string `json:"name"`
	}

	result, err := Encode(item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be empty for nil pointer
	if result != "" {
		t.Errorf("expected empty string for nil, got:\n%s", result)
	}
}

func TestEncodeBottleneckLikeStruct(t *testing.T) {
	// Simulates the actual Bottleneck struct from the analyzer
	type Bottleneck struct {
		Type          string  `json:"type"`
		Severity      string  `json:"severity"`
		Count         int     `json:"count"`
		TotalDuration float64 `json:"total_duration_ms"`
		Description   string  `json:"description"`
	}

	type Report struct {
		Score       int          `json:"score"`
		Summary     string       `json:"summary"`
		Bottlenecks []Bottleneck `json:"bottlenecks"`
	}

	input := Report{
		Score:   75,
		Summary: "3 issues detected",
		Bottlenecks: []Bottleneck{
			{Type: "long_tasks", Severity: "high", Count: 42, TotalDuration: 1234.5, Description: "Long running tasks"},
			{Type: "gc_pressure", Severity: "medium", Count: 10, TotalDuration: 567.8, Description: "GC pressure detected"},
		},
	}

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check structure
	if !strings.Contains(result, "score: 75") {
		t.Errorf("expected 'score: 75', got:\n%s", result)
	}
	if !strings.Contains(result, "summary: 3 issues detected") {
		t.Errorf("expected summary, got:\n%s", result)
	}
	if !strings.Contains(result, "bottlenecks[2]{type,severity,count,total_duration_ms,description}:") {
		t.Errorf("expected tabular header for bottlenecks, got:\n%s", result)
	}

	t.Logf("Output:\n%s", result)
}

func TestEncodeComplexNestedStruct(t *testing.T) {
	type CategoryStats struct {
		Name    string  `json:"name"`
		TimeMs  float64 `json:"time_ms"`
		Percent float64 `json:"percent"`
	}

	type WorkerStats struct {
		ThreadName    string          `json:"thread_name"`
		CPUTimeMs     float64         `json:"cpu_time_ms"`
		TopCategories []CategoryStats `json:"top_categories,omitempty"`
	}

	type Analysis struct {
		TotalWorkers int           `json:"total_workers"`
		Workers      []WorkerStats `json:"workers"`
	}

	input := Analysis{
		TotalWorkers: 2,
		Workers: []WorkerStats{
			{
				ThreadName: "Worker 1",
				CPUTimeMs:  100.5,
				TopCategories: []CategoryStats{
					{Name: "JavaScript", TimeMs: 80.0, Percent: 79.6},
					{Name: "GC", TimeMs: 20.5, Percent: 20.4},
				},
			},
			{
				ThreadName: "Worker 2",
				CPUTimeMs:  50.0,
			},
		},
	}

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "total_workers: 2") {
		t.Errorf("expected 'total_workers: 2', got:\n%s", result)
	}

	t.Logf("Output:\n%s", result)
}

func TestEncodeIntSlice(t *testing.T) {
	input := struct {
		Numbers []int `json:"numbers"`
	}{
		Numbers: []int{1, 2, 3, 4, 5},
	}

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "numbers[5]: 1,2,3,4,5\n"
	if result != expected {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{10.0, "10"},
		{10.5, "10.5"},
		{10.50, "10.5"},
		{10.500, "10.5"},
		{0.1, "0.1"},
		{0.10, "0.1"},
		{1234.5678, "1234.5678"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatFloat(tt.input, 64)
			if result != tt.expected {
				t.Errorf("formatFloat(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
