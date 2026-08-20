package transport_test

import (
	"bytes"
	"fmt"
	"net"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jcr-byte/rudp-lab/internal/packet"
	"github.com/jcr-byte/rudp-lab/internal/transport"
)

func TestSendReceive(t *testing.T) {
	cases := []struct {
		name string
		size int
		loss float64
	}{
		{"clean", 5000, 0},
		{"loss", 5000, 0.1},
		{"tiny", 1, 0},
		{"exact_chunk", 1024, 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ready := make(chan net.Addr, 1)
			errCh := make(chan error, 1)
			var buf bytes.Buffer

			reccfg := transport.ReceiveConfig{
				Addr:   "127.0.0.1:0",
				Loss:   c.loss,
				Out:    &buf,
				Ready:  ready,
				Linger: 50 * time.Millisecond,
			}

			payload := make([]byte, c.size)
			for i := range payload {
				payload[i] = byte(i)
			}

			go func() { errCh <- transport.Receive(reccfg) }()
			var addr net.Addr
			select {
			case addr = <-ready:
			case err := <-errCh:
				t.Fatal("receiver failed to start:", err)
			}

			sendcfg := transport.DefaultSendConfig()
			sendcfg.Addr = addr.String()
			sendcfg.Payload = payload
			sendcfg.Loss = c.loss

			_, err := transport.Send(sendcfg)
			if err != nil {
				t.Fatal("send failed:", err)
			}

			if err := <-errCh; err != nil {
				t.Fatal("receive failed:", err)
			}

			if !bytes.Equal(payload, buf.Bytes()) {
				t.Errorf("mismatch: got %d bytes, want %d", buf.Len(), len(payload))
			}
		})
	}
}

func TestSendGivesUpWithoutAck(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	go func() {
		buf := make([]byte, 2048)
		for {
			if _, _, err := conn.ReadFromUDP(buf); err != nil {
				return
			}
		}
	}()

	cfg := transport.DefaultSendConfig()
	cfg.Addr = conn.LocalAddr().String()
	cfg.Payload = []byte("test payload")
	cfg.Timeout = 10 * time.Millisecond
	cfg.Retries = 3

	_, err = transport.Send(cfg)
	if err == nil {
		t.Fatal("expected sender to give up, got nil")
	}

	if !strings.Contains(err.Error(), "giving up on seq") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendFillsWindowBeforeAck(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	type receiverResult struct {
		sequences []uint16
		err       error
	}

	resultCh := make(chan receiverResult, 1)

	go func() {
		buf := make([]byte, 2048)
		sequences := make([]uint16, 0, 4)

		if err := conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond)); err != nil {
			resultCh <- receiverResult{err: err}
			return
		}

		var senderAddr *net.UDPAddr

		for len(sequences) < 4 {
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				resultCh <- receiverResult{
					sequences: sequences,
					err:       fmt.Errorf("reading window: %w", err),
				}
				return
			}

			if !packet.Verify(buf[:n]) {
				resultCh <- receiverResult{err: fmt.Errorf("invalid packet checksum")}
				return
			}

			p, err := packet.Decode(buf[:n])
			if err != nil {
				resultCh <- receiverResult{err: err}
				return
			}

			if p.Flag != packet.FlagData {
				resultCh <- receiverResult{
					err: fmt.Errorf("got packet flag %d, want data", p.Flag),
				}
				return
			}

			sequences = append(sequences, p.Seq)
			senderAddr = addr
		}

		ack := packet.Packet{
			Flag: packet.FlagAck,
			Seq:  5,
		}
		if _, err := conn.WriteToUDP(ack.Encode(), senderAddr); err != nil {
			resultCh <- receiverResult{sequences: sequences, err: err}
			return
		}

		n, senderAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			resultCh <- receiverResult{
				sequences: sequences,
				err:       fmt.Errorf("reading FIN: %w", err),
			}
			return
		}

		fin, err := packet.Decode(buf[:n])
		if err != nil {
			resultCh <- receiverResult{sequences: sequences, err: err}
			return
		}

		if fin.Flag != packet.FlagFin || fin.Seq != 5 {
			resultCh <- receiverResult{
				sequences: sequences,
				err:       fmt.Errorf("got flag=%d seq=%d, want FIN seq=5", fin.Flag, fin.Seq),
			}
			return
		}

		finAck := packet.Packet{
			Flag: packet.FlagAck,
			Seq:  fin.Seq + 1,
		}
		if _, err := conn.WriteToUDP(finAck.Encode(), senderAddr); err != nil {
			resultCh <- receiverResult{sequences: sequences, err: err}
			return
		}

		resultCh <- receiverResult{sequences: sequences}
	}()

	cfg := transport.DefaultSendConfig()
	cfg.Addr = conn.LocalAddr().String()
	cfg.Payload = make([]byte, 4*1024)
	cfg.Timeout = time.Second
	cfg.Retries = 3
	cfg.WindowSize = 4

	if _, err = transport.Send(cfg); err != nil {
		t.Fatalf("send failed: %v", err)
	}

	result := <-resultCh
	if result.err != nil {
		t.Fatal(result.err)
	}

	want := []uint16{1, 2, 3, 4}
	if !slices.Equal(result.sequences, want) {
		t.Fatalf(
			"received sequences %v before ACK, want %v",
			result.sequences,
			want,
		)
	}
}

