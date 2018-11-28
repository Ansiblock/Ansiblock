package network

import (
	"bytes"
	"testing"
	"time"
)

func TestBlobGenerator(t *testing.T) {
	var pc PCon
	pc.B = make([]byte, 10)
	out := BlobGenerator(&pc, 1)
	index := 1
	for blob := range out {
		if blob.Bs[0].Data[0] != 1 {
			t.Errorf("PacketGenerator: wrong packet %v", blob)
		}
		if index == 9 {
			break
		}
		index++
	}
}

func TestBlobSender(t *testing.T) {
	input := make(chan *Blobs)
	messagingCon := NewSocketMock(nil, nil, nil)

	BlobSender(messagingCon, input)

	blobs := new(Blobs)
	blobs.Bs = make([]Blob, 10)
	resBuf := make([]byte, 10000)
	index := 0
	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {
			blobs.Bs[i].Data[j] = byte(j)
			resBuf[index] = byte(j)
			index++
		}
		blobs.Bs[i].Addr = nil
		blobs.Bs[i].Size = 1000
	}

	input <- blobs
	time.Sleep(1 * time.Second)
	writeBuff := messagingCon.EmptyWriteBuff()
	if !bytes.Equal(writeBuff, resBuf) {
		t.Errorf(`BlobSender sent wrong blobs. \n
		Original: %v ...
		Sent %v... `, resBuf[:10], writeBuff[:10])
	}
}
