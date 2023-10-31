package d4

//go:generate go run ../../cmd/structgen/structgen.go ../../d4data/definitions.json generated_types.go

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/Dakota628/d4parse/pkg/bin"
	"golang.org/x/exp/slog"
	"hash"
	"io"
	"math"
	"strconv"
)

var (
	ErrInvalidPadding      = errors.New("invalid value in padding")
	ErrArrayLengthRequired = errors.New("array length option required")
	ErrGroupRequired       = errors.New("group option required")
)

type (
	TypeOptions struct {
		Flags           int
		Alignment       int
		TagMapAlignment int
	}

	FieldOptions struct {
		Flags                int
		ArrayLength          int
		Group                int32
		OverrideTypeInstance Object
		TagMapType           Object
	}

	Object interface {
		UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error
		TypeHash() int
		SubTypeHash() int
		Hash(h hash.Hash) error
	}

	MaybeExternal interface {
		IsExternal() bool
	}
)

func (o *FieldOptions) CopyForChild() *FieldOptions {
	if o == nil {
		return nil
	}
	return &FieldOptions{
		Group: o.Group,
	}
}

// DT_NULL ..
type DT_NULL struct{}

func (d *DT_NULL) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	return nil
}

func (d *DT_NULL) Hash(_ hash.Hash) error {
	return nil
}

// DT_BYTE ...
type DT_BYTE struct {
	Value uint8
}

func (d *DT_BYTE) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	return r.Uint8(&d.Value)
}

func (d *DT_BYTE) Hash(h hash.Hash) error {
	_, err := h.Write([]byte{d.Value})
	return err
}

// DT_WORD ...
type DT_WORD struct {
	Value uint16
}

func (d *DT_WORD) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	return r.Uint16LE(&d.Value)
}

func (d *DT_WORD) Hash(h hash.Hash) error {
	bs := make([]byte, 2)
	binary.LittleEndian.PutUint16(bs, d.Value)
	_, err := h.Write(bs)
	return err
}

// DT_ENUM ...
type DT_ENUM struct {
	Value int32
}

func (d *DT_ENUM) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	return r.Int32LE(&d.Value)
}

func (d *DT_ENUM) Hash(h hash.Hash) error {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(d.Value))
	_, err := h.Write(bs)
	return err
}

// DT_INT ...
type DT_INT struct {
	Value int32
}

func (d *DT_INT) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	return r.Int32LE(&d.Value)
}

func (d *DT_INT) Hash(h hash.Hash) error {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(d.Value))
	_, err := h.Write(bs)
	return err
}

// DT_FLOAT ...
type DT_FLOAT struct {
	Value float32
}

func (d *DT_FLOAT) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	return r.Float32LE(&d.Value)
}

func (d *DT_FLOAT) Hash(h hash.Hash) error {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, math.Float32bits(d.Value))
	_, err := h.Write(bs)
	return err
}

// DT_OPTIONAL ...
type DT_OPTIONAL[T Object] struct {
	Exists int32
	Value  T
}

func (d *DT_OPTIONAL[T]) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	if err := r.Int32LE(&d.Exists); err != nil {
		return err
	}

	if d.Exists > 0 {
		var err error
		if d.Value, err = newElemWithOpts(d.Value, o); err != nil {
			return err
		}

		return d.Value.UnmarshalD4(r, o.CopyForChild())
	}

	return nil
}

func (d *DT_OPTIONAL[T]) Walk(cb WalkCallback, data ...any) {
	if d.Exists > 0 {
		cb.Do("", d.Value, data...)
	}
}

func (d *DT_OPTIONAL[T]) Hash(h hash.Hash) error {
	if d.Exists > 0 {
		return d.Value.Hash(h)
	}
	return nil
}

// DT_SNO ...
type DT_SNO struct {
	Id int32
}

func (d *DT_SNO) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	return r.Int32LE(&d.Id)
}

