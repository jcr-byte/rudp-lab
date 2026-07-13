package netsim

import (
	"math/rand"
	"net"
	"fmt"
)

type LossyConn struct {
	conn *net.UDPConn
	loss float64
	rng *rand.Rand
}

func NewLossyConn(conn *net.UDPConn, loss float64, seed int64) *LossyConn {
	return &LossyConn{
		conn: conn,
		loss: loss,
		rng: rand.New(rand.NewSource(seed)),
	}
}

func (connection *LossyConn) Write(bytes []byte) (int, error) {
	if connection.rng.Float64() <= connection.loss {
		fmt.Printf("DROP %d bytes\n", len(bytes))
		return len(bytes), nil
	} else {
		return connection.conn.Write(bytes)
	}
}

func (connection *LossyConn) WriteToUDP(bytes []byte, addr *net.UDPAddr) (int, error) {
	if connection.rng.Float64() <= connection.loss {
		fmt.Printf("DROP %d bytes\n", len(bytes))
		return len(bytes), nil
	} else {
		return connection.conn.WriteToUDP(bytes, addr)
	}
}