package d4

//go:generate go run ../../cmd/structgen/structgen.go ../../d4data/definitions.json generated_types.go

import (
	"errors"
	"fmt"
	"github.com/Dakota628/d4parse/pkg/bin"
	"golang.org/x/exp/slog"
	"io"
	"strconv"
)

var (
	ErrInvalidPadding      = errors.New("invalid value in padding")
	ErrArrayLengthRequired = errors.New("array length option required")
	ErrGroupRequired       = errors.New("group option required")
)

// TODO: implement Walk for iterable types

type (
	Options struct {
		Flags       int
		ArrayLength int
		Group       int32
	}

	Object interface {
		UnmarshalD4(r *bin.BinaryReader, o *Options) error
	}

	MaybeExternal interface {
		IsExternal() bool
	}
)

// DT_NULL ..
type DT_NULL struct{}

func (d *DT_NULL) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	return nil
}

// DT_BYTE ...
type DT_BYTE struct {
	Value uint8
}

func (d *DT_BYTE) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	return r.Uint8(&d.Value)
}

// DT_WORD ...
type DT_WORD struct {
	Value uint16
}

func (d *DT_WORD) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	return r.Uint16LE(&d.Value)
}

// DT_ENUM ...
type DT_ENUM struct {
	Value int32
}

func (d *DT_ENUM) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	return r.Int32LE(&d.Value)
}

// DT_INT ...
type DT_INT struct {
	Value int32
}

func (d *DT_INT) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	return r.Int32LE(&d.Value)
}

// DT_FLOAT ...
type DT_FLOAT struct {
	Value float32
}

func (d *DT_FLOAT) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	return r.Float32LE(&d.Value)
}

// DT_OPTIONAL ...
type DT_OPTIONAL[T Object] struct {
	Exists int32
	Value  T
}

func (d *DT_OPTIONAL[T]) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	if err := r.Int32LE(&d.Exists); err != nil {
		return err
	}

	if d.Exists > 0 {
		return d.Value.UnmarshalD4(r, o)
	}

	return nil
}

func (d *DT_OPTIONAL[T]) Walk(cb WalkCallback, data ...any) {
	if d.Exists > 0 {
		cb.Do("", d.Value, data...)
	}
}

// DT_SNO ...
type DT_SNO struct {
	Id int32
}

func (d *DT_SNO) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	return r.Int32LE(&d.Id)
}

// DT_SNO_NAME ...
type DT_SNO_NAME struct {
	Group int32
	Id    int32
}

func (d *DT_SNO_NAME) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	if err := r.Int32LE(&d.Group); err != nil {
		return err
	}
	return r.Int32LE(&d.Id)
}

// DT_GBID ...
type DT_GBID struct {
	Group int32
	Value uint32
}

func (d *DT_GBID) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	if o == nil {
		return ErrGroupRequired
	}

	d.Group = o.Group
	return r.Uint32LE(&d.Value)
}

// DT_STARTLOC_NAME ...
type DT_STARTLOC_NAME struct {
	Value uint32
}

func (d *DT_STARTLOC_NAME) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	return r.Uint32LE(&d.Value)
}

// DT_UINT ...
type DT_UINT struct {
	Value uint32
}

func (d *DT_UINT) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	return r.Uint32LE(&d.Value)
}

// DT_ACD_NETWORK_NAME ...
type DT_ACD_NETWORK_NAME struct {
	Value uint64
}

func (d *DT_ACD_NETWORK_NAME) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	return r.Uint64LE(&d.Value)
}

// DT_SHARED_SERVER_DATA_ID ...
type DT_SHARED_SERVER_DATA_ID struct {
	Value uint64
}

func (d *DT_SHARED_SERVER_DATA_ID) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	return r.Uint64LE(&d.Value)
}

// DT_INT64 ...
type DT_INT64 struct {
	Value int64
}

func (d *DT_INT64) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	return r.Int64LE(&d.Value)
}

// DT_RANGE ...
type DT_RANGE[T Object] struct {
	LowerBound T
	UpperBound T
}

func (d *DT_RANGE[T]) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	d.LowerBound = newElem(d.LowerBound)
	if err := d.LowerBound.UnmarshalD4(r, o); err != nil {
		return err
	}
	d.UpperBound = newElem(d.LowerBound)
	return d.UpperBound.UnmarshalD4(r, o)
}

func (d *DT_RANGE[T]) Walk(cb WalkCallback, data ...any) {
	cb.Do("LowerBound", d.LowerBound, data...)
	cb.Do("UpperBound", d.UpperBound, data...)
}

