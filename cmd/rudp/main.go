package main

import (
	"fmt"
	"os"
	"flag"	
	"log"

	"github.com/jcr-byte/rudp-lab/internal/transport"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "send":
	case "receive":
		err := runReceive(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
			
	default: 
		usage()
	}
}

func runReceive(args []string) error {

	fs := flag.NewFlagSet("recv", flag.ExitOnError)

	var cfg transport.ReceiveConfig
	
	fs.StringVar(&cfg.Addr, "addr", ":9000", "address to listen on, as host:port")
	fs.Float64Var(&cfg.Loss, "loss", 0, "fraction of packets to drop, 0.0-1.0")
	fs.Int64Var(&cfg.Seed, "seed", 2, "RNG seed for loss simulation")

	fs.Parse(args)

	if cfg.Loss < 0.0 || cfg.Loss > 1.0 {
		return fmt.Errorf("loss must be between 0 and 1. Got %v", cfg.Loss)
	}

	return transport.Receive(cfg)
}	

func usage() {
	fmt.Fprint(os.Stderr, `usage: rudp <command> [flags]
Commands:
  send    Sends a message reliably over UDP
  recv    Listen for and receive messages

Run "rudp <command> -h" for flag details.
`)

	os.Exit(2)
}

