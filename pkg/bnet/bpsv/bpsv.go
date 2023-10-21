package bpsv

import (
	"bytes"
	"errors"
	"io"
	"strconv"
)

var (
	ErrInvalidFieldFormat    = errors.New("invalid field format")
	ErrInvalidRowFormat      = errors.New("invalid row format")
	ErrInvalidDocumentFormat = errors.New("invalid document format")

	SeqnPrefix = []byte("## seqn = ")
)

// Field ...
type Field struct {
	Name   string
	Type   string // STRING => string, HEX => []byte, DEC => int64
	Length int
}

func ParseField(bs []byte) (Field, []byte, error) {
	if len(bs) == 0 || bytes.Equal(bs, []byte{'\n'}) {
		return Field{}, nil, io.EOF
	}

	var f Field

	// Parse name
	idx := bytes.IndexByte(bs, '!')
	if idx == -1 {
		return f, bs, ErrInvalidFieldFormat
	}

	f.Name = string(bs[:idx])
	bs = bs[idx+1:]

	// Parse type
	idx = bytes.IndexByte(bs, ':')
	if idx == -1 {
		return f, bs, ErrInvalidFieldFormat
	}

	f.Type = string(bs[:idx])
	bs = bs[idx+1:]

	// Parse length
	var lenStr string
	idx = bytes.IndexAny(bs, "|\n")
	if idx == -1 {
		lenStr = string(bs)
		bs = []byte{}
	} else {
		lenStr = string(bs[:idx])
		bs = bs[idx+1:]
	}

	var err error
	f.Length, err = strconv.Atoi(lenStr)
	if err != nil {
		return f, bs, ErrInvalidFieldFormat
	}

	return f, bs, nil
}

func ParseFields(bs []byte) (fields []Field, err error) {
	bs = bytes.TrimRight(bs, "\n")

	var f Field
	for {
		f, bs, err = ParseField(bs)

		// If eof, no more fields to parse
		if errors.Is(err, io.EOF) {
			err = nil
			break
		}

		// Other err; failure
		if err != nil {
			return
		}

		fields = append(fields, f)
	}

	return fields, nil
}

// Row ...
type Row map[string]string

func ParseRow(bs []byte, fields []Field) (Row, error) {
	bs = bytes.TrimRight(bs, "\n")
	values := bytes.Split(bs, []byte{'|'})

	if len(values) != len(fields) {
		return nil, ErrInvalidRowFormat
	}

	row := make(Row, len(values))
	for i := range values {
		row[fields[i].Name] = string(values[i])
	}

	return row, nil
}

// Document ...
type Document struct {
	Fields []Field
	Rows   []Row
	Seqn   int
}

func ParseSeqn(bs []byte) (int, bool) {
	if !bytes.HasPrefix(bs, SeqnPrefix) {
		return 0, false
	}

	bs = bytes.TrimRight(bs, "\n")

	seqn, err := strconv.Atoi(string(bs[len(SeqnPrefix):]))
	return seqn, err == nil
}

func ParseDocument(bs []byte) (Document, error) {
	var doc Document

	// Split by new line
	lines := bytes.Split(bs, []byte{'\n'})

	if len(lines) < 1 {
		return doc, ErrInvalidDocumentFormat
	}

	// Parse fields
	var err error
	if doc.Fields, err = ParseFields(lines[0]); err != nil {
		return doc, err
	}
	lines = lines[1:]

	// Check for seqn
	if len(lines) == 0 {
		return doc, nil
	}

	var ok bool
	if doc.Seqn, ok = ParseSeqn(lines[0]); ok {
		lines = lines[1:]
	}

	// Parse rows
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		row, err := ParseRow(line, doc.Fields)
		if err != nil {
			return doc, err
		}

		doc.Rows = append(doc.Rows, row)
	}

	return doc, nil
}
