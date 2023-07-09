package d4

//go:generate go run ../../cmd/structgen/structgen.go ../../d4data/definitions.json generated_types.go

import (
	"errors"
	"fmt"
	"github.com/Dakota628/d4parse/pkg/bin"
	"io"
)

var (
	ErrInvalidPadding = errors.New("invalid value in padding")
)

type UnmarshalBinary interface { // TODO: rename this
	UnmarshalBinary(r *bin.BinaryReader) error
}

// DT_NULL ..
type DT_NULL struct{}

func (d *DT_NULL) UnmarshalBinary(r *bin.BinaryReader) error {
	return nil
}

// DT_BYTE ...
type DT_BYTE struct {
	Value uint8
}

func (d *DT_BYTE) UnmarshalBinary(r *bin.BinaryReader) error {
	return r.Uint8(&d.Value)
}

// DT_WORD ...
type DT_WORD struct {
	Value uint16
}

func (d *DT_WORD) UnmarshalBinary(r *bin.BinaryReader) error {
	return r.Uint16LE(&d.Value)
}

// DT_ENUM ...
type DT_ENUM struct {
	Value int32
}

func (d *DT_ENUM) UnmarshalBinary(r *bin.BinaryReader) error {
	return r.Int32LE(&d.Value)
}

// DT_INT ...
type DT_INT struct {
	Value int32
}

func (d *DT_INT) UnmarshalBinary(r *bin.BinaryReader) error {
	return r.Int32LE(&d.Value)
}

// DT_FLOAT ...
type DT_FLOAT struct {
	Value float32
}

func (d *DT_FLOAT) UnmarshalBinary(r *bin.BinaryReader) error {
	return r.Float32LE(&d.Value)
}

// DT_OPTIONAL ...
type DT_OPTIONAL[T UnmarshalBinary] struct {
	Exists int32
	Value  T
}

func (d *DT_OPTIONAL[T]) UnmarshalBinary(r *bin.BinaryReader) error {
	if err := r.Int32LE(&d.Exists); err != nil {
		return err
	}

	if d.Exists > 0 {
		return d.Value.UnmarshalBinary(r)
	}

	return nil
}

// DT_SNO ...
type DT_SNO struct {
	Id int32
}

func (d *DT_SNO) UnmarshalBinary(r *bin.BinaryReader) error {
	return r.Int32LE(&d.Id)
}

// DT_SNO_NAME ...
type DT_SNO_NAME struct {
	Group int32
	Id    int32
}

func (d *DT_SNO_NAME) UnmarshalBinary(r *bin.BinaryReader) error {
	if err := r.Int32LE(&d.Group); err != nil {
		return err
	}
	return r.Int32LE(&d.Id)
}

// DT_GBID ...
type DT_GBID struct {
	Value uint32
}

func (d *DT_GBID) UnmarshalBinary(r *bin.BinaryReader) error {
	return r.Uint32LE(&d.Value)
}

// DT_STARTLOC_NAME ...
type DT_STARTLOC_NAME struct {
	Value uint32
}

func (d *DT_STARTLOC_NAME) UnmarshalBinary(r *bin.BinaryReader) error {
	return r.Uint32LE(&d.Value)
}

// DT_UINT ...
type DT_UINT struct {
	Value uint32
}

func (d *DT_UINT) UnmarshalBinary(r *bin.BinaryReader) error {
	return r.Uint32LE(&d.Value)
}

// DT_ACD_NETWORK_NAME ...
type DT_ACD_NETWORK_NAME struct {
	Value uint64
}

func (d *DT_ACD_NETWORK_NAME) UnmarshalBinary(r *bin.BinaryReader) error {
	return r.Uint64LE(&d.Value)
}

// DT_SHARED_SERVER_DATA_ID ...
type DT_SHARED_SERVER_DATA_ID struct {
	Value uint64
}

func (d *DT_SHARED_SERVER_DATA_ID) UnmarshalBinary(r *bin.BinaryReader) error {
	return r.Uint64LE(&d.Value)
}

// DT_INT64 ...
type DT_INT64 struct {
	Value int64
}

func (d *DT_INT64) UnmarshalBinary(r *bin.BinaryReader) error {
	return r.Int64LE(&d.Value)
}

// DT_RANGE ...
type DT_RANGE[T UnmarshalBinary] struct {
	LowerBound T
	UpperBound T
}

func (d *DT_RANGE[T]) UnmarshalBinary(r *bin.BinaryReader) error {
	if err := d.LowerBound.UnmarshalBinary(r); err != nil {
		return err
	}
	return d.UpperBound.UnmarshalBinary(r)
}

// DT_FIXEDARRAY ...
type DT_FIXEDARRAY[T UnmarshalBinary] struct {
	Length uint32
	Value  []T
}

