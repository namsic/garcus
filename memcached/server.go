package memcached

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"strconv"
	"strings"
)

type Server struct {
	conn      net.Conn
	writeChan chan operation
	readChan  chan operation
}

func (self *Server) AsyncOperation() -> chan []byte {
	responseCh := make(chan []byte)

	return responseCh
}

func (self *Server) AsyncOperation() -> []byte {
	return <-self.AsyncOperation()
}

type operation struct {
	raw          []byte
	receiverChan chan<- []byte // FIXME: Use Response struct instead of slice to contain more information
}

var END_OF_RESPONSE map[string]struct{} = map[string]struct{}{
	"END\r\n":    {},
	"STORED\r\n": {},
}

func Connect(addr string) *Server {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}

	self := new(Server)
	self.conn = conn
	self.writeChan = make(chan operation)
	self.readChan = make(chan operation)
	go self.writeOperation()
	go self.readResponse()

	return self
}

func (self *Server) Close() {
	if self.conn != nil {
		self.conn.Close()
		self.conn = nil
	}
	if self.writeChan != nil {
		close(self.writeChan)
		self.writeChan = nil
	}
	if self.writeChan != nil {
		close(self.readChan)
		self.readChan = nil
	}
}

func (self *Server) Request(data []byte, receiver chan []byte) {
	self.writeChan <- operation{raw: data, receiverChan: receiver}
}

func (self *Server) writeOperation() {
	writer := bufio.NewWriter(self.conn)
	for op := range self.writeChan {
		_, err := writer.Write(op.raw)
		if err != nil {
			panic(err)
		}
		writer.Flush()
		self.readChan <- op
	}
}

func (self *Server) readResponse() {
	reader := bufio.NewReader(self.conn)
	for op := range self.readChan {
		response := []byte{}
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				panic(err)
			}
			response = append(response, line...)
			if bytes.HasPrefix(line, []byte("VALUE")) {
				valueLen, err := strconv.Atoi(strings.Split(strings.TrimRight(string(line), "\r\n"), " ")[3])
				if err != nil {
					panic(err)
				}
				buf := make([]byte, valueLen+2)
				_, err = io.ReadFull(reader, buf)
				if err != nil {
					panic(err)
				}
				response = append(response, buf...)
				continue
			}
			if _, ok := END_OF_RESPONSE[string(line)]; ok {
				break
			}
		}

		op.receiverChan <- response
	}
}
