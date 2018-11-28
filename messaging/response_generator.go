package messaging

import (
	"net"

	"github.com/Ansiblock/Ansiblock/books"
	"github.com/Ansiblock/Ansiblock/network"
)

const messageCapacity = 1

// ResponseGenerator thread is responsible for generating Responses
// from network Packets.
func ResponseGenerator(input <-chan *network.Packets, am *books.Accounts) <-chan *Responses {
	out := make(chan *Responses, messageCapacity)
	go func(input <-chan *network.Packets) {
		for {
			batchPackets := network.PacketBatch(input)
			for _, batch := range batchPackets {
				var requests Requests
				requests.Requests = make([]Request, len(batch.Ps))
				requests.Deserialize(batch)
				responses := ProcessMessages(requests.Requests, am)
				out <- responses
			}
		}
	}(input)
	return out
}

// ResponseSender thread is responsible to Serialize and
// send Responses to the socket
func ResponseSender(writer net.PacketConn, input <-chan *Responses) {
	go func() {
		for {
			responses := <-input
			packets := responses.Serialize()
			packets.WriteTo(writer)
		}
	}()
}
