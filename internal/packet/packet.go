package packet

import "encoding/binary"

// Wire format (big-endian):
//	byte 0: Flag
//	bytes 1-2: Seq (uint16)
//	bytes 3-4: Checksum (uint16)
//	bytes 5+: Payload
type Packet struct {
	Flag byte
	Seq uint16
	Checksum uint16
	Payload []byte
}

func (packet *Packet) Encode() []byte {
	buf := make([]byte, 5+len(packet.Payload))
	buf[0] = packet.Flag
	binary.BigEndian.PutUint16(buf[1:3], packet.Seq)
	binary.BigEndian.PutUint16(buf[3:5], packet.Checksum)
	copy(buf[5:], packet.Payload)
	return buf
}