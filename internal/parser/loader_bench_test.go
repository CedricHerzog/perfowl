package parser_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

// simpleProfile creates a minimal Firefox profile for benchmarking.
// Inlined here to avoid import cycle with testutil.
func simpleProfile(sampleCount int) *parser.Profile {
	strings := make([]string, 50)
	for i := 0; i < 50; i++ {
		strings[i] = fmt.Sprintf("function_%d", i)
	}

	// Build minimal stack/frame/func tables
	stackPrefix := make([]int, 50)
	stackFrame := make([]int, 50)
	stackCategory := make([]int, 50)
	for i := 0; i < 50; i++ {
		stackFrame[i] = i
		stackCategory[i] = 2
		if i > 0 {
			stackPrefix[i] = i - 1
		} else {
			stackPrefix[i] = -1
		}
	}

	frameFunc := make([]int, 50)
	frameCategory := make([]int, 50)
	for i := 0; i < 50; i++ {
		frameFunc[i] = i
		frameCategory[i] = 2
	}

	funcName := make([]int, 50)
	funcIsJS := make([]bool, 50)
	funcResource := make([]int, 50)
	for i := 0; i < 50; i++ {
		funcName[i] = i
		funcIsJS[i] = true
		funcResource[i] = -1
	}

	// Build samples
	sampleStack := make([]int, sampleCount)
	sampleTime := make([]float64, sampleCount)
	sampleCPUDelta := make([]int, sampleCount)
	for i := 0; i < sampleCount; i++ {
		sampleStack[i] = i % 50
		sampleTime[i] = float64(i)
		sampleCPUDelta[i] = 1000
	}

	return &parser.Profile{
		Meta: parser.Meta{
			StartTime: 0,
			Product:   "Firefox",
		},
		Threads: []parser.Thread{
			{
				Name:         "GeckoMain",
				TID:          "1",
				IsMainThread: true,
				StringArray:  strings,
				StackTable: parser.StackTable{
					Prefix:   stackPrefix,
					Frame:    stackFrame,
					Category: stackCategory,
				},
				FrameTable: parser.FrameTable{
					Func:     frameFunc,
					Category: frameCategory,
				},
				FuncTable: parser.FuncTable{
					Name:     funcName,
					IsJS:     funcIsJS,
					Resource: funcResource,
				},
				Samples: parser.Samples{
					Stack:          sampleStack,
					Time:           sampleTime,
					ThreadCPUDelta: sampleCPUDelta,
				},
			},
		},
	}
}

// chromeTraceData creates a Chrome trace JSON with N events for benchmarking.
func chromeTraceData(count int) []byte {
	events := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		events[i] = map[string]interface{}{
			"pid":  1,
			"tid":  1 + (i % 4),
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
			"product":   "Chrome",
			"userAgent": "Mozilla/5.0 Chrome/120.0.0.0",
		},
	}

	data, _ := json.Marshal(trace)
	return data
}

// tempProfileFile creates a temporary file with the profile JSON.
func tempProfileFile(b *testing.B, profile *parser.Profile) string {
	b.Helper()
	data, err := json.Marshal(profile)
	if err != nil {
		b.Fatalf("failed to marshal profile: %v", err)
	}

	f, err := os.CreateTemp("", "bench-profile-*.json")
	if err != nil {
		b.Fatalf("failed to create temp file: %v", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(data); err != nil {
		b.Fatalf("failed to write profile: %v", err)
	}

	b.Cleanup(func() { _ = os.Remove(f.Name()) })
	return f.Name()
}

// tempGzipProfileFile creates a gzipped temporary file with the profile JSON.
func tempGzipProfileFile(b *testing.B, profile *parser.Profile) string {
	b.Helper()
	data, err := json.Marshal(profile)
	if err != nil {
		b.Fatalf("failed to marshal profile: %v", err)
	}

	f, err := os.CreateTemp("", "bench-profile-*.json.gz")
	if err != nil {
		b.Fatalf("failed to create temp file: %v", err)
	}
	defer func() { _ = f.Close() }()

	gz := gzip.NewWriter(f)
	if _, err := gz.Write(data); err != nil {
		b.Fatalf("failed to write profile: %v", err)
	}
	_ = gz.Close()

	b.Cleanup(func() { _ = os.Remove(f.Name()) })
	return f.Name()
}

// tempChromeFile creates a temporary file with Chrome trace data.
func tempChromeFile(b *testing.B, data []byte) string {
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

func BenchmarkLoadProfile(b *testing.B) {
	b.Run("Small", func(b *testing.B) {
		profile := simpleProfile(100)
		path := tempProfileFile(b, profile)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = parser.LoadProfile(path)
		}
	})

	b.Run("Medium", func(b *testing.B) {
		profile := simpleProfile(1000)
		path := tempProfileFile(b, profile)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = parser.LoadProfile(path)
		}
	})

	b.Run("Large", func(b *testing.B) {
		profile := simpleProfile(10000)
		path := tempProfileFile(b, profile)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = parser.LoadProfile(path)
		}
	})
}

func BenchmarkLoadProfileFromReader(b *testing.B) {
	b.Run("Small", func(b *testing.B) {
		profile := simpleProfile(100)
		data, _ := json.Marshal(profile)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = parser.LoadProfileFromReader(bytes.NewReader(data))
		}
	})

	b.Run("Medium", func(b *testing.B) {
		profile := simpleProfile(1000)
		data, _ := json.Marshal(profile)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = parser.LoadProfileFromReader(bytes.NewReader(data))
		}
	})

	b.Run("Large", func(b *testing.B) {
		profile := simpleProfile(10000)
		data, _ := json.Marshal(profile)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = parser.LoadProfileFromReader(bytes.NewReader(data))
		}
	})
}

func BenchmarkLoadProfileGzip(b *testing.B) {
	b.Run("Small", func(b *testing.B) {
		profile := simpleProfile(100)
		path := tempGzipProfileFile(b, profile)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = parser.LoadProfile(path)
		}
	})

	b.Run("Medium", func(b *testing.B) {
		profile := simpleProfile(1000)
		path := tempGzipProfileFile(b, profile)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = parser.LoadProfile(path)
		}
	})

	b.Run("Large", func(b *testing.B) {
		profile := simpleProfile(10000)
		path := tempGzipProfileFile(b, profile)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = parser.LoadProfile(path)
		}
	})
}

func BenchmarkDetectBrowserType(b *testing.B) {
	b.Run("Firefox", func(b *testing.B) {
		profile := simpleProfile(100)
		path := tempProfileFile(b, profile)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = parser.DetectBrowserType(path)
		}
	})

	b.Run("Chrome", func(b *testing.B) {
		path := tempChromeFile(b, chromeTraceData(100))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = parser.DetectBrowserType(path)
		}
	})
}

func BenchmarkLoadProfileAuto(b *testing.B) {
	b.Run("Firefox", func(b *testing.B) {
		profile := simpleProfile(1000)
		path := tempProfileFile(b, profile)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = parser.LoadProfileAuto(path)
		}
	})

	b.Run("Chrome", func(b *testing.B) {
		path := tempChromeFile(b, chromeTraceData(1000))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = parser.LoadProfileAuto(path)
		}
	})
}
