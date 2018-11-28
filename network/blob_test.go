package network_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/Ansiblock/Ansiblock/block"
	. "github.com/Ansiblock/Ansiblock/network"
)

func TestBlobNew(t *testing.T) {
	b := NewBlobs()
	if len(b.Bs) != NumBlobs {
		t.Errorf("PacketNew: packets size wanted %v was %v", NumBlobs, len(b.Bs))
	}
}

func TestBlobRead(t *testing.T) {
	b := NewBlobs()
	sm := NewSocketMock(nil, nil, nil)
	sm.AddToReadBuff([]byte{2, 2, 2})
	n := b.ReadFrom(sm)
	if n != 1 {
		t.Errorf("ReadFrom: expected to read %v byte but read %v", 1, n)
	}
	for _, ba := range b.Bs {
		if ba.Size != 3 {
			t.Errorf("ReadFrom: expected size %v but was %v", 3, ba.Size)
		}
		if ba.Addr != nil {
			t.Errorf("ReadFrom: expected addr %v but was %v", "nil", ba.Addr)
		}
		for j := uint32(0); j < ba.Size; j++ {
			if ba.Data[j] != 2 {
				t.Fatalf("ReadFrom: unexpected data %v", ba.Data)
			}
		}
	}
}

func TestBlobReadFromError(t *testing.T) {
	b := NewBlobs()
	for i := range b.Bs {
		b.Bs[i].Data[0] = 2
	}
	var pc PCon
	pc.B = make([]byte, 10)

	n := b.ReadFrom(&pc)
	if n != 0 {
		t.Errorf("ReadFrom: Did not Error")
	}
}

func TestBlobWriteTo(t *testing.T) {
	sm := NewSocketMock(nil, nil, nil)
	b := NewBlobs()
	resBuf := make([]byte, 0)
	for i := range b.Bs {
		b.Bs[i].Data[0] = 2
		b.Bs[i].Data[1] = 2
		b.Bs[i].Data[2] = 2
		b.Bs[i].Size = 3
		resBuf = append(resBuf, 2, 2, 2)
	}
	n := b.WriteTo(sm)
	if n != len(b.Bs)*3 {
		t.Errorf("WriteTo: expected to write %v byte but wrote %v", len(b.Bs)*3, n)
	}

	writeBuff := sm.EmptyWriteBuff()
	if !bytes.Equal(writeBuff, resBuf) || len(writeBuff) != len(resBuf) {
		t.Errorf("WriteTo: Wrong buffer")
	}

}

// func TestBlobWriteTo(t *testing.T) {
// 	var pc PCon
// 	pc.B = make([]byte, 2)

// 	trs1 := block.CreateRealTransactions(10)
// 	packets := trs1.ToPackets(nil)
// 	n := packets.WriteTo(&pc)
// 	if n != 7*10 {
// 		t.Errorf("ReadFrom: expected to write %v byte but wrote %v", 70, n)
// 	}
// }

// func TestBlobWriteToError(t *testing.T) {
// 	var pc PCon
// 	pc.B = make([]byte, 2)
// 	pc.B[0] = 1
// 	trs1 := block.CreateRealTransactions(10)
// 	packets := trs1.ToPackets(nil)
// 	n := packets.WriteTo(&pc)
// 	if n != 0 {
// 		t.Errorf("ReadFrom: expected to write %v byte but wrote %v", 0, n)
// 	}
// }

func TestIndex(t *testing.T) {
	blob := new(Blob)
	num := uint64(1)
	for i := 0; i < 100; i++ {
		blob.SetIndex(num)
		if blob.Index() != num {
			t.Errorf("Blob.Index expected %v got %v", num, blob.Index())
		}
		num = num * 173
	}

	for i := uint64(0); i < 100000; i++ {
		blob.SetIndex(i)
		if blob.Index() != i {
			t.Errorf("Blob.Index expected %v got %v", i, blob.Index())
		}
	}

}

func TestFrom(t *testing.T) {
	blob := new(Blob)
	pub := block.NewKeyPair().Public
	blob.SetFrom(pub)
	if !bytes.Equal(blob.From(), pub) {
		t.Errorf("Blob.From expected %v got %v", pub, blob.From())
	}
}

func TestFlags(t *testing.T) {
	blob := new(Blob)
	num := uint32(1)
	for i := 0; i < 100; i++ {
		blob.SetFlags(num)
		if blob.Flags() != num {
			t.Errorf("Blob.Flags expected %v got %v", num, blob.Flags())
			fmt.Println(blob.Data[8+32 : 44])
		}
		num = num * 1173
	}
}

func TestCoding(t *testing.T) {
	blob := new(Blob)
	if blob.IsCoding() {
		t.Error("Blob.Coding wrong initial coding")
	}
	blob.SetCoding()
	if !blob.IsCoding() {
		t.Error("Blob.Coding wrong coding")
	}
}

func TestIndexBlobs(t *testing.T) {
	blobs := NewBlobs()
	from := block.NewKeyPair().Public
	start := uint64(100)
	blobs.IndexBlobs(from, start)
	for i, blob := range blobs.Bs {
		if blob.Index() != uint64(100)+uint64(i) {
			t.Errorf("Blobs.IndexBlobs wrong index: expected %v was %v", start+uint64(i), blob.Index())
		}
		if !bytes.Equal(blob.From(), from) {
			t.Errorf("Blobs.IndexBlobs wrong From: expected %v was %v", from, blob.From())
		}
	}
}

func TestNewNumBlobs(t *testing.T) {
	if len(NewNumBlobs(100).Bs) != 100 {
		t.Error("NewNumBlobs failed")
	}
}

// func TestIndexBlobs(t *testing.T) {
// 	blobs := NewNumBlobs(100)

// }
