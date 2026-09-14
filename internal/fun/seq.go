package fun

import (
	"iter"
)

func FromRange(from, to int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := from; i < to; i++ {
			if !yield(i) {
				return
			}
		}
	}
}

func Map[T, R any](seq iter.Seq[T], f func(T) R) iter.Seq[R] {
	return func(yield func(R) bool) {
		for x := range seq {
			if !yield(f(x)) {
				return
			}
		}
	}
}

func FlatMap[T, R any](seq iter.Seq[T], f func(T) iter.Seq[R]) iter.Seq[R] {
	return func(yield func(R) bool) {
		for x := range seq {
			for y := range f(x) {
				if !yield(y) {
					return
				}
			}
		}
	}
}

func SliceMap[T, R any](xs []T, f func(T) R) []R {
	res := make([]R, len(xs))
	for i, x := range xs {
		res[i] = f(x)
	}

	return res
}
