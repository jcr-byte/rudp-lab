package main

import (
	"errors"
	"fmt"
	"net"
	"time"
	"log"

	"github.com/jcr-byte/rudp-lab/internal/packet"
)

const (
	timeout    = 500 * time.Millisecond
	maxRetries = 5
	maxPayload = 5
)

func main() {
	raddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:9000")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	data := []byte("hello world")
	var currentSeq uint16 = 1
	for offset := 0; offset < len(data); offset += maxPayload {

		end := min(offset+maxPayload, len(data))
		p := packet.Packet{Flag: packet.FlagData, Seq: currentSeq, Checksum: 0xBEEF, Payload: data[offset:end]}
		encoded := p.Encode()

		buf := make([]byte, 2048)
		acked := false
		for attempt := 0; attempt < maxRetries; attempt++ {
			_, err := conn.Write(encoded)
			if err != nil {
				log.Fatal(err)
			}

			conn.SetReadDeadline(time.Now().Add(timeout))

			n, err := conn.Read(buf)
			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					fmt.Println("no ack, retransmitting")
					continue
				}
				log.Fatal("read failed", err)
			}

			if !packet.Verify(buf[:n]) {
				fmt.Println("Recieved packet is corrupted")
				continue
			}

			decodedPacket, err := packet.Decode(buf[:n])
			if err != nil {
				log.Fatal(err)
			}
			if decodedPacket.Flag == packet.FlagAck && decodedPacket.Seq == p.Seq {
				fmt.Println("Ack arrived and is valid")
				currentSeq++
				acked = true
				break
			}
		}

		if !acked {
			log.Fatalf("giving up on seq %d", p.Seq)
		}
	}
}
