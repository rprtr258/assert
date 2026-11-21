package assert

import (
	"strings"
	"sync"
	"testing"

	"github.com/rprtr258/assert/internal/ass"
)

func TestTableSlice(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	table := []string{"a", "b", "c"}
	names := map[string]struct{}{
		"case_a": {},
		"case_b": {},
		"case_c": {},
	}

	t.Cleanup(func() { ass.Equal(t, 0, len(names)) })
	TableSlice(t, table, "case_{{.}}", func(t *testing.T, tc string) { //nolint:thelper // subtest callback, not a helper
		mu.Lock()
		_, name, _ := strings.Cut(t.Name(), "/")
		_, ok := names[name]
		ass.True(t, ok)
		delete(names, name)
		mu.Unlock()
	})
}

func TestTableSlice2(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	type testcase = struct {
		A, B string
		C    int
	}
	table := []testcase{
		{"a", "b", 1},
		{"A", "B", 2},
	}
	names := map[string]struct{}{
		"a_*_b_=_1": {},
		"A_*_B_=_2": {},
	}

	t.Cleanup(func() { ass.Equal(t, 0, len(names)) })
	TableSlice(t, table, "{{.A}} * {{.B}} = {{.C}}",
		func(t *testing.T, tc testcase) { //nolint:thelper // subtest callback, not a helper
			mu.Lock()
			_, name, _ := strings.Cut(t.Name(), "/")
			_, ok := names[name]
			ass.True(t, ok)
			delete(names, name)
			mu.Unlock()
		})
}
