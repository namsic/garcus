package memcached_test

import (
	"testing"

	"github.com/namsic/arcus-go/internal/memcached"
)

func TestOperation(t *testing.T) {
	testCases := [][]string{
		{
			"set garcus:kv1 0 0 5\r\ndata1\r\n",
			"STORED\r\n",
		},
		{
			"get garcus:kv1\r\n",
			"VALUE garcus:kv1 0 5\r\ndata1\r\nEND\r\n",
		},
		{
			"lop insert garcus:list1 0 5 create 0 0 100 pipe\r\nelem1\r\nlop insert garcus:list1 1 5 create 0 0 100\r\nelem2\r\n",
			"RESPONSE 2\r\nCREATED_STORED\r\nSTORED\r\nEND\r\n",
		},
		{
			"lop get garcus:list1 0..5\r\n",
			"VALUE 0 2\r\n5 elem1\r\n5 elem2\r\nEND\r\n",
		},
		{
			"bop insert garcus:btree1 10 3 create 0 0 100 pipe\r\nv10\r\nbop insert garcus:btree1 20 3 create 0 0 100\r\nv20\r\n",
			"RESPONSE 2\r\nCREATED_STORED\r\nSTORED\r\nEND\r\n",
		},
		{
			"bop insert garcus:btree1 30 0x01 3 create 0 0 100\r\nv30\r\n",
			"STORED\r\n",
		},
		{
			"bop get garcus:btree1 0..50\r\n",
			"VALUE 0 3\r\n10 3 v10\r\n20 3 v20\r\n30 0x01 3 v30\r\nEND\r\n",
		},
		{
			"bop get garcus:btree1 0..50 0 EQ 0x01\r\n",
			"VALUE 0 1\r\n30 0x01 3 v30\r\nEND\r\n",
		},
	}

	server, err := memcached.Connect("127.0.0.1:11211")
	if err != nil {
		t.Errorf("Failed to connect: %v\n", err)
		return
	}

	if _, err := server.Operation([]byte("flush_prefix garcus\r\n")); err != nil {
		t.Errorf("\n[command]\nfulsh_prefix garcus\n%v", err)
		return
	}

	for _, tc := range testCases {
		command, expected := tc[0], tc[1]
		if response, err := server.Operation([]byte(command)); err != nil {
			t.Errorf("\n[command]\n%v\n%v", command, err)
		} else if string(response) != expected {
			t.Errorf("\n[command]\n%v\n[response]\n%s\n[expected]\n%v", command, response, expected)
		}
	}
}
