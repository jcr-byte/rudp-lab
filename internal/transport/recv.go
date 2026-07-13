package transport

import (
	"fmt"
	"net"

	"github.com/jcr-byte/rudp-lab/internal/netsim"
	"github.com/jcr-byte/rudp-lab/internal/packet"
)

type ReceiveConfig struct {
	Addr string
	Loss float64
	Seed int64
}

func Receive(cfg ReceiveConfig) error {
	addr, err := net.ResolveUDPAddr("udp", cfg.Addr)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	lossyConn := netsim.NewLossyConn(conn, cfg.Loss, cfg.Seed)
	defer conn.Close()

	var last uint16
	haveDelivered := false
	buf := make([]byte, 2048)
	for {
		n, senderAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return err
		}

		if !packet.Verify(buf[:n]) {
			fmt.Println("Received corrupted packet")
			continue
		}

		data, err := packet.Decode(buf[:n])
		if err != nil {
			fmt.Println(err)
			continue
		}
		if data.Flag == packet.FlagData {
			if !(haveDelivered && data.Seq == last) {
				fmt.Println("received", string(data.Payload), "from", senderAddr)
				last = data.Seq
				haveDelivered = true
			}
			ackPacket := packet.Packet{Flag: packet.FlagAck, Seq: data.Seq, Checksum: 0}
			n, err = lossyConn.WriteToUDP(ackPacket.Encode(), senderAddr)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}
}
