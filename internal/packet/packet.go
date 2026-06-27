package packet

import (
	"encoding/binary"
	"fmt"
)

// Wire format (big-endian):
//
//	byte 0: Flag
//	bytes 1-2: Seq (uint16)
//	bytes 3-4: Checksum (uint16)
//	bytes 5+: Payload
type Packet struct {
	Flag     byte
	Seq      uint16
	Checksum uint16
	Payload  []byte
}

const (
	FlagData byte = 1
	FlagAck  byte = 2
)

func (packet *Packet) Encode() []byte {
	buf := make([]byte, 5+len(packet.Payload))
	buf[0] = packet.Flag
	binary.BigEndian.PutUint16(buf[1:3], packet.Seq)
	binary.BigEndian.PutUint16(buf[3:5], packet.Checksum)
	copy(buf[5:], packet.Payload)
	return buf
}

func Decode(buf []byte) (Packet, error) {
	if len(buf) < 5 {
		return Packet{}, fmt.Errorf("packet too short: %d bytes", len(buf))
	}
	payload := make([]byte, len(buf)-5)
	copy(payload, buf[5:])
	return Packet{
		Flag:     buf[0],
		Seq:      binary.BigEndian.Uint16(buf[1:3]),
		Checksum: binary.BigEndian.Uint16(buf[3:5]),
		Payload:  payload,
	}, nil
}
