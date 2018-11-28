package network_test

import (
	"testing"

	"github.com/Ansiblock/Ansiblock/block"
	. "github.com/Ansiblock/Ansiblock/network"
)

func TestPacketNew(t *testing.T) {
	p := NewPackets()
	if len(p.Ps) != NumPackets {
		t.Errorf("PacketNew: packets size wanted %v was %v", NumPackets, len(p.Ps))
	}
}

func TestNewNumPackets(t *testing.T) {
	p := NewNumPackets(10)
	if len(p.Ps) != 10 {
		t.Errorf("PacketNew: packets size wanted %v was %v", NumPackets, len(p.Ps))
	}
}

func TestReadFrom(t *testing.T) {
	p := NewPackets()
	// b := make([]byte, packetDataSize)
	var pc PCon
	pc.B = make([]byte, 10)
	n := p.ReadFrom(&pc)
	if n != 1*8*1024 {
		t.Errorf("ReadFrom: expected to read %v byte but read %v", 1*8*1024, n)
	}
	for _, pa := range p.Ps {
		if pa.Size != 1 {
			t.Errorf("ReadFrom: expected size %v but was %v", 1, pa.Size)
		}
		if pa.Addr == nil {
			t.Errorf("ReadFrom: expected addr %v but was %v", "127.0.0.1:8000", pa.Addr)
		}
		for j := uint8(0); j < pa.Size; j++ {
			if pa.Data[j] != 1 {
				t.Fatalf("ReadFrom: unexpected data %v", pa.Data)
			}
		}
	}
}

// func TestReadFromPanic(t *testing.T) {
// 	defer func() {
// 		if r := recover(); r == nil {
// 			t.Errorf("ReadFrom: Did not panic")
// 		}
// 	}()
// 	p := NewPackets()
// 	for i := range p.Ps {
// 		p.Ps[i].Data[0] = 1
// 	}
// 	var pc PCon
// 	pc.B = make([]byte, 10)

// 	_ = p.ReadFrom(&pc)
// }

func TestReadFromError(t *testing.T) {
	p := NewPackets()
	for i := range p.Ps {
		p.Ps[i].Data[0] = 2
	}
	var pc PCon
	pc.B = make([]byte, 10)

	n := p.ReadFrom(&pc)
	if n != 0 {
		t.Errorf("ReadFrom: Did not Error")
	}
}

// func TestReadFromTimeout(t *testing.T) {
// 	p := NewPackets()
// 	// for i := range p.Ps {
// 	// 	p.Ps[i].Data[0] = 0
// 	// }
// 	var pc PCon
// 	pc.B = make([]byte, 10)
// 	pc.B[9] = 1
// 	n := p.ReadFrom(&pc)
// 	if n != 1 {
// 		t.Errorf("ReadFrom: expected to read %v byte but read %v", 1*8*1024, n)
// 	}
// }

func TestWriteTo(t *testing.T) {
	var pc PCon
	pc.B = make([]byte, 2)

	trs1 := block.CreateRealTransactions(10)
	packets := trs1.ToPackets(nil)
	n := packets.WriteTo(&pc)
	if n != 7*10 {
		t.Errorf("ReadFrom: expected to write %v byte but wrote %v", 70, n)
	}
}

func TestWriteToError(t *testing.T) {
	var pc PCon
	pc.B = make([]byte, 2)
	pc.B[0] = 1
	trs1 := block.CreateRealTransactions(10)
	packets := trs1.ToPackets(nil)
	n := packets.WriteTo(&pc)
	if n != 0 {
		t.Errorf("ReadFrom: expected to write %v byte but wrote %v", 0, n)
	}
}
