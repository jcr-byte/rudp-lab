package main

import (
	"fmt"
	"net"
	"os"

	"github.com/jcr-byte/rudp-lab/internal/packet"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:9000")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	buf := make([]byte, 2048)
	for {
		n, senderAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println(err)
			continue
		}

		data, err := packet.Decode(buf[:n])
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("recieved", string(data.Payload), "from", senderAddr)
	}
}
