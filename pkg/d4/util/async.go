package util

import (
	"errors"
	"sync"
)

func DoWorkChan[T any](workers uint, c chan T, f func(T)) {
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

func DoWorkSlice[T any](workers uint, data []T, f func(T)) {
	workers = Min(workers, uint(len(data)))

	// Add data to the channel
	c := make(chan T, len(data))
	for _, d := range data {
		c <- d
	}
	close(c)

	// Do the work
	DoWorkChan(workers, c, f)
}

type doWorkMapItem[K comparable, V any] struct {
	key   K
	value V
}

func DoWorkMap[K comparable, V any](workers uint, data map[K]V, f func(K, V)) {
	workers = Min(workers, uint(len(data)))

	// Add data to the channel
	c := make(chan doWorkMapItem[K, V], len(data))
	for k, v := range data {
		c <- doWorkMapItem[K, V]{k, v}
	}
	close(c)

	// Do the work
	DoWorkChan(workers, c, func(i doWorkMapItem[K, V]) {
		f(i.key, i.value)
	})
}

type Errors struct {
	errs []error
	mu   sync.Mutex
}

func NewErrors() *Errors {
	return &Errors{}
}

func (e *Errors) Add(err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.errs = append(e.errs, err)
}

func (e *Errors) Err() error {
	if len(e.errs) == 0 {
		return nil
	}
	return errors.Join(e.errs...)
}
