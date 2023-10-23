package bin

import (
	"errors"
	"golang.org/x/exp/slices"
	"io"
)

// StreamReadSeeker ...
// deprecated: not currently tested
type StreamReadSeeker struct {
	stream io.Reader
	offset int64
	buf    []byte
	eof    bool
}

func NewStreamReadSeeker(stream io.Reader) *StreamReadSeeker {
	return &StreamReadSeeker{
		stream: stream,
	}
}

func (s *StreamReadSeeker) grow(want int) error {
	if s.eof {
		return io.EOF
	}

	available := int(int64(len(s.buf)) - s.offset)

	if need := available - want; need > 0 {
		i := len(s.buf)
		s.buf = slices.Grow(s.buf, need)
		s.buf = s.buf[:i]
		m, err := s.stream.Read(s.buf[i:cap(s.buf)])
		s.buf = s.buf[:i+m]

		if errors.Is(err, io.EOF) {
			s.eof = true
		} else if err != nil {
			return err
		}
	}

	return nil
}

func (s *StreamReadSeeker) Read(p []byte) (int, error) {
	want := len(p)
	if err := s.grow(want); err != nil {
		return 0, err
	}
	copy(p, s.buf[s.offset:int(s.offset)+want])
	s.offset += int64(want)
	return want, nil
}

func (s *StreamReadSeeker) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekEnd:
		more, err := io.ReadAll(s.stream)
		if err != nil {
			return s.offset, err
		}
		s.buf = append(s.buf, more...)
		s.offset = int64(len(s.buf)) - offset
	case io.SeekCurrent:
		if s.offset+offset >= int64(len(s.buf)) {
			if err := s.grow(int(s.offset) + int(offset) - len(s.buf)); err != nil {
				return 0, err
			}
		}
		s.offset += offset
	case io.SeekStart:
		if offset >= int64(len(s.buf)) {
			if err := s.grow(int(offset) - len(s.buf)); err != nil {
				return 0, err
			}
		}
		s.offset = offset
	default:
		return 0, errors.New("invalid seek")
	}

	return s.offset, nil
}
