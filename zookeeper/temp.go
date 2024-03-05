package zookeeper

import (
	"fmt"
	"runtime"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/namsic/arcus-go/pkg/internal"
	"github.com/namsic/arcus-go/pkg/memcached"
)

type Ensemble struct {
	conn     *zk.Conn
	clusters map[string][]serverInfo
}

type serverInfo struct {
	ip       string
	port     string
	hostname string
	group    string
	conn     memcached.Operator
}

func (e *Ensemble) CreateEphemeralZnode(serviceCode string) {
	info := ""
	info += fmt.Sprintf("%v\t%v\n", "runtimeVer", runtime.Version())
	info += fmt.Sprintf("%v\t%v/%v\n", "runtimeArch", runtime.GOOS, runtime.GOARCH)

	e.conn.Create(
		fmt.Sprintf("/arcus/client_list/%v/%v_%v_%v_go_%v_%v_%x",
			serviceCode, "127.0.0.1", "localhost", 1, internal.VERSION, time.Now().Format("20060102150405"), e.conn.SessionID()),
		[]byte(info), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
}

func ConnectEnsemble(address string) Ensemble {
	addr := []string{"127.0.0.1:2181"}
	conn, watcher, err := zk.Connect(addr, time.Second*30)
	if err != nil {
		panic(err)
	}
	for event := range watcher {
		fmt.Println(event)
		if event.State == zk.StateHasSession {
			break
		}
	}
	return Ensemble{conn: conn, clusters: map[string][]serverInfo{}}
}

func Watcher(conn *zk.Conn) {
	watcher := printChild(conn)
	for {
		event := <-watcher
		if event.Err != nil {
			panic(event.Err)
		}
		watcher = printChild(conn)
	}
}

func printChild(conn *zk.Conn) <-chan zk.Event {
	child, _, watcher, err := conn.ChildrenW("/namsic")
	if err != nil {
		panic(err)
	}
	fmt.Println(child)
	return watcher
}

func (e *Ensemble) UpdateCacheServerMapping(replication bool) {
	prefix := "/arcus"
	if replication {
		prefix += "_repl"
	}
	prefix += "/cache_server_mapping"

	addrs, _, err := e.conn.Children(prefix)
	if err != nil {
		panic(err)
	}
	for _, addr := range addrs {
		e.conn.Children(prefix + "/" + addr)
	}
}
