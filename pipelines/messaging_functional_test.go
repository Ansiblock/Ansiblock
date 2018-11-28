package pipelines_test

import (
	"encoding/json"
	"fmt"
	"net"
	"testing"

	"github.com/Ansiblock/Ansiblock/mint"
	"github.com/Ansiblock/Ansiblock/pipelines"
	"github.com/Ansiblock/Ansiblock/user"

	"github.com/Ansiblock/Ansiblock/books"
	"github.com/Ansiblock/Ansiblock/log"
	"github.com/Ansiblock/Ansiblock/network"
	"go.uber.org/zap"
)

func parseMint() mint.Mint {

	// dat := []byte{123, 34, 75, 101, 121, 80, 97, 105, 114, 34, 58, 123, 34, 80, 117, 98, 108, 105, 99, 34, 58, 34, 66, 87, 84, 85, 71, 48, 55, 89, 112, 65, 67, 50, 66, 73, 43, 98, 97, 122, 74, 69, 69, 97, 115, 104, 71, 72, 67, 66, 54, 120, 108, 53, 66, 105, 74, 108, 81, 118, 66, 51, 88, 75, 107, 61, 34, 44, 34, 80, 114, 105, 118, 97, 116, 101, 34, 58, 34, 69, 106, 49, 86, 120, 117, 110, 55, 81, 85, 66, 107, 90, 76, 90, 120, 54, 113, 86, 115, 109, 114, 51, 114, 101, 110, 119, 87, 109, 70, 84, 78, 87, 87, 48, 66, 110, 106, 89, 55, 107, 57, 89, 70, 90, 78, 81, 98, 84, 116, 105, 107, 65, 76, 89, 69, 106, 53, 116, 114, 77, 107, 81, 82, 113, 121, 69, 89, 99, 73, 72, 114, 71, 88, 107, 71, 73, 109, 86, 67, 56, 72, 100, 99, 113, 81, 61, 61, 34, 125, 44, 34, 80, 114, 105, 118, 97, 116, 101, 75, 101, 121, 34, 58, 34, 69, 106, 49, 86, 120, 117, 110, 55, 81, 85, 66, 107, 90, 76, 90, 120, 54, 113, 86, 115, 109, 114, 51, 114, 101, 110, 119, 87, 109, 70, 84, 78, 87, 87, 48, 66, 110, 106, 89, 55, 107, 57, 89, 70, 90, 78, 81, 98, 84, 116, 105, 107, 65, 76, 89, 69, 106, 53, 116, 114, 77, 107, 81, 82, 113, 121, 69, 89, 99, 73, 72, 114, 71, 88, 107, 71, 73, 109, 86, 67, 56, 72, 100, 99, 113, 81, 61, 61, 34, 44, 34, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 66, 87, 84, 85, 71, 48, 55, 89, 112, 65, 67, 50, 66, 73, 43, 98, 97, 122, 74, 69, 69, 97, 115, 104, 71, 72, 67, 66, 54, 120, 108, 53, 66, 105, 74, 108, 81, 118, 66, 51, 88, 75, 107, 61, 34, 44, 34, 84, 111, 107, 101, 110, 115, 34, 58, 49, 48, 48, 48, 48, 48, 48, 125}
	dat := []byte{123, 34, 75, 101, 121, 80, 97, 105, 114, 34, 58, 123, 34, 80, 117, 98, 108, 105, 99, 34, 58, 34, 56, 109, 112, 55, 120, 83, 120, 97, 79, 55, 88, 53, 55, 81, 79, 69, 108, 103, 75, 77, 83, 119, 80, 71, 110, 118, 70, 80, 57, 118, 100, 49, 104, 100, 107, 121, 88, 102, 120, 109, 65, 103, 56, 61, 34, 44, 34, 80, 114, 105, 118, 97, 116, 101, 34, 58, 34, 67, 75, 103, 118, 105, 121, 66, 120, 54, 112, 99, 104, 78, 107, 111, 121, 118, 49, 87, 104, 107, 70, 71, 119, 43, 99, 71, 75, 112, 69, 99, 68, 74, 87, 56, 88, 65, 53, 120, 84, 65, 119, 84, 121, 97, 110, 118, 70, 76, 70, 111, 55, 116, 102, 110, 116, 65, 52, 83, 87, 65, 111, 120, 76, 65, 56, 97, 101, 56, 85, 47, 50, 57, 51, 87, 70, 50, 84, 74, 100, 47, 71, 89, 67, 68, 119, 61, 61, 34, 125, 44, 34, 80, 114, 105, 118, 97, 116, 101, 75, 101, 121, 34, 58, 34, 67, 75, 103, 118, 105, 121, 66, 120, 54, 112, 99, 104, 78, 107, 111, 121, 118, 49, 87, 104, 107, 70, 71, 119, 43, 99, 71, 75, 112, 69, 99, 68, 74, 87, 56, 88, 65, 53, 120, 84, 65, 119, 84, 121, 97, 110, 118, 70, 76, 70, 111, 55, 116, 102, 110, 116, 65, 52, 83, 87, 65, 111, 120, 76, 65, 56, 97, 101, 56, 85, 47, 50, 57, 51, 87, 70, 50, 84, 74, 100, 47, 71, 89, 67, 68, 119, 61, 61, 34, 44, 34, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 56, 109, 112, 55, 120, 83, 120, 97, 79, 55, 88, 53, 55, 81, 79, 69, 108, 103, 75, 77, 83, 119, 80, 71, 110, 118, 70, 80, 57, 118, 100, 49, 104, 100, 107, 121, 88, 102, 120, 109, 65, 103, 56, 61, 34, 44, 34, 84, 111, 107, 101, 110, 115, 34, 58, 49, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 125}

	var m mint.Mint
	err := json.Unmarshal([]byte(dat), &m)

	if err != nil {
		log.Fatal(err.Error())
	}
	return m
}

