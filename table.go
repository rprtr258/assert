package assert

import (
	"fmt"
	"iter"
	"testing"
)

func TableMap[K comparable, V any](t *testing.T, table map[K]V, test func(*testing.T, K, V)) {
	t.Helper()
	t.Parallel()

	for k, testcase := range table {
		t.Run(fmt.Sprint(k), func(t *testing.T) {
			t.Helper()
			t.Parallel()

			test(t, k, testcase)
		})
	}
}

func Table[T any](t *testing.T, table map[string]T, test func(*testing.T, T)) {
	t.Helper()
	t.Parallel()

	for name, testcase := range table {
		t.Run(name, func(t *testing.T) {
			t.Helper()
			t.Parallel()

			test(t, testcase)
		})
	}
}

func TableSlice[T any](t *testing.T, table []T, test func(*testing.T, T)) {
	t.Helper()
	t.Parallel()

	for i, testcase := range table {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			t.Helper()
			t.Parallel()

			test(t, testcase)
		})
	}
}

func TableSeq[T any](t *testing.T, table iter.Seq[T], test func(*testing.T, T)) {
	t.Helper()
	t.Parallel()

	i := 0
	for testcase := range table {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			t.Helper()
			t.Parallel()

			test(t, testcase)
		})
		i++
	}
}
