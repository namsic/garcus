package memcached

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
)

type Protocol interface {
	ParseCommand(*bufio.Reader) ([]byte, error)
	ParseResponse(*bufio.Reader, []byte) ([]byte, error)
	GetKey([]byte) ([]byte, error)
}

func ProtocolAscii() Protocol {
	return protocolAscii{}
}

type protocolAscii struct{}

func (p protocolAscii) GetKey(command []byte) ([]byte, error) {
	header, _, _ := bytes.Cut(command, []byte("\n"))
	tokens := bytes.Split(bytes.TrimSuffix(header, []byte("\r")), []byte(" "))
	switch string(tokens[0]) {
	case "set", "get":
		if len(tokens) < 2 {
			return nil, os.ErrInvalid
		}
		return tokens[1], nil
	case "lop", "sop", "mop", "bop":
		if len(tokens) < 3 {
			return nil, os.ErrInvalid
		}
		return tokens[2], nil
	default:
		return nil, nil
	}
}

func (p protocolAscii) ParseCommand(reader *bufio.Reader) ([]byte, error) {
	// TODO
	return nil, errors.ErrUnsupported
}

func (p protocolAscii) ParseResponse(reader *bufio.Reader, command []byte) ([]byte, error) {
	response := []byte{}
	restResponse := 1

	for restResponse > 0 {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return response, err
		}
		response = append(response, line...)
		tokens := strings.Split(strings.TrimSpace(string(line)), " ")
		cmd, _, _ := strings.Cut(string(command), " ")
		switch tokens[0] {
		case "STAT":
		case "RESPONSE":
			n, err := strconv.Atoi(tokens[1])
			if err != nil {
				return response, err
			}
			restResponse += n
		case "VALUE":
			switch cmd {
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
					if cmd == "mop" || cmd == "bop" {
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
							return response, err
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
