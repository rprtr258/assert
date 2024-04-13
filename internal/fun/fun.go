// Internal fun replacement, just to remove dependency on lo and fun modules
package fun

func Ternary[T any](predicate bool, ifTrue, ifFalse T) T {
	if predicate {
		return ifTrue
	}
	return ifFalse
}
