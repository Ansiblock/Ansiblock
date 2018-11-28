package messaging

import (
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/network"
)

func TestRequestSerializationDeserialization(t *testing.T) {
	messagingAddr := net.UDPAddr{Port: 12345, IP: net.ParseIP("127.0.0.1")}
	acc1 := block.NewKeyPair()
	var requests Requests
	requests.Requests = []Request{Request{Type: Balance, Addr: &messagingAddr,
		PublicKey: acc1.Public}}
	fmt.Println(requests)

	serializedRequests := requests.Serialize()
	fmt.Println(serializedRequests)

	var deserializedRequests Requests
	deserializedRequests.Deserialize(serializedRequests)
	fmt.Println(deserializedRequests)

	if !reflect.DeepEqual(requests, deserializedRequests) {
		t.Errorf(`deserialized requests does not equal original. \n
		Original: %v
		Deserialized %v`, requests, deserializedRequests)
	}
}

func TestRequestWithEmptyRequest(t *testing.T) {
	messagingAddr := net.UDPAddr{Port: 12345, IP: net.ParseIP("127.0.0.1")}
	acc1 := block.NewKeyPair()
	var requests Requests
	requests.Requests = []Request{Request{Type: Balance, Addr: &messagingAddr,
		PublicKey: acc1.Public}}
	fmt.Println(requests)

	serializedRequests := requests.Serialize()
	fmt.Println(serializedRequests)
	serializedRequests.Ps = append(serializedRequests.Ps,
		network.Packet{Addr: &messagingAddr})
	fmt.Println(serializedRequests)

	var deserializedRequests Requests
	deserializedRequests.Deserialize(serializedRequests)
	fmt.Println(deserializedRequests)

	if !reflect.DeepEqual(requests, deserializedRequests) {
		t.Errorf(`deserialized requests does not equal original. \n
		Original: %v
		Deserialized %v`, requests, deserializedRequests)
	}
}
