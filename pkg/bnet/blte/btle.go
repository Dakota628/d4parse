package blte

import (
	"errors"
	"github.com/Dakota628/d4parse/pkg/bin"
	"io"
)

var (
	signature = [4]uint8{'B', 'T', 'L', 'E'}

	ErrInvalidSignature = errors.New("invalid signature")
	ErrInvalidChunk     = errors.New("invalid chunk")
)

type Header struct {
	bin.BEStruct
	FileSignature [4]uint8 `offset:"0x0"`
	HeaderSize    uint32   `offset:"0x04"`
}

func (h *Header) UnmarshalBinary(r *bin.BinaryReader) error {
	if err := bin.UnmarshalStruct(h, r); err != nil {
		return err
	}
	if signature != h.FileSignature {
		return ErrInvalidSignature
	}
	return nil
}

type ChunkInfoEntry struct {
	bin.BEStruct
	CompressedSize   uint32    `offset:"0x0"`
	DecompressedSize uint32    `offset:"0x04"`
	Checksum         [16]uint8 `offset:"0x08"`
}

func (e *ChunkInfoEntry) UnmarshalBinary(r *bin.BinaryReader) error {
	return bin.UnmarshalStruct(e, r)
}

type ChunkInfo struct {
	bin.BEStruct
	Flags      uint8            `offset:"0x0"`
	ChunkCount [3]uint8         `offset:"0x02"` // type:uint24
	ChunkInfos []ChunkInfoEntry // len:ChunkCount
}

func (c *ChunkInfo) UnmarshalBinary(r *bin.BinaryReader) error {
	// Read static fields
	if err := bin.UnmarshalStruct(c, r); err != nil {
		return err
	}

	// Read chunks
	c.ChunkInfos = make([]ChunkInfoEntry, bin.Uint24(c.ChunkCount))
	for i := range c.ChunkInfos {
		if err := c.ChunkInfos[i].UnmarshalBinary(r); err != nil {
			return err
		}
	}

	return nil
}

type EncodingMode byte

var (
	PlainEncodingMode      EncodingMode = 'N'
	ZLibEncodingMode       EncodingMode = 'Z'
	LZ4EncodingMode        EncodingMode = '4'
	RecursiveEncodingMode  EncodingMode = 'F'
	EncryptionEncodingMode EncodingMode = 'E'
)

type Chunk struct {
	EncodingMode EncodingMode
	Data         []byte
}

func (c Chunk) Decode() ([]byte, error) {
	switch c.EncodingMode {
	case PlainEncodingMode:
		return c.Data, nil
	case ZLibEncodingMode:
		return nil, errors.New("not implemented") // TODO
	case LZ4EncodingMode:
		return nil, errors.New("not implemented") // TODO
	case RecursiveEncodingMode:
		return nil, errors.New("not implemented") // TODO
	case EncryptionEncodingMode:
		return nil, errors.New("not implemented") // TODO
	default:
		return nil, ErrInvalidChunk
	}
}

type Reader struct {
	binReader *bin.BinaryReader
	header    Header
	chunkInfo *ChunkInfo
	offset    int
}

func NewReader(binReader *bin.BinaryReader) (*Reader, error) {
	r := &Reader{
		binReader: binReader,
	}

	// Read header
	if err := r.header.UnmarshalBinary(binReader); err != nil {
		return nil, err
	}

	if r.header.HeaderSize > 0 {
		r.chunkInfo = &ChunkInfo{}
		if err := r.chunkInfo.UnmarshalBinary(binReader); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (r *Reader) Offset() int {
	return r.offset
}

func (r *Reader) Len() int {
	if r.chunkInfo != nil {
		return len(r.chunkInfo.ChunkInfos)
	}
	return 1
}

func (r *Reader) ReadChunk() (Chunk, error) {
	if r.Offset() >= r.Len() {
		return Chunk{}, io.EOF
	}

	// Single chunk if no chunk info
	if r.chunkInfo == nil {
		buf, err := io.ReadAll(r.binReader)
		if err != nil {
			return Chunk{}, err
		}

		if len(buf) < 1 {
			return Chunk{}, ErrInvalidChunk
		}

		r.offset++
		return Chunk{
			EncodingMode: EncodingMode(buf[0]),
			Data:         buf[1:],
		}, nil
	}

	// Otherwise, read next chunk
	chunkInfo := r.chunkInfo.ChunkInfos[r.offset]

	buf := make([]byte, chunkInfo.CompressedSize)
	if _, err := r.binReader.Read(buf); err != nil {
		return Chunk{}, err
	}

	if len(buf) < 1 {
		return Chunk{}, ErrInvalidChunk
	}

	r.offset++
	return Chunk{
		EncodingMode: EncodingMode(buf[0]),
		Data:         buf[1:],
	}, nil
}
