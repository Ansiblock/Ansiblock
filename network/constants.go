package network

import "net"

const (
	// TransactionAddrServer is an address for transaction socket
	TransactionAddrServer = "0.0.0.0:40051"
	// BlockAddrServer is an address for incoming blocks
	BlockAddrServer = "127.0.0.1:40052"
	// MessagingAddrServer is an address for messaging socket
	MessagingAddrServer = "0.0.0.0:40057"
)

// BlockAddrUserUDP is UDP address for Block transfering
var BlockAddrUserUDP = net.UDPAddr{
	Port: 50053,
	IP:   net.ParseIP("127.0.0.1"),
}
