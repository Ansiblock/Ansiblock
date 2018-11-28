package network

import (
	"net"
	"time"

	"github.com/Ansiblock/Ansiblock/log"
	"go.uber.org/zap"
)

const (
	maxBatchSize = 100000
)

// PacketGenerator thread is responsible for reading packets from socket
// and sending to the output channel
func PacketGenerator(reader net.PacketConn, capacity int) <-chan *Packets {
	// packetCount := 0
	out := make(chan *Packets, capacity)
	go func(reader net.PacketConn, p chan<- *Packets) {
		for {
			packets := NewPackets()
			n := packets.ReadFrom(reader)

			if n > 0 {
				// packetCount += len(packets.Ps)
				log.Info("PacketGenerator: ", zap.Int("Total Packets", n))
				// fmt.Println(n)
				out <- packets
			}
		}
	}(reader, out)
	return out
}

// PacketBatch is responsible for reading packets from
// channel and sending batches of packets to the output
// maximum batch size should be maxBatchSize packets
// PacketBatch waits for 1 second for the incoming packets
func PacketBatch(input <-chan *Packets) []*Packets {
	batch := make([]*Packets, 0, 200)
	size := 0
	ok := true
	for ok {
		select {
		case packets := <-input:
			batch = append(batch, packets)
			size += len(packets.Ps)
			if size > maxBatchSize {
				ok = false
			}

		case <-time.After(300 * time.Millisecond): //TODO refactor to wait less
			ok = false
		}
	}
	return batch
}
