package assert

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, []diffLine{
		{expected: "a", actual: "d", selector: ".Login"},
		{expected: "b", actual: "e", selector: ".pass.Payload"},
		{expected: "c", actual: "f", selector: ".pass.salt"},
	}, diffImpl("",
		User{"a", Pass{"b", "c"}},
		User{"d", Pass{"e", "f"}},
	).ToSlice())
}
