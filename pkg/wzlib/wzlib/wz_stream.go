package wzlib

import (
	"fmt"
	"io"
	"os"
)

type PartialStream struct {
	Base     *os.File
	Offset   int64
	Length   int64
	position int64
}

func NewPartialStream(base *os.File, offset, length int64) (io.ReadSeeker, error) {
	p := &PartialStream{
		Base:     base,
		Offset:   offset,
		Length:   length,
		position: 0,
	}
	return p, nil
}

func (ps *PartialStream) Read(p []byte) (int, error) {
	if ps.position >= ps.Length {
		return 0, io.EOF
	}
	remain := ps.Length - ps.position
	if int64(len(p)) > remain {
		p = p[:remain]
	}
	n, err := ps.Base.ReadAt(p, ps.Offset+ps.position)
	ps.position += int64(n)
	return n, err
}

func (ps *PartialStream) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = ps.position + offset
	case io.SeekEnd:
		abs = ps.Length + offset
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}
	if abs < 0 || abs > ps.Length {
		return 0, fmt.Errorf("seek out of bounds")
	}
	ps.position = abs
	return abs, nil
}
