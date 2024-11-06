package pa_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/rprtr258/assert/pa"
)

func TestMain(m *testing.M) {
	pa.Fuse()
	os.Exit(m.Run())
}

func TestExample(t *testing.T) {
	pa.Assert(t, 2+2 == 5)

	two := 1 + 1
	pa.Assert(t, two+2 == 5)

	xs := []int{1, 2}
	pa.Assert(t, reflect.DeepEqual(append(xs, 3), []int{1, 2, 4}))
	pa.Assert(t, reflect.DeepEqual(append(xs, 3)[1:], []int{2, 4}))

	factorial := func(n int) int {
		res := 1
		for i := 2; i <= n; i++ {
			res *= i
		}
		return res
	}
	pa.Assert(t, factorial(5) == 60)

	pa.Assert(t, xs[1] == 1)
	pa.Assert(t, -1 == 1)
	pa.Assert(t, (1+1) == 1)

	var s struct{ x struct{ y int } }
	pa.Assert(t, &s == nil)
	pa.Assert(t, s.x.y != 0)

	pa.Assert(t, (*int)(nil) != nil)
	pa.Assert(t, len(make([]int, 0)) == 1)
	pa.Assert(t, new(struct{}) == nil)
	pa.Assert(t, *new(int) == 1)
	pa.Assert(t, len(map[int]int{1: two}) == 0)
	pa.Assert(t, &s == &s)

	t.Run("require", func(t *testing.T) {
		pa.Require(t, two != 1+1)
	})
}
