package assert_test

import (
	"reflect"
	"testing"

	"github.com/rprtr258/assert"
)

func TestExample(t *testing.T) {
	assert.Assert(t, 2+2 == 5)

	two := 1 + 1
	assert.Assert(t, two+2 == 5)

	xs := []int{1, 2}
	assert.Assert(t, reflect.DeepEqual(append(xs, 3), []int{1, 2, 4}))
	assert.Assert(t, reflect.DeepEqual(append(xs, 3)[1:], []int{2, 4}))

	factorial := func(n int) int {
		res := 1
		for i := 2; i <= n; i++ {
			res *= i
		}
		return res
	}
	assert.Assert(t, factorial(5) == 60)

	assert.Assert(t, xs[1] == 1)
	assert.Assert(t, -1 == 1)
	assert.Assert(t, (1+1) == 1)

	var s struct{ x struct{ y int } }
	assert.Assert(t, &s == nil)
	assert.Assert(t, s.x.y != 0)

	assert.Assert(t, (*int)(nil) != nil)
	assert.Assert(t, len(make([]int, 0)) == 1)
	assert.Assert(t, new(struct{}) == nil)
	assert.Assert(t, *new(int) == 1)
	assert.Assert(t, len(map[int]int{1: two}) == 0)
	assert.Assert(t, &s == &s)

	t.Run("require", func(t *testing.T) {
		assert.Require(t, two != 1+1)
	})
}
