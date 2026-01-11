package analyzer

import (
	"testing"

	"github.com/CedricHerzog/perfowl/internal/testutil"
)

func BenchmarkDetectBottlenecks(b *testing.B) {
	b.Run("Small", func(b *testing.B) {
		profile := testutil.SmallProfile()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			DetectBottlenecks(profile)
		}
	})

	b.Run("Medium", func(b *testing.B) {
		profile := testutil.MediumProfile()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			DetectBottlenecks(profile)
		}
	})

	b.Run("Large", func(b *testing.B) {
		profile := testutil.LargeProfile()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			DetectBottlenecks(profile)
		}
	})
}

func BenchmarkDetectBottlenecks_ManyMarkers(b *testing.B) {
	b.Run("100Markers", func(b *testing.B) {
		profile := testutil.ProfileWithNMarkers(100)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			DetectBottlenecks(profile)
		}
	})

	b.Run("1000Markers", func(b *testing.B) {
		profile := testutil.ProfileWithNMarkers(1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			DetectBottlenecks(profile)
		}
	})

	b.Run("5000Markers", func(b *testing.B) {
		profile := testutil.ProfileWithNMarkers(5000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			DetectBottlenecks(profile)
		}
	})
}

func BenchmarkGenerateSummary(b *testing.B) {
	profile := testutil.MediumProfile()
	bottlenecks := DetectBottlenecks(profile)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateSummary(bottlenecks, profile)
	}
}

func BenchmarkCalculateScore(b *testing.B) {
	profile := testutil.MediumProfile()
	bottlenecks := DetectBottlenecks(profile)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateScore(bottlenecks)
	}
}
