package testutil

import (
	"strings"
	"testing"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

// AssertNoError fails the test if err is not nil.
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// AssertError fails the test if err is nil.
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Error("expected an error, got nil")
	}
}

// AssertErrorContains fails the test if err is nil or doesn't contain the substring.
func AssertErrorContains(t *testing.T, err error, substring string) {
	t.Helper()
	if err == nil {
		t.Errorf("expected error containing %q, got nil", substring)
		return
	}
	if !strings.Contains(err.Error(), substring) {
		t.Errorf("expected error containing %q, got %q", substring, err.Error())
	}
}

// AssertEqual fails the test if got != want.
func AssertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

// AssertNotEqual fails the test if got == want.
func AssertNotEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got == want {
		t.Errorf("got %v, want different value", got)
	}
}

// AssertTrue fails the test if condition is false.
func AssertTrue(t *testing.T, condition bool, msg string) {
	t.Helper()
	if !condition {
		t.Errorf("expected true: %s", msg)
	}
}

// AssertFalse fails the test if condition is true.
func AssertFalse(t *testing.T, condition bool, msg string) {
	t.Helper()
	if condition {
		t.Errorf("expected false: %s", msg)
	}
}

// AssertNil fails the test if v is not nil.
func AssertNil(t *testing.T, v interface{}) {
	t.Helper()
	if v != nil {
		t.Errorf("expected nil, got %v", v)
	}
}

// AssertNotNil fails the test if v is nil.
func AssertNotNil(t *testing.T, v interface{}) {
	t.Helper()
	if v == nil {
		t.Error("expected non-nil value, got nil")
	}
}

// AssertStringContains fails the test if str doesn't contain substring.
func AssertStringContains(t *testing.T, str, substring string) {
	t.Helper()
	if !strings.Contains(str, substring) {
		t.Errorf("expected string to contain %q, got %q", substring, str)
	}
}

// AssertStringNotContains fails the test if str contains substring.
func AssertStringNotContains(t *testing.T, str, substring string) {
	t.Helper()
	if strings.Contains(str, substring) {
		t.Errorf("expected string to not contain %q, got %q", substring, str)
	}
}

// AssertStringHasPrefix fails the test if str doesn't start with prefix.
func AssertStringHasPrefix(t *testing.T, str, prefix string) {
	t.Helper()
	if !strings.HasPrefix(str, prefix) {
		t.Errorf("expected string to start with %q, got %q", prefix, str)
	}
}

// AssertStringHasSuffix fails the test if str doesn't end with suffix.
func AssertStringHasSuffix(t *testing.T, str, suffix string) {
	t.Helper()
	if !strings.HasSuffix(str, suffix) {
		t.Errorf("expected string to end with %q, got %q", suffix, str)
	}
}

// AssertIntInRange fails the test if value is not in [min, max].
func AssertIntInRange(t *testing.T, value, min, max int) {
	t.Helper()
	if value < min || value > max {
		t.Errorf("expected value in range [%d, %d], got %d", min, max, value)
	}
}

// AssertFloatInRange fails the test if value is not in [min, max].
func AssertFloatInRange(t *testing.T, value, min, max float64) {
	t.Helper()
	if value < min || value > max {
		t.Errorf("expected value in range [%f, %f], got %f", min, max, value)
	}
}

// AssertFloatApproxEqual fails the test if |got - want| > epsilon.
func AssertFloatApproxEqual(t *testing.T, got, want, epsilon float64) {
	t.Helper()
	diff := got - want
	if diff < 0 {
		diff = -diff
	}
	if diff > epsilon {
		t.Errorf("expected %f to be approximately %f (epsilon=%f), diff=%f", got, want, epsilon, diff)
	}
}

// AssertSliceLen fails the test if len(slice) != expected.
func AssertSliceLen[T any](t *testing.T, slice []T, expected int) {
	t.Helper()
	if len(slice) != expected {
		t.Errorf("expected slice length %d, got %d", expected, len(slice))
	}
}

// AssertSliceEmpty fails the test if slice is not empty.
func AssertSliceEmpty[T any](t *testing.T, slice []T) {
	t.Helper()
	if len(slice) != 0 {
		t.Errorf("expected empty slice, got %d elements", len(slice))
	}
}

// AssertSliceNotEmpty fails the test if slice is empty.
func AssertSliceNotEmpty[T any](t *testing.T, slice []T) {
	t.Helper()
	if len(slice) == 0 {
		t.Error("expected non-empty slice")
	}
}

// AssertMapHasKey fails the test if the map doesn't have the key.
func AssertMapHasKey[K comparable, V any](t *testing.T, m map[K]V, key K) {
	t.Helper()
	if _, ok := m[key]; !ok {
		t.Errorf("expected map to have key %v", key)
	}
}

// AssertMapNotHasKey fails the test if the map has the key.
func AssertMapNotHasKey[K comparable, V any](t *testing.T, m map[K]V, key K) {
	t.Helper()
	if _, ok := m[key]; ok {
		t.Errorf("expected map to not have key %v", key)
	}
}

// AssertMapLen fails the test if len(m) != expected.
func AssertMapLen[K comparable, V any](t *testing.T, m map[K]V, expected int) {
	t.Helper()
	if len(m) != expected {
		t.Errorf("expected map length %d, got %d", expected, len(m))
	}
}

