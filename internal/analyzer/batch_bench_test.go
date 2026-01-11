package analyzer

import (
	"testing"

	"github.com/CedricHerzog/perfowl/internal/testutil"
)

func BenchmarkAnalyzeBatch(b *testing.B) {
	b.Run("1Profile", func(b *testing.B) {
		profile := testutil.MediumProfile()
		path := testutil.TempProfileFile(b, profile)
		profiles := []ProfileEntry{
			{
				Path:        path,
				WorkerCount: 1,
				Label:       "test",
			},
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = AnalyzeBatch(profiles)
		}
	})

	b.Run("5Profiles", func(b *testing.B) {
		profiles := make([]ProfileEntry, 5)
		for j := 0; j < 5; j++ {
			profile := testutil.MediumProfile()
			path := testutil.TempProfileFile(b, profile)
			profiles[j] = ProfileEntry{
				Path:        path,
				WorkerCount: j + 1,
				Label:       "test",
			}
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = AnalyzeBatch(profiles)
		}
	})

	b.Run("10Profiles", func(b *testing.B) {
		profiles := make([]ProfileEntry, 10)
		for j := 0; j < 10; j++ {
			profile := testutil.MediumProfile()
			path := testutil.TempProfileFile(b, profile)
			profiles[j] = ProfileEntry{
				Path:        path,
				WorkerCount: j + 1,
				Label:       "test",
			}
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = AnalyzeBatch(profiles)
		}
	})
}

func BenchmarkAnalyzeBatch_WithMeasurement(b *testing.B) {
	profiles := make([]ProfileEntry, 5)
	for j := 0; j < 5; j++ {
		profile := testutil.ProfileWithNMarkers(500)
		path := testutil.TempProfileFile(b, profile)
		profiles[j] = ProfileEntry{
			Path:        path,
			WorkerCount: j + 1,
			Label:       "test",
		}
	}

	b.Run("NoMeasurement", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = AnalyzeBatch(profiles)
		}
	})

	b.Run("WithMeasurement", func(b *testing.B) {
		// Add measurement patterns to each entry
		profilesWithMeasure := make([]ProfileEntry, len(profiles))
		copy(profilesWithMeasure, profiles)
		for j := range profilesWithMeasure {
			profilesWithMeasure[j].StartPattern = "GCMajor"
			profilesWithMeasure[j].EndPattern = "LongTask"
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = AnalyzeBatch(profilesWithMeasure)
		}
	})
}

func BenchmarkAnalyzeBatch_LargeProfiles(b *testing.B) {
	b.Run("3LargeProfiles", func(b *testing.B) {
		profiles := make([]ProfileEntry, 3)
		for j := 0; j < 3; j++ {
			profile := testutil.LargeProfile()
			path := testutil.TempProfileFile(b, profile)
			profiles[j] = ProfileEntry{
				Path:        path,
				WorkerCount: j + 1,
				Label:       "large",
			}
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = AnalyzeBatch(profiles)
		}
	})
}
