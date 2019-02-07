package rtsp

import (
	"bufio"
	"encoding/binary"
	"fmt"
)

type Binary struct {
	Channel int
	Data    []byte
}

func ReadInterleaved(b *bufio.Reader) (Binary, error) {
	magic, err := b.ReadByte()
	if err != nil {
		return Binary{}, err
	}
	if magic != '$' {
		return Binary{}, fmt.Errorf("invalid magic prefix: %s", magic)
	}
	channel, err := b.ReadByte()
	if err != nil {
		return Binary{}, err
	}
	var length uint16
	if err := binary.Read(b, binary.LittleEndian, &length); err != nil {
		return Binary{}, err
	}
	data := make([]byte, length)
	if _, err := b.Read(data); err != nil {
		return Binary{}, err
	}
	return Binary{
		Channel: int(channel),
		Data:    data,
	}, nil
}
