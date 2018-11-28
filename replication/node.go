package replication

import (
	"net"

	"github.com/Ansiblock/Ansiblock/block"
)

// Sockets struct saves all sockets of the node
type Sockets struct {
	Sync        net.PacketConn
	SyncSend    net.PacketConn
	Messages    net.PacketConn
	Replicate   net.PacketConn
	Transaction net.PacketConn
	Respond     net.PacketConn
	Broadcast   net.PacketConn
	Repair      net.PacketConn
	Transport   net.PacketConn
}

// Node saves the data of a user node
type Node struct {
	Data    *NodeData
	Sockets Sockets
}

// NewNode creates new Node
func NewNode(nodeType string, name string) Node {
	sync, _ := net.ListenPacket("udp", "127.0.0.1:0")
	syncSend, _ := net.ListenPacket("udp", "127.0.0.1:0")
	messages, _ := net.ListenPacket("udp", "127.0.0.1:0")
	replicate, _ := net.ListenPacket("udp", "127.0.0.1:0")
	transaction, _ := net.ListenPacket("udp", "127.0.0.1:0")
	respond, _ := net.ListenPacket("udp", "127.0.0.1:0")
	broadcast, _ := net.ListenPacket("udp", "127.0.0.1:0")
	repair, _ := net.ListenPacket("udp", "127.0.0.1:0")
	transport, _ := net.ListenPacket("udp", "127.0.0.1:0")

	pubKey := block.NewKeyPair().Public

	data := NewNodeData(pubKey, nodeType, name, *sync.LocalAddr().(*net.UDPAddr), *replicate.LocalAddr().(*net.UDPAddr), *messages.LocalAddr().(*net.UDPAddr), *transaction.LocalAddr().(*net.UDPAddr), *repair.LocalAddr().(*net.UDPAddr))

	return Node{Data: data, Sockets: Sockets{sync, syncSend, messages, replicate, transaction, respond, broadcast, repair, transport}}
}

func NewProducerNode(nodeType string, name string) Node {
	sync, _ := net.ListenPacket("udp", "127.0.0.1:0")
	syncSend, _ := net.ListenPacket("udp", "127.0.0.1:0")
	messages, _ := net.ListenPacket("udp", "127.0.0.1:59133")
	replicate, _ := net.ListenPacket("udp", "127.0.0.1:0")
	transaction, _ := net.ListenPacket("udp", "127.0.0.1:59135")
	respond, _ := net.ListenPacket("udp", "127.0.0.1:0")
	broadcast, _ := net.ListenPacket("udp", "127.0.0.1:0")
	repair, _ := net.ListenPacket("udp", "127.0.0.1:0")
	transport, _ := net.ListenPacket("udp", "127.0.0.1:0")

	pubKey := block.NewKeyPair().Public

	data := NewNodeData(pubKey, nodeType, name, *sync.LocalAddr().(*net.UDPAddr), *replicate.LocalAddr().(*net.UDPAddr), *messages.LocalAddr().(*net.UDPAddr), *transaction.LocalAddr().(*net.UDPAddr), *repair.LocalAddr().(*net.UDPAddr))

	return Node{Data: data, Sockets: Sockets{sync, syncSend, messages, replicate, transaction, respond, broadcast, repair, transport}}
}

// DestroyNode closes all packet connections of the node
// func (tn *Node) DestroyNode() {
// 	tn.Sockets.Sync.Close()
// 	tn.Sockets.SyncSend.Close()
// 	tn.Sockets.Messages.Close()
// 	tn.Sockets.Replicate.Close()
// 	tn.Sockets.Transaction.Close()
// 	tn.Sockets.Respond.Close()
// 	tn.Sockets.Broadcast.Close()
// 	tn.Sockets.Repair.Close()
// }
