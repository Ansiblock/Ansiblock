package messaging

import (
	"net"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/network"
	"github.com/Ansiblock/Ansiblock/utils"
	"golang.org/x/crypto/ed25519"
)

// Response interface is implemented by response concrete types
// ResponseBalance, ResponseValidVDFValue, ResponseTransactionsTotal
type Response interface {
	Serialize() network.Packet
	Deserialize(network.Packet)
}

// Responses is slice of Response types
type Responses struct {
	Responses []Response
}

// ResponseBalance stores balance on requested account
type ResponseBalance struct {
	Value     int64
	Addr      net.Addr
	PublicKey ed25519.PublicKey
}

// ResponseValidVDFValue stores last vdf value on producer node
type ResponseValidVDFValue struct {
	Value block.VDFValue
	Addr  net.Addr
}

// ResponseTransactionsTotal stores processed transactions count
type ResponseTransactionsTotal struct {
	Value uint64
	Addr  net.Addr
}

// Serialize method converts ResponseBalance to Packet
func (rb *ResponseBalance) Serialize() network.Packet {
	packet := new(network.Packet)
	packet.Addr = rb.Addr
	packet.Size = ed25519.PublicKeySize + 8 + 1
	packet.Data[0] = Balance
	start := 1
	//copy balance value
	packet.Data[start] = byte(rb.Value >> 56)
	packet.Data[start+1] = byte(rb.Value >> 48)
	packet.Data[start+2] = byte(rb.Value >> 40)
	packet.Data[start+3] = byte(rb.Value >> 32)
	packet.Data[start+4] = byte(rb.Value >> 24)
	packet.Data[start+5] = byte(rb.Value >> 16)
	packet.Data[start+6] = byte(rb.Value >> 8)
	packet.Data[start+7] = byte(rb.Value)
	start += 8
	//copy public key into packet data
	for j := 0; j < ed25519.PublicKeySize; j++ {
		packet.Data[j+start] = rb.PublicKey[j]
	}
	return *packet
}

// Deserialize method converts Packet to ResponseBalance
//TODO: error handling?? what if it is not balance response packet
func (rb *ResponseBalance) Deserialize(packet network.Packet) {
	rb.Addr = packet.Addr
	rb.Value = utils.ByteToInt64(packet.Data[:], 1)
	rb.PublicKey = make([]byte, ed25519.PublicKeySize)
	for j := 0; j < ed25519.PublicKeySize; j++ {
		rb.PublicKey[j] = packet.Data[j+9]
	}
}

// Serialize method converts ResponseValidVDFValue to Packet
func (rl *ResponseValidVDFValue) Serialize() network.Packet {
	packet := new(network.Packet)
	packet.Addr = rl.Addr
	packet.Size = 1 + 32 //Type size plus VDF size
	packet.Data[0] = ValidVDFValue
	for j := 0; j < block.VDFSize; j++ {
		packet.Data[j+1] = rl.Value[j]
	}
	return *packet
}

// Deserialize method converts Packet to ResponseValidVDFValue
func (rl *ResponseValidVDFValue) Deserialize(packet network.Packet) {
	rl.Addr = packet.Addr
	rl.Value = make([]byte, block.VDFSize)
	for j := 0; j < block.VDFSize; j++ {
		rl.Value[j] = packet.Data[j+1]
	}
}

// Serialize method converts ResponseTransactionsTotal to Packet
func (rt *ResponseTransactionsTotal) Serialize() network.Packet {
	packet := new(network.Packet)
	packet.Addr = rt.Addr
	packet.Size = 1 + 8 //Type size plus uint64 size
	packet.Data[0] = TransactionsTotal
	v := utils.Uint64toByte(rt.Value)
	for j := 0; j < 8; j++ {
		packet.Data[j+1] = v[j]
	}
	return *packet
}

// Deserialize method converts Packet to ResponseTransactionsTotal
func (rt *ResponseTransactionsTotal) Deserialize(packet network.Packet) {
	rt.Addr = packet.Addr
	rt.Value = uint64(utils.ByteToInt64(packet.Data[:10], 1))
}

// Serialize method converts Responses to Packets
func (rs *Responses) Serialize() *network.Packets {
	var result network.Packets
	result.Ps = make([]network.Packet, len(rs.Responses))
	counter := make(chan bool)
	for i, response := range rs.Responses {
		go func(i int, response Response) {
			result.Ps[i] = response.Serialize()
			counter <- true
		}(i, response)
	}
	for range rs.Responses {
		<-counter
	}
	return &result
}

// Deserialize method converts Packets to Responses
func (rs *Responses) Deserialize(packets *network.Packets) {
	rs.Responses = make([]Response, len(packets.Ps))
	counter := make(chan bool)
	for i := range packets.Ps {
		go func(i int) {
			switch packets.Ps[i].Data[0] {
			case Balance:
				var deserializedResponse ResponseBalance
				deserializedResponse.Deserialize(packets.Ps[i])
				rs.Responses[i] = &deserializedResponse
			case ValidVDFValue:
				var deserializedResponse ResponseValidVDFValue
				deserializedResponse.Deserialize(packets.Ps[i])
				rs.Responses[i] = &deserializedResponse
			case TransactionsTotal:
				var deserializedResponse ResponseTransactionsTotal
				deserializedResponse.Deserialize(packets.Ps[i])
				rs.Responses[i] = &deserializedResponse
			}
			counter <- true
		}(i)
	}
	for range packets.Ps {
		<-counter
	}
}
