package d4

import (
	"encoding/binary"
	"github.com/Dakota628/d4parse/pkg/bin"
	"hash"
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

func nilObject[T Object]() (obj T) {
	return
}

func hashField(h hash.Hash, fieldHash uint32, obj Object) error {
	// Write field hash
	bs := make([]byte, fieldHash)
	binary.LittleEndian.PutUint32(bs, fieldHash)
	if _, err := h.Write(bs); err != nil {
		return err
	}

	// Write field value
	if err := obj.Hash(h); err != nil {
		return err
	}

	return nil
}
