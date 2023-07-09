package d4

import "github.com/Dakota628/d4parse/pkg/bin"

// TODO: impl toc parser

type SnoGroup int32

const (
	SnoGroupActor SnoGroup = 1
	// TODO: generate rest of enums
	SnoGroupFace    SnoGroup = 140
	SnoGroupUnknown SnoGroup = -3
	SnoGroupCode    SnoGroup = -2
	SnoGroupNone    SnoGroup = -1
)

type TocEntry struct {
	SnoGroup SnoGroup
	SnoId    int32
	PName    int32

	SnoName string
}

func (t *TocEntry) UnmarshalBinary(r *bin.BinaryReader) error {
	return nil // TODO
}

type Toc struct {
	NumSnoGroups   int32
	EntryCounts    []int32 // n = numSnoGroups
	EntryOffsets   []int32 // n = numSnoGroups
	EntryUnkCounts []int32 // n = numSnoGroups
	Unk1           int32
	Entries        []TocEntry
}

func (t *Toc) UnmarshalBinary(r *bin.BinaryReader) error {
	return nil // TODO
}
