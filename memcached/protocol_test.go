package memcached_test

import (
	"testing"

	"github.com/namsic/arcus-go/memcached"
)

func TestAsciiKV(t *testing.T) {
	response, err := memcached.Operation{Operator: o, Command: []byte("delete arcus-go:kv01\r\n")}.Result()
	if err != nil {
		t.Error(err)
	} else if string(response) != "DELETED\r\n" && string(response) != "NOT_FOUND\r\n" {
		t.Errorf("\n%s", response)
	}

	response, err = memcached.Operation{Operator: o, Command: []byte("set arcus-go:kv01 0 0 1\r\n9\r\n")}.Result()
	if err != nil {
		t.Error(err)
	} else if string(response) != "STORED\r\n" {
		t.Errorf("\n%s", response)
	}

	response, err = memcached.Operation{Operator: o, Command: []byte("get arcus-go:kv01\r\n")}.Result()
	if err != nil {
		t.Error(err)
	} else if string(response) != "VALUE arcus-go:kv01 0 1\r\n9\r\nEND\r\n" {
		t.Errorf("\n%s", response)
	}

	response, err = memcached.Operation{Operator: o, Command: []byte("delete arcus-go:kv01\r\n")}.Result()
	if err != nil {
		t.Error(err)
	} else if string(response) != "DELETED\r\n" {
		t.Errorf("\n%s", response)
	}
}

func TestAsciiList(t *testing.T) {
	response, err := memcached.Operation{Operator: o, Command: []byte("delete arcus-go:list01\r\n")}.Result()
	if err != nil {
		t.Error(err)
	} else if string(response) != "DELETED\r\n" && string(response) != "NOT_FOUND\r\n" {
		t.Errorf("%s", response)
	}

	response, err = memcached.Operation{Operator: o, Command: []byte(
		"lop insert arcus-go:list01 0 5 create 0 0 100 pipe\r\nelem1\r\n" +
			"lop insert arcus-go:list01 1 5 create 0 0 100\r\nelem2\r\n")}.Result()
	if err != nil {
		t.Error(err)
	} else if string(response) != "RESPONSE 2\r\nCREATED_STORED\r\nSTORED\r\nEND\r\n" {
		t.Errorf("\n%s", response)
	}

	response, err = memcached.Operation{Operator: o, Command: []byte("lop get arcus-go:list01 0..5\r\n")}.Result()
	if err != nil {
		t.Error(err)
	} else if string(response) != "VALUE 0 2\r\n5 elem1\r\n5 elem2\r\nEND\r\n" {
		t.Errorf("\n%s", response)
	}

	response, err = memcached.Operation{Operator: o, Command: []byte("delete arcus-go:list01\r\n")}.Result()
	if err != nil {
		t.Error(err)
	} else if string(response) != "DELETED\r\n" {
		t.Errorf("\n%s", response)
	}
}

func TestAsciiBTree(t *testing.T) {
	response, err := memcached.Operation{Operator: o, Command: []byte("delete arcus-go:btree01\r\n")}.Result()
	if err != nil {
		t.Error(err)
	} else if string(response) != "DELETED\r\n" && string(response) != "NOT_FOUND\r\n" {
		t.Errorf("%s", response)
	}

	response, err = memcached.Operation{Operator: o, Command: []byte(
		"bop insert arcus-go:btree01 10 3 create 0 0 100 pipe\r\nv10\r\n" +
			"bop insert arcus-go:btree01 20 3 create 0 0 100\r\nv20\r\n")}.Result()
	if err != nil {
		t.Error(err)
	} else if string(response) != "RESPONSE 2\r\nCREATED_STORED\r\nSTORED\r\nEND\r\n" {
		t.Errorf("\n%s", response)
	}

	response, err = memcached.Operation{Operator: o, Command: []byte("bop insert arcus-go:btree01 30 0x01 3 create 0 0 100\r\nv30\r\n")}.Result()
	if err != nil {
		t.Error(err)
	} else if string(response) != "STORED\r\n" {
		t.Errorf("\n%s", response)
	}

	response, err = memcached.Operation{Operator: o, Command: []byte("bop get arcus-go:btree01 0..50\r\n")}.Result()
	if err != nil {
		t.Error(err)
	} else if string(response) != "VALUE 0 3\r\n10 3 v10\r\n20 3 v20\r\n30 0x01 3 v30\r\nEND\r\n" {
		t.Errorf("\n%s", response)
	}

	response, err = memcached.Operation{Operator: o, Command: []byte("bop get arcus-go:btree01 0..50 0 EQ 0x01\r\n")}.Result()
	if err != nil {
		t.Error(err)
	} else if string(response) != "VALUE 0 1\r\n30 0x01 3 v30\r\nEND\r\n" {
		t.Errorf("\n%s", response)
	}

	response, err = memcached.Operation{Operator: o, Command: []byte("delete arcus-go:btree01\r\n")}.Result()
	if err != nil {
		t.Error(err)
	} else if string(response) != "DELETED\r\n" {
		t.Errorf("\n%s", response)
	}
}
