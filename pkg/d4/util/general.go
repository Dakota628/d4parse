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

func CombineSlices[T any](slices ...[]T) (out []T) {
	l := 0
	for _, slice := range slices {
		l += len(slice)
	}
	out = make([]T, 0, l)
	for _, slice := range slices {
		out = append(out, slice...)
	}
	return
}
