package d4

import (
	"github.com/Dakota628/d4parse/pkg/bin"
	"io"
)

type (
	WalkNext func(d ...any)

	WalkCallback func(k string, v Object, next WalkNext, d ...any)

	Walkable interface {
		Walk(cb WalkCallback, d ...any)
	}
)

func (cb WalkCallback) Do(k string, t Object, d ...any) {
	cb(k, t, func(sd ...any) {
		if x, ok := t.(Walkable); ok {
			x.Walk(cb, sd...)
		}
	}, d...)
}

func UnmarshalAt(offset int64, t Object, r *bin.BinaryReader, o *FieldOptions) error {
	if _, err := r.Seek(offset, io.SeekStart); err != nil {
		return err
	}
	if err := t.UnmarshalD4(r, o); err != nil {
		return err
	}
	return nil
}