func (d *DT_FIXEDARRAY[T]) UnmarshalBinary(r *bin.BinaryReader) error {
	d.Value = make([]T, d.Length)
	for i := uint32(0); i < d.Length; i++ {
		d.Value[i] = newElem(d.Value[i])
		if err := d.Value[i].UnmarshalBinary(r); err != nil {
			return err
		}
	}
	return nil
}

// DT_TAGMAP ...
type DT_TAGMAP[T UnmarshalBinary] struct {
	Padding1   int64
	DataOffset int32
	DataSize   int32

	// TODO: figure out how to implement this fully
}

func (d *DT_TAGMAP[T]) UnmarshalBinary(r *bin.BinaryReader) error {
	if err := r.Int64LE(&d.Padding1); err != nil {
		return err
	}

	if err := r.Int32LE(&d.DataOffset); err != nil {
		return err
	}

	return r.Int32LE(&d.DataSize)
}

// DT_VARIABLEARRAY ...
type DT_VARIABLEARRAY[T UnmarshalBinary] struct {
	Padding1   int64
	DataOffset int32
	DataSize   int32

	Value []T
}

func (d *DT_VARIABLEARRAY[T]) UnmarshalBinary(r *bin.BinaryReader) error {
	if err := r.Int64LE(&d.Padding1); err != nil {
		return err
	}

	if err := r.Int32LE(&d.DataOffset); err != nil {
		return err
	}

	if err := r.Int32LE(&d.DataSize); err != nil {
		return err
	}

	return r.AtPos(int64(d.DataOffset), io.SeekStart, func(r *bin.BinaryReader) error {
		curr := int64(d.DataOffset)

		for int64(d.DataOffset+d.DataSize) > curr {
			var elem T
			elem = newElem(elem)
			if err := elem.UnmarshalBinary(r); err != nil {
				return err
			}
			d.Value = append(d.Value, elem)

			var err error
			if curr, err = r.Pos(); err != nil {
				return err
			}
		}

		return nil
	})
}

// DT_POLYMORPHIC_VARIABLEARRAY ...
type DT_POLYMORPHIC_VARIABLEARRAY[T UnmarshalBinary] struct {
	Padding1   int64
	DataOffset int32
	DataSize   int32
	DataCount  int32
	Padding2   int32

	Value []UnmarshalBinary
}

func (d *DT_POLYMORPHIC_VARIABLEARRAY[T]) UnmarshalBinary(r *bin.BinaryReader) error {
	if err := r.Int64LE(&d.Padding1); err != nil {
		return err
	}

	if err := r.Int32LE(&d.DataOffset); err != nil {
		return err
	}

	if err := r.Int32LE(&d.DataSize); err != nil {
		return err
	}

	if err := r.Int32LE(&d.DataCount); err != nil {
		return err
	}

	if err := r.Int32LE(&d.Padding2); err != nil {
		return err
	}

	if d.Padding1 != 0 || d.Padding2 != 0 {
		return ErrInvalidPadding
	}

	// Skip 8 bytes per entry at the start. We're not sure what this is.
	nSkipped := int64(d.DataCount) * 8

	// Read the data
	d.Value = make([]UnmarshalBinary, d.DataCount)
	return r.AtPos(int64(d.DataOffset)+nSkipped, io.SeekStart, func(r *bin.BinaryReader) error {
		for i := int32(0); i < d.DataCount; i++ {
			// Note: technically we should read T, but polymorphic base should cover every case

			// Read polymorphic base to get type info before reading real type
			var base PolymorphicBase
			if err := r.AtPos(0, io.SeekCurrent, base.UnmarshalBinary); err != nil {
				return err
			}

			// Read real type based on base
			elemTypeHash := int(base.DwType.Value)

			// Use DT_NULL as subtype for now as we don't know if it's possible to nest a third type atm
			d.Value[i] = NewByTypeHash[*DT_NULL](elemTypeHash)
			if d.Value[i] == nil {
				return fmt.Errorf("could not find type for type hash: %d", elemTypeHash)
			}

			if err := d.Value[i].UnmarshalBinary(r); err != nil {
				return err
			}
		}

		return nil
	})
}

// DT_STRING_FORMULA ...
type DT_STRING_FORMULA struct {
	FormulaOffset  int32
	FormulaSize    int32
	CompiledOffset int32
	CompiledSize   int32

	Value    string
	Compiled string
}