func (d *DT_SNO) Hash(h hash.Hash) error {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(d.Id))
	_, err := h.Write(bs)
	return err
}

// DT_SNO_NAME ...
type DT_SNO_NAME struct {
	Group int32
	Id    int32
}

func (d *DT_SNO_NAME) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	if err := r.Int32LE(&d.Group); err != nil {
		return err
	}
	return r.Int32LE(&d.Id)
}

func (d *DT_SNO_NAME) Hash(h hash.Hash) error {
	bs := make([]byte, 4)

	binary.LittleEndian.PutUint32(bs, uint32(d.Group))
	if _, err := h.Write(bs); err != nil {
		return err
	}

	binary.LittleEndian.PutUint32(bs, uint32(d.Id))
	if _, err := h.Write(bs); err != nil {
		return err
	}

	return nil
}

// DT_GBID ...
type DT_GBID struct {
	Group int32
	Value uint32
}

func (d *DT_GBID) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	if o == nil {
		return ErrGroupRequired
	}

	d.Group = o.Group
	return r.Uint32LE(&d.Value)
}

func (d *DT_GBID) Hash(h hash.Hash) error {
	bs := make([]byte, 4)

	binary.LittleEndian.PutUint32(bs, uint32(d.Group))
	if _, err := h.Write(bs); err != nil {
		return err
	}

	binary.LittleEndian.PutUint32(bs, d.Value)
	if _, err := h.Write(bs); err != nil {
		return err
	}

	return nil
}

// DT_STARTLOC_NAME ...
type DT_STARTLOC_NAME struct {
	Value uint32
}

func (d *DT_STARTLOC_NAME) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	return r.Uint32LE(&d.Value)
}

func (d *DT_STARTLOC_NAME) Hash(h hash.Hash) error {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, d.Value)
	_, err := h.Write(bs)
	return err
}

// DT_UINT ...
type DT_UINT struct {
	Value uint32
}

func (d *DT_UINT) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	return r.Uint32LE(&d.Value)
}

func (d *DT_UINT) Hash(h hash.Hash) error {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, d.Value)
	_, err := h.Write(bs)
	return err
}

// DT_ACD_NETWORK_NAME ...
type DT_ACD_NETWORK_NAME struct {
	Value uint64
}

func (d *DT_ACD_NETWORK_NAME) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	return r.Uint64LE(&d.Value)
}

func (d *DT_ACD_NETWORK_NAME) Hash(h hash.Hash) error {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, d.Value)
	_, err := h.Write(bs)
	return err
}

// DT_SHARED_SERVER_DATA_ID ...
type DT_SHARED_SERVER_DATA_ID struct {
	Value uint64
}

func (d *DT_SHARED_SERVER_DATA_ID) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	return r.Uint64LE(&d.Value)
}

func (d *DT_SHARED_SERVER_DATA_ID) Hash(h hash.Hash) error {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, d.Value)
	_, err := h.Write(bs)
	return err
}

// DT_INT64 ...
type DT_INT64 struct {
	Value int64
}

func (d *DT_INT64) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	return r.Int64LE(&d.Value)
}

func (d *DT_INT64) Hash(h hash.Hash) error {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, uint64(d.Value))
	_, err := h.Write(bs)
	return err
}

// DT_RANGE ...
type DT_RANGE[T Object] struct {
	LowerBound T
	UpperBound T
}

func (d *DT_RANGE[T]) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) (err error) {
	if d.LowerBound, err = newElemWithOpts(d.LowerBound, o); err != nil {
		return
	}
	if err = d.LowerBound.UnmarshalD4(r, o.CopyForChild()); err != nil {
		return
	}
	if d.UpperBound, err = newElemWithOpts(d.UpperBound, o); err != nil {
		return
	}
	return d.UpperBound.UnmarshalD4(r, o.CopyForChild())
}

