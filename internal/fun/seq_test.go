package fun

import (
	"iter"
	"maps"
	"slices"
	"sort"
	"testing"

	"github.com/rprtr258/assert/internal/ass"
)

func TestFlatMap(t *testing.T) {
	seq := FromMany(1, 2, 3)

	ass.Equal(t, []int{1, 1, 2, 4, 3, 9}, slices.Collect(FlatMap(seq, func(n int) iter.Seq[int] {
		return FromMany(n, n*n)
	})))
}

func TestFromMany(t *testing.T) {
	ass.Equal(t, []int{1, 2, 3}, slices.Collect(FromMany(1, 2, 3)))
}

func TestMap(t *testing.T) {
	seq := FromMany(1, 2, 3)

	ass.Equal(t, []int{1, 4, 9}, slices.Collect(Map(seq, func(n int) int {
		return n * n
	})))
}

func TestFromDictKeys(t *testing.T) {
	dict := map[string]int{"one": 1, "two": 2}
	actual := slices.Collect(maps.Keys(dict))
	sort.Strings(actual)
	ass.Equal(t, []string{"one", "two"}, actual)
}

func TestFromRange(t *testing.T) {
	ass.Equal(t, []int{1, 2, 3}, slices.Collect(FromRange(1, 4)))
}

func TestToSlice(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		seq      iter.Seq[int]
		expected []int
	}{
		"Empty": {
			seq:      func(func(int) bool) {},
			expected: nil,
		},
		"Integers": {
			seq: func(yield func(int) bool) {
				for _, item := range []int{1, 2, 3} {
					if !yield(item) {
						break
					}
				}
			},
			expected: []int{1, 2, 3},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ass.Equal(t, test.expected, slices.Collect(test.seq))
		})
	}
}
