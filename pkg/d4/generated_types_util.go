package d4

import (
	"github.com/Dakota628/d4parse/pkg/bin"
	"io"
)

type (
	WalkNext func()

	WalkCallback func(k string, v Object, next WalkNext)

	Walkable interface {
		Walk(cb WalkCallback)
	}
)

func (cb WalkCallback) Do(k string, t Object) {
	cb(k, t, func() {
		if x, ok := t.(Walkable); ok {
			x.Walk(cb)
		}
	})
}

func UnmarshalAt(offset int64, t Object, r *bin.BinaryReader, o *Options) error {
	if _, err := r.Seek(offset, io.SeekStart); err != nil {
		return err
	}
	if err := t.UnmarshalD4(r, o); err != nil {
		return err
	}
	return nil
}
