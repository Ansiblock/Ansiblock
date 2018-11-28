package messaging

import (
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/books"
	"github.com/Ansiblock/Ansiblock/network"
)

func TestResponseGeneratorCheckReturnType(t *testing.T) {
	accounts := books.NewBookManager()
	acc1 := block.NewKeyPair()
	acc2 := block.NewKeyPair()
	accounts.CreateAccount(acc1.Public, 100)
	accounts.CreateAccount(acc2.Public, 100)
	vdf := []byte{1, 2, 3}
	accounts.AddValidVDFValue(vdf)

	input := make(chan *network.Packets)
	out := ResponseGenerator(input, accounts)

	// create request
	messagingAddr := net.UDPAddr{
		Port: 12345,
		IP:   net.ParseIP("127.0.0.1"),
	}
	var requests Requests
	requests.Requests = []Request{Request{Type: Balance, Addr: &messagingAddr,
		PublicKey: acc1.Public}}
	fmt.Println(requests)

	// serialize and send
	serializedRequests := requests.Serialize()
	fmt.Println(serializedRequests)
	// send
	input <- serializedRequests

	responses := <-out
	fmt.Printf("%v", responses)
	r := responses.Responses[0].(*ResponseBalance)
	if r.Value != 100 {
		t.Errorf(`response balance does not match expacted value, 
		should be 104 got %d`, r.Value)
	}

}

func TestResponseSender(t *testing.T) {
	input := make(chan *Responses)
	// var writer *network.PCon
	messagingCon := network.NewSocketMock(nil, nil, nil)

	ResponseSender(messagingCon, input)

	responseAddr := net.UDPAddr{Port: 12345, IP: net.ParseIP("127.0.0.1")}
	keyPair := block.NewKeyPair()
	responseBalance := ResponseBalance{Value: 1024, Addr: &responseAddr,
		PublicKey: keyPair.Public}

	var responses Responses
	responses.Responses = append(responses.Responses, &responseBalance)
	input <- &responses

	serializedResponse := responseBalance.Serialize()
	fmt.Println(serializedResponse)

	// packetRespose := 0
	time.Sleep(100 * time.Millisecond)
	writeBuff := messagingCon.EmptyWriteBuff()[:41]
	if !reflect.DeepEqual(writeBuff, serializedResponse.Data[:41]) {
		t.Errorf(`deserialized response does not equal original. \n
		Original: %v
		Deserialized %v`, writeBuff, serializedResponse.Data)
	}
}
