package d4

import (
	"errors"
	"reflect"
	"unicode"
)

var (
	ErrInvalidOverrideType = errors.New("override type instance type does not implement generic type")
)

func newElem[T Object](t T) T {
	elemType := reflect.TypeOf(t).Elem()
	elemPtr := reflect.New(elemType)
	return elemPtr.Interface().(T)
}

func newElemWithOpts[T Object](t T, o *FieldOptions) (T, error) {
	if o != nil && o.OverrideTypeInstance != nil {
		if t, ok := newElem(o.OverrideTypeInstance).(T); ok {
			return t, nil
		}
		var zero T
		return zero, ErrInvalidOverrideType
	}
	return newElem(t), nil
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