// AssertGreater fails the test if a <= b.
func AssertGreater[T int | int64 | float64](t *testing.T, a, b T) {
	t.Helper()
	if a <= b {
		t.Errorf("expected %v > %v", a, b)
	}
}

// AssertGreaterOrEqual fails the test if a < b.
func AssertGreaterOrEqual[T int | int64 | float64](t *testing.T, a, b T) {
	t.Helper()
	if a < b {
		t.Errorf("expected %v >= %v", a, b)
	}
}

// AssertLess fails the test if a >= b.
func AssertLess[T int | int64 | float64](t *testing.T, a, b T) {
	t.Helper()
	if a >= b {
		t.Errorf("expected %v < %v", a, b)
	}
}

// AssertLessOrEqual fails the test if a > b.
func AssertLessOrEqual[T int | int64 | float64](t *testing.T, a, b T) {
	t.Helper()
	if a > b {
		t.Errorf("expected %v <= %v", a, b)
	}
}

// AssertPositive fails the test if value <= 0.
func AssertPositive[T int | int64 | float64](t *testing.T, value T) {
	t.Helper()
	if value <= 0 {
		t.Errorf("expected positive value, got %v", value)
	}
}

// AssertNonNegative fails the test if value < 0.
func AssertNonNegative[T int | int64 | float64](t *testing.T, value T) {
	t.Helper()
	if value < 0 {
		t.Errorf("expected non-negative value, got %v", value)
	}
}

// AssertZero fails the test if value != 0.
func AssertZero[T int | int64 | float64](t *testing.T, value T) {
	t.Helper()
	if value != 0 {
		t.Errorf("expected zero, got %v", value)
	}
}

// --- Parser-specific assertions ---

// AssertMarkerType fails the test if marker.Type != expected.
func AssertMarkerType(t *testing.T, marker parser.ParsedMarker, expected parser.MarkerType) {
	t.Helper()
	if marker.Type != expected {
		t.Errorf("expected marker type %s, got %s", expected, marker.Type)
	}
}

// AssertMarkerName fails the test if marker.Name != expected.
func AssertMarkerName(t *testing.T, marker parser.ParsedMarker, expected string) {
	t.Helper()
	if marker.Name != expected {
		t.Errorf("expected marker name %q, got %q", expected, marker.Name)
	}
}

// AssertMarkerCategory fails the test if marker.Category != expected.
func AssertMarkerCategory(t *testing.T, marker parser.ParsedMarker, expected string) {
	t.Helper()
	if marker.Category != expected {
		t.Errorf("expected marker category %q, got %q", expected, marker.Category)
	}
}

// AssertMarkerHasDuration fails the test if marker doesn't have duration.
func AssertMarkerHasDuration(t *testing.T, marker parser.ParsedMarker) {
	t.Helper()
	if !marker.IsDuration() {
		t.Error("expected marker to have duration")
	}
}

// AssertMarkerNoDuration fails the test if marker has duration.
func AssertMarkerNoDuration(t *testing.T, marker parser.ParsedMarker) {
	t.Helper()
	if marker.IsDuration() {
		t.Errorf("expected marker to have no duration, got %fms", marker.DurationMs())
	}
}

// AssertMarkerDurationInRange fails the test if duration is not in [min, max].
func AssertMarkerDurationInRange(t *testing.T, marker parser.ParsedMarker, min, max float64) {
	t.Helper()
	d := marker.DurationMs()
	if d < min || d > max {
		t.Errorf("expected marker duration in range [%f, %f]ms, got %fms", min, max, d)
	}
}

// AssertProfileDuration fails the test if profile duration != expected.
func AssertProfileDuration(t *testing.T, profile *parser.Profile, expected float64) {
	t.Helper()
	if profile.Duration() != expected {
		t.Errorf("expected profile duration %fms, got %fms", expected, profile.Duration())
	}
}

// AssertProfileDurationInRange fails the test if profile duration is not in [min, max].
func AssertProfileDurationInRange(t *testing.T, profile *parser.Profile, min, max float64) {
	t.Helper()
	d := profile.Duration()
	if d < min || d > max {
		t.Errorf("expected profile duration in range [%f, %f]ms, got %fms", min, max, d)
	}
}

// AssertThreadCount fails the test if profile thread count != expected.
func AssertThreadCount(t *testing.T, profile *parser.Profile, expected int) {
	t.Helper()
	if profile.ThreadCount() != expected {
		t.Errorf("expected %d threads, got %d", expected, profile.ThreadCount())
	}
}

// AssertExtensionCount fails the test if profile extension count != expected.
func AssertExtensionCount(t *testing.T, profile *parser.Profile, expected int) {
	t.Helper()
	if profile.ExtensionCount() != expected {
		t.Errorf("expected %d extensions, got %d", expected, profile.ExtensionCount())
	}
}

// AssertHasMainThread fails the test if profile has no main thread.
func AssertHasMainThread(t *testing.T, profile *parser.Profile) {
	t.Helper()
	if len(profile.GetMainThreads()) == 0 {
		t.Error("expected profile to have at least one main thread")
	}
}

// AssertCategoryExists fails the test if the category doesn't exist.
func AssertCategoryExists(t *testing.T, profile *parser.Profile, categoryName string) {
	t.Helper()
	for _, c := range profile.Meta.Categories {
		if c.Name == categoryName {
			return
		}
	}
	t.Errorf("expected category %q to exist", categoryName)
}
