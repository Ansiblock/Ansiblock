// +build !cuda

package books

import (
	"github.com/Ansiblock/Ansiblock/network"
	"golang.org/x/crypto/ed25519"
)

// type empty struct{}

// func verifiedTransactions(packets *network.Packets) *network.Packets {
// 	var res network.Packets
// 	res.Ps = make([]network.Packet, 0, len(packets.Ps))
// 	pack := make(chan empty, 1)
// 	lock := new(sync.Mutex)
// 	for _, packet := range packets.Ps {
// 		go func(packet network.Packet, pack chan empty) {
// 			if verifyPacket(&packet) {
// 				lock.Lock()
// 				res.Ps = append(res.Ps, packet)
// 				lock.Unlock()
// 				pack <- empty{}
// 			} else {
// 				log.Warn("packet did not pass verification")
// 			}
// 		}(packet, pack)
// 	}
// 	for range packets.Ps {
// 		<-pack
// 	}
// 	return &res
// 	// return packets
// }

// func verifiedTransactions(packets *network.Packets) (*network.Packets, int) {
// 	var res network.Packets
// 	res.Ps = make([]network.Packet, 0, len(packets.Ps))
// 	pack := make(chan *network.Packet, 1)
// 	for _, packet := range packets.Ps {
// 		go func(packet *network.Packet, pack chan *network.Packet) {
// 			if verifyPacket(packet) {
// 				pack <- packet
// 			} else {
// 				log.Warn("packet did not pass verification")

// 			}
// 		}(&packet, pack)
// 	}
// 	ok := true
// 	for ok {
// 		select {
// 		case p := <-pack:
// 			res.Ps = append(res.Ps, *p)
// 		case <-time.After(20 * time.Millisecond):
// 			ok = false
// 		}
// 	}
// 	return &res, len(res.Ps)
// 	// return packets
// }

// func verifiedTransactions(packets *network.Packets) (*network.Packets, int) {
// 	res := make(chan byte, len(packets.Ps))
// 	for i := range packets.Ps {
// 		go func(packet *network.Packet) {
// 			if !ed25519.Verify(packet.Data[:32], packet.Data[:112], packet.Data[112:176]) {
// 				packet.Size = 0
// 				res <- 0
// 				log.Warn(fmt.Sprintf("packet did not pass verification - %v\n", packet))
// 			} else {
// 				res <- 1
// 			}
// 		}(&packets.Ps[i])
// 	}
// 	ans := 0
// 	for range packets.Ps {
// 		ans += int(<-res)
// 	}
// 	return packets, ans
// }

// SignatureVerification accepts Transactions verifies them and sends verified only
// transactions to the output channel
// func SignatureVerification(packetReceiver <-chan *network.Packets) <-chan *network.Packets {
// 	out := make(chan *network.Packets, cap(packetReceiver))
// 	go func(out chan<- *network.Packets, packetReceiver <-chan *network.Packets) {
// 		for {
// 			// pacs := network.PacketBatch(packetReceiver)
// 			packets, ok := <-packetReceiver
// 			if !ok {
// 				log.Error("SignatureVerification receiver failed, closing channel")
// 				close(out)
// 				return
// 			}
// 			log.Info(fmt.Sprintf("SignatureVerification: %v packets received", len(packets.Ps)))
// 			packs, num := verifiedTransactions(packets)
// 			out <- packs
// 			log.Info(fmt.Sprintf("SignatureVerification: %v transactions verified", num))
// 		}
// 	}(out, packetReceiver)
// 	return out
// }

// verifyPackets recives array of pointers to network packets.
// each packet contains several transaction. The signature of each transation is verified.
// If signature is not valied, packets size is set to zero.
func verifyPackets(packets []*network.Packets) ([]*network.Packets, int) {
	res := make(chan byte, len(packets))
	for i := range packets {
		go func(packet *network.Packets) {
			for j := range packet.Ps {
				go func(pa *network.Packet) {
					if !ed25519.Verify(pa.Data[64:96], pa.Data[64:176], pa.Data[:64]) {
						pa.Size = 0
						res <- 0
					} else {
						res <- 1
					}
				}(&packet.Ps[j])
			}
		}(packets[i])
	}
	ans := 0
	for _, packet := range packets {
		for range packet.Ps {
			ans += int(<-res)
		}
	}
	return packets, ans
}
