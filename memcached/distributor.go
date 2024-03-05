package memcached

import (
	"crypto/md5"
	"fmt"
	"sort"
)

type Distributor interface {
	Distribute([]byte) Operator
}

type distributionKetamaSpy struct {
	hashPoints map[uint32]Operator
	sortedHash []uint32
}

func DistributionKetamaSpy(operators []Operator) Distributor {
	k := distributionKetamaSpy{
		hashPoints: map[uint32]Operator{},
		sortedHash: []uint32{},
	}

	for _, o := range operators {
		for i := 0; i < 40; i++ {
			digest := md5.Sum([]byte(fmt.Sprintf("%s-%v", o.Identifier(), i)))

			hashPoint := bytes2uint32([4]byte{digest[0], digest[1], digest[2], digest[3]})
			if _, exists := k.hashPoints[hashPoint]; !exists {
				k.hashPoints[hashPoint] = o
				k.sortedHash = append(k.sortedHash, hashPoint)
			}
			hashPoint = bytes2uint32([4]byte{digest[4], digest[5], digest[6], digest[7]})
			if _, exists := k.hashPoints[hashPoint]; !exists {
				k.hashPoints[hashPoint] = o
				k.sortedHash = append(k.sortedHash, hashPoint)
			}
			hashPoint = bytes2uint32([4]byte{digest[8], digest[9], digest[10], digest[11]})
			if _, exists := k.hashPoints[hashPoint]; !exists {
				k.hashPoints[hashPoint] = o
				k.sortedHash = append(k.sortedHash, hashPoint)
			}
			hashPoint = bytes2uint32([4]byte{digest[12], digest[13], digest[14], digest[15]})
			if _, exists := k.hashPoints[hashPoint]; !exists {
				k.hashPoints[hashPoint] = o
				k.sortedHash = append(k.sortedHash, hashPoint)
			}
		}
		sort.Slice(k.sortedHash, func(i, j int) bool { return k.sortedHash[i] < k.sortedHash[j] })
	}
	return &k
}

func (k *distributionKetamaSpy) Distribute(key []byte) Operator {
	digest := md5.Sum(key)
	hashValue := bytes2uint32([4]byte{digest[0], digest[1], digest[2], digest[3]})

	low, mid, high := 0, 0, len(k.sortedHash)-1
	for low <= high {
		mid = (low + high) / 2
		if hashValue <= k.sortedHash[mid] {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	if low == len(k.sortedHash) {
		low = 0
	}
	return k.hashPoints[k.sortedHash[low]]
}

func bytes2uint32(b [4]byte) uint32 {
	u := uint32(b[3])
	u = (u << 8) | uint32(b[2])
	u = (u << 8) | uint32(b[1])
	u = (u << 8) | uint32(b[0])
	return u
}
