package main

import (
	"log"
	"time"

	"github.com/jcr-byte/rudp-lab/internal/transport"
)

func main() {
	cfg := transport.SendConfig{
		Addr:    "127.0.0.1:9000",
		Msg:     "hello world",
		Loss:    0,
		Seed:    2,
		Timeout: 500 * time.Millisecond,
		Retries: 10,
	}
	if err := transport.Send(cfg); err != nil {
		log.Fatal(err)
	}
}
