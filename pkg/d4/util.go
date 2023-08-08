package d4

import (
	"golang.org/x/exp/constraints"
	"reflect"
	"sync"
	"unicode"
)

func newElem[T Object](t T) T {
	elemType := reflect.TypeOf(t).Elem()
	elemPtr := reflect.New(elemType)
	return elemPtr.Interface().(T)
}

func TrimNullTerminated(x []rune) string {
	for i, c := range x {
		if c == 0 {
			return string(x[:i])
		}
	}
	return string(x)
}

func IsIndex(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

func Max[T constraints.Ordered](a T, b T) T {
	if b > a {
		return b
	}
	return a
}

func Work[T any](workers uint, data []T, f func(T)) {
	// Add data to the channel
	c := make(chan T, len(data))
	for _, d := range data {
		c <- d
	}
	close(c)

	// Create worker
	wg := sync.WaitGroup{}

	worker := func() {
		defer wg.Done()
		for {
			item, ok := <-c
			if !ok {
				return
			}
			f(item)
		}
	}

	// Launch the workers
	for thread := uint(0); thread < workers; thread++ {
		wg.Add(1)
		go worker()
	}

	wg.Wait()
}
