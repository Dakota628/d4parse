package blte

import (
	"github.com/Dakota628/d4parse/pkg/bin"
)

type EncodingTableHeader struct {
	bin.BEStruct
	Signature                   [2]uint8 `offset:"0x0"`
	Version                     uint8    `offset:"0x02"`
	HashSizeCKey                uint8    `offset:"0x03"`
	HashSizeEKey                uint8    `offset:"0x04"`
	CEKeyPageTablePageSizeKb    uint16   `offset:"0x05"`
	EKeySpecPageTablePageSizeKb uint16   `offset:"0x07"`
	CEKeyPageTablePageCount     uint32   `offset:"0x09"`
	EKeyPageTablePageCount      uint32   `offset:"0x0D"`
	UnkX11                      uint8    `offset:"0x11"`
	ESpecBlockSize              uint32   `offset:"0x12"`
}

func (h *EncodingTableHeader) UnmarshalBinary(r *bin.BinaryReader) error {
	return bin.UnmarshalStruct(h, r)
}

type EncodingTableESpec struct {
}

type EncodingTable struct {
	bin.BEStruct
	Header EncodingTableHeader `offset:"0x0"`
}

func (t *EncodingTable) UnmarshalBinary(r *bin.BinaryReader) error {
	return bin.UnmarshalStruct(t, r)
}
