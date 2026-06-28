package packet_test

import (
	"bytes"
	"testing"

	"github.com/jcr-byte/rudp-lab/internal/packet"
)

func TestRoundTrip(t *testing.T) {
	p := packet.Packet{Flag: 0xA5, Seq: 300, Checksum: 0xBEEF, Payload: []byte("hello")}
	encoded := p.Encode()
	decoded, err := packet.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.Flag != p.Flag {
		t.Errorf("Flag: got %d, want %d", decoded.Flag, p.Flag)
	}

	if decoded.Seq != p.Seq {
		t.Errorf("Seq: got %d, want %d", decoded.Seq, p.Seq)
	}

	if decoded.Checksum != p.Checksum {
		t.Errorf("Checksum: got %d, want %d", decoded.Checksum, p.Checksum)
	}

	if !bytes.Equal(decoded.Payload, p.Payload) {
		t.Errorf("Payload: got %q, want %q", decoded.Payload, p.Payload)
	}
}

func TestChecksum(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04}
	checksum := packet.Checksum(data)
	if checksum != 10 {
		t.Errorf("Checksum: got %d, want 10", checksum)
	}
	data[0] = 0x02

	checksum = packet.Checksum(data)
	if checksum != 11 {
		t.Errorf("Checksum: got %d, want 11", checksum)
	}
}
