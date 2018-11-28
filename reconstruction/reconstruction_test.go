package reconstruction

import (
	"bytes"
	"testing"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/network"
	"github.com/Ansiblock/Ansiblock/replication"
)

func TestHandleReconstruct(t *testing.T) {
	self := block.NewKeyPair().Public
	node := replication.NewNodeData(self, "producer", "self", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	s, _ := replication.NewSync(node)
	frame := network.NewFrame()

	simpleRequest := new(Request)
	simpleRequest.Index = 1
	simpleRequest.From = node
	blob := simpleRequest.Serialize()

	res := handleReconstruct(s, frame, blob)
	if res != nil {
		t.Errorf("HandleReconstruct not nil")
	}

	fromSelf := block.NewKeyPair().Public
	fromNode := replication.NewNodeData(fromSelf, "signer", "from", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)

	request := new(Request)
	request.Index = 1
	request.From = fromNode
	blob = request.Serialize()

	res = handleReconstruct(s, frame, blob)
	if res != nil {
		t.Errorf("HandleReconstruct not nil - found some blob")
	}

	myBlob := new(network.Blob)
	myBlob.SetIndex(request.Index)
	myBlob.Size = uint32(len(myBlob.Data))
	frame.Blobs[request.Index] = myBlob
	res = handleReconstruct(s, frame, blob)
	if res.Index() != request.Index {
		t.Errorf("HandleReconstruct returned blob has wrong index")
	}
	if !bytes.Equal(res.From(), fromSelf) {
		t.Errorf("HandleReconstruct returned blob has wrong from")
	}
}

func TestListener(t *testing.T) {
	self := block.NewKeyPair().Public
	node := replication.NewNodeData(self, "producer", "self", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	s, _ := replication.NewSync(node)
	frame := network.NewFrame()
	for i := 0; i < 10; i++ {
		blob := new(network.Blob)
		blob.SetIndex(uint64(i))
		blob.Size = uint32(len(blob.Data))
		frame.Blobs[i] = blob
	}
	fromSelf := block.NewKeyPair().Public
	fromNode := replication.NewNodeData(fromSelf, "signer", "from", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)

	input := make(chan *network.Blobs)
	go func() {
		for i := 0; i < 10; i++ {
			request := new(Request)
			request.Index = uint64(i)
			request.From = fromNode
			blob := request.Serialize()
			blobs := network.NewNumBlobs(1)
			blobs.Bs[0] = *blob
			input <- blobs
		}
	}()
	output := Listener(s, frame, input)
	for i := 0; i < 10; i++ {
		blobs := <-output
		if blobs.Bs[0].Index() != uint64(i) {
			t.Errorf("HandleReconstruct returned blob has wrong index")
		}
		if !bytes.Equal(blobs.Bs[0].From(), fromSelf) {
			t.Errorf("HandleReconstruct returned blob has wrong from")
		}
	}
}
