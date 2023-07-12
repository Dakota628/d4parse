package d4

import (
	"fmt"
	"github.com/Dakota628/d4parse/pkg/bin"
	"io"
	"os"
)

const SNOFileHeaderSize = 16 // TODO: calculate from struct in case it ever changes

type SnoMeta struct {
	Header SNOFileHeader
	Id     DT_INT
	Meta   Object
}

func (t *SnoMeta) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	// Read SNOFileHeader
	if err := t.Header.UnmarshalD4(r, nil); err != nil {
		return err
	}

	// Offset the reader by the header length
	if err := r.Offset(SNOFileHeaderSize); err != nil {
		return err
	}

	// Read Id, but don't advance pointer as Meta padding will skip the Id
	if err := r.AtPos(0, io.SeekCurrent, func(r *bin.BinaryReader) error {
		return t.Id.UnmarshalD4(r, nil)
	}); err != nil {
		return err
	}

	// Read meta
	if t.Meta = NewByFormatHash(int(t.Header.DwFormatHash.Value)); t.Meta == nil {
		return fmt.Errorf("could not find type for format hash: %d", t.Header.DwFormatHash)
	}

	if err := t.Meta.UnmarshalD4(r, nil); err != nil {
		return err
	}

	return nil
}

func ReadSnoMetaFile(path string) (SnoMeta, error) {
	var snoMeta SnoMeta

	// Open file
	f, err := os.Open(path)
	if err != nil {
		return snoMeta, err
	}
	defer f.Close()

	// Create binary reader
	r := bin.NewBinaryReader(f)

	// Unmarshal meta
	return snoMeta, snoMeta.UnmarshalD4(r, nil)
}
