// Internal testing helpers in order to not depend on testify/assert

package ass

import (
	"reflect"
	"strings"
	"testing"
)

func True(t *testing.T, condition bool) {
	t.Helper()

	if !condition {
		t.Errorf("Must be true")
	}
}

func False(t *testing.T, condition bool) {
	t.Helper()

	if condition {
		t.Errorf("Must be false")
	}
}

func Equal[T any](t *testing.T, expected, actula T) {
	t.Helper()

	if !reflect.DeepEqual(expected, actula) {
		t.Errorf("Not Equal\nExpected: %v\nActual: %v", expected, actula)
	}
}

func NotEqual[T any](t *testing.T, expected, actula T) {
	t.Helper()

	if reflect.DeepEqual(expected, actula) {
		t.Errorf("Expected and actual are equal, while should not:\n%v", expected)
	}
}

func Contains[T any](t *testing.T, elem T, collection ...T) {
	t.Helper()

	for _, item := range collection {
		if reflect.DeepEqual(elem, item) {
			return
		}
	}
	t.Errorf("Not contains\nElement: %v\nCollection: %v", elem, collection)
}

func ContainsNot[T any](t *testing.T, elem T, collection ...T) {
	t.Helper()

	for _, item := range collection {
		if reflect.DeepEqual(elem, item) {
			t.Errorf("Contains\nElement: %v\nCollection: %v", elem, collection)
			return
		}
	}
}

func ContainsIs[T any](t *testing.T, shouldContain bool, elem T, collection ...T) {
	t.Helper()

	if shouldContain {
		Contains(t, elem, collection...)
	} else {
		ContainsNot(t, elem, collection...)
	}
}

func SContains(t *testing.T, needle, haystack string) {
	t.Helper()

	if !strings.Contains(haystack, needle) {
		t.Errorf("Not scontains\nNeedle: %q\nHaystack: %q", needle, haystack)
	}
}

func SContainsNot(t *testing.T, needle, haystack string) {
	t.Helper()

	if strings.Contains(haystack, needle) {
		t.Errorf("Scontains\nNeedle: %q\nHaystack: %q", needle, haystack)
	}
}

func SContainsIs(t *testing.T, shouldContain bool, needle, haystack string) {
	t.Helper()

	if shouldContain {
		SContains(t, needle, haystack)
	} else {
		SContainsNot(t, needle, haystack)
	}
}