func TestSendRetransmitsWindowOnTimeout(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	type receiverResult struct {
		firstWindow         []uint16
		retransmittedWindow []uint16
		err                 error
	}

	resultCh := make(chan receiverResult, 1)

	go func() {
		buf := make([]byte, 2048)
		firstWindow := make([]uint16, 0, 4)

		if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			resultCh <- receiverResult{err: err}
			return
		}

		var senderAddr *net.UDPAddr

		for len(firstWindow) < 4 {
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				resultCh <- receiverResult{
					firstWindow: firstWindow,
					err:         fmt.Errorf("reading window: %w", err),
				}
				return
			}

			if !packet.Verify(buf[:n]) {
				resultCh <- receiverResult{err: fmt.Errorf("invalid packet checksum")}
				return
			}

			p, err := packet.Decode(buf[:n])
			if err != nil {
				resultCh <- receiverResult{err: err}
				return
			}

			if p.Flag != packet.FlagData {
				resultCh <- receiverResult{
					err: fmt.Errorf("got packet flag %d, want data", p.Flag),
				}
				return
			}

			firstWindow = append(firstWindow, p.Seq)
			senderAddr = addr
		}

		retransmittedWindow := make([]uint16, 0, 4)
		for len(retransmittedWindow) < 4 {
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				resultCh <- receiverResult{
					retransmittedWindow: retransmittedWindow,
					err:                 fmt.Errorf("reading window: %w", err),
				}
				return
			}

			if !packet.Verify(buf[:n]) {
				resultCh <- receiverResult{err: fmt.Errorf("invalid packet checksum")}
				return
			}

			p, err := packet.Decode(buf[:n])
			if err != nil {
				resultCh <- receiverResult{err: err}
				return
			}

			if p.Flag != packet.FlagData {
				resultCh <- receiverResult{
					err: fmt.Errorf("got packet flag %d, want data", p.Flag),
				}
				return
			}

			retransmittedWindow = append(retransmittedWindow, p.Seq)
			senderAddr = addr
		}

		ack := packet.Packet{
			Flag: packet.FlagAck,
			Seq:  5,
		}
		if _, err := conn.WriteToUDP(ack.Encode(), senderAddr); err != nil {
			resultCh <- receiverResult{retransmittedWindow: retransmittedWindow, err: err}
			return
		}

		n, senderAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			resultCh <- receiverResult{
				retransmittedWindow: retransmittedWindow,
				err:                 fmt.Errorf("reading FIN: %w", err),
			}
			return
		}

		fin, err := packet.Decode(buf[:n])
		if err != nil {
			resultCh <- receiverResult{retransmittedWindow: retransmittedWindow, err: err}
			return
		}

		if fin.Flag != packet.FlagFin || fin.Seq != 5 {
			resultCh <- receiverResult{
				retransmittedWindow: retransmittedWindow,
				err:                 fmt.Errorf("got flag=%d seq=%d, want FIN seq=5", fin.Flag, fin.Seq),
			}
			return
		}

		finAck := packet.Packet{
			Flag: packet.FlagAck,
			Seq:  fin.Seq + 1,
		}
		if _, err := conn.WriteToUDP(finAck.Encode(), senderAddr); err != nil {
			resultCh <- receiverResult{retransmittedWindow: retransmittedWindow, err: err}
			return
		}

		resultCh <- receiverResult{
			firstWindow:         firstWindow,
			retransmittedWindow: retransmittedWindow,
		}
	}()

	cfg := transport.DefaultSendConfig()
	cfg.Addr = conn.LocalAddr().String()
	cfg.Payload = make([]byte, 4*1024)
	cfg.Timeout = 100 * time.Millisecond
	cfg.Retries = 3
	cfg.WindowSize = 4

	stats, err := transport.Send(cfg)
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}

	if stats.Retransmissions != 4 {
		t.Fatalf("retransmissions = %d, want 4", stats.Retransmissions)
	}

	result := <-resultCh
	if result.err != nil {
		t.Fatal(result.err)
	}

	want := []uint16{1, 2, 3, 4}

	if !slices.Equal(result.firstWindow, want) {
		t.Fatalf(
			"received first window sequences %v before ACK, want %v",
			result.firstWindow,
			want,
		)
	}

	if !slices.Equal(result.retransmittedWindow, want) {
		t.Fatalf(
			"received retransmitted sequences %v before ACK, want %v",
			result.retransmittedWindow,
			want,
		)
	}
}

