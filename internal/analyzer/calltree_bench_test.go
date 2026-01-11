package analyzer

import (
	"testing"

	"github.com/CedricHerzog/perfowl/internal/testutil"
)

func BenchmarkAnalyzeCallTree(b *testing.B) {
	b.Run("Small", func(b *testing.B) {
		profile := testutil.SmallProfile()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeCallTree(profile, "", 20)
		}
	})

	b.Run("Medium", func(b *testing.B) {
		profile := testutil.MediumProfile()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeCallTree(profile, "", 20)
		}
	})

	b.Run("Large", func(b *testing.B) {
		profile := testutil.LargeProfile()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeCallTree(profile, "", 20)
		}
	})
}

func BenchmarkAnalyzeCallTree_DeepStack(b *testing.B) {
	b.Run("Depth50", func(b *testing.B) {
		profile := testutil.ProfileWithDeepCallStack(50, 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeCallTree(profile, "", 20)
		}
	})

	b.Run("Depth100", func(b *testing.B) {
		profile := testutil.ProfileWithDeepCallStack(100, 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeCallTree(profile, "", 20)
		}
	})

	b.Run("Depth200", func(b *testing.B) {
		profile := testutil.ProfileWithDeepCallStack(200, 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeCallTree(profile, "", 20)
		}
	})
}

func BenchmarkAnalyzeCallTree_ThreadFilter(b *testing.B) {
	profile := testutil.MediumProfile()

	b.Run("NoFilter", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeCallTree(profile, "", 20)
		}
	})

	b.Run("WithFilter", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeCallTree(profile, "GeckoMain", 20)
		}
	})
}

func BenchmarkAnalyzeCallTree_Limit(b *testing.B) {
	profile := testutil.LargeProfile()

	b.Run("Limit10", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeCallTree(profile, "", 10)
		}
	})

	b.Run("Limit50", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeCallTree(profile, "", 50)
		}
	})

	b.Run("Limit100", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeCallTree(profile, "", 100)
		}
	})
}

func BenchmarkFormatCallTree(b *testing.B) {
	profile := testutil.MediumProfile()
	analysis := AnalyzeCallTree(profile, "", 20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatCallTree(analysis)
	}
}
