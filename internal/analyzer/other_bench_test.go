package analyzer

import (
	"testing"

	"github.com/CedricHerzog/perfowl/internal/testutil"
)

// Category Analysis Benchmarks

func BenchmarkAnalyzeCategories(b *testing.B) {
	b.Run("Small", func(b *testing.B) {
		profile := testutil.SmallProfile()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeCategories(profile, "")
		}
	})

	b.Run("Medium", func(b *testing.B) {
		profile := testutil.MediumProfile()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeCategories(profile, "")
		}
	})

	b.Run("Large", func(b *testing.B) {
		profile := testutil.LargeProfile()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeCategories(profile, "")
		}
	})
}

func BenchmarkAnalyzeCategories_ThreadFilter(b *testing.B) {
	profile := testutil.MediumProfile()

	b.Run("NoFilter", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeCategories(profile, "")
		}
	})

	b.Run("WithFilter", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeCategories(profile, "GeckoMain")
		}
	})
}

// Thread Analysis Benchmarks

func BenchmarkAnalyzeThreads(b *testing.B) {
	b.Run("Small", func(b *testing.B) {
		profile := testutil.SmallProfile()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeThreads(profile)
		}
	})

	b.Run("Medium", func(b *testing.B) {
		profile := testutil.MediumProfile()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeThreads(profile)
		}
	})

	b.Run("Large", func(b *testing.B) {
		profile := testutil.LargeProfile()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeThreads(profile)
		}
	})
}

func BenchmarkAnalyzeThreads_ManyWorkers(b *testing.B) {
	b.Run("4Workers", func(b *testing.B) {
		profile := testutil.ProfileWithNWorkers(4, 500)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeThreads(profile)
		}
	})

	b.Run("16Workers", func(b *testing.B) {
		profile := testutil.ProfileWithNWorkers(16, 500)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeThreads(profile)
		}
	})
}

// Extension Analysis Benchmarks

func BenchmarkAnalyzeExtensions(b *testing.B) {
	b.Run("2Extensions", func(b *testing.B) {
		profile := testutil.ProfileWithManyExtensions(2, 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeExtensions(profile)
		}
	})

	b.Run("10Extensions", func(b *testing.B) {
		profile := testutil.ProfileWithManyExtensions(10, 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeExtensions(profile)
		}
	})

	b.Run("50Extensions", func(b *testing.B) {
		profile := testutil.ProfileWithManyExtensions(50, 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeExtensions(profile)
		}
	})
}

// Contention Analysis Benchmarks

func BenchmarkAnalyzeContention(b *testing.B) {
	b.Run("Small", func(b *testing.B) {
		profile := testutil.SmallProfile()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeContention(profile)
		}
	})

	b.Run("Medium", func(b *testing.B) {
		profile := testutil.MediumProfile()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeContention(profile)
		}
	})

	b.Run("Large", func(b *testing.B) {
		profile := testutil.LargeProfile()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeContention(profile)
		}
	})
}

// Scaling Analysis Benchmarks

func BenchmarkAnalyzeScaling(b *testing.B) {
	b.Run("1Worker", func(b *testing.B) {
		profile := testutil.ProfileWithNWorkers(1, 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeScaling(profile)
		}
	})

	b.Run("4Workers", func(b *testing.B) {
		profile := testutil.ProfileWithNWorkers(4, 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeScaling(profile)
		}
	})

	b.Run("16Workers", func(b *testing.B) {
		profile := testutil.ProfileWithNWorkers(16, 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeScaling(profile)
		}
	})
}

func BenchmarkCompareScaling(b *testing.B) {
	baseline := testutil.ProfileWithNWorkers(1, 1000)
	comparison := testutil.ProfileWithNWorkers(4, 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CompareScaling(baseline, comparison)
	}
}

// Profile Comparison Benchmarks

func BenchmarkCompareProfiles(b *testing.B) {
	b.Run("Small", func(b *testing.B) {
		baseline := testutil.SmallProfile()
		comparison := testutil.SmallProfile()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			CompareProfiles(baseline, comparison)
		}
	})

	b.Run("Medium", func(b *testing.B) {
		baseline := testutil.MediumProfile()
		comparison := testutil.MediumProfile()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			CompareProfiles(baseline, comparison)
		}
	})

	b.Run("Large", func(b *testing.B) {
		baseline := testutil.LargeProfile()
		comparison := testutil.LargeProfile()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			CompareProfiles(baseline, comparison)
		}
	})
}

// Delimiter/Measurement Benchmarks

func BenchmarkGetDelimiterMarkers(b *testing.B) {
	b.Run("100Markers", func(b *testing.B) {
		profile := testutil.ProfileWithNMarkers(100)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			GetDelimiterMarkers(profile, nil)
		}
	})

	b.Run("1000Markers", func(b *testing.B) {
		profile := testutil.ProfileWithNMarkers(1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			GetDelimiterMarkers(profile, nil)
		}
	})

	b.Run("5000Markers", func(b *testing.B) {
		profile := testutil.ProfileWithNMarkers(5000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			GetDelimiterMarkers(profile, nil)
		}
	})
}

func BenchmarkMeasureOperationAdvanced(b *testing.B) {
	profile := testutil.ProfileWithNMarkers(1000)

	b.Run("PatternBased", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = MeasureOperationAdvanced(profile, MeasureOptions{
				StartPattern: "GCMajor",
				EndPattern:   "MainThreadLongTask",
			})
		}
	})

	b.Run("WithTimeBounds", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = MeasureOperationAdvanced(profile, MeasureOptions{
				StartPattern: "GCMajor",
				EndPattern:   "MainThreadLongTask",
				StartAfterMs: 100,
				EndBeforeMs:  500,
			})
		}
	})
}
