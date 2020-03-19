package rtsp

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

// Frame of interleaved binary data.
// This is encoded in RTP format.
type Frame struct {
	Channel int
	Data    []byte
}

type frameHeader struct {
	Magic   byte
	Channel uint8
	Length  uint16
}

// Write the interleaved frame to the provided writer.
func (f Frame) Write(w io.Writer) error {
	hdr := frameHeader{
		Magic:   '$',
		Channel: uint8(f.Channel),
		Length:  uint16(len(f.Data)),
	}
	if err := binary.Write(w, binary.BigEndian, hdr); err != nil {
		return err
	}
	if _, err := w.Write(f.Data); err != nil {
		return err
	}
	return nil
}

// ReadFrame reads an interleaved binary frame from the reader.
func ReadFrame(r io.Reader) (Frame, error) {
	var hdr frameHeader
	if err := binary.Read(r, binary.BigEndian, &hdr); err != nil {
		return Frame{}, err
	}
	if hdr.Magic != '$' {
		return Frame{}, fmt.Errorf("invalid magic prefix: %v", hdr.Magic)
	}
	data := make([]byte, hdr.Length)
	if _, err := io.ReadFull(r, data); err != nil {
		return Frame{}, err
	}
	return Frame{
		Channel: int(hdr.Channel),
		Data:    data,
	}, nil
}

// IsFrameNext returns true when the next message is an interleaved frame
func IsFrameNext(r *bufio.Reader) (bool, error) {
	first, err := r.Peek(1)
	if err != nil {
		return false, err
	}
	return first[0] == '$', nil
}
