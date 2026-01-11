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

func TestEncodeIndent(t *testing.T) {
	type Nested struct {
		Inner string `json:"inner"`
	}
	input := struct {
		Name   string `json:"name"`
		Value  int    `json:"value"`
		Nested Nested `json:"nested"`
	}{
		Name:   "test",
		Value:  42,
		Nested: Nested{Inner: "deep"},
	}

	// Test with prefix
	result, err := EncodeIndent(input, ">>", "    ")
	if err != nil {
		t.Fatalf("EncodeIndent error: %v", err)
	}

	// Should have prefix
	if !strings.Contains(result, ">>name: test") {
		t.Errorf("expected prefix in output, got:\n%s", result)
	}
}

func TestEncodeIndent_TabIndent(t *testing.T) {
	type Inner struct {
		Value int `json:"value"`
	}
	input := struct {
		Name  string `json:"name"`
		Inner Inner  `json:"inner"`
	}{
		Name:  "test",
		Inner: Inner{Value: 42},
	}

	result, err := EncodeIndent(input, "", "\t")
	if err != nil {
		t.Fatalf("EncodeIndent error: %v", err)
	}

	if !strings.Contains(result, "name: test") {
		t.Errorf("expected output, got:\n%s", result)
	}
}

// Test isZero behavior through omitempty
func TestOmitEmptyBehavior(t *testing.T) {
	type TestStruct struct {
		String    string            `json:"string,omitempty"`
		Int       int               `json:"int,omitempty"`
		Float     float64           `json:"float,omitempty"`
		Bool      bool              `json:"bool,omitempty"`
		Slice     []string          `json:"slice,omitempty"`
		Map       map[string]string `json:"map,omitempty"`
		Ptr       *int              `json:"ptr,omitempty"`
		NonEmpty  string            `json:"non_empty,omitempty"`
		NonZero   int               `json:"non_zero,omitempty"`
		TrueBool  bool              `json:"true_bool,omitempty"`
		FullSlice []string          `json:"full_slice,omitempty"`
	}

	input := TestStruct{
		NonEmpty:  "value",
		NonZero:   42,
		TrueBool:  true,
		FullSlice: []string{"a"},
	}

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	// Zero values should be omitted
	if strings.Contains(result, "string:") && !strings.Contains(result, "non_empty") {
		t.Errorf("empty string should be omitted")
	}
	if strings.Contains(result, "int:") && !strings.Contains(result, "non_zero") {
		t.Errorf("zero int should be omitted")
	}

	// Non-zero values should be present
	if !strings.Contains(result, "non_empty: value") {
		t.Errorf("expected 'non_empty: value', got:\n%s", result)
	}
	if !strings.Contains(result, "non_zero: 42") {
		t.Errorf("expected 'non_zero: 42', got:\n%s", result)
	}
}

// Test various primitive types encoding
func TestEncodePrimitiveTypes(t *testing.T) {
	type AllTypes struct {
		String  string  `json:"string"`
		Int     int     `json:"int"`
		Int8    int8    `json:"int8"`
		Int16   int16   `json:"int16"`
		Int32   int32   `json:"int32"`
		Int64   int64   `json:"int64"`
		Uint    uint    `json:"uint"`
		Uint8   uint8   `json:"uint8"`
		Uint16  uint16  `json:"uint16"`
		Uint32  uint32  `json:"uint32"`
		Uint64  uint64  `json:"uint64"`
		Float32 float32 `json:"float32"`
		Float64 float64 `json:"float64"`
		Bool    bool    `json:"bool"`
	}

	input := AllTypes{
		String:  "hello",
		Int:     -42,
		Int8:    -8,
		Int16:   -16,
		Int32:   -32,
		Int64:   -64,
		Uint:    42,
		Uint8:   8,
		Uint16:  16,
		Uint32:  32,
		Uint64:  64,
		Float32: 3.14,
		Float64: 2.718,
		Bool:    true,
	}

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	expectations := []string{
		"string: hello",
		"int: -42",
		"int8: -8",
		"int16: -16",
		"int32: -32",
		"int64: -64",
		"uint: 42",
		"uint8: 8",
		"uint16: 16",
		"uint32: 32",
		"uint64: 64",
		"bool: true",
	}

	for _, exp := range expectations {
		if !strings.Contains(result, exp) {
			t.Errorf("expected %q in output, got:\n%s", exp, result)
		}
	}
}

