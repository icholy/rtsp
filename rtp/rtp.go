package rtp

import (
	"encoding/binary"
	"errors"
	"time"
)

const (
	// RtpVersion is used to verify compliance with current specification of the RTP protocol.
	RtpVersion = 2 << 6
	// HeaderSize defines the size of the fixed part of the packet, up to and inclding SSRC.
	HeaderSize = 12
)

// Packet encapsulates RTP packet structure.
//  0                   1                   2                   3
//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |V=2|P|X|  CC   |M|     PT      |       sequence number         |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                           timestamp                           |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |           synchronization source (SSRC) identifier            |
// +=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
// |            contributing source (CSRC) identifiers             |
// |                             ....                              |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
type Packet struct {
	VPXCC   byte     // Version, Padding, Extension, Contributing Source Count
	MPT     byte     // Marker, Payload Type
	SN      uint16   // Sequense Number
	TS      uint32   // Timestamp
	SSRC    uint32   // Synchronization Source Identifier
	CSRC    []uint32 // Contributing Source Identifiers
	XH      uint16   // Extension Header (profile dependent)
	XL      uint16   // Extension Length (in `uint`s not inclusing this header)
	XD      []byte   // Extension Data
	Payload []byte   // Payload
}

var order = binary.BigEndian

// Parse validates a packed RTP packet and converts it into a sparse structure.
func Parse(buf []byte) (*Packet, error) {
	if len(buf) < HeaderSize {
		return nil, errors.New("RTP header too short")
	}
	if (buf[0] & 0xC0) != RtpVersion {
		return nil, errors.New("RTP version not supported")
	}
	packet := &Packet{
		VPXCC: buf[0],
		MPT:   buf[1],
		SN:    order.Uint16(buf[2:]),
		TS:    order.Uint32(buf[4:]),
		SSRC:  order.Uint32(buf[8:]),
	}
	off := HeaderSize
	packet.CSRC = make([]uint32, packet.ContributingCount())
	if len(buf[off:]) < len(packet.CSRC)*4 {
		return nil, errors.New("RTP incorrect contributing count")
	}
	for i := range packet.CSRC {
		packet.CSRC[i] = order.Uint32(buf[off:])
		off += 4
	}
	if packet.Extension() {
		if len(buf[off:]) < 4 {
			return nil, errors.New("RTP extension header missing")
		}
		packet.XH = order.Uint16(buf[off:])
		packet.XL = order.Uint16(buf[off+2:])
		off += 4
		if packet.XL > 0 {
			if len(buf[off:]) < int(packet.XL)*4 {
				return nil, errors.New("RTP extension data missing")
			}
			packet.XD = buf[off : off+int(packet.XL)*4]
			off += int(packet.XL) * 4
		}
	}
	packet.Payload = buf[off:]
	return packet, nil
}

// Time returns Timestamp value
func (p Packet) Time() time.Time {
	return time.Unix(int64(p.TS), 0)
}

// Padding returns Padding flag value of the packet.
func (p Packet) Padding() bool {
	return (p.VPXCC & 0x20) != 0
}

// Extension returns Extension flag value of the packet.
func (p Packet) Extension() bool {
	return (p.VPXCC & 0x10) != 0
}

// ContributingCount returns Contributing Source Count of the packet.
func (p Packet) ContributingCount() int {
	return int(p.VPXCC & 0x0F)
}

// Marker returns Marker value of the packet.
func (p Packet) Marker() bool {
	return (p.MPT & 0x80) != 0
}

// PayloadType returns Payload Type of the packet.
func (p Packet) PayloadType() int {
	return int(p.MPT & 0x7F)
}
