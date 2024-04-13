package fun

import (
	"sort"
	"testing"

	"github.com/rprtr258/assert/internal/ass"
)

func toSlice[T any](seq Seq[T]) []T {
	var xs []T
	seq(func(x T) bool {
		xs = append(xs, x)
		return true
	})
	return xs
}

func TestFlatMap(t *testing.T) {
	seq := FromMany(1, 2, 3)

	ass.Equal(t, []int{1, 1, 2, 4, 3, 9}, toSlice(FlatMap(seq, func(n int) Seq[int] {
		return FromMany(n, n*n)
	})))
}

func TestFromMany(t *testing.T) {
	ass.Equal(t, []int{1, 2, 3}, toSlice(FromMany(1, 2, 3)))
}

func TestMap(t *testing.T) {
	seq := FromMany(1, 2, 3)

	ass.Equal(t, []int{1, 4, 9}, toSlice(Map(seq, func(n int) int {
		return n * n
	})))
}

func TestFromDictKeys(t *testing.T) {
	dict := map[string]int{"one": 1, "two": 2}
	actual := toSlice(FromDictKeys(dict))
	sort.Strings(actual)
	ass.Equal(t, []string{"one", "two"}, actual)
}

func TestFromRange(t *testing.T) {
	ass.Equal(t, []int{1, 2, 3}, toSlice(FromRange(1, 4)))
}
