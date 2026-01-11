package parser_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

// profileWithMarkers creates a profile with N markers of various types for benchmarking.
// Inlined here to avoid import cycle with testutil.
func profileWithMarkers(count int) *parser.Profile {
	// Build marker data
	markerNames := make([]int, count)
	markerStartTimes := make([]float64, count)
	markerEndTimes := make([]interface{}, count)
	markerPhases := make([]int, count)
	markerCategories := make([]int, count)
	markerData := make([]json.RawMessage, count)

	strings := []string{
		"GCMajor", "GCMinor", "MainThreadLongTask", "Styles",
		"IPC", "DOMEvent", "Network", "Paint",
	}

	for i := 0; i < count; i++ {
		markerNames[i] = i % 8
		markerStartTimes[i] = float64(i * 2)
		markerEndTimes[i] = float64(i*2 + 10 + i%50)
		markerPhases[i] = 1 // Interval
		markerCategories[i] = (i % 4) + 1

		// Add marker-specific data as JSON
		var data interface{}
		switch i % 8 {
		case 0: // GCMajor
			data = map[string]interface{}{"type": "GCMajor"}
		case 1: // GCMinor
			data = map[string]interface{}{"type": "GCMinor"}
		case 2: // LongTask
			data = map[string]interface{}{"type": "MainThreadLongTask"}
		case 3: // Styles
			data = map[string]interface{}{"type": "Styles"}
		case 4: // IPC
			data = map[string]interface{}{"type": "IPC", "sync": true}
		case 5: // DOMEvent
			data = map[string]interface{}{"type": "DOMEvent", "eventType": "click"}
		case 6: // Network
			data = map[string]interface{}{
				"type": "Network",
				"URI":  fmt.Sprintf("https://example.com/api/%d", i),
			}
		case 7: // Paint
			data = map[string]interface{}{"type": "Paint"}
		}
		markerData[i], _ = json.Marshal(data)
	}

	categories := []parser.Category{
		{Name: "Other", Color: "grey"},
		{Name: "JavaScript", Color: "yellow"},
		{Name: "Layout", Color: "purple"},
		{Name: "Graphics", Color: "green"},
		{Name: "Network", Color: "blue"},
		{Name: "GC / CC", Color: "orange"},
	}

	return &parser.Profile{
		Meta: parser.Meta{
			StartTime:  0,
			Product:    "Firefox",
			Categories: categories,
		},
		Threads: []parser.Thread{
			{
				Name:         "GeckoMain",
				TID:          "1",
				IsMainThread: true,
				StringArray:  strings,
				Markers: parser.Markers{
					Length:    count,
					Name:      markerNames,
					StartTime: markerStartTimes,
					EndTime:   markerEndTimes,
					Phase:     markerPhases,
					Category:  markerCategories,
					Data:      markerData,
				},
			},
		},
	}
}

func BenchmarkExtractMarkers(b *testing.B) {
	b.Run("100Markers", func(b *testing.B) {
		profile := profileWithMarkers(100)
		thread := &profile.Threads[0]
		categories := profile.Meta.Categories
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			parser.ExtractMarkers(thread, categories)
		}
	})

	b.Run("1000Markers", func(b *testing.B) {
		profile := profileWithMarkers(1000)
		thread := &profile.Threads[0]
		categories := profile.Meta.Categories
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			parser.ExtractMarkers(thread, categories)
		}
	})

	b.Run("5000Markers", func(b *testing.B) {
		profile := profileWithMarkers(5000)
		thread := &profile.Threads[0]
		categories := profile.Meta.Categories
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			parser.ExtractMarkers(thread, categories)
		}
	})
}

func BenchmarkFilterMarkersByType(b *testing.B) {
	profile := profileWithMarkers(1000)
	thread := &profile.Threads[0]
	markers := parser.ExtractMarkers(thread, profile.Meta.Categories)

	b.Run("GCMajor", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			parser.FilterMarkersByType(markers, parser.MarkerTypeGCMajor)
		}
	})

	b.Run("LongTask", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			parser.FilterMarkersByType(markers, parser.MarkerTypeMainThreadLongTask)
		}
	})
}

func BenchmarkFilterMarkersByCategory(b *testing.B) {
	profile := profileWithMarkers(1000)
	thread := &profile.Threads[0]
	markers := parser.ExtractMarkers(thread, profile.Meta.Categories)

	b.Run("GC", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			parser.FilterMarkersByCategory(markers, "GC")
		}
	})

	b.Run("Network", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			parser.FilterMarkersByCategory(markers, "Network")
		}
	})
}

func BenchmarkFilterMarkersByDuration(b *testing.B) {
	profile := profileWithMarkers(1000)
	thread := &profile.Threads[0]
	markers := parser.ExtractMarkers(thread, profile.Meta.Categories)

	b.Run("Above10ms", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			parser.FilterMarkersByDuration(markers, 10)
		}
	})

	b.Run("Above50ms", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			parser.FilterMarkersByDuration(markers, 50)
		}
	})
}

func BenchmarkGetMarkerStats(b *testing.B) {
	profile := profileWithMarkers(1000)
	thread := &profile.Threads[0]
	markers := parser.ExtractMarkers(thread, profile.Meta.Categories)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.GetMarkerStats(markers)
	}
}
