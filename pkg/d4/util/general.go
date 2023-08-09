package util

import (
	"golang.org/x/exp/constraints"
)

func Max[T constraints.Ordered](a T, b T) T {
	if b > a {
		return b
	}
	return a
}
