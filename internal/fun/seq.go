package fun

import (
	"iter"
	"slices"
)

func FromMany[T any](xs ...T) iter.Seq[T] {
	return slices.Values(xs)
}

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