// Test quoting behavior through encoding
func TestQuotingBehavior(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldQuote bool
	}{
		{"plain", "hello", false},
		{"with colon", "hello: world", true},
		{"with newline", "hello\nworld", true},
		{"with quote", `say "hi"`, true},
		{"with tab", "a\tb", true},
		{"looks like number", "123", true},
		{"looks like bool", "true", true},
		{"looks like null", "null", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := struct {
				Value string `json:"value"`
			}{Value: tt.input}

			result, err := Encode(input)
			if err != nil {
				t.Fatalf("Encode error: %v", err)
			}

			hasQuotes := strings.Contains(result, `"`)
			if hasQuotes != tt.shouldQuote {
				t.Errorf("value %q: hasQuotes=%v, expected %v\nresult: %s", tt.input, hasQuotes, tt.shouldQuote, result)
			}
		})
	}
}

// Test table quoting behavior
func TestTableQuotingBehavior(t *testing.T) {
	type Item struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	tests := []struct {
		name        string
		value       string
		shouldQuote bool
	}{
		{"plain", "hello", false},
		{"with comma", "a,b", true},
		{"with newline", "a\nb", true},
		{"colon ok in table", "a:b", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := struct {
				Items []Item `json:"items"`
			}{
				Items: []Item{{Name: "test", Value: tt.value}},
			}

			result, err := Encode(input)
			if err != nil {
				t.Fatalf("Encode error: %v", err)
			}

			// Check if the value in the table row is quoted
			// The row format is "name,value" so we look for quotes around the value
			hasQuotes := strings.Contains(result, `"`)
			if hasQuotes != tt.shouldQuote {
				t.Errorf("table value %q: hasQuotes=%v, expected %v\nresult: %s", tt.value, hasQuotes, tt.shouldQuote, result)
			}
		})
	}
}

func TestEncodeFloatSlice(t *testing.T) {
	input := struct {
		Values []float64 `json:"values"`
	}{
		Values: []float64{1.1, 2.2, 3.3},
	}

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "values[3]: 1.1,2.2,3.3\n"
	if result != expected {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}

func TestEncodeBoolSlice(t *testing.T) {
	input := struct {
		Flags []bool `json:"flags"`
	}{
		Flags: []bool{true, false, true},
	}

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "flags[3]: true,false,true\n"
	if result != expected {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}

func TestEncodeEmptyMap(t *testing.T) {
	input := struct {
		Data map[string]string `json:"data"`
	}{
		Data: map[string]string{},
	}

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "data: {}") {
		t.Errorf("expected 'data: {}', got:\n%s", result)
	}
}

func TestEncodeMapWithIntValues(t *testing.T) {
	input := struct {
		Counts map[string]int `json:"counts"`
	}{
		Counts: map[string]int{
			"a": 1,
			"b": 2,
		},
	}

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "counts:") {
		t.Errorf("expected 'counts:' header, got:\n%s", result)
	}
}

func TestEncodeInterface(t *testing.T) {
	// Test encoding an interface{} value
	var data interface{} = map[string]interface{}{
		"name":  "test",
		"value": 42,
	}

	result, err := Encode(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "name: test") {
		t.Errorf("expected 'name: test', got:\n%s", result)
	}
}

func TestEncodeSliceOfInterface(t *testing.T) {
	input := struct {
		Items []interface{} `json:"items"`
	}{
		Items: []interface{}{"a", 1, true},
	}

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Non-uniform slice should have individual elements
	if !strings.Contains(result, "items") {
		t.Errorf("expected 'items', got:\n%s", result)
	}
}

func TestEncodeTabularWithQuotedValues(t *testing.T) {
	type Item struct {
		Name    string `json:"name"`
		Comment string `json:"comment"`
	}

	input := struct {
		Items []Item `json:"items"`
	}{
		Items: []Item{
			{Name: "item1", Comment: "no special chars"},
			{Name: "item2", Comment: "has,comma"},
			{Name: "item3", Comment: "has\nnewline"},
		},
	}

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Values with special chars should be quoted in tabular output
	if !strings.Contains(result, "items[3]{name,comment}:") {
		t.Errorf("expected tabular header, got:\n%s", result)
	}
}

func TestEncodeSliceOfPointers(t *testing.T) {
	type Item struct {
		Name string `json:"name"`
	}

	item1 := &Item{Name: "first"}
	item2 := &Item{Name: "second"}

	input := struct {
		Items []*Item `json:"items"`
	}{
		Items: []*Item{item1, item2},
	}

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "items[2]{name}:") {
		t.Errorf("expected tabular header, got:\n%s", result)
	}
}

