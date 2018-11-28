package books

import (
	"github.com/Ansiblock/Ansiblock/network"
)

// SignatureVerification accepts Packets verifies them and sends verified only
// Packets to the output channel
func SignatureVerification(packetReceiver <-chan *network.Packets) <-chan *network.Packets {
	out := make(chan *network.Packets, cap(packetReceiver))
	go func(out chan<- *network.Packets, packetReceiver <-chan *network.Packets) {
		for {
			packets := network.PacketBatch(packetReceiver)
			packets, _ = verifyPackets(packets)
			// log.Info(fmt.Sprintf("SignatureVerification: %v packets verified", num))
			for _, packet := range packets {
				out <- packet
			}
		}
	}(out, packetReceiver)
	return out
}
