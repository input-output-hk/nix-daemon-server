package main

import (
	"bytes"
	"io"

	"github.com/nix-community/go-nix/pkg/wire"
)

type framedSource struct {
	from    io.Reader
	pending *bytes.Buffer
	eof     bool
}

func newFramedSource(from io.Reader) *framedSource {
	return &framedSource{from: from, pending: &bytes.Buffer{}}
}

func (s framedSource) Read(buf []byte) (int, error) {
	if s.eof {
		return 0, io.EOF
	}

	if s.pending.Len() == 0 {
		size, err := wire.ReadUint64(s.from)
		if size == 0 {
			s.eof = true
			return 0, io.EOF
		}
		if err != nil {
			if err == io.EOF {
				s.eof = true
			}
			return int(size), err
		}
		io.Copy(s.pending, io.LimitReader(s.from, int64(size)))
	}

	return s.pending.Read(buf)
}
