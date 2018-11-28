package user

import (
	"net"
	"reflect"
	"testing"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/messaging"
	"github.com/Ansiblock/Ansiblock/network"
)

var MessagingAddrServerUDP = net.UDPAddr{
	Port: 50056,
	IP:   net.ParseIP("127.0.0.1"),
}

func TestBalance(t *testing.T) {
	publicKey := block.KeyPairs(1)[0].Public
	// var requests messaging.Requests
	// requests.Requests = []messaging.Request{messaging.Request{Type: messaging.Balance, Addr: &network.MessagingAddrServerUDP, PublicKey: publicKey}}

	// create balance response and serialize it
	response := messaging.ResponseBalance{Value: 100, Addr: &MessagingAddrServerUDP, PublicKey: publicKey}
	packet := response.Serialize()

	// create mock socket
	messagingCon := network.NewSocketMock(nil, nil, &MessagingAddrServerUDP)
	messagingCon.AddToReadBuff(packet.Data[:packet.Size])

	// create user api
	us := NewUserAPI(&MessagingAddrServerUDP, nil, messagingCon, nil)
	balance, err := us.Balance(publicKey)

	if err != nil || balance != 100 {
		t.Errorf(`balance request error! error: %v, balance: %v`, err, balance)
	}

	balance, err = us.Balance([]byte{1, 2})
	if err == nil {
		t.Errorf("Got balance from unknown account")
	}

}

func TestTransactionsTotal(t *testing.T) {
	// create transaction count response and serialize it
	response := messaging.ResponseTransactionsTotal{Addr: &MessagingAddrServerUDP, Value: 156}
	packet := response.Serialize()

	// create mock socket
	messagingCon := network.NewSocketMock(nil, nil, &MessagingAddrServerUDP)
	messagingCon.AddToReadBuff(packet.Data[:packet.Size])

	// create user api
	us := NewUserAPI(&MessagingAddrServerUDP, nil, messagingCon, nil)
	count := us.TransactionsTotal()

	if count != 156 {
		t.Errorf(`incorrect transaction count was returned! count: %v`, count)
	}
}

func TestValidVDF(t *testing.T) {
	// create valid vdf response and serialize it
	testValidVDFValue := block.VDF([]byte("TestResponsResponse"))
	response := messaging.ResponseValidVDFValue{Value: testValidVDFValue, Addr: &MessagingAddrServerUDP}
	packet := response.Serialize()

	// create mock socket
	messagingCon := network.NewSocketMock(nil, nil, &MessagingAddrServerUDP)
	messagingCon.AddToReadBuff(packet.Data[:packet.Size])

	// create user api
	us := NewUserAPI(&MessagingAddrServerUDP, nil, messagingCon, nil)
	vdf := us.ValidVDFValue()

	if !reflect.DeepEqual(vdf, testValidVDFValue) {
		t.Errorf(`incorrect valid vdf value`)
	}

	response2 := messaging.ResponseBalance{Value: 10, Addr: &MessagingAddrServerUDP, PublicKey: testValidVDFValue[:32]}
	packet2 := response2.Serialize()

	messagingCon.AddToReadBuff(packet.Data[:packet2.Size])
	vdf = us.ValidVDFValue()
	if !reflect.DeepEqual(vdf, testValidVDFValue) {
		t.Errorf(`incorrect valid vdf value`)
	}

}

func TestTransfer(t *testing.T) {
	pk := block.NewKeyPair()
	// create mock socket
	messagingCon := network.NewSocketMock(nil, nil, &MessagingAddrServerUDP)
	messagingCon.AddToReadBuff(make([]byte, 0))

	// create user api
	us := NewUserAPI(&MessagingAddrServerUDP, nil, nil, messagingCon)
	us.Transfer(&pk, pk.Public, 1, block.VDF([]byte("ee")))
	if messagingCon.WriteBuffSize() != 176 {
		t.Errorf("Incorrect transfer")
	}
}
