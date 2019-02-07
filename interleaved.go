package rtsp

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Frame struct {
	Channel int
	Data    []byte
}

type frameHeader struct {
	Magic   byte
	Channel uint8
	Length  uint16
}

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