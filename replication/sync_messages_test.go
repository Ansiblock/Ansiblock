package replication

import (
	"bytes"
	"testing"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/network"
)

func TestGetUpdates(t *testing.T) {
	self := block.NewKeyPair().Public
	node := NewNodeData(self, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	ru := GetUpdates{LastUpdateIndex: 10, MyInfo: node}
	blob := ru.Serialize()

	req := new(GetUpdates)
	req.Deserialize(blob)

	if req.LastUpdateIndex != ru.LastUpdateIndex {
		t.Errorf("GetUpdatesSerialize: %v != %v\n", req.LastUpdateIndex, ru.LastUpdateIndex)
	}
	if !req.MyInfo.Equals(ru.MyInfo) {
		t.Errorf("GetUpdatesSerialize: %v\n!=\n%v\n", req.MyInfo, ru.MyInfo)

	}
}

func TestUpdates(t *testing.T) {
	self := block.NewKeyPair().Public
	nodes := make([]*NodeData, 10)
	for i := uint64(0); i < 10; i++ {
		npk := block.NewKeyPair().Public
		nodes[i] = NewNodeData(npk, "signer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
		nodes[i].Version = i
		producer := block.NewKeyPair().Public
		nodes[i].Producer = producer
	}

	ru := Updates{From: self, LastUpdateIndex: 10000, Updates: nodes}
	blob := ru.Serialize()

	rec := new(Updates)
	rec.Deserialize(blob)

	if !bytes.Equal(ru.From, rec.From) || ru.LastUpdateIndex != rec.LastUpdateIndex {
		t.Errorf("Updates: %v != %v\n", ru.LastUpdateIndex, rec.LastUpdateIndex)
	}
	for i := 0; i < 10; i++ {
		if !ru.Updates[i].Equals(rec.Updates[i]) {
			t.Errorf("Updates: Error in updates slice")
		}
	}
}
