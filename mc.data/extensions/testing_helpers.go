package extensions

import (
	"math"
	"testing"
)

func AssertAreEqual[T comparable](t *testing.T, name string, expected T, actual T) {
	t.Helper()
	if expected != actual {
		t.Fatalf("value mismatch for %s, expected %v, got %v", name, expected, actual)
	}
}

func AssertNillability[T comparable](t *testing.T, name string, expected bool, actual *T) {
	t.Helper()
	if (actual == nil) != expected {
		t.Fatalf("value mismatch for %s, expected %v, got %v", name, expected, (actual == nil))
	}
}

// Helper: Check if slices are equal
func AssertSlicesEqual[T Number](t *testing.T, a, b []T) bool {
	t.Helper()
	if len(a) != len(b) {
		t.Fatalf("provided slices do not have the same lenth")
	}
	for i := range a {
		if math.Abs(float64(a[i])-float64(b[i])) > 1e-10 {
			t.Fatalf("provided slices have differing items: %v != %v at index %v", a[i], b[i], i)
		}
	}
	return true
}
