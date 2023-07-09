package d4

import "reflect"

func newElem[T UnmarshalBinary](t T) T {
	elemType := reflect.TypeOf(t).Elem()
	elemPtr := reflect.New(elemType)
	return elemPtr.Interface().(T)
}

// TODO: function dynamically determine size of struct
