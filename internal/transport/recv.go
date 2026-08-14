package transport

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/jcr-byte/rudp-lab/internal/netsim"
	"github.com/jcr-byte/rudp-lab/internal/packet"
)

type ReceiveConfig struct {
	Addr string
	Loss float64
	Seed int64
	Out  io.Writer
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
			fmt.Fprintln(os.Stderr, "Received corrupted packet")
			continue
		}

		data, err := packet.Decode(buf[:n])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		if data.Flag == packet.FlagData {
			if !(haveDelivered && data.Seq == last) {
				fmt.Fprintln(os.Stderr, "received payload of length", len(data.Payload), "from", senderAddr)
				last = data.Seq
				haveDelivered = true
				if _, err := cfg.Out.Write(data.Payload); err != nil {
					return err
				}
			}
			ackPacket := packet.Packet{Flag: packet.FlagAck, Seq: data.Seq}
			n, err = lossyConn.WriteToUDP(ackPacket.Encode(), senderAddr)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
		}

		if data.Flag == packet.FlagFin {
			fmt.Fprintln(os.Stderr, "received fin packet from", senderAddr)

			ackPacket := packet.Packet{Flag: packet.FlagAck, Seq: data.Seq}
			_, err = lossyConn.WriteToUDP(ackPacket.Encode(), senderAddr)
			if err != nil {
				return err
			}
			return nil
		}
	}
}