// DT_FIXEDARRAY ...
type DT_FIXEDARRAY[T Object] struct {
	Value []T
}

func (d *DT_FIXEDARRAY[T]) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	if o == nil {
		return ErrArrayLengthRequired
	}

	d.Value = make([]T, o.ArrayLength)
	for i := 0; i < o.ArrayLength; i++ {
		d.Value[i] = newElem(d.Value[i])
		if err := d.Value[i].UnmarshalD4(r, o); err != nil {
			return err
		}
	}
	return nil
}

func (d *DT_FIXEDARRAY[T]) Walk(cb WalkCallback, data ...any) {
	for i, v := range d.Value {
		cb.Do(strconv.Itoa(i), v, data...)
	}
}

// DT_TAGMAP ...
type TagMapEntry struct {
	Name string
	Value Object
}
type DT_TAGMAP[T Object] struct {
	Padding1   int64
	DataOffset int32
	DataSize   int32

	DataCount int32
	Value     []TagMapEntry
}

func (d *DT_TAGMAP[T]) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	if err := r.Int64LE(&d.Padding1); err != nil {
		return err
	}

	if err := r.Int32LE(&d.DataOffset); err != nil {
		return err
	}

	if err := r.Int32LE(&d.DataSize); err != nil {
		return err
	}

	if d.Padding1 != 0 {
		return ErrInvalidPadding
	}

	if d.DataOffset < 1 || d.DataSize < 1 {
		return nil
	}

	return r.AtPos(int64(d.DataOffset), io.SeekStart, func(r *bin.BinaryReader) error {
		if err := r.Int32LE(&d.DataCount); err != nil {
			return err
		}
		d.Value = make([]TagMapEntry, d.DataCount)

		for i := int32(0); i < d.DataCount; i++ {
			var elemFieldHash uint32
			var elemTypeHash uint32
			var elemSubTypeHash uint32
			var elemSubType Object
			elemSubType = &DT_NULL{}
			if err := r.Uint32LE(&elemFieldHash); err != nil {
				return err
			}
			if err := r.Uint32LE(&elemTypeHash); err != nil {
				return err
			}

			// Type flag 0x8000
			if elemTypeHash == 1683664497 || // DT_POLYMORPHIC_VARIABLEARRAY
			   elemTypeHash == 2388214534 || // DT_FIXEDARRAY
			   elemTypeHash == 3121633597 || // DT_OPTIONAL
			   elemTypeHash == 3244749660 || // DT_VARIABLEARRAY
			   elemTypeHash == 3493213809 || // DT_TAGMAP
			   elemTypeHash == 3846829457 || // DT_CSTRING
			   elemTypeHash == 3877855748 {  // DT_RANGE
				if err := r.Uint32LE(&elemSubTypeHash); err != nil {
					return err
				}
				elemSubType = NewByTypeHash(int(elemSubTypeHash), &DT_NULL{})
			}

			d.Value[i].Name = NameByFieldHash(int(elemFieldHash))
			d.Value[i].Value = NewByTypeHash(int(elemTypeHash), elemSubType)
			if d.Value[i].Value == nil {
				return fmt.Errorf("could not find type for type hash: %d", elemTypeHash)
			}
		}

		for i := int32(0); i < d.DataCount; i++ {
			if err := d.Value[i].Value.UnmarshalD4(r, o); err != nil {
				return err
			}
		}

		return nil
	})
}

func (d *DT_TAGMAP[T]) Walk(cb WalkCallback, data ...any) {
	for _, v := range d.Value {
		cb.Do(v.Name, v.Value, data...)
	}
}

// DT_VARIABLEARRAY ...
type DT_VARIABLEARRAY[T Object] struct {
	Padding1   int64
	DataOffset int32
	DataSize   int32

	external bool
	Value    []T
}