func TestEncodeNilSlice(t *testing.T) {
	input := struct {
		Items []string `json:"items,omitempty"`
	}{
		Items: nil,
	}

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be omitted due to omitempty
	if strings.Contains(result, "items") {
		t.Errorf("expected items to be omitted, got:\n%s", result)
	}
}

func TestEncodeUnexportedFields(t *testing.T) {
	type Item struct {
		Name     string `json:"name"`
		internal string // unexported, should be ignored
	}

	input := Item{Name: "test", internal: "hidden"}
	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "name: test") {
		t.Errorf("expected 'name: test', got:\n%s", result)
	}
	if strings.Contains(result, "hidden") {
		t.Errorf("unexported field should not appear, got:\n%s", result)
	}
}

func TestEncodeStructWithNoJSONTag(t *testing.T) {
	type Item struct {
		Name  string `json:"name"`
		Value int    // no json tag, uses field name
	}

	input := Item{Name: "test", Value: 42}
	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "name: test") {
		t.Errorf("expected 'name: test', got:\n%s", result)
	}
	// Field without json tag might use lowercase name
	if !strings.Contains(result, "Value: 42") && !strings.Contains(result, "value: 42") {
		t.Errorf("expected 'Value: 42' or 'value: 42', got:\n%s", result)
	}
}

func TestEncodeJSONDashTag(t *testing.T) {
	type Item struct {
		Name   string `json:"name"`
		Secret string `json:"-"` // should be ignored
	}

	input := Item{Name: "test", Secret: "hidden"}
	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "name: test") {
		t.Errorf("expected 'name: test', got:\n%s", result)
	}
	if strings.Contains(result, "hidden") || strings.Contains(result, "Secret") {
		t.Errorf("json:\"-\" field should not appear, got:\n%s", result)
	}
}

// Test string escaping through encoding
func TestStringEscaping(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{"with quotes", `say "hi"`, `\"hi\"`},
		{"with newline", "line1\nline2", `\n`},
		{"with tab", "a\tb", `\t`},
		{"with carriage return", "a\rb", `\r`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := struct {
				Value string `json:"value"`
			}{Value: tt.input}

			result, err := Encode(input)
			if err != nil {
				t.Fatalf("Encode error: %v", err)
			}

			if !strings.Contains(result, tt.contains) {
				t.Errorf("expected %q in output, got:\n%s", tt.contains, result)
			}
		})
	}
}

// Test primitive slices work correctly
func TestPrimitiveSlices(t *testing.T) {
	t.Run("string slice", func(t *testing.T) {
		input := struct {
			Values []string `json:"values"`
		}{Values: []string{"a", "b", "c"}}

		result, err := Encode(input)
		if err != nil {
			t.Fatalf("Encode error: %v", err)
		}

		if !strings.Contains(result, "values[3]: a,b,c") {
			t.Errorf("unexpected output:\n%s", result)
		}
	})

	t.Run("int slice", func(t *testing.T) {
		input := struct {
			Values []int `json:"values"`
		}{Values: []int{1, 2, 3}}

		result, err := Encode(input)
		if err != nil {
			t.Fatalf("Encode error: %v", err)
		}

		if !strings.Contains(result, "values[3]: 1,2,3") {
			t.Errorf("unexpected output:\n%s", result)
		}
	})

	t.Run("float slice", func(t *testing.T) {
		input := struct {
			Values []float64 `json:"values"`
		}{Values: []float64{1.1, 2.2}}

		result, err := Encode(input)
		if err != nil {
			t.Fatalf("Encode error: %v", err)
		}

		if !strings.Contains(result, "values[2]: 1.1,2.2") {
			t.Errorf("unexpected output:\n%s", result)
		}
	})

	t.Run("bool slice", func(t *testing.T) {
		input := struct {
			Values []bool `json:"values"`
		}{Values: []bool{true, false}}

		result, err := Encode(input)
		if err != nil {
			t.Fatalf("Encode error: %v", err)
		}

		if !strings.Contains(result, "values[2]: true,false") {
			t.Errorf("unexpected output:\n%s", result)
		}
	})
}

func TestEncodeMapWithStructValues(t *testing.T) {
	type Value struct {
		Count int `json:"count"`
	}

	input := struct {
		Data map[string]Value `json:"data"`
	}{
		Data: map[string]Value{
			"a": {Count: 1},
			"b": {Count: 2},
		},
	}

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "data:") {
		t.Errorf("expected 'data:' header, got:\n%s", result)
	}
}
