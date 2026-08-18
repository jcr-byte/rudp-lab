package transport_test

import (
	"bytes"
	"net"
	"testing"
	"time"

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
