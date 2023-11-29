package main

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/namsic/garcus/memcached"
)

func clientErrorResponse(conn net.Conn) error {
	_, err := conn.Write([]byte("CLIENT_ERROR"))
	return err
}

func handler(conn net.Conn) {
	defer conn.Close()
	connCh := make(chan []byte)

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	for {
		raw, err := rw.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		tokens := bytes.Split(bytes.TrimRight(raw, "\r\n"), []byte(" "))
		switch string(tokens[0]) {
		case "set":
			valueLen, err := strconv.Atoi(string(tokens[4]))
			if err != nil {
				clientErrorResponse(conn)
				continue
			}
			buf := make([]byte, valueLen+2)
			_, err = io.ReadFull(rw, buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				panic(err)
			}
			raw = append(raw, buf...)
			node := memcachedCluster.FindNodeByKey(tokens[1])
			node.Request(raw, connCh)
			response := <-connCh
			_, err = rw.Write(response)
			if err != nil {
				if err == io.EOF {
					break
				}
				panic(err)
			}
		case "get":
			node := memcachedCluster.FindNodeByKey(tokens[1])
			node.Request(raw, connCh)
			response := <-connCh
			_, err = rw.Write(response)
			if err != nil {
				if err == io.EOF {
					break
				}
				panic(err)
			}
		}
	}
}

var memcachedCluster *memcached.Cluster = nil

func main() {
	cluster := flag.String("cluster", "", "zkConnectString/serviceCode")
	flag.Parse()
	c := strings.Split(*cluster, "/")
	if len(c) != 2 {
		panic("--help")
	}

	zkConn, _, err := zk.Connect(strings.Split(c[0], ","), time.Second)
	if err != nil {
		panic(err)
	}
	memcachedCluster = memcached.NewCluster(zkConn, c[1])

	listener, err := net.Listen("tcp", ":22122")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go handler(conn)
	}

}
