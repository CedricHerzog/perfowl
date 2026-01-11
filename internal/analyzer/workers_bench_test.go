package analyzer

import (
	"testing"

	"github.com/CedricHerzog/perfowl/internal/testutil"
)

func BenchmarkAnalyzeWorkers(b *testing.B) {
	b.Run("Small", func(b *testing.B) {
		profile := testutil.SmallProfile()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeWorkers(profile)
		}
	})

	b.Run("Medium", func(b *testing.B) {
		profile := testutil.MediumProfile()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeWorkers(profile)
		}
	})

	b.Run("Large", func(b *testing.B) {
		profile := testutil.LargeProfile()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeWorkers(profile)
		}
	})
}

func BenchmarkAnalyzeWorkers_ManyWorkers(b *testing.B) {
	b.Run("1Worker", func(b *testing.B) {
		profile := testutil.ProfileWithNWorkers(1, 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeWorkers(profile)
		}
	})

	b.Run("4Workers", func(b *testing.B) {
		profile := testutil.ProfileWithNWorkers(4, 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeWorkers(profile)
		}
	})

	b.Run("8Workers", func(b *testing.B) {
		profile := testutil.ProfileWithNWorkers(8, 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeWorkers(profile)
		}
	})

	b.Run("16Workers", func(b *testing.B) {
		profile := testutil.ProfileWithNWorkers(16, 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeWorkers(profile)
		}
	})
}

func BenchmarkFormatWorkerAnalysis(b *testing.B) {
	profile := testutil.ProfileWithNWorkers(4, 1000)
	analysis := AnalyzeWorkers(profile)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatWorkerAnalysis(analysis)
	}
}
