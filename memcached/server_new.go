package memcached

import (
	"fmt"
	"net"
	"sync/atomic"
)

type _command struct {
	raw      []byte
	response <-chan []byte
	err      <-chan error
}

type server_ struct {
	aliveConnection int32
	op2write        chan _command
}

func (s *server_) io() {
	_, err := net.Dial("tcp", "127.0.0.1:11211")
	if err != nil {
		return
	}
	fmt.Println("aliveConn:", atomic.AddInt32(&s.aliveConnection, 1))
	defer fmt.Println("aliveConn:", atomic.AddInt32(&s.aliveConnection, -1))

	op2read := make(chan _command, 100)
	defer close(op2read)
}
