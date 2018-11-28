// Package user is responsible for user communication.
// This is an API user should use to connect to the test node.
// Later we will add wallet API
package user

import (
	"bytes"
	"errors"
	"fmt"
	"net"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/log"
	"github.com/Ansiblock/Ansiblock/messaging"
	"github.com/Ansiblock/Ansiblock/network"

	"go.uber.org/zap"
	"golang.org/x/crypto/ed25519"
)

// API object is for querying and sending transactions to the network.
type API struct {
	messagingAddr      net.Addr
	transactionsAddr   net.Addr
	messagingSocket    net.PacketConn
	transactionsSocket net.PacketConn
	validVDFValue      block.VDFValue
	transactionsTotal  uint64
	balances           map[string]int64
}

// NewUserAPI creates new API object.
func NewUserAPI(ra net.Addr, ta net.Addr, rs net.PacketConn, ts net.PacketConn) *API {
	user := &API{messagingAddr: ra, transactionsAddr: ta, messagingSocket: rs, transactionsSocket: ts}
	user.balances = make(map[string]int64)
	return user
}

// Balance method requests balance of user holding 'publicKey' and blocks
// until the server sends a response. If the response packet is dropped
// by the network, this method will hang indefinitely.
func (tc *API) Balance(publicKey ed25519.PublicKey) (int64, error) {
	log.Info("Balance")
	if len(publicKey) != ed25519.PublicKeySize {
		return 0, errors.New("public key not found")
	}
	requestsArr := []messaging.Request{messaging.Request{Type: messaging.Balance, Addr: tc.messagingAddr, PublicKey: publicKey}}
	requests := messaging.Requests{Requests: requestsArr}
	// fmt.Printf("requests %v\n", requests)

	requestPackets := requests.Serialize()
	// fmt.Printf("packets %v\n", requestPackets)
	requestPackets.WriteTo(tc.messagingSocket)
	//TODO: what if connection brakes? this loop will continue infinitely
	for {
		responsePackets := network.NewNumPackets(1)
		n := responsePackets.ReadFrom(tc.messagingSocket)
		// fmt.Printf("readfrom packets %v\n", responsePackets)

		if n > 0 {
			log.Info("user.API read balance response: ", zap.Int("bytes", n))
			// fmt.Printf("user.API read balance response: %v\n bytes", n)

			response := messaging.ResponseBalance{}
			response.Deserialize(responsePackets.Ps[0])
			if bytes.Equal(response.PublicKey, publicKey) {
				tc.balances[string(publicKey)] = response.Value
				break
			}
			log.Warn("user.API's response PublicKey is different from initial one, continue reading!")
			// fmt.Printf("user.API's response PublicKey is different from initial one, continue reading!: %v\n", response)

		}
	}

	if val, ok := tc.balances[string(publicKey)]; ok {
		return val, nil
	}
	return 0, errors.New("public key not found")

}

// TransactionsTotal requests the transaction count from server.
// If the response packet is dropped by the network, this method will hang.
func (tc *API) TransactionsTotal() uint64 {
	log.Info("Transactions Total")
	requestsArr := []messaging.Request{messaging.Request{Type: messaging.TransactionsTotal, Addr: tc.messagingAddr}}
	requests := messaging.Requests{Requests: requestsArr}

	packets := requests.Serialize()
	packets.WriteTo(tc.messagingSocket)

	for {
		packets := network.NewPackets()
		n := packets.ReadFrom(tc.messagingSocket)

		//TODO: do we need check if response type is correct, if so packet must have request type byte!
		if n > 0 {
			log.Info("user.API read transaction count response: ", zap.Int("bytes", n))
			response := messaging.ResponseTransactionsTotal{}
			response.Deserialize(packets.Ps[0])
			tc.transactionsTotal = response.Value
			break
		}
	}
	return tc.transactionsTotal
}

// TransferTransaction will transfer transaction to the transactionSocket
func (tc *API) TransferTransaction(tran block.Transaction) {
	trans := block.Transactions{Ts: []block.Transaction{tran}}
	packets := trans.ToPackets(tc.transactionsAddr)
	packets.WriteTo(tc.transactionsSocket)
}

// Transfer will create Transaction, sign and transfer to the transactionSocket
func (tc *API) Transfer(from *block.KeyPair, to ed25519.PublicKey, token int64, vdf block.VDFValue) {
	tran := block.NewTransaction(from, to, token, 0, vdf)
	tc.TransferTransaction(tran)
}

// ValidVDFValue method queries producer for the valid vdf saved in the ledger
// and returns the result.
// If producer will not respond or the request will be lost this method will hang.
func (tc *API) ValidVDFValue() block.VDFValue {
	log.Info("get valid vdf value from producer")
	requestsArr := []messaging.Request{messaging.Request{Type: messaging.ValidVDFValue, Addr: tc.messagingAddr}}
	requests := messaging.Requests{Requests: requestsArr}

	// fmt.Printf("requests %v\n", requests)
	packets := requests.Serialize()
	packets.WriteTo(tc.messagingSocket)
	// fmt.Printf("packets %v\n", packets)

	ok := true
	for ok {
		packets := network.NewPackets()
		n := packets.ReadFrom(tc.messagingSocket)
		if n > 0 {
			log.Info("user.API read validVDFValue response: ", zap.Int("bytes", n))
			responses := messaging.Responses{Responses: []messaging.Response{}}
			responses.Deserialize(packets)
			switch resp := responses.Responses[0].(type) {
			case *messaging.ResponseValidVDFValue:
				tc.validVDFValue = resp.Value
				ok = false
			default:
				log.Warn("unexpected response for ValidVDF")
				fmt.Println(resp)
				ok = false
				tc.validVDFValue = block.VDF([]byte("hello"))
			}
		}
	}
	return tc.validVDFValue
}
