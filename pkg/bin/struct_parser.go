package bin

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// Endianness ...
type Endianness bool

const (
	BE Endianness = false
	LE Endianness = true
)

// LEStruct ...
type LEStruct struct{}

func (s LEStruct) isStruct() bool {
	return true
}

func (s LEStruct) endianness() Endianness {
	return LE
}

// BEStruct ...
type BEStruct struct{}

func (s BEStruct) isStruct() bool {
	return true
}

func (s BEStruct) endianness() Endianness {
	return BE
}

type internalStruct interface {
	isStruct() bool
	endianness() Endianness
}

func parseOffset(s string) (int64, error) {
	if strings.HasPrefix(s, "0x") {
		return strconv.ParseInt(s[2:], 16, 64)
	}
	return strconv.ParseInt(s, 10, 64)
}

// UnmarshalStruct ...
// Tag Format: `offset:"(hex_or_dec)"`
func UnmarshalStruct[T internalStruct](x T, r *BinaryReader) error {
	baseOffset, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(x)
	rt := reflect.TypeOf(x)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
		rt = rt.Elem()
	}

	for i := 0; i < rt.NumField(); i++ {
		fv := rv.Field(i)
		ff := rt.Field(i)
		ft := ff.Type

		// Get offset
		offsetTag, ok := ff.Tag.Lookup("offset")
		if !ok {
			continue
		}

		offset, err := parseOffset(offsetTag)
		if err != nil {
			return fmt.Errorf("invalid offset struct tag for '%s': %w", ft.Name, err)
		}

		// TODO: allow field-by-field endianness switch?

		// Seek to offset
		if _, err := r.Seek(baseOffset+offset, io.SeekStart); err != nil {
			return err
		}

		// Read field
		if err := unmarshalField(r, x.endianness(), fv, ft); err != nil {
			return err
		}
	}

	return nil
}

func unmarshalField(r *BinaryReader, e Endianness, fv reflect.Value, ft reflect.Type) error {
	// Recurse for each elem if array
	if ft.Kind() == reflect.Array {
		currT := ft.Elem()

		for i := 0; i < fv.Len(); i++ {
			currV := fv.Index(i)
			if err := unmarshalField(r, e, currV, currT); err != nil {
				return nil
			}
		}
		return nil
	}

	switch ft.Kind() {
	case reflect.Array:
		currT := ft.Elem()
		for i := 0; i < fv.Len(); i++ {
			currV := fv.Index(i)
			if err := unmarshalField(r, e, currV, currT); err != nil {
				return nil
			}
		}
		return nil
	case reflect.Slice:
		break // TODO: could support slice with len from another value in the future
	}

	// If it's another struct type, call UnmarshalStruct again
	if ft.Implements(reflect.TypeOf((*internalStruct)(nil)).Elem()) {
		if ft.Kind() == reflect.Ptr {
			if fv.IsNil() {
				fv.Set(reflect.New(ft.Elem()))
			}
		} else {
			fv = fv.Addr()
			ft = fv.Type()
		}
		return UnmarshalStruct(fv.Interface().(internalStruct), r)
	}

	// It's a base type
	switch e {
	case LE:
		switch x := fv.Addr().Interface().(type) {
		case *uint8:
			return r.Uint8(x)
		case *uint16:
			return r.Uint16LE(x)
		case *uint32:
			return r.Uint32LE(x)
		case *uint64:
			return r.Uint64LE(x)
		case *int8:
			return r.Int8(x)
		case *int16:
			return r.Int16LE(x)
		case *int32:
			return r.Int32LE(x)
		case *int64:
			return r.Int64LE(x)
		case *float32:
			return r.Float32LE(x)
		default:
			return fmt.Errorf("unsupported little endian field type: %s", ft.Name())
		}
	case BE:
		switch x := fv.Addr().Interface().(type) {
		case *uint8:
			return r.Uint8(x)
		case *uint16:
			return r.Uint16BE(x)
		case *uint32:
			return r.Uint32BE(x)
		case *uint64:
			return r.Uint64BE(x)
		case *int8:
			return r.Int8(x)
		case *int16:
			return r.Int16BE(x)
		case *int32:
			return r.Int32BE(x)
		case *int64:
			return r.Int64BE(x)
		case *float32:
			return r.Float32BE(x)
		default:
			return fmt.Errorf("unsupported big endian field type: %s", ft.Name())
		}
	default:
		return fmt.Errorf("invalid endianess") // Not possible
	}
}
