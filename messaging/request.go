package messaging

import (
	"net"

	"github.com/Ansiblock/Ansiblock/network"

	"golang.org/x/crypto/ed25519"
)

// Type describes message types, balance, last vdf value and transactions count
type Type = byte

const (
	// Balance message
	Balance Type = iota

	// ValidVDFValue message
	ValidVDFValue

	// TransactionsTotal message
	TransactionsTotal
)

// Request stores message request and sender address
type Request struct {
	Type      Type
	Addr      net.Addr
	PublicKey ed25519.PublicKey
}

// Requests is a slice of Request types
type Requests struct {
	Requests []Request
}

// Serialize method converts Requests to Packets
func (r *Requests) Serialize() *network.Packets {
	packets := new(network.Packets)
	packets.Ps = make([]network.Packet, len(r.Requests))
	counter := make(chan bool)
	for i := range r.Requests {
		go func(i int) {
			packets.Ps[i].Addr = r.Requests[i].Addr
			packets.Ps[i].Size = 1
			packets.Ps[i].Data[0] = r.Requests[i].Type
			if r.Requests[i].Type == Balance {
				packets.Ps[i].Size += ed25519.PublicKeySize
				for j := 0; j < ed25519.PublicKeySize; j++ {
					packets.Ps[i].Data[j+1] = r.Requests[i].PublicKey[j]
				}
			}
			counter <- true
		}(i)
	}
	for range r.Requests {
		<-counter
	}
	return packets
}

// Deserialize method converts Packets to Requests
// TODO fix the bug!!
func (r *Requests) Deserialize(packets *network.Packets) {
	count := 0
	for i := range packets.Ps {
		if packets.Ps[i].Size == 0 {
			continue
		}
		count++
	}
	r.Requests = make([]Request, count)
	for i := range packets.Ps {
		if packets.Ps[i].Size == 0 {
			continue
		}
		r.Requests[i].Addr = packets.Ps[i].Addr
		r.Requests[i].Type = packets.Ps[i].Data[0]
		r.Requests[i].PublicKey = make([]byte, ed25519.PublicKeySize)
		for j := 0; j < ed25519.PublicKeySize; j++ {
			r.Requests[i].PublicKey[j] = packets.Ps[i].Data[j+1]
		}
	}
}
