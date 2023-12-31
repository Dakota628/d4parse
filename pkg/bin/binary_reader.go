package bin

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"strings"
)

type BinaryReader struct {
	io.ReadSeeker
	offset int64
}

func NewBinaryReader(rs io.ReadSeeker) *BinaryReader {
	return &BinaryReader{
		ReadSeeker: rs,
		offset:     0,
	}
}

func (r *BinaryReader) AtPos(offset int64, whence int, f func(r *BinaryReader) error) error {
	// Save current pos
	startPos, err := r.Pos()
	if err != nil {
		return err
	}

	// Change pos
	if _, err = r.Seek(offset, whence); err != nil {
		return err
	}

	// Call callback
	if err = f(r); err != nil {
		return err
	}

	// Restore pos
	if _, err = r.Seek(startPos, io.SeekStart); err != nil {
		return err
	}

	return nil
}

func (r *BinaryReader) Pos() (int64, error) {
	return r.Seek(0, io.SeekCurrent)
}

func (r *BinaryReader) Size() (int64, error) {
	// Get current pos
	currPos, err := r.Pos()
	if err != nil {
		return 0, err
	}

	// Get end of file pos
	endPos, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}

	// Restore pos
	if _, err = r.Seek(currPos, io.SeekStart); err != nil {
		return 0, err
	}

	return endPos + 1, nil
}

func (r *BinaryReader) Uint8(x *uint8) error {
	buf := make([]byte, 1)
	if _, err := r.Read(buf); err != nil {
		return err
	}
	*x = buf[0]
	return nil
}

func (r *BinaryReader) Uint16LE(x *uint16) error {
	buf := make([]byte, 2)
	if _, err := r.Read(buf); err != nil {
		return err
	}
	*x = binary.LittleEndian.Uint16(buf)
	return nil
}

func (r *BinaryReader) Uint16BE(x *uint16) error {
	buf := make([]byte, 2)
	if _, err := r.Read(buf); err != nil {
		return err
	}
	*x = binary.BigEndian.Uint16(buf)
	return nil
}

func (r *BinaryReader) Uint32LE(x *uint32) error {
	buf := make([]byte, 4)
	if _, err := r.Read(buf); err != nil {
		return err
	}
	*x = binary.LittleEndian.Uint32(buf)
	return nil
}

func (r *BinaryReader) Uint32BE(x *uint32) error {
	buf := make([]byte, 4)
	if _, err := r.Read(buf); err != nil {
		return err
	}
	*x = binary.BigEndian.Uint32(buf)
	return nil
}

func (r *BinaryReader) Uint64LE(x *uint64) error {
	buf := make([]byte, 8)
	if _, err := r.Read(buf); err != nil {
		return err
	}
	*x = binary.LittleEndian.Uint64(buf)
	return nil
}

func (r *BinaryReader) Uint64BE(x *uint64) error {
	buf := make([]byte, 8)
	if _, err := r.Read(buf); err != nil {
		return err
	}
	*x = binary.BigEndian.Uint64(buf)
	return nil
}

func (r *BinaryReader) Int8(x *int8) error {
	var temp uint8
	if err := r.Uint8(&temp); err != nil {
		return err
	}
	*x = int8(temp)
	return nil
}

func (r *BinaryReader) Int16LE(x *int16) error {
	var temp uint16
	if err := r.Uint16LE(&temp); err != nil {
		return err
	}
	*x = int16(temp)
	return nil
}

func (r *BinaryReader) Int16BE(x *int16) error {
	var temp uint16
	if err := r.Uint16BE(&temp); err != nil {
		return err
	}
	*x = int16(temp)
	return nil
}

func (r *BinaryReader) Int32LE(x *int32) error {
	var temp uint32
	if err := r.Uint32LE(&temp); err != nil {
		return err
	}
	*x = int32(temp)
	return nil
}

func (r *BinaryReader) Int32BE(x *int32) error {
	var temp uint32
	if err := r.Uint32BE(&temp); err != nil {
		return err
	}
	*x = int32(temp)
	return nil
}

func (r *BinaryReader) Int64LE(x *int64) error {
	var temp uint64
	if err := r.Uint64LE(&temp); err != nil {
		return err
	}
	*x = int64(temp)
	return nil
}

func (r *BinaryReader) Int64BE(x *int64) error {
	var temp uint64
	if err := r.Uint64BE(&temp); err != nil {
		return err
	}
	*x = int64(temp)
	return nil
}

func (r *BinaryReader) Float32LE(x *float32) error {
	buf := make([]byte, 4)
	if _, err := r.Read(buf); err != nil {
		return err
	}
	*x = math.Float32frombits(binary.LittleEndian.Uint32(buf))
	return nil
}

func (r *BinaryReader) Float32BE(x *float32) error {
	buf := make([]byte, 4)
	if _, err := r.Read(buf); err != nil {
		return err
	}
	*x = math.Float32frombits(binary.BigEndian.Uint32(buf))
	return nil
}

func (r *BinaryReader) Offset(offset int64) error {
	offset += r.offset

	// Seek to new offset
	if _, err := r.Seek(offset, io.SeekStart); err != nil {
		return err
	}

	// Set offset
	r.offset = offset
	return nil
}

func (r *BinaryReader) Seek(offset int64, whence int) (int64, error) {
	// Adjust seeks by r.offset
	if whence == io.SeekStart {
		offset += r.offset
	}

	newOffset, err := r.ReadSeeker.Seek(offset, whence)
	newOffset -= r.offset

	if newOffset < 0 {
		return newOffset - r.offset, errors.New("invalid seek before offset ")
	}

	return newOffset, err
}

func (r *BinaryReader) NullTermString(x *string) error {
	bufReader := bufio.NewReader(r)
	sb := strings.Builder{}
	var b byte
	var err error

	for {
		if b, err = bufReader.ReadByte(); err == io.EOF || (err == nil && b == 0) {
			*x = sb.String()
			return nil
		} else if err != nil {
			return err
		}
		if err := sb.WriteByte(b); err != nil {
			return err
		}
	}
}
