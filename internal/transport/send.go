package transport

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/jcr-byte/rudp-lab/internal/netsim"
	"github.com/jcr-byte/rudp-lab/internal/packet"
)

type SendConfig struct {
	Addr    string
	Payload []byte
	Loss    float64
	Seed    int64
	Timeout time.Duration
	Retries int
}

func DefaultSendConfig() SendConfig {
	cfg := SendConfig{
		Addr:    "127.0.0.1:9000",
		Payload: nil,
		Loss:    0,
		Seed:    2,
		Timeout: 500 * time.Millisecond,
		Retries: 5,
	}

	return cfg
}

const maxPayload = 1024

func Send(cfg SendConfig) error {
	raddr, err := net.ResolveUDPAddr("udp", cfg.Addr)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// declare lossy connection for simulation
	lossyConn := netsim.NewLossyConn(conn, cfg.Loss, cfg.Seed)

	data := cfg.Payload
	var currentSeq uint16 = 1
	for offset := 0; offset < len(data); offset += maxPayload {

		end := min(offset+maxPayload, len(data))
		p := packet.Packet{Flag: packet.FlagData, Seq: currentSeq, Payload: data[offset:end]}
		encoded := p.Encode()

		buf := make([]byte, 2048)
		acked := false
		for attempt := 0; attempt < cfg.Retries; attempt++ {
			_, err := lossyConn.Write(encoded)
			if err != nil {
				return err
			}

			conn.SetReadDeadline(time.Now().Add(cfg.Timeout))

			n, err := conn.Read(buf)
			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					fmt.Println("no ack, retransmitting")
					continue
				}
				return err
			}

			if !packet.Verify(buf[:n]) {
				fmt.Println("Received packet is corrupted")
				continue
			}

			decodedPacket, err := packet.Decode(buf[:n])
			if err != nil {
				return err
			}
			if decodedPacket.Flag == packet.FlagAck && decodedPacket.Seq == p.Seq {
				fmt.Println("Ack arrived and is valid")
				currentSeq++
				acked = true
				break
			}
		}

		if !acked {
			return fmt.Errorf("giving up on seq %d", p.Seq)
		}
	}
	return nil
}
