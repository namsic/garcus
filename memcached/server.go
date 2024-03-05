package memcached

import (
	"bufio"
	"net"
	"os"
	"time"
)

// [Operator] implementation for single arcus-memcached node.
type server struct {
	address    string
	op2write   chan operation
	protocol   Protocol
	identifier []byte
}

func NewServer(addr string, protocol Protocol, identifier []byte) Operator {
	s := &server{
		address:    addr,
		op2write:   make(chan operation, 100),
		protocol:   protocol,
		identifier: identifier,
	}
	go s.io() // FIXME
	return s
}

func (c *server) ProcessCommand(command []byte) (<-chan []byte, <-chan error) {
	responseChan := make(chan []byte)
	errorChan := make(chan error)
	c.op2write <- operation{
		command:      command,
		responseChan: responseChan,
		errorChan:    errorChan,
	}
	return responseChan, errorChan
}

func (c *server) Identifier() []byte {
	return c.identifier
}

func (c *server) io() {
	conn, err := net.DialTimeout("tcp", c.address, time.Second)
	if err != nil {
		panic(err) // FIXME
	}

	op2read := make(chan operation, 100)
	failed := false
	for !failed {
		select {
		case op := <-c.op2write:
			if err := writeOperation(bufio.NewWriter(conn), op, op2read); err != nil {
				op.fail(err)
				close(op2read)
			}
		case op := <-op2read:
			if response, err := c.protocol.ParseResponse(bufio.NewReader(conn), op.command); err == nil {
				op.success(response)
			} else {
				op.fail(err)
				return
			}
		}
	}
	for op := range op2read {
		op.fail(os.ErrProcessDone)
	}
}

func writeOperation(writer *bufio.Writer, op operation, readChan chan<- operation) error {
	if _, err := writer.Write(op.command); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	readChan <- op
	return nil
}
