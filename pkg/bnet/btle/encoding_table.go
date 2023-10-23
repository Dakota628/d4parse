package btle

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
	//// Signature
	//if err := r.Uint8(&h.Signature[0]); err != nil {
	//	return err
	//}
	//if err := r.Uint8(&h.Signature[1]); err != nil {
	//	return err
	//}
	//
	//// Version
	//if err := r.Uint8(&h.Version); err != nil {
	//	return err
	//}
	//
	//// Hash Sizes
	//if err := r.Uint8(&h.HashSizeCKey); err != nil {
	//	return err
	//}
	//if err := r.Uint8(&h.HashSizeEKey); err != nil {
	//	return err
	//}
	//
	//// Page Sizes
	//if err := r.Uint16BE(&h.CEKeyPageTablePageSizeKb); err != nil {
	//	return err
	//}
	//if err := r.Uint16BE(&h.EKeySpecPageTablePageSizeKb); err != nil {
	//	return err
	//}
	//
	//// Page Counts
	//if err := r.Uint32BE(&h.CEKeyPageTablePageCount); err != nil {
	//	return err
	//}
	//if err := r.Uint32BE(&h.EKeyPageTablePageCount); err != nil {
	//	return err
	//}
	//
	//// Skip unk
	//if _, err := r.Seek(1, io.SeekCurrent); err != nil {
	//	return err
	//}
	//
	//// ESpec Block Size
	//if err := r.Uint32BE(&h.ESpecBlockSize); err != nil {
	//	return err
	//}
	//
	//// Skip padding
	//if _, err := r.Seek(4, io.SeekCurrent); err != nil {
	//	return err
	//}
	//
	//return nil
	return bin.UnmarshalStruct(h, r)
}

type EncodingTable struct {
	bin.BEStruct
	Header EncodingTableHeader `offset:"0x0"`
}

func (t *EncodingTable) UnmarshalBinary(r *bin.BinaryReader) error {
	return bin.UnmarshalStruct(t, r)
}