func (d *DT_VARIABLEARRAY[T]) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	d.external = o != nil && ((o.Flags&0x200000) > 0 || (o.Flags&0x400000) > 0)

	if err := r.Int64LE(&d.Padding1); err != nil {
		return err
	}

	if err := r.Int32LE(&d.DataOffset); err != nil {
		return err
	}

	if err := r.Int32LE(&d.DataSize); err != nil {
		return err
	}

	if d.Padding1 != 0 {
		return ErrInvalidPadding
	}

	if d.external {
		// There's probably a way to get the external data from a payload file, but we don't know how yet.
		return nil
	}

	if d.DataOffset < 1 || d.DataSize < 1 {
		return nil
	}

	return r.AtPos(int64(d.DataOffset), io.SeekStart, func(r *bin.BinaryReader) error {
		//for (curr - int64(d.DataOffset)) < int64(d.DataSize) {
		for curr := int64(d.DataOffset); curr < int64(d.DataOffset+d.DataSize); {
			var elem T
			elem = newElem(elem)
			if err := elem.UnmarshalD4(r, o); err != nil {
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

func (d *DT_VARIABLEARRAY[T]) Walk(cb WalkCallback, data ...any) {
	for i, v := range d.Value {
		cb.Do(strconv.Itoa(i), v, data...)
	}
}

func (d *DT_VARIABLEARRAY[T]) IsExternal() bool {
	return d.external
}

// DT_POLYMORPHIC_VARIABLEARRAY ...
type DT_POLYMORPHIC_VARIABLEARRAY[T Object] struct {
	Padding1   int64
	DataOffset int32
	DataSize   int32
	DataCount  int32
	Padding2   int32

	external bool
	Value    []Object
}

func (d *DT_POLYMORPHIC_VARIABLEARRAY[T]) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	d.external = o != nil && ((o.Flags&0x200000) > 0 || (o.Flags&0x400000) > 0)

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

	if d.external {
		// There's probably a way to get the external data from a payload file, but we don't know how yet.
		return nil
	}

	if d.DataOffset < 1 || d.DataSize < 1 || d.DataCount < 1 {
		return nil
	}

	// Skip 8 bytes per entry at the start. We're not sure what this is.
	nSkipped := int64(d.DataCount) * 8

	// Read the data
	d.Value = make([]Object, d.DataCount)
	return r.AtPos(int64(d.DataOffset)+nSkipped, io.SeekStart, func(r *bin.BinaryReader) error {
		for i := int32(0); i < d.DataCount; i++ {
			// Note: technically we should read T (T is the actual polymorphic base), but as far as we can tell
			// polymorphic base is the basis of every non-basic type. BuffCallbackBase does use DT_INT64 as it's base,
			// however, we currently don't understand why.

			// Read polymorphic base to get type info before reading real type
			var base PolymorphicBase
			if err := r.AtPos(0, io.SeekCurrent, func(r *bin.BinaryReader) error {
				return base.UnmarshalD4(r, o)
			}); err != nil {
				// TODO: this is definitely not right, remove once GameBalanceTable issue solved
				if err == io.EOF {
					slog.Warn(
						"Allowing invalid polymorphic array",
						slog.Any("error", err),
						slog.String("type", fmt.Sprintf("%T", d)),
					)
					return nil
				}

				return err
			}

			// Read real type based on base
			elemTypeHash := int(base.DwType.Value)

			// Use DT_NULL as subtype for now as we don't know if it's possible to nest a third type atm
			d.Value[i] = NewByTypeHash(elemTypeHash, &DT_NULL{})
			if d.Value[i] == nil {
				return fmt.Errorf("could not find type for type hash: %d", elemTypeHash)
			}

			if err := d.Value[i].UnmarshalD4(r, o); err != nil {
				return err
			}
		}

		return nil
	})
}

func (d *DT_POLYMORPHIC_VARIABLEARRAY[T]) Walk(cb WalkCallback, data ...any) {
	for i, v := range d.Value {
		cb.Do(strconv.Itoa(i), v, data...)
	}
}

func (d *DT_POLYMORPHIC_VARIABLEARRAY[T]) IsExternal() bool {
	return d.external
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

func (d *DT_STRING_FORMULA) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
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
type DT_CSTRING[Unused Object] struct {
	Offset int32
	Size   int32

	Value string
}

func (d *DT_CSTRING[Unused]) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	// Skip 8 bytes
	if _, err := r.Seek(8, io.SeekCurrent); err != nil {
		return err
	}

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

func (d *DT_CSTRING[Unused]) String() string {
	return d.Value
}

// DT_CHARARRAY ...
type DT_CHARARRAY struct {
	Value []rune
}

func (d *DT_CHARARRAY) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	if o == nil {
		return ErrArrayLengthRequired
	}

	buf := make([]byte, o.ArrayLength)
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

func (d *DT_RGBACOLOR) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
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

func (d *DT_RGBACOLORVALUE) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
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

func (d *DT_BCVEC2I) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
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

func (d *DT_VECTOR2D) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
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

func (d *DT_VECTOR3D) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
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

func (d *DT_VECTOR4D) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
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
