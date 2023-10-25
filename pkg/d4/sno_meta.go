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

func (m *SnoMeta) UnmarshalD4(r *bin.BinaryReader, o *Options) error {
	// Read SNOFileHeader
	if err := m.Header.UnmarshalD4(r, nil); err != nil {
		return err
	}

	// Offset the reader by the header length
	if err := r.Offset(SNOFileHeaderSize); err != nil {
		return err
	}

	// Read Id, but don't advance pointer as Meta padding will skip the Id
	if err := r.AtPos(0, io.SeekCurrent, func(r *bin.BinaryReader) error {
		return m.Id.UnmarshalD4(r, nil)
	}); err != nil {
		return err
	}

	// Read meta
	if m.Meta = NewByFormatHash(int(m.Header.DwFormatHash.Value)); m.Meta == nil {
		return fmt.Errorf("could not find type for format hash: %d", m.Header.DwFormatHash)
	}

	if err := m.Meta.UnmarshalD4(r, nil); err != nil {
		return err
	}

	return nil
}

func (m *SnoMeta) Walk(cb WalkCallback, d ...any) {
	cb.Do("", m.Meta, d...)
}

func (m *SnoMeta) GetFlags() int {
	return 0
}

// GetReferences gets a list of SNO IDs referenced by this SNO. Will also add GameBalance SNO references if gbData is
// not nil.
func (m *SnoMeta) GetReferences(gbData *GbData) (refs []int32) {
	// Assert Walkable
	x, ok := m.Meta.(Walkable)
	if !ok {
		return
	}

	// Walk SNO and keep track of referenced SNOs
	x.Walk(func(_ string, v Object, next WalkNext, d ...any) {
		var id int32

		switch t := v.(type) {
		case *DT_SNO:
			id = t.Id
		case *DT_SNO_NAME:
			id = t.Id
		case *DT_GBID:
			if gbData != nil {
				if gbInfoIfc, ok := gbData.Load(*t); ok {
					if gbInfo, ok := gbInfoIfc.(GbInfo); ok {
						id = gbInfo.SnoId
					}
				}
			}
		}

		if id > 0 {
			refs = append(refs, id)
		}

		next(d...)
	})

	return
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

func ReadSnoMetaHeader(path string) (header SNOFileHeader, err error) {
	// Open file
	f, err := os.Open(path)
	if err != nil {
		return header, err
	}
	defer f.Close()

	// Create binary reader
	r := bin.NewBinaryReader(f)

	// Read SNOFileHeader
	if err := header.UnmarshalD4(r, nil); err != nil {
		return header, err
	}

	return header, nil
}

func GetDefinition[T Object](meta SnoMeta) (T, error) {
	def, ok := meta.Meta.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("meta does not contain definition of type %T", zero)
	}
	return def, nil
}
