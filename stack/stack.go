package stack

import (
	. "github.com/cvanloo/blog-go/assert"
)

type (
	Stack[T any] []T
	Maybe[T any] struct {
		HasValue bool
		Value T
	}
)

func (s Stack[T]) Push(v T) Stack[T] {
	return append(s, v)
}

func (s Stack[T]) Pop() (Stack[T], T) {
	l := len(s)
	Assert(l > 0, "Pop called on empty stack (maybe you want to use SafePop?)")
	return s[:l-1], s[l-1]
}

func (s Stack[T]) SafePop() (Stack[T], Maybe[T]) {
	l := len(s)
	if l > 0 {
		return s[:l-1], Maybe[T]{true, s[l-1]}
	}
	return s, Maybe[T]{HasValue: false}
}

func (s Stack[T]) Peek() T {
	l := len(s)
	return s[l-1]
}

func (s Stack[T]) Empty() (empty Stack[T]) {
	return
}
