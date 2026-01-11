package parser_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

// chromeTraceWithNEvents creates a Chrome trace JSON with N events for benchmarking.
// Inlined here to avoid import cycle with testutil.
func chromeTraceWithNEvents(count int) []byte {
	events := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		events[i] = map[string]interface{}{
			"pid":  1,
			"tid":  1 + (i % 4), // Spread across 4 threads
			"ts":   i * 1000,
			"ph":   "X",
			"cat":  "devtools.timeline",
			"name": fmt.Sprintf("Function_%d", i%100),
			"dur":  500 + (i % 500),
		}
	}

	trace := map[string]interface{}{
		"traceEvents": events,
		"metadata": map[string]interface{}{
			"product":     "Chrome",
			"userAgent":   "Mozilla/5.0 Chrome/120.0.0.0",
			"cpuProfile":  true,
			"networkData": true,
		},
	}

	data, _ := json.Marshal(trace)
	return data
}

// tempChromeTraceFile creates a temporary file with Chrome trace data.
func tempChromeTraceFile(b *testing.B, data []byte) string {
	b.Helper()

	f, err := os.CreateTemp("", "bench-chrome-*.json")
	if err != nil {
		b.Fatalf("failed to create temp file: %v", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(data); err != nil {
		b.Fatalf("failed to write chrome trace: %v", err)
	}

	b.Cleanup(func() { _ = os.Remove(f.Name()) })
	return f.Name()
}

func BenchmarkLoadChromeProfile(b *testing.B) {
	b.Run("1000Events", func(b *testing.B) {
		data := chromeTraceWithNEvents(1000)
		path := tempChromeTraceFile(b, data)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = parser.LoadChromeProfile(path)
		}
	})

	b.Run("5000Events", func(b *testing.B) {
		data := chromeTraceWithNEvents(5000)
		path := tempChromeTraceFile(b, data)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = parser.LoadChromeProfile(path)
		}
	})

	b.Run("10000Events", func(b *testing.B) {
		data := chromeTraceWithNEvents(10000)
		path := tempChromeTraceFile(b, data)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = parser.LoadChromeProfile(path)
		}
	})
}

func BenchmarkConvertChromeToProfile(b *testing.B) {
	b.Run("1000Events", func(b *testing.B) {
		data := chromeTraceWithNEvents(1000)
		path := tempChromeTraceFile(b, data)
		chromeProfile, _ := parser.LoadChromeProfile(path)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = parser.ConvertChromeToProfile(chromeProfile)
		}
	})

	b.Run("5000Events", func(b *testing.B) {
		data := chromeTraceWithNEvents(5000)
		path := tempChromeTraceFile(b, data)
		chromeProfile, _ := parser.LoadChromeProfile(path)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = parser.ConvertChromeToProfile(chromeProfile)
		}
	})

	b.Run("10000Events", func(b *testing.B) {
		data := chromeTraceWithNEvents(10000)
		path := tempChromeTraceFile(b, data)
		chromeProfile, _ := parser.LoadChromeProfile(path)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = parser.ConvertChromeToProfile(chromeProfile)
		}
	})
}
