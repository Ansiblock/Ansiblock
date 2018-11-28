package replication

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/network"
)

func TestNewNodeData(t *testing.T) {
	self := block.NewKeyPair().Public
	node := NewNodeData(self, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	a := Addresses{network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP}
	if node.Version != 0 || !bytes.Equal(node.Self, self) {
		t.Errorf("NewNodeData wrong struct")
	}
	if !node.Addresses.Equals(&a) {
		t.Errorf("NewNodeData wrong struct")
	}
}

func TestNewSync(t *testing.T) {
	self := block.NewKeyPair().Public
	node := NewNodeData(self, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	sync, err := NewSync(node)
	if err != nil || reflect.DeepEqual(sync.me, *node) {
		t.Errorf("NewSync wrong struct")
	}
	node.Version = 1
	sync, err = NewSync(node)
	if err == nil {
		t.Errorf("NewSync wrong version")
	}

	if sync != nil {
		t.Errorf("Incorrect sync created")
	}

}

func TestInsert(t *testing.T) {
	self := block.NewKeyPair().Public
	node := NewNodeData(self, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	sync, _ := NewSync(node)
	newNode := NewNodeData(self, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	newNode.Version = 2
	producer := block.NewKeyPair().Public
	newNode.Producer = producer
	sync.Insert(newNode)
	if !bytes.Equal(sync.table[string(self)].Producer, producer) {
		t.Errorf("Insert wrong producer")
	}
}

func TestChangeProducer(t *testing.T) {
	self := block.NewKeyPair().Public
	node := NewNodeData(self, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	sync, _ := NewSync(node)
	producer := block.NewKeyPair().Public
	sync.ChangeProducer(producer)

	if !bytes.Equal(sync.table[string(self)].Producer, producer) {
		t.Errorf("ChangeProducer wrong producer")
	}

}

func TestUpdatesSince(t *testing.T) {
	self := block.NewKeyPair().Public
	node := NewNodeData(self, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	sync, _ := NewSync(node)
	for i := uint64(0); i < 10; i++ {
		newNode := NewNodeData(self, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
		newNode.Version = i
		producer := block.NewKeyPair().Public
		newNode.Producer = producer
		sync.Insert(newNode)
	}
	pk, ind, res := sync.updatesSince(5)
	if !bytes.Equal(pk, self) || ind != 10 || len(res) != 1 {
		t.Errorf(fmt.Sprintf("UpdatesSince: %v, %v, %v", pk, ind, len(res)))
	}
}

func TestUpdatesSince2(t *testing.T) {
	self := block.NewKeyPair().Public
	node := NewNodeData(self, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	sync, _ := NewSync(node)
	for i := uint64(0); i < 10; i++ {
		npk := block.NewKeyPair().Public
		newNode := NewNodeData(npk, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
		newNode.Version = i
		producer := block.NewKeyPair().Public
		newNode.Producer = producer
		sync.Insert(newNode)
	}
	pk, ind, res := sync.updatesSince(5)
	if !bytes.Equal(pk, self) || ind != 11 || len(res) != 6 {
		t.Errorf(fmt.Sprintf("UpdatesSince: %v, %v, %v", pk, ind, len(res)))
	}
}

func TestRandomNode(t *testing.T) {
	self := block.NewKeyPair().Public
	node := NewNodeData(self, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	sync, _ := NewSync(node)
	pk := block.NewKeyPair().Public
	node2 := NewNodeData(pk, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	sync.Insert(node2)
	n, _ := sync.RandomNode()
	if !n.Equals(node2) {
		t.Errorf("Error in RandomNode")
	}
}

func TestRandomNode2(t *testing.T) {
	self := block.NewKeyPair().Public
	node := NewNodeData(self, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	sync, _ := NewSync(node)
	pk2 := block.NewKeyPair().Public
	node2 := NewNodeData(pk2, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	sync.Insert(node2)

	pk3 := block.NewKeyPair().Public
	node3 := NewNodeData(pk3, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	sync.Insert(node3)

	n, _ := sync.RandomNode()
	if !n.Equals(node2) && !n.Equals(node3) {
		t.Errorf("Error in RandomNode")
	}
}

func TestRequestSync(t *testing.T) {
	self := block.NewKeyPair().Public
	node := NewNodeData(self, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	sync, _ := NewSync(node)
	pk2 := block.NewKeyPair().Public
	node2 := NewNodeData(pk2, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	sync.Insert(node2)

	pk3 := block.NewKeyPair().Public
	node3 := NewNodeData(pk3, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	sync.Insert(node3)

	addr, n := sync.requestSync()
	if addr.Network() != network.BlockAddrUserUDP.Network() || addr.String() != network.BlockAddrUserUDP.String() {
		t.Errorf("Error RequestSync addr")
	}
	if !n.MyInfo.Equals(node) {
		t.Errorf("Error RequestSync MyInfo")
	}
}

func TestProducerNodeData(t *testing.T) {
	self := block.NewKeyPair().Public
	node := NewNodeData(self, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	sync, err := NewSync(node)
	// with veriosn 0 sync.ProducerNodeData() should be nil
	if err != nil || sync.ProducerNodeData() != nil {
		t.Errorf("NewSync wrong struct")
	}
}

func TestConnectedNodes(t *testing.T) {
	self := block.NewKeyPair().Public
	node := NewNodeData(self, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	sync, _ := NewSync(node)
	if sync.ConnectedNodes() != 0 {
		t.Errorf("Sync.ConnectedNodes() wrong count")
	}
}

func TestAllNodesExceptMe(t *testing.T) {
	self := block.NewKeyPair().Public
	node := NewNodeData(self, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	sync, _ := NewSync(node)
	nodes := sync.AllNodesExceptMe(self)
	if len(nodes) != 0 {
		t.Errorf("Sync.TestAllNodesExceptMe() wrong count")
	}
}

func TestRemoteTableCopy(t *testing.T) {

}
