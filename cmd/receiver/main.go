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

	var last uint16
	haveDelivered := false
	buf := make([]byte, 2048)
	for {
		n, senderAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if !packet.Verify(buf[:n]) {
			fmt.Println("Recieved corrupted packet")
			continue
		}

		data, err := packet.Decode(buf[:n])
		if err != nil {
			fmt.Println(err)
			continue
		}
		if data.Flag == packet.FlagData {
			if !(haveDelivered && data.Seq == last) {
				fmt.Println("recieved", string(data.Payload), "from", senderAddr)
				last = data.Seq
				haveDelivered = true
			}
			ackPacket := packet.Packet{Flag: packet.FlagAck, Seq: data.Seq, Checksum: 0}
			n, err = conn.WriteToUDP(ackPacket.Encode(), senderAddr)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}

	}
}
