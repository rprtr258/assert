// Package fun is just a little fun/lo replacement.
package fun

func Ternary[T any](predicate bool, ifTrue, ifFalse T) T {
	if predicate {
		return ifTrue
	}

	return ifFalse
}
