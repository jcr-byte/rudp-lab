package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/jcr-byte/rudp-lab/internal/packet"
)

const (
	timeout    = 500 * time.Millisecond
	maxRetries = 5
)

func main() {
	raddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:9000")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	p := packet.Packet{Flag: packet.FlagData, Seq: 300, Checksum: 0xBEEF, Payload: []byte("hello")}
	encoded := p.Encode()

	buf := make([]byte, 2048)
	for attempt := 0; attempt < maxRetries; attempt++ {
		_, err := conn.Write(encoded)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		conn.SetReadDeadline(time.Now().Add(timeout))

		n, err := conn.Read(buf)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				fmt.Println("no ack, retransmitting")
				continue
			}
			fmt.Println("read failed:", err)
			os.Exit(1)
		}

		if !packet.Verify(buf[:n]) {
			fmt.Println("Recieved packet is corrupted")
			continue
		}

		decodedPacket, err := packet.Decode(buf[:n])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if decodedPacket.Flag == packet.FlagAck && decodedPacket.Seq == p.Seq {
			fmt.Println("Ack arrived and is valid")
			break
		}
	}
}
