package assert

import (
	"fmt"
	"testing"
)

func TableMap[K comparable, V any](t *testing.T, table map[K]V, test func(*testing.T, K, V)) {
	t.Parallel()

	for k, testcase := range table {
		k := k
		testcase := testcase
		t.Run(fmt.Sprint(k), func(t *testing.T) {
			t.Parallel()

			test(t, k, testcase)
		})
	}
}

func Table[T any](t *testing.T, table map[string]T, test func(*testing.T, T)) {
	t.Parallel()

	for name, testcase := range table {
		testcase := testcase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			test(t, testcase)
		})
	}
}

func TableSlice[T any](t *testing.T, table []T, test func(*testing.T, T)) {
	t.Parallel()

	for i, testcase := range table {
		testcase := testcase
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			t.Parallel()

			test(t, testcase)
		})
	}
}
