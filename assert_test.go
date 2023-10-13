package assert

import (
	"testing"

	a "github.com/stretchr/testify/assert"
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
	a.Len(t, diffImpl("", User{"a", Pass{"b", "c"}}, User{"d", Pass{"e", "f"}}).ToSlice(), 0)
}
