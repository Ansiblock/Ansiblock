package replication

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/network"
)

func contains(node *NodeData, s1 []*NodeData) bool {
	for _, n := range s1 {
		if n.Equals(node) {
			return true
		}
	}
	return false
}

func compareNodeSets(s1 []*NodeData, s2 []*NodeData) bool {
	for _, node1 := range s1 {
		if !contains(node1, s2) {
			return false
		}
	}
	for _, node2 := range s2 {
		if !contains(node2, s1) {
			return false
		}
	}
	return true
}

func TestSyncGenerator(t *testing.T) {
	self := block.NewKeyPair().Public
	node := NewNodeData(self, "producer", "test1", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	sync, _ := NewSync(node)
	nodes := make([]*NodeData, 10)
	for i := uint64(0); i < 10; i++ {
		npk := block.NewKeyPair().Public
		nodes[i] = NewNodeData(npk, "signer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
		nodes[i].Version = i
		producer := block.NewKeyPair().Public
		nodes[i].Producer = producer
		sync.Insert(nodes[i])
	}

	out := SyncGenerator(sync, 1*time.Microsecond)
	res := <-out
	fmt.Printf("From %v\n\n", res.Data[:res.Size])

	if res.Data[0] != getUpdatesType {
		t.Error("SyncGenerator wrong sync")
	}
}
func TestSyncListener(t *testing.T) {
	self := block.NewKeyPair().Public
	node := NewNodeData(self, "producer", "test1", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	sync, _ := NewSync(node)
	nodes := make([]*NodeData, 10)
	for i := uint64(0); i < 10; i++ {
		npk := block.NewKeyPair().Public
		nodes[i] = NewNodeData(npk, "signer", "test2", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
		nodes[i].Version = i
		producer := block.NewKeyPair().Public
		nodes[i].Producer = producer
		sync.Insert(nodes[i])
	}
	input := make(chan *network.Blobs)
	out := SyncListener(sync, input)
	syncs := SyncGenerator(sync, 1*time.Microsecond)
	syncMessage := <-syncs
	syncMessageBlob := network.Blobs{Bs: []network.Blob{*syncMessage}}
	input <- &syncMessageBlob
	res1 := <-out
	fmt.Printf("len res1  - %v", len(res1.Bs))
	res := &res1.Bs[0]
	if res.Data[0] != updatesType {
		fmt.Println(res.Size)
		t.Error("SyncGenerator wrong type")
	}
	updates := new(Updates)
	updates.Deserialize(res)
	if !bytes.Equal(updates.From, self) {
		t.Error("SyncGenerator wrong from")
	}
	input <- res1
}
