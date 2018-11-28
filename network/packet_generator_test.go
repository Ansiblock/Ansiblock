package network

import "testing"

func TestPacketGenerator(t *testing.T) {
	var pc PCon
	pc.B = make([]byte, 10)
	out := PacketGenerator(&pc, 1)
	index := 1
	for packet := range out {
		if packet.Ps[0].Data[0] != 1 {
			t.Errorf("PacketGenerator: wrong packet %v", packet)
		}
		if index == 9 {
			break
		}
		index++
	}
}

func TestPacketBatch(t *testing.T) {
	input := make(chan *Packets, 1)
	packets := NewPackets()
	input <- packets
	out := PacketBatch(input)
	if len(out) != 1 {
		t.Errorf("PacketBatch: wrong number of packets %v", out)
	}
}

func TestPacketBatchMany(t *testing.T) {
	input := make(chan *Packets, 1)
	go func() {
		for i := 0; i < 1000000; i++ {
			packets := NewPackets()
			packets.Ps = packets.Ps[:1000]
			input <- packets
		}
	}()
	out := PacketBatch(input)
	if len(out) != 101 {
		t.Errorf("PacketBatch: wrong number of packets %v", out)
	}
}
