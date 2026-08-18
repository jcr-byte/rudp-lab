package transport

import (
	"errors"
	"fmt"
	"net"
	"os"
	"slices"
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

type Stats struct {
	Elapsed         time.Duration
	RTTMin          time.Duration
	RTTMean         time.Duration
	RTTMax          time.Duration
	Packets         int
	Retransmissions int
	Bytes           int
}

func (s Stats) Goodput() float64 {
	return float64(s.Bytes) / s.Elapsed.Seconds()
}

const maxPayload = 1024

func Send(cfg SendConfig) (stats Stats, err error) {
	start := time.Now()
	var samples []time.Duration

	raddr, err := net.ResolveUDPAddr("udp", cfg.Addr)
	if err != nil {
		return stats, err
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return stats, err
	}
	defer conn.Close()

	// declare lossy connection for simulation
	lossyConn := netsim.NewLossyConn(conn, cfg.Loss, cfg.Seed)

	data := cfg.Payload
	var currentSeq uint16 = 1
	for offset := 0; offset < len(data); offset += maxPayload {
		end := min(offset+maxPayload, len(data))
		p := packet.Packet{Flag: packet.FlagData, Seq: currentSeq, Payload: data[offset:end]}
		attempts, rtt, err := sendReliably(conn, lossyConn, cfg, p)
		if err != nil {
			return stats, err
		}
		if attempts == 1 {
			samples = append(samples, rtt)
		}
		stats.Retransmissions += attempts - 1
		stats.Packets++
		currentSeq++
	}

	finPacket := packet.Packet{Flag: packet.FlagFin, Seq: currentSeq}
	attempts, rtt, err := sendReliably(conn, lossyConn, cfg, finPacket)
	if err == nil && attempts == 1 {
		samples = append(samples, rtt)
	}
	stats.Retransmissions += attempts - 1
	stats.Packets++
	if err != nil {
		fmt.Fprintln(os.Stderr, "warning: transfer complete but close unconfirmed:", err)
	}

	stats.Bytes = len(cfg.Payload)
	stats.Elapsed = time.Since(start)
	if len(samples) > 0 {
		var sum time.Duration
		for _, s := range samples {
			sum += s
		}
		stats.RTTMean = sum / time.Duration(len(samples))
		stats.RTTMin = slices.Min(samples)
		stats.RTTMax = slices.Max(samples)
	}
	return stats, nil
}

func sendReliably(conn *net.UDPConn, lossyConn *netsim.LossyConn, cfg SendConfig, p packet.Packet) (attempts int, rtt time.Duration, err error) {
	buf := make([]byte, 2048)
	encoded := p.Encode()
	for attempt := 0; attempt < cfg.Retries; attempt++ {
		t0 := time.Now()
		_, err := lossyConn.Write(encoded)
		if err != nil {
			return attempt + 1, rtt, err
		}

		conn.SetReadDeadline(time.Now().Add(cfg.Timeout))

		n, err := conn.Read(buf)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				fmt.Println("no ack, retransmitting")
				continue
			}
			return attempt + 1, rtt, err
		}

		if !packet.Verify(buf[:n]) {
			fmt.Println("Received packet is corrupted")
			continue
		}

		decodedPacket, err := packet.Decode(buf[:n])
		if err != nil {
			return attempt + 1, rtt, err
		}

		if decodedPacket.Flag == packet.FlagAck && decodedPacket.Seq == p.Seq {
			fmt.Println("Ack arrived and is valid")
			rtt = time.Since(t0)
			return attempt + 1, rtt, err
		}
	}

	return cfg.Retries, rtt, fmt.Errorf("giving up on seq %d", p.Seq)
}