func TestReceiveDiscardsOutOfOrderPacket(t *testing.T) {
	ready := make(chan net.Addr, 1)
	errCh := make(chan error, 1)
	var output bytes.Buffer

	cfg := transport.ReceiveConfig{
		Addr:   "127.0.0.1:0",
		Out:    &output,
		Ready:  ready,
		Linger: 50 * time.Millisecond,
	}

	go func() {
		errCh <- transport.Receive(cfg)
	}()

	var receiverAddr net.Addr
	select {
	case receiverAddr = <-ready:
	case err := <-errCh:
		t.Fatalf("receiver failed to start: %v", err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", receiverAddr.String())
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	sendAndExpectAck := func(p packet.Packet, wantAck uint16) {
		t.Helper()

		if _, err := conn.Write(p.Encode()); err != nil {
			t.Fatalf("sending seq %d: %v", p.Seq, err)
		}

		if err := conn.SetReadDeadline(
			time.Now().Add(time.Second),
		); err != nil {
			t.Fatal(err)
		}

		buf := make([]byte, 2048)
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("reading ACK checksum for seq %d: %v", p.Seq, err)
		}

		if !packet.Verify(buf[:n]) {
			t.Fatalf("invalid ACK checksum for seq %d", p.Seq)
		}

		ack, err := packet.Decode(buf[:n])
		if err != nil {
			t.Fatal(err)
		}

		if ack.Flag != packet.FlagAck || ack.Seq != wantAck {
			t.Fatalf(
				"after sending seq %d, got flag=%d ACK=%d, want ACK=%d",
				p.Seq,
				ack.Flag,
				ack.Seq,
				wantAck,
			)
		}
	}

	sendAndExpectAck(packet.Packet{
		Flag:    packet.FlagData,
		Seq:     2,
		Payload: []byte("second"),
	}, 1)

	sendAndExpectAck(packet.Packet{
		Flag:    packet.FlagData,
		Seq:     1,
		Payload: []byte("first"),
	}, 2)

	sendAndExpectAck(packet.Packet{
		Flag:    packet.FlagData,
		Seq:     2,
		Payload: []byte("second"),
	}, 3)

	sendAndExpectAck(packet.Packet{
		Flag: packet.FlagFin,
		Seq:  3,
	}, 4)

	if err := <-errCh; err != nil {
		t.Fatalf("receiver failed: %v", err)
	}

	want := []byte("firstsecond")
	if !bytes.Equal(output.Bytes(), want) {
		t.Fatalf("received payload %q, want %q", output.Bytes(), want)
	}

}
