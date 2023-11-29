package memcached

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

const operationBufferLength = 100

type server struct {
	op2write chan operation
	op2read  chan operation
}

func Connect(address string) (Operator, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	s := server{
		op2write: make(chan operation, operationBufferLength),
		op2read:  make(chan operation, operationBufferLength),
	}
	go s.handleOperation(conn)
	return &s, nil
}

func (self *server) AsyncOperation(asciiCommand []byte) (<-chan []byte, <-chan error) {
	responseChan := make(chan []byte)
	errorChan := make(chan error)
	op := operation{
		asciiCommand: asciiCommand,
		responseChan: responseChan,
		errorChan:    errorChan,
	}
	if bytes.HasSuffix(asciiCommand, []byte("\r\n")) {
		self.op2write <- op
	} else {
		op.errorResponse(os.ErrInvalid)
	}
	return responseChan, errorChan
}

func (self *server) Operation(asciiCommand []byte) ([]byte, error) {
	responseChan, errorChan := self.AsyncOperation(asciiCommand)
	select {
	case response := <-responseChan:
		return response, nil
	case err := <-errorChan:
		return nil, err
	}
}

func (self *server) handleOperation(conn net.Conn) {
	defer conn.Close()
	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	for {
		select {
		case op := <-self.op2write:
			if err := self.writeOperation(writer, op); err != nil {
				op.errorResponse(err)
				return
			}
		case op := <-self.op2read:
			if response, err := self.readResponse(reader, op); err != nil {
				op.errorResponse(err)
				return
			} else {
				op.bytesResponse(response)
			}
		}
	}
}

func (self *server) writeOperation(writer *bufio.Writer, op operation) error {
	if _, err := writer.Write(op.asciiCommand); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	self.op2read <- op
	return nil
}

func (self *server) readResponse(reader *bufio.Reader, op operation) ([]byte, error) {
	response := []byte{}
	restResponse := 1

	for restResponse > 0 {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return response, err
		}
		response = append(response, line...)
		tokens := strings.Split(strings.TrimSpace(string(line)), " ")
		switch tokens[0] {
		case "STAT":
		case "RESPONSE":
			n, err := strconv.Atoi(tokens[1])
			if err != nil {
				return response, err
			}
			restResponse += n
		case "VALUE":
			command := strings.Split(string(op.asciiCommand), " ")[0]
			switch command {
			case "get", "mget":
				valueLen, err := strconv.Atoi(tokens[3])
				if err != nil {
					return response, err
				}
				buf := make([]byte, valueLen+2) // \r\n
				if _, err = io.ReadFull(reader, buf); err != nil {
					return response, err
				}
				response = append(response, buf...)
			case "lop", "sop", "mop", "bop":
				elemCount, err := strconv.Atoi(tokens[2])
				if err != nil {
					return response, err
				}
				for i := 0; i < elemCount; i++ {
					if command == "mop" || command == "bop" {
						token, err := reader.ReadBytes(' ')
						if err != nil {
							return response, err
						}
						response = append(response, token...)
					}
					token, err := reader.ReadBytes(' ')
					if err != nil {
						return response, err
					}
					response = append(response, token...)
					if strings.HasPrefix(string(token), "0x") {
						token, err = reader.ReadBytes(' ')
						if err != nil {
						}
						response = append(response, token...)
					}
					elemLen, err := strconv.Atoi(strings.TrimSpace(string(token)))
					if err != nil {
						return response, err
					}
					buf := make([]byte, elemLen+2) // \r\n
					if _, err = io.ReadFull(reader, buf); err != nil {
						return response, err
					}
					response = append(response, buf...)
				}
			}
		default:
			restResponse -= 1
		}
	}
	return response, nil
}
