package rtsp

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// Frame of interleaved binary data.
// This is encoded in RTP format.
type Frame struct {
	Channel int
	Data    []byte
}

// Write the interleaved frame to the provided writer.
func (f Frame) Write(w io.Writer) error {
	if f.Channel < 0 || f.Channel > math.MaxUint8 {
		return fmt.Errorf("invalid channel: %d", f.Channel)
	}
	var hdr [4]byte
	hdr[0] = '$'
	hdr[1] = uint8(f.Channel)
	binary.BigEndian.PutUint16(hdr[2:], uint16(len(f.Data)))
	if _, err := w.Write(hdr[:]); err != nil {
		return err
	}
	if _, err := w.Write(f.Data); err != nil {
		return err
	}
	return nil
}

// ReadFrame reads an interleaved binary frame from the reader.
func ReadFrame(r io.Reader) (Frame, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return Frame{}, err
	}
	if hdr[0] != '$' {
		return Frame{}, fmt.Errorf("invalid magic prefix: %v", hdr[0])
	}
	data := make([]byte, binary.BigEndian.Uint16(hdr[2:]))
	if _, err := io.ReadFull(r, data); err != nil {
		return Frame{}, err
	}
	return Frame{
		Channel: int(hdr[1]),
		Data:    data,
	}, nil
}

// IsFrame returns true when the next message is an interleaved frame.
// This will block until at least one byte is available in the reader
func IsFrame(r *bufio.Reader) (bool, error) {
	first, err := r.Peek(1)
	if err != nil {
		return false, err
	}
	return first[0] == '$', nil
}