func (d *DT_STRING_FORMULA) UnmarshalBinary(r *bin.BinaryReader) error {
	// Skip 8 bytes: https://github.com/blizzhackers/d4data/blob/master/parse.js#L548
	if _, err := r.Seek(8, io.SeekCurrent); err != nil {
		return err
	}

	if err := r.Int32LE(&d.FormulaOffset); err != nil {
		return err
	}

	if err := r.Int32LE(&d.FormulaSize); err != nil {
		return err
	}

	if err := r.Int32LE(&d.CompiledOffset); err != nil {
		return err
	}

	if err := r.Int32LE(&d.CompiledSize); err != nil {
		return err
	}

	if err := r.AtPos(int64(d.FormulaOffset), io.SeekStart, func(r *bin.BinaryReader) error {
		buf := make([]byte, d.FormulaSize)
		if _, err := r.Read(buf); err != nil {
			return err
		}
		d.Value = string(buf)
		return nil
	}); err != nil {
		return err
	}

	if err := r.AtPos(int64(d.CompiledOffset), io.SeekStart, func(r *bin.BinaryReader) error {
		buf := make([]byte, d.CompiledSize)
		if _, err := r.Read(buf); err != nil {
			return err
		}
		d.Compiled = string(buf)
		return nil
	}); err != nil {
		return err
	}

	return nil
}

// DT_CSTRING ...
type DT_CSTRING[Unused UnmarshalBinary] struct {
	Offset int32
	Size   int32

	Value string
}

func (d *DT_CSTRING[Unused]) UnmarshalBinary(r *bin.BinaryReader) error {
	if err := r.Int32LE(&d.Offset); err != nil {
		return err
	}

	if err := r.Int32LE(&d.Size); err != nil {
		return err
	}

	return r.AtPos(int64(d.Offset), io.SeekStart, func(r *bin.BinaryReader) error {
		buf := make([]byte, d.Size)
		if _, err := r.Read(buf); err != nil {
			return err
		}
		d.Value = string(buf)
		return nil
	})
}

// DT_CHARARRAY ...
type DT_CHARARRAY struct {
	Length uint32
	Value  []rune
}

func (d *DT_CHARARRAY) UnmarshalBinary(r *bin.BinaryReader) error {
	buf := make([]byte, d.Length)
	if _, err := r.Read(buf); err != nil {
		return nil
	}
	d.Value = []rune(string(buf))
	return nil
}

// DT_RGBACOLOR ...
type DT_RGBACOLOR struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

func (d *DT_RGBACOLOR) UnmarshalBinary(r *bin.BinaryReader) error {
	if err := r.Uint8(&d.R); err != nil {
		return err
	}

	if err := r.Uint8(&d.G); err != nil {
		return err
		return err
	}

	if err := r.Uint8(&d.B); err != nil {
		return err
	}

	return r.Uint8(&d.A)
}

// DT_RGBACOLORVALUE ...
type DT_RGBACOLORVALUE struct {
	R float32
	G float32
	B float32
	A float32
}

func (d *DT_RGBACOLORVALUE) UnmarshalBinary(r *bin.BinaryReader) error {
	if err := r.Float32LE(&d.R); err != nil {
		return err
	}

	if err := r.Float32LE(&d.G); err != nil {
		return err
	}

	if err := r.Float32LE(&d.B); err != nil {
		return err
	}

	return r.Float32LE(&d.A)
}

// DT_BCVEC2I ...
type DT_BCVEC2I struct {
	X float32
	Y float32
}

func (d *DT_BCVEC2I) UnmarshalBinary(r *bin.BinaryReader) error {
	if err := r.Float32LE(&d.X); err != nil {
		return err
	}

	return r.Float32LE(&d.Y)
}

// DT_VECTOR2D ...
type DT_VECTOR2D struct {
	X float32
	Y float32
}

func (d *DT_VECTOR2D) UnmarshalBinary(r *bin.BinaryReader) error {
	if err := r.Float32LE(&d.X); err != nil {
		return err
	}

	return r.Float32LE(&d.Y)
}

// DT_VECTOR3D ...
type DT_VECTOR3D struct {
	X float32
	Y float32
	Z float32
}

func (d *DT_VECTOR3D) UnmarshalBinary(r *bin.BinaryReader) error {
	if err := r.Float32LE(&d.X); err != nil {
		return err
	}

	if err := r.Float32LE(&d.Y); err != nil {
		return err
	}

	return r.Float32LE(&d.Z)
}

// DT_VECTOR4D ...
type DT_VECTOR4D struct {
	X float32
	Y float32
	Z float32
	W float32
}

func (d *DT_VECTOR4D) UnmarshalBinary(r *bin.BinaryReader) error {
	if err := r.Float32LE(&d.X); err != nil {
		return err
	}

	if err := r.Float32LE(&d.Y); err != nil {
		return err
	}

	if err := r.Float32LE(&d.Z); err != nil {
		return err
	}

	return r.Float32LE(&d.W)
}
