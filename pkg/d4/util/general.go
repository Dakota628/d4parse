package util

import (
	"golang.org/x/exp/constraints"
)

func Min[T constraints.Ordered](a T, b T) T {
	if b < a {
		return b
	}
	return a
}

func Max[T constraints.Ordered](a T, b T) T {
	if b > a {
		return b
	}
	return a
}