// processMintAndCreateAccounts will process genesis blocks and create mint account
func processMintAndCreateAccounts() (*books.Accounts, mint.Mint) {
	bm := books.NewBookManager()
	m := parseMint()
	blocks := m.CreateBlocks()

	log.Info("Create genesis blocks")

	aliceKP := m.KeyPair
	bm.CreateAccount(aliceKP.Public, m.Tokens)

	bm.AddValidVDFValue(blocks[0].Val)
	bm.AddValidVDFValue(blocks[1].Val)

	log.Info("Process genesis blocks")
	err := bm.ProcessBlocks(blocks[1:])
	if err != nil {
		log.Fatal(err.Error())
	}
	return bm, m
}

func TestMessaging(t *testing.T) {
	messagingCon, err := net.ListenPacket("udp", "127.0.0.1:50016")
	if err != nil {
		fmt.Println(err)
		log.Error("can't create ", zap.String("udp socket", network.MessagingAddrServer))
	}
	defer messagingCon.Close()
	responseCon, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		fmt.Println(err)
		log.Error("can't create ", zap.String("udp socket", network.MessagingAddrServer))
	}
	defer responseCon.Close()

	bm, m := processMintAndCreateAccounts()
	pipelines.Messaging(bm, messagingCon, responseCon)
	MessagingAddrServerUDP := net.UDPAddr{Port: 50016, IP: net.ParseIP("127.0.0.1")}
	fmt.Println(messagingCon)
	iUser := user.NewUserAPI(&MessagingAddrServerUDP, nil, responseCon, nil)
	bal, err := iUser.Balance(m.PublicKey)
	if err != nil {
		t.Errorf("Something wrong with initial balance")

	}
	if bal != 1000000000000 {
		t.Errorf("Wrong initial balance")
	}
	if iUser.TransactionsTotal() != 1 {
		t.Errorf("Wrong initial transaction count")
	}

}
