package testutil

import (
	"errors"
	"testing"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

// Note: These tests verify the assertion functions work correctly.
// They call assertions with valid data to exercise the code paths.
// Assertions that would fail are tested by verifying the failure condition.

func TestAssertNoError_Pass(t *testing.T) {
	AssertNoError(t, nil)
}

func TestAssertError_Pass(t *testing.T) {
	AssertError(t, errors.New("test error"))
}

func TestAssertErrorContains_Pass(t *testing.T) {
	AssertErrorContains(t, errors.New("test error message"), "error")
}

func TestAssertEqual_Pass(t *testing.T) {
	AssertEqual(t, 5, 5)
	AssertEqual(t, "test", "test")
	AssertEqual(t, 3.14, 3.14)
}

func TestAssertNotEqual_Pass(t *testing.T) {
	AssertNotEqual(t, 5, 10)
	AssertNotEqual(t, "test", "other")
}

func TestAssertTrue_Pass(t *testing.T) {
	AssertTrue(t, true, "should be true")
	AssertTrue(t, 1 == 1, "1 equals 1")
}

func TestAssertFalse_Pass(t *testing.T) {
	AssertFalse(t, false, "should be false")
	AssertFalse(t, 1 == 2, "1 does not equal 2")
}

func TestAssertNil_Pass(t *testing.T) {
	// Only test untyped nil - typed nil pointers don't compare equal to nil
	// due to Go's interface nil semantics
	AssertNil(t, nil)
}

func TestAssertNotNil_Pass(t *testing.T) {
	AssertNotNil(t, "not nil")
	AssertNotNil(t, 42)
	val := 5
	AssertNotNil(t, &val)
}

func TestAssertStringContains_Pass(t *testing.T) {
	AssertStringContains(t, "hello world", "world")
	AssertStringContains(t, "hello world", "hello")
	AssertStringContains(t, "testing", "test")
}

func TestAssertStringNotContains_Pass(t *testing.T) {
	AssertStringNotContains(t, "hello world", "xyz")
	AssertStringNotContains(t, "testing", "foo")
}

func TestAssertStringHasPrefix_Pass(t *testing.T) {
	AssertStringHasPrefix(t, "hello world", "hello")
	AssertStringHasPrefix(t, "testing", "test")
}

func TestAssertStringHasSuffix_Pass(t *testing.T) {
	AssertStringHasSuffix(t, "hello world", "world")
	AssertStringHasSuffix(t, "testing", "ing")
}

func TestAssertIntInRange_Pass(t *testing.T) {
	AssertIntInRange(t, 5, 1, 10)
	AssertIntInRange(t, 1, 1, 10)
	AssertIntInRange(t, 10, 1, 10)
}

func TestAssertFloatInRange_Pass(t *testing.T) {
	AssertFloatInRange(t, 5.5, 1.0, 10.0)
	AssertFloatInRange(t, 1.0, 1.0, 10.0)
	AssertFloatInRange(t, 10.0, 1.0, 10.0)
}

func TestAssertFloatApproxEqual_Pass(t *testing.T) {
	AssertFloatApproxEqual(t, 1.001, 1.0, 0.01)
	AssertFloatApproxEqual(t, 0.999, 1.0, 0.01)
	AssertFloatApproxEqual(t, 1.0, 1.0, 0.001)
}

func TestAssertSliceLen_Pass(t *testing.T) {
	AssertSliceLen(t, []int{1, 2, 3}, 3)
	AssertSliceLen(t, []string{}, 0)
	AssertSliceLen(t, []float64{1.0, 2.0}, 2)
}

func TestAssertSliceEmpty_Pass(t *testing.T) {
	AssertSliceEmpty(t, []int{})
	AssertSliceEmpty(t, []string{})
}

func TestAssertSliceNotEmpty_Pass(t *testing.T) {
	AssertSliceNotEmpty(t, []int{1, 2, 3})
	AssertSliceNotEmpty(t, []string{"a"})
}

func TestAssertMapHasKey_Pass(t *testing.T) {
	m := map[string]int{"key": 1, "other": 2}
	AssertMapHasKey(t, m, "key")
	AssertMapHasKey(t, m, "other")
}

func TestAssertMapNotHasKey_Pass(t *testing.T) {
	m := map[string]int{"key": 1}
	AssertMapNotHasKey(t, m, "nokey")
	AssertMapNotHasKey(t, m, "missing")
}

func TestAssertMapLen_Pass(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	AssertMapLen(t, m, 2)
	AssertMapLen(t, map[string]int{}, 0)
}

func TestAssertGreater_Pass(t *testing.T) {
	AssertGreater(t, 10, 5)
	AssertGreater(t, int64(100), int64(50))
	AssertGreater(t, 3.14, 2.71)
}

func TestAssertGreaterOrEqual_Pass(t *testing.T) {
	AssertGreaterOrEqual(t, 10, 10)
	AssertGreaterOrEqual(t, 10, 5)
	AssertGreaterOrEqual(t, 3.14, 3.14)
}

func TestAssertLess_Pass(t *testing.T) {
	AssertLess(t, 5, 10)
	AssertLess(t, int64(50), int64(100))
	AssertLess(t, 2.71, 3.14)
}

func TestAssertLessOrEqual_Pass(t *testing.T) {
	AssertLessOrEqual(t, 10, 10)
	AssertLessOrEqual(t, 5, 10)
	AssertLessOrEqual(t, 2.71, 2.71)
}

func TestAssertPositive_Pass(t *testing.T) {
	AssertPositive(t, 5)
	AssertPositive(t, int64(100))
	AssertPositive(t, 3.14)
}

func TestAssertNonNegative_Pass(t *testing.T) {
	AssertNonNegative(t, 0)
	AssertNonNegative(t, 5)
	AssertNonNegative(t, 0.0)
}

func TestAssertZero_Pass(t *testing.T) {
	AssertZero(t, 0)
	AssertZero(t, int64(0))
	AssertZero(t, 0.0)
}

// Parser-specific assertion tests

func TestAssertMarkerType_Pass(t *testing.T) {
	marker := parser.ParsedMarker{Type: parser.MarkerTypeGCMajor}
	AssertMarkerType(t, marker, parser.MarkerTypeGCMajor)
}

func TestAssertMarkerName_Pass(t *testing.T) {
	marker := parser.ParsedMarker{Name: "TestMarker"}
	AssertMarkerName(t, marker, "TestMarker")
}

func TestAssertMarkerCategory_Pass(t *testing.T) {
	marker := parser.ParsedMarker{Category: "JavaScript"}
	AssertMarkerCategory(t, marker, "JavaScript")
}

func TestAssertMarkerHasDuration_Pass(t *testing.T) {
	marker := parser.ParsedMarker{StartTime: 100, EndTime: 110, Duration: 10}
	AssertMarkerHasDuration(t, marker)
}

func TestAssertMarkerNoDuration_Pass(t *testing.T) {
	marker := parser.ParsedMarker{StartTime: 100, Duration: 0}
	AssertMarkerNoDuration(t, marker)
}

func TestAssertMarkerDurationInRange_Pass(t *testing.T) {
	marker := parser.ParsedMarker{StartTime: 100, EndTime: 110, Duration: 10}
	AssertMarkerDurationInRange(t, marker, 5, 15)
}

func TestAssertProfileDuration_Pass(t *testing.T) {
	profile := NewProfileBuilder().WithDuration(1000).Build()
	AssertProfileDuration(t, profile, 1000)
}

func TestAssertProfileDurationInRange_Pass(t *testing.T) {
	profile := NewProfileBuilder().WithDuration(1000).Build()
	AssertProfileDurationInRange(t, profile, 500, 1500)
}

func TestAssertThreadCount_Pass(t *testing.T) {
	profile := ProfileWithWorkers(3)
	AssertThreadCount(t, profile, 4) // 1 main + 3 workers
}

func TestAssertExtensionCount_Pass(t *testing.T) {
	profile := ProfileWithExtensions()
	AssertExtensionCount(t, profile, 2)
}

func TestAssertHasMainThread_Pass(t *testing.T) {
	profile := ProfileWithMainThread()
	AssertHasMainThread(t, profile)
}

func TestAssertCategoryExists_Pass(t *testing.T) {
	profile := NewProfileBuilder().WithCategories(DefaultCategories()).Build()
	AssertCategoryExists(t, profile, "JavaScript")
}
