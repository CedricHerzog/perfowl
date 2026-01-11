package analyzer

import (
	"testing"

	"github.com/CedricHerzog/perfowl/internal/testutil"
)

func BenchmarkAnalyzeCrypto(b *testing.B) {
	b.Run("Small", func(b *testing.B) {
		profile := testutil.ProfileWithCryptoOperations(100)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeCrypto(profile)
		}
	})

	b.Run("Medium", func(b *testing.B) {
		profile := testutil.ProfileWithCryptoOperations(1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeCrypto(profile)
		}
	})

	b.Run("Large", func(b *testing.B) {
		profile := testutil.ProfileWithCryptoOperations(10000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeCrypto(profile)
		}
	})
}

func BenchmarkAnalyzeJSCrypto(b *testing.B) {
	b.Run("Small", func(b *testing.B) {
		profile := testutil.ProfileWithCryptoOperations(100)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeJSCrypto(profile)
		}
	})

	b.Run("Medium", func(b *testing.B) {
		profile := testutil.ProfileWithCryptoOperations(1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeJSCrypto(profile)
		}
	})

	b.Run("Large", func(b *testing.B) {
		profile := testutil.ProfileWithCryptoOperations(10000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			AnalyzeJSCrypto(profile)
		}
	})
}

func BenchmarkFormatCryptoAnalysis(b *testing.B) {
	profile := testutil.ProfileWithCryptoOperations(1000)
	analysis := AnalyzeCrypto(profile)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatCryptoAnalysis(analysis)
	}
}
