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
	Meta   UnmarshalBinary
}

func (t *SnoMeta) UnmarshalBinary(r *bin.BinaryReader) error {
	// Read SNOFileHeader
	if err := t.Header.UnmarshalBinary(r); err != nil {
		return err
	}

	// Offset the reader by the header length
	if err := r.Offset(SNOFileHeaderSize); err != nil {
		return err
	}

	// Read Id, but don't advance pointer as Meta padding will skip the Id
	if err := r.AtPos(0, io.SeekCurrent, t.Id.UnmarshalBinary); err != nil {
		return err
	}

	// Read meta
	if t.Meta = NewByFormatHash(int(t.Header.DwFormatHash.Value)); t.Meta == nil {
		return fmt.Errorf("could not find type for format hash: %d", t.Header.DwFormatHash)
	}
	return t.Meta.UnmarshalBinary(r)
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
	return snoMeta, snoMeta.UnmarshalBinary(r)
}
