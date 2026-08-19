package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/jcr-byte/rudp-lab/internal/transport"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "send":
		err := runSend(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
	case "recv":
		err := runReceive(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}

	default:
		usage()
	}
}

func runSend(args []string) error {

	fs := flag.NewFlagSet("send", flag.ExitOnError)

	cfg := transport.DefaultSendConfig()

	var size int

	fs.StringVar(&cfg.Addr, "addr", cfg.Addr, "address to send to, as host:port")
	fs.Float64Var(&cfg.Loss, "loss", cfg.Loss, "fraction of packets to drop, 0.0-1.0")
	fs.Int64Var(&cfg.Seed, "seed", cfg.Seed, "RNG seed for reproducible loss simulation")
	fs.IntVar(&size, "size", 15000, "size of payload to send")

	fs.Parse(args)

	if cfg.Loss < 0.0 || cfg.Loss > 1.0 {
		return fmt.Errorf("loss must be between 0 and 1. Got %v", cfg.Loss)
	}

	if size <= 0 {
		return fmt.Errorf("size must be greater than 0. Got %d", size)
	}

	r := rand.New(rand.NewSource(0))
	cfg.Payload = make([]byte, size)
	r.Read(cfg.Payload)
	fmt.Fprintf(os.Stderr, "sending %d bytes, sha256=%x\n", size, sha256.Sum256(cfg.Payload))

	stats, err := transport.Send(cfg)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "sent %d bytes in %v, %d packets, %d retransmissions\n", stats.Bytes, stats.Elapsed, stats.Packets, stats.Retransmissions)
	fmt.Fprintf(os.Stderr, "RTT Min: %v, RTT Max: %v, RTT Mean: %v\n", stats.RTTMin, stats.RTTMax, stats.RTTMean)
	fmt.Fprintf(os.Stderr, "%.0f B/s goodput \n", stats.Goodput())
	return nil
}

func runReceive(args []string) error {

	fs := flag.NewFlagSet("recv", flag.ExitOnError)

	var cfg transport.ReceiveConfig

	fs.StringVar(&cfg.Addr, "addr", ":9000", "address to listen on, as host:port")
	fs.Float64Var(&cfg.Loss, "loss", 0, "fraction of packets to drop, 0.0-1.0")
	fs.Int64Var(&cfg.Seed, "seed", 2, "RNG seed for loss simulation")

	file, err := os.Create("received.bin")
	if err != nil {
		return err
	}
	defer file.Close()
	cfg.Out = file

	fs.Parse(args)

	if cfg.Loss < 0.0 || cfg.Loss > 1.0 {
		return fmt.Errorf("loss must be between 0 and 1. Got %v", cfg.Loss)
	}

	if err := transport.Receive(cfg); err != nil {
		return err
	}

	got, err := os.ReadFile("received.bin")
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "received %d bytes, sha256=%x\n", len(got), sha256.Sum256(got))

	return nil
}

func usage() {
	fmt.Fprint(os.Stderr, `usage: rudp <command> [flags]
Commands:
  send    Sends random bytes reliably over udp  
  recv    Listen for and receive messages

Run "rudp <command> -h" for flag details.
`)

	os.Exit(2)
}
