package assert

import (
	"testing"

	"github.com/rprtr258/assert/internal/ass"
	"github.com/rprtr258/assert/internal/fun"
)

type Pass struct {
	Payload, salt string
}

type User struct {
	Login string
	pass  Pass
}

func toSlice[T any](seq fun.Seq[T]) []T {
	var xs []T
	seq(func(x T) bool {
		xs = append(xs, x)
		return true
	})
	return xs
}

func TestDiffImpl(t *testing.T) {
	// must not panic on comparing structs in private field User.pass
	expected := []diffLine{
		{expected: "a", actual: "d", selector: ".Login"},
		{expected: "b", actual: "e", selector: ".pass.Payload"},
		{expected: "c", actual: "f", selector: ".pass.salt"},
	}
	actual := toSlice(diffImpl("",
		User{"a", Pass{"b", "c"}},
		User{"d", Pass{"e", "f"}},
	))
	ass.Equal(t, expected, actual)
}
