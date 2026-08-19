package transport

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/jcr-byte/rudp-lab/internal/netsim"
	"github.com/jcr-byte/rudp-lab/internal/packet"
)

type ReceiveConfig struct {
	Addr   string
	Loss   float64
	Seed   int64
	Out    io.Writer
	Ready  chan<- net.Addr
	Linger time.Duration
}

const defaultLinger = 2 * time.Second

func Receive(cfg ReceiveConfig) error {
	addr, err := net.ResolveUDPAddr("udp", cfg.Addr)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	if cfg.Ready != nil {
		cfg.Ready <- conn.LocalAddr()
	}

	lossyConn := netsim.NewLossyConn(conn, cfg.Loss, cfg.Seed)
	defer conn.Close()

	var nextExpectedSeq uint16 = 1
	buf := make([]byte, 2048)

	lingerFor := cfg.Linger
	if lingerFor <= 0 {
		lingerFor = defaultLinger
	}
	for {
		n, senderAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return err
		}

		if !packet.Verify(buf[:n]) {
			continue
		}

		data, err := packet.Decode(buf[:n])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		if data.Flag == packet.FlagData {
			if nextExpectedSeq == data.Seq {
				if _, err := cfg.Out.Write(data.Payload); err != nil {
					return err
				}
				nextExpectedSeq++
			}
			ackPacket := packet.Packet{Flag: packet.FlagAck, Seq: data.Seq}
			n, err = lossyConn.WriteToUDP(ackPacket.Encode(), senderAddr)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
		}

		if data.Flag == packet.FlagFin {

			ackPacket := packet.Packet{Flag: packet.FlagAck, Seq: data.Seq}
			_, err = lossyConn.WriteToUDP(ackPacket.Encode(), senderAddr)
			if err != nil {
				return err
			}

			return linger(conn, lossyConn, data.Seq, lingerFor)
		}
	}
}

func linger(conn *net.UDPConn, lossyConn *netsim.LossyConn, finSeq uint16, lingerDuration time.Duration) error {
	conn.SetReadDeadline(time.Now().Add(lingerDuration))
	buf := make([]byte, 2048)
	for {
		n, senderAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				return nil
			}
			return err
		}

		if !packet.Verify(buf[:n]) {
			continue
		}

		p, err := packet.Decode(buf[:n])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		if p.Flag == packet.FlagFin && p.Seq == finSeq {
			ackPacket := packet.Packet{Flag: packet.FlagAck, Seq: p.Seq}
			if _, err := lossyConn.WriteToUDP(ackPacket.Encode(), senderAddr); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}
}