func (d *DT_RANGE[T]) Walk(cb WalkCallback, data ...any) {
	cb.Do("LowerBound", d.LowerBound, data...)
	cb.Do("UpperBound", d.UpperBound, data...)
}

func (d *DT_RANGE[T]) Hash(h hash.Hash) error {
	if err := d.LowerBound.Hash(h); err != nil {
		return err
	}
	return d.UpperBound.Hash(h)
}

// DT_FIXEDARRAY ...
type DT_FIXEDARRAY[T Object] struct {
	Value []T
}

func (d *DT_FIXEDARRAY[T]) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	if o == nil {
		return ErrArrayLengthRequired
	}

	d.Value = make([]T, o.ArrayLength)
	for i := 0; i < o.ArrayLength; i++ {
		var err error
		if d.Value[i], err = newElemWithOpts(d.Value[i], o); err != nil {
			return err
		}

		if err := d.Value[i].UnmarshalD4(r, o.CopyForChild()); err != nil {
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

func (d *DT_FIXEDARRAY[T]) Hash(h hash.Hash) error {
	for _, v := range d.Value {
		if err := v.Hash(h); err != nil {
			return err
		}
	}
	return nil
}

// DT_TAGMAP ...
type (
	TagMapEntry struct {
		Name      string
		FieldHash uint32
		Value     Object
	}

	DT_TAGMAP[T Object] struct {
		Padding1   int64
		DataOffset int32
		DataSize   int32

		DataCount int32
		Value     []TagMapEntry
		Type      Object
	}
)

func (d *DT_TAGMAP[T]) getTypeAlignment(fieldType Object, fieldSubType Object) int32 {
	switch fieldType.TypeHash() {
	case TypeHashByName("DT_FIXEDARRAY"),
		TypeHashByName("DT_OPTIONAL"),
		TypeHashByName("DT_RANGE"):
		subtypeOptions := OptionsForType(fieldSubType.TypeHash())
		return int32(subtypeOptions.TagMapAlignment)
	default:
		typeOptions := OptionsForType(fieldType.TypeHash())
		return int32(typeOptions.TagMapAlignment)
	}
}

func (d *DT_TAGMAP[T]) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	// Note: this is probably not fully correct. In order to support possibilities such as nested tag maps and fixed arr
	// in a tag map, we would need to look at the associated type and follow the flags and such for each field.

	if o != nil {
		d.Type = o.TagMapType
	}

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
		subTypeInstances := make([]Object, d.DataCount)

		for i := int32(0); i < d.DataCount; i++ {
			var elemFieldHash uint32
			if err := r.Uint32LE(&elemFieldHash); err != nil {
				return err
			}

			var elemTypeHash uint32
			if err := r.Uint32LE(&elemTypeHash); err != nil {
				return err
			}

			name := NameByFieldHash(int(elemFieldHash))
			value := NewByTypeHash[Object](int(elemTypeHash))
			if value == nil {
				return fmt.Errorf("could not find type for type hash: %d", elemTypeHash)
			}

			// Type flag 0x8000
			elemTypeOptions := OptionsForType(value.TypeHash())
			if DefFlagHasSubType.In(elemTypeOptions.Flags) {
				var elemSubTypeHash uint32
				if err := r.Uint32LE(&elemSubTypeHash); err != nil {
					return err
				}
				subTypeInstances[i] = NewByTypeHash[Object](int(elemSubTypeHash))
			}

			d.Value[i].Name = name
			d.Value[i].FieldHash = elemFieldHash // TODO: maybe get rid of this and just hash the string when needed
			d.Value[i].Value = value
		}

		for i := int32(0); i < d.DataCount; i++ {
			currentOffset, err := r.Pos()
			if err != nil {
				return err
			}

			requiredAlignment := d.getTypeAlignment(d.Value[i].Value, subTypeInstances[i])
			currentAlignment := int32(currentOffset) % requiredAlignment
			if currentAlignment > 0 {
				padding := requiredAlignment - currentAlignment
				if _, err = r.Seek(int64(padding), io.SeekCurrent); err != nil {
					return err
				}
			}

			currOpts := o.CopyForChild()
			if subTypeInstances[i] != nil {
				currOpts.OverrideTypeInstance = subTypeInstances[i]
			}

			if err := d.Value[i].Value.UnmarshalD4(r, currOpts); err != nil {
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

func (d *DT_TAGMAP[T]) Hash(h hash.Hash) error {
	// Note: values should already be sorted by hash; that's how they're stored
	for _, entry := range d.Value {
		if err := hashField(h, entry.FieldHash, entry.Value); err != nil {
			return err
		}
	}
	return nil
}

// DT_VARIABLEARRAY ...
type DT_VARIABLEARRAY[T Object] struct {
	Padding1   int64
	DataOffset int32
	DataSize   int32

	external bool
	Value    []T
}

func (d *DT_VARIABLEARRAY[T]) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	d.external = o != nil && (FieldFlagPayload.In(o.Flags) || FieldFlagPayload2.In(o.Flags))

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
		for curr := int64(d.DataOffset); curr < int64(d.DataOffset+d.DataSize); {
			var err error
			var elem T

			elem, err = newElemWithOpts(elem, o)
			if err != nil {
				return err
			}

			if err = elem.UnmarshalD4(r, o.CopyForChild()); err != nil {
				return err
			}
			d.Value = append(d.Value, elem)

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

func (d *DT_VARIABLEARRAY[T]) Hash(h hash.Hash) error {
	for _, v := range d.Value {
		if err := v.Hash(h); err != nil {
			return err
		}
	}
	return nil
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

func (d *DT_POLYMORPHIC_VARIABLEARRAY[T]) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	d.external = o != nil && (FieldFlagPayload.In(o.Flags) || FieldFlagPayload2.In(o.Flags))

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
				return base.UnmarshalD4(r, o.CopyForChild())
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
			d.Value[i] = NewByTypeHash[*DT_NULL](elemTypeHash)
			if d.Value[i] == nil {
				return fmt.Errorf("could not find type for type hash: %d", elemTypeHash)
			}

			if err := d.Value[i].UnmarshalD4(r, o.CopyForChild()); err != nil {
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

func (d *DT_POLYMORPHIC_VARIABLEARRAY[T]) Hash(h hash.Hash) error {
	for _, v := range d.Value {
		if err := v.Hash(h); err != nil {
			return err
		}
	}
	return nil
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

func (d *DT_STRING_FORMULA) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
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

func (d *DT_STRING_FORMULA) Hash(h hash.Hash) error {
	if _, err := h.Write([]byte(d.Value)); err != nil {
		return err
	}
	_, err := h.Write([]byte(d.Compiled))
	return err
}

// DT_CSTRING ...
type DT_CSTRING[Unused Object] struct {
	Offset int32
	Size   int32

	Value string
}

func (d *DT_CSTRING[Unused]) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
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

func (d *DT_CSTRING[Unused]) Hash(h hash.Hash) error {
	_, err := h.Write([]byte(d.Value))
	return err
}

func (d *DT_CSTRING[Unused]) String() string {
	return d.Value
}

// DT_CHARARRAY ...
type DT_CHARARRAY struct {
	Value []byte
}

func (d *DT_CHARARRAY) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	if o == nil {
		return ErrArrayLengthRequired
	}

	d.Value = make([]byte, o.ArrayLength)
	if _, err := r.Read(d.Value); err != nil {
		return nil
	}
	return nil
}

func (d *DT_CHARARRAY) Hash(h hash.Hash) error {
	_, err := h.Write(d.Value)
	return err
}

// DT_RGBACOLOR ...
type DT_RGBACOLOR struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

func (d *DT_RGBACOLOR) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	if err := r.Uint8(&d.R); err != nil {
		return err
	}

	if err := r.Uint8(&d.G); err != nil {
		return err
	}

	if err := r.Uint8(&d.B); err != nil {
		return err
	}

	return r.Uint8(&d.A)
}

func (d *DT_RGBACOLOR) Hash(h hash.Hash) error {
	_, err := h.Write([]byte{
		d.R,
		d.G,
		d.B,
		d.A,
	})
	return err
}

// DT_RGBACOLORVALUE ...
type DT_RGBACOLORVALUE struct {
	R float32
	G float32
	B float32
	A float32
}

func (d *DT_RGBACOLORVALUE) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
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

func (d *DT_RGBACOLORVALUE) Hash(h hash.Hash) error {
	bs := make([]byte, 16)

	binary.LittleEndian.PutUint32(bs, math.Float32bits(d.R))
	binary.LittleEndian.PutUint32(bs[4:], math.Float32bits(d.G))
	binary.LittleEndian.PutUint32(bs[8:], math.Float32bits(d.B))
	binary.LittleEndian.PutUint32(bs[12:], math.Float32bits(d.A))

	_, err := h.Write(bs)
	return err
}

// DT_BCVEC2I ...
type DT_BCVEC2I struct {
	X float32
	Y float32
}

func (d *DT_BCVEC2I) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	if err := r.Float32LE(&d.X); err != nil {
		return err
	}

	return r.Float32LE(&d.Y)
}

func (d *DT_BCVEC2I) Hash(h hash.Hash) error {
	bs := make([]byte, 8)

	binary.LittleEndian.PutUint32(bs, math.Float32bits(d.X))
	binary.LittleEndian.PutUint32(bs[4:], math.Float32bits(d.Y))

	_, err := h.Write(bs)
	return err
}

// DT_VECTOR2D ...
type DT_VECTOR2D struct {
	X float32
	Y float32
}

func (d *DT_VECTOR2D) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	if err := r.Float32LE(&d.X); err != nil {
		return err
	}

	return r.Float32LE(&d.Y)
}

func (d *DT_VECTOR2D) Hash(h hash.Hash) error {
	bs := make([]byte, 8)

	binary.LittleEndian.PutUint32(bs, math.Float32bits(d.X))
	binary.LittleEndian.PutUint32(bs[4:], math.Float32bits(d.Y))

	_, err := h.Write(bs)
	return err
}

// DT_VECTOR3D ...
type DT_VECTOR3D struct {
	X float32
	Y float32
	Z float32
}

func (d *DT_VECTOR3D) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
	if err := r.Float32LE(&d.X); err != nil {
		return err
	}

	if err := r.Float32LE(&d.Y); err != nil {
		return err
	}

	return r.Float32LE(&d.Z)
}

func (d *DT_VECTOR3D) Hash(h hash.Hash) error {
	bs := make([]byte, 12)

	binary.LittleEndian.PutUint32(bs, math.Float32bits(d.X))
	binary.LittleEndian.PutUint32(bs[4:], math.Float32bits(d.Y))
	binary.LittleEndian.PutUint32(bs[8:], math.Float32bits(d.Z))

	_, err := h.Write(bs)
	return err
}

// DT_VECTOR4D ...
type DT_VECTOR4D struct {
	X float32
	Y float32
	Z float32
	W float32
}

func (d *DT_VECTOR4D) UnmarshalD4(r *bin.BinaryReader, o *FieldOptions) error {
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

func (d *DT_VECTOR4D) Hash(h hash.Hash) error {
	bs := make([]byte, 16)

	binary.LittleEndian.PutUint32(bs, math.Float32bits(d.X))
	binary.LittleEndian.PutUint32(bs[4:], math.Float32bits(d.Y))
	binary.LittleEndian.PutUint32(bs[8:], math.Float32bits(d.Z))
	binary.LittleEndian.PutUint32(bs[12:], math.Float32bits(d.W))

	_, err := h.Write(bs)
	return err
}
