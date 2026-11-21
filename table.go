package assert

import (
	"bytes"
	"fmt"
	"iter"
	"testing"
	"text/template"
)

func TableMap[K comparable, V any](t *testing.T, table map[K]V, test func(*testing.T, K, V)) {
	t.Helper()

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

	for name, testcase := range table {
		t.Run(name, func(t *testing.T) {
			t.Helper()
			t.Parallel()

			test(t, testcase)
		})
	}
}

// TableSlice runs test for each element of table in a parallel subtest named
// by executing the nameFormat template on the testcase.
func TableSlice[T any](t *testing.T, table []T, nameFormat string, test func(*testing.T, T)) {
	t.Helper()

	tmpl, err := template.New("testname").Parse(nameFormat)
	if err != nil {
		t.Fatalf("invalid test name format %q: %v", nameFormat, err)
	}

	for i, testcase := range table {
		var name bytes.Buffer
		if err := tmpl.Execute(&name, testcase); err != nil {
			t.Fatalf("test name format %q failed on testcase #%d: %v", nameFormat, i, err)
		}

		t.Run(name.String(), func(t *testing.T) {
			t.Helper()
			t.Parallel()

			test(t, testcase)
		})
	}
}

func TableSeq[T any](t *testing.T, table iter.Seq[T], test func(*testing.T, T)) {
	t.Helper()

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
