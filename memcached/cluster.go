package memcached

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/go-zookeeper/zk"
)

type Cluster struct {
	servers      map[string]*Server
	serversMutex sync.RWMutex
	points       []point
	pointsMutex  sync.RWMutex
}

func NewCluster(zkConn *zk.Conn, serviceCode string) *Cluster {
	cluster := new(Cluster)
	cluster.serversMutex = sync.RWMutex{}
	cluster.pointsMutex = sync.RWMutex{}
	go cluster.updateCacheList(zkConn, serviceCode)
	return cluster
}

func (self *Cluster) FindNodeByKey(key []byte) *Server {
	digest := md5.Sum(key)

	keyHash := bytes2uint32([4]byte(digest[:4]))

	return self.findNodeByHash(keyHash)
}

type point struct {
	hashValue uint32
	node      *Server
}

func (self *Cluster) findNodeByHash(hashValue uint32) *Server {
	self.pointsMutex.RLock()
	defer self.pointsMutex.RUnlock()
	low, mid, high := 0, 0, len(self.points)-1
	for low <= high {
		mid = (low + high) / 2
		if hashValue <= self.points[mid].hashValue {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	if low == len(self.points) {
		low = 0
	}
	return self.points[low].node
}

func (self *Cluster) updateCacheList(zkConn *zk.Conn, serviceCode string) {
	for {
		cacheListZnodes, _, eventChan, err := zkConn.ChildrenW("/arcus/cache_list/" + serviceCode)
		if err != nil {
			panic(err)
		}
		newServers := map[string]*Server{}
		newPoints := []point{}

		sort.Strings(cacheListZnodes)
		for _, cacheListZnode := range cacheListZnodes {
			addr, _, _ := strings.Cut(cacheListZnode, "-")
			server := Connect(addr)
			newServers[addr] = server
			for i := 0; i < 40; i++ {
				digest := md5.Sum([]byte(fmt.Sprintf("%v-%v", addr, i)))
				insertHashPoint(&newPoints, bytes2uint32([4]byte(digest[:4])), server)
				insertHashPoint(&newPoints, bytes2uint32([4]byte(digest[4:8])), server)
				insertHashPoint(&newPoints, bytes2uint32([4]byte(digest[8:12])), server)
				insertHashPoint(&newPoints, bytes2uint32([4]byte(digest[12:])), server)
			}
		}

		self.serversMutex.Lock()
		self.servers = newServers
		self.serversMutex.Unlock()
		self.pointsMutex.Lock()
		self.points = newPoints
		self.pointsMutex.Unlock()

		for event := range eventChan {
			if event.Err != nil {
				panic(event.Err)
			}
		}
	}
}

func bytes2uint32(b [4]byte) uint32 {
	u := uint32(b[3])
	u = (u << 8) | uint32(b[2])
	u = (u << 8) | uint32(b[1])
	u = (u << 8) | uint32(b[0])
	return u
}

func insertHashPoint(points *[]point, hashValue uint32, node *Server) {
	for i, point := range *points {
		if hashValue == point.hashValue {
			// Duplicate points.
			return
		}
		if hashValue < point.hashValue {
			*points = append((*points)[:i+1], (*points)[i:]...)
			(*points)[i].hashValue = hashValue
			(*points)[i].node = node
			return
		}
	}
	*points = append(*points, point{hashValue: hashValue, node: node})
}
