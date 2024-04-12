// Internal fun replacement, just to remove dependency on lo and fun modules
package fun

func Ternary[T any](predicate bool, ifTrue, ifFalse T) T {
	if predicate {
		return ifTrue
	}
	return ifFalse
}

type Seq[T any] func(func(T) bool)

func FromMany[T any](xs ...T) Seq[T] {
	return func(yield func(T) bool) {
		for _, x := range xs {
			if !yield(x) {
				return
			}
		}
	}
}

func FromRange(from, to int) Seq[int] {
	return func(yield func(int) bool) {
		for i := from; i < to; i++ {
			if !yield(i) {
				return
			}
		}
	}
}

func FromDictKeys[K comparable, V any](m map[K]V) Seq[K] {
	return func(yield func(K) bool) {
		for k := range m {
			if !yield(k) {
				return
			}
		}
	}
}

func Map[T, R any](seq Seq[T], f func(T) R) Seq[R] {
	return func(yield func(R) bool) {
		seq(func(x T) bool {
			return yield(f(x))
		})
	}
}

func FlatMap[T, R any](seq Seq[T], f func(T) Seq[R]) Seq[R] {
	return func(yield func(R) bool) {
		seq(func(x T) bool {
			cont := true
			f(x)(func(y R) bool {
				if !yield(y) {
					cont = false
				}
				return cont
			})
			return cont
		})
	}
}
