package zookeeper_test

import (
	"testing"
	"time"

	"github.com/namsic/arcus-go/pkg/zookeeper"
)

func TestConnect(t *testing.T) {
	zookeeper.ConnectEnsemble("")
	return
	for {
		time.Sleep(time.Minute)
	}
}
