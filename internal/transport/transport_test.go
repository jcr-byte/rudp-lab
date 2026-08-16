package transport_test

import (
	"bytes"
	"net"
	"testing"

	"github.com/jcr-byte/rudp-lab/internal/transport"
)

func TestSendReceive(t *testing.T) {
	ready := make(chan net.Addr, 1)
	errCh := make(chan error, 1)
	var buf bytes.Buffer

	reccfg := transport.ReceiveConfig{
		Addr:  "127.0.0.1:0",
		Out:   &buf,
		Ready: ready,
	}

	payload := make([]byte, 5000)
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

	if err := transport.Send(sendcfg); err != nil {
		t.Fatal("send failed:", err)
	}

	if err := <-errCh; err != nil {
		t.Fatal("receive failed:", err)
	}

	if !bytes.Equal(payload, buf.Bytes()) {
		t.Errorf("mismatch: got %d bytes, want %d", buf.Len(), len(payload))
	}
}
