package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	raddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:9000")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	n, err := conn.Write([]byte("Hello from sender"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("sent %d bytes\n", n)
}
