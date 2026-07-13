package main

import (
	"log"

	"github.com/jcr-byte/rudp-lab/internal/transport"
)

func main() {
	cfg := transport.ReceiveConfig{
		Addr: "127.0.0.1:9000",
		Loss: 0,
		Seed: 2,
	}
	if err := transport.Receive(cfg); err != nil {
		log.Fatal(err)
	}
}
