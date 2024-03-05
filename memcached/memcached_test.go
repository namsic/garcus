package memcached_test

import (
	"fmt"
	"testing"

	"github.com/namsic/arcus-go/memcached"
)

func ExampleOperation_Async() {
	responseChan, errorChan := memcached.Operation{Operator: o, Command: []byte("set arcus-go:example 0 10 5\r\narcus\r\n")}.Async()
	select {
	case response := <-responseChan:
		fmt.Printf("%s", response)
	case err := <-errorChan:
		fmt.Printf("Failed to Operation.Async(): %v", err)
	}
	// Output: STORED
}

func ExampleOperation_Result() {
	response, err := memcached.Operation{Operator: o, Command: []byte("set arcus-go:example 0 10 5\r\narcus\r\n")}.Result()
	if err != nil {
		fmt.Printf("Failed to Operation.Result(): %v", err)
	} else {
		fmt.Printf("%s", response)
	}
	// Output: STORED
}

var o memcached.Operator

func TestMain(m *testing.M) {
	o = memcached.NewServer("127.0.0.1:11211", memcached.ProtocolAscii(), nil)
	m.Run()
}
