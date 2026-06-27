package main

import (
	"fmt"
	"net"
	"os"

	"github.com/jcr-byte/rudp-lab/internal/packet"
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

	p := packet.Packet{Flag: 0xA5, Seq: 300, Checksum: 0xBEEF, Payload: []byte("hello")}
	encoded := p.Encode()

	n, err := conn.Write(encoded)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("sent %d bytes\n", n)
}
