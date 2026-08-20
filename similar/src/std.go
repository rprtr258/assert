package src

import (
	"fmt"
	"slices"
)

type Usize int

func (self Usize) CheckedSub(rhs Usize) Option[Usize] {
	return Option[Usize]{self >= rhs, self - rhs}
}

func (self Usize) abs() Usize {
	return max(self, -self)
}

type Vec[T any] []T

func (self Vec[T]) len() Usize {
	return Usize(len(self))
}

func (self Vec[T]) Get(index Usize) Option[T] {
	if index < Usize(len(self)) {
		return Some(self[index])
	} else {
		return None[T]()
	}
}

func (self *Vec[T]) remove(index Usize) {
	*self = slices.Delete(*self, int(index), int(index)+1)
}

func (self *Vec[T]) insert(index Usize, v T) {
	*self = slices.Insert(*self, int(index), v)
}

func (self *Vec[T]) swap(a, b Usize) {
	(*self)[a], (*self)[b] = (*self)[b], (*self)[a]
}

// https://doc.rust-lang.org/stable/std/ops/struct.Range.html
type Range struct {
	start, end Usize
}

func (self Range) len() Usize {
	return Usize(self.end - self.start)
}

type Option[T any] struct {
	Valid bool
	Value T
}

func None[T any]() Option[T] {
	return Option[T]{Valid: false}
}

func Some[T any](value T) Option[T] {
	return Option[T]{Valid: true, Value: value}
}

func (self Option[T]) Unpack() (T, bool) {
	return self.Value, self.Valid
}

func Map[T, R any](self Option[T], f func(T) R) Option[R] {
	if self.Valid {
		return Some(f(self.Value))
	} else {
		return None[R]()
	}
}

func AndThen[T, R any](self Option[T], f func(T) Option[R]) Option[R] {
	if self.Valid {
		return f(self.Value)
	} else {
		return None[R]()
	}
}

type Result[T any] struct {
	Ok  bool
	Val T
	Err error
}

func Ok[T any](val T) Result[T] {
	return Result[T]{Ok: true, Val: val}
}

func (self Result[T]) unwrap() T {
	if !self.Ok {
		panic(fmt.Sprint(self.Err))
	}
	return self.Val
}

type Unit struct{}
