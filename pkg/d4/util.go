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

func newElemWithOpts[T Object](t T, o *Options) (T, error) {
	if o.OverrideTypeInstance != nil {
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

func NewCollectionBySubTypeInstance[T Object](h int, _ T) Object {
	return NewByTypeHash[T](h)
}

func NewCollectionBySubTypeHash(h int, sh int) Object {
	st := NewByTypeHash[Object](sh)
	return NewCollectionBySubTypeInstance(h, st)
}

func NewPolymorphicVariableArray[B Object](_ B) *DT_POLYMORPHIC_VARIABLEARRAY[B] {
	return &DT_POLYMORPHIC_VARIABLEARRAY[B]{}
}

func TypeNameByHash(h int) string {
	t := NewByTypeHash[Object](h)
	return reflect.TypeOf(t).Elem().Name()
}
