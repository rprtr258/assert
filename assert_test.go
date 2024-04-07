package assert

import (
	"testing"

	"github.com/rprtr258/assert/internal/ass"
)

type Pass struct {
	Payload, salt string
}

type User struct {
	Login string
	pass  Pass
}

func TestDiffImpl(t *testing.T) {
	// must not panic on comparing structs in private field User.pass
	expected := []diffLine{
		{expected: "a", actual: "d", selector: ".Login"},
		{expected: "b", actual: "e", selector: ".pass.Payload"},
		{expected: "c", actual: "f", selector: ".pass.salt"},
	}
	actual := diffImpl("",
		User{"a", Pass{"b", "c"}},
		User{"d", Pass{"e", "f"}},
	).ToSlice()
	ass.Equal(t, expected, actual)
}
