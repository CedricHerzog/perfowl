package toon

import (
	"testing"

	"github.com/CedricHerzog/perfowl/internal/analyzer"
	"github.com/CedricHerzog/perfowl/internal/testutil"
)

func BenchmarkEncode_Small(b *testing.B) {
	profile := testutil.SmallProfile()
	analysis := analyzer.AnalyzeCallTree(profile, "", 20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Encode(analysis)
	}
}

func BenchmarkEncode_Medium(b *testing.B) {
	profile := testutil.MediumProfile()
	analysis := analyzer.AnalyzeCallTree(profile, "", 50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Encode(analysis)
	}
}

func BenchmarkEncode_Large(b *testing.B) {
	profile := testutil.LargeProfile()
	analysis := analyzer.AnalyzeCallTree(profile, "", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Encode(analysis)
	}
}

func BenchmarkEncode_Bottlenecks(b *testing.B) {
	profile := testutil.ProfileWithNMarkers(500)
	bottlenecks := analyzer.DetectBottlenecks(profile)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Encode(bottlenecks)
	}
}

func BenchmarkEncode_Workers(b *testing.B) {
	profile := testutil.ProfileWithNWorkers(8, 1000)
	analysis := analyzer.AnalyzeWorkers(profile)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Encode(analysis)
	}
}

func BenchmarkEncode_TabularArray(b *testing.B) {
	// Create a struct slice that will trigger tabular encoding
	type Item struct {
		Name     string
		Value    int
		Duration float64
	}

	b.Run("10Items", func(b *testing.B) {
		items := make([]Item, 10)
		for i := 0; i < 10; i++ {
			items[i] = Item{Name: "item", Value: i, Duration: float64(i) * 1.5}
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = Encode(items)
		}
	})

	b.Run("100Items", func(b *testing.B) {
		items := make([]Item, 100)
		for i := 0; i < 100; i++ {
			items[i] = Item{Name: "item", Value: i, Duration: float64(i) * 1.5}
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = Encode(items)
		}
	})

	b.Run("1000Items", func(b *testing.B) {
		items := make([]Item, 1000)
		for i := 0; i < 1000; i++ {
			items[i] = Item{Name: "item", Value: i, Duration: float64(i) * 1.5}
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = Encode(items)
		}
	})
}

func BenchmarkEncode_DeepNesting(b *testing.B) {
	type Inner struct {
		Value int
	}
	type Middle struct {
		Inner Inner
		Name  string
	}
	type Outer struct {
		Middle Middle
		ID     int
	}

	data := Outer{
		ID: 1,
		Middle: Middle{
			Name:  "test",
			Inner: Inner{Value: 42},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Encode(data)
	}
}

func BenchmarkEncode_Map(b *testing.B) {
	b.Run("SmallMap", func(b *testing.B) {
		m := make(map[string]float64)
		for i := 0; i < 10; i++ {
			m["key"+string(rune('A'+i))] = float64(i) * 1.5
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = Encode(m)
		}
	})

	b.Run("LargeMap", func(b *testing.B) {
		m := make(map[string]float64)
		for i := 0; i < 100; i++ {
			m["key"+string(rune('A'+i%26))+string(rune('0'+i/26))] = float64(i) * 1.5
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = Encode(m)
		}
	})
}

func BenchmarkEncodeIndent(b *testing.B) {
	profile := testutil.MediumProfile()
	analysis := analyzer.AnalyzeCallTree(profile, "", 20)

	b.Run("NoIndent", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = EncodeIndent(analysis, "", "")
		}
	})

	b.Run("WithIndent", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = EncodeIndent(analysis, "", "  ")
		}
	})
}
