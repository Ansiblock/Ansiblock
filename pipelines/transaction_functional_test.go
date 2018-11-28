package pipelines_test

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/replication"
	"github.com/Ansiblock/Ansiblock/user"

	"github.com/Ansiblock/Ansiblock/log"
	"github.com/Ansiblock/Ansiblock/network"
	"github.com/Ansiblock/Ansiblock/pipelines"
	"go.uber.org/zap"
)

func TestTransactions(t *testing.T) {
	messagingCon, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		fmt.Println(err)
		log.Error("can't create ", zap.String("udp socket", network.MessagingAddrServer))
		return
	}
	defer messagingCon.Close()
	responseCon, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		fmt.Println(err)
		log.Error("can't create ", zap.String("udp socket", network.MessagingAddrServer))
		return
	}
	defer responseCon.Close()

	transactionCon, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		fmt.Println(err)
		log.Error("can't create ", zap.String("udp socket", network.MessagingAddrServer))
		return
	}
	defer transactionCon.Close()

	am, m := processMintAndCreateAccounts()
	go pipelines.Messaging(am, messagingCon, responseCon)
	fmt.Println("Messaging run")
	producer := replication.NewNode("producer", "test1")
	producer.Data.Producer = producer.Data.Self
	syncL, _ := replication.NewSync(producer.Data)
	go pipelines.BlockGeneration(am, syncL, transactionCon, nil, nil, 2, nil)
	fmt.Println("Transaction run")

	messagingCons := strings.Split(messagingCon.LocalAddr().String(), ":")
	reqPort, _ := strconv.Atoi(messagingCons[1])
	transactionCons := strings.Split(transactionCon.LocalAddr().String(), ":")
	transPort, _ := strconv.Atoi(transactionCons[1])
	MessagingAddrServerUDP := net.UDPAddr{Port: reqPort, IP: net.ParseIP(messagingCons[0])}
	transactionAddrServerUDP := net.UDPAddr{Port: transPort, IP: net.ParseIP(transactionCons[0])}
	fmt.Println(messagingCons)
	fmt.Println(reqPort)
	fmt.Println(transactionCons)
	fmt.Println(transPort)
	iUser := user.NewUserAPI(&MessagingAddrServerUDP, &transactionAddrServerUDP, responseCon, transactionCon)
	validVDFValue := iUser.ValidVDFValue()
	num := 10
	randUsers := block.KeyPairs(num)
	for _, user := range randUsers {
		tran := block.NewTransaction(&m.KeyPair, user.Public, 1, 0, validVDFValue)
		fmt.Printf("Transaction %v\n", tran)
		iUser.TransferTransaction(tran)
	}
	time.Sleep(3000 * time.Millisecond)
	for _, user := range randUsers {
		bal, err := iUser.Balance(user.Public)
		if err != nil {
			t.Errorf("Something wrong balance")
		}
		if bal != 1 {
			t.Errorf("Wrong balance")
		}
	}
	bal, err := iUser.Balance(m.PublicKey)
	if err != nil {
		t.Errorf("Something wrong balance")
	}
	if bal != int64(1000000000000-num) {
		t.Errorf("Wrong balance")
	}
	if iUser.TransactionsTotal() != 1+uint64(num) {
		t.Errorf("Wrong transaction count")
	}

	for _, user := range randUsers {
		tran := block.NewTransaction(&user, randUsers[0].Public, 1, 0, validVDFValue)
		fmt.Printf("Transaction %v\n", tran)
		iUser.TransferTransaction(tran)
	}

	time.Sleep(3000 * time.Millisecond)
	for i, user := range randUsers {
		ans := 0
		if i == 0 {
			ans = num
		}
		bal, err := iUser.Balance(user.Public)
		if err != nil {
			t.Errorf("Something wrong balance")
		}
		if bal != int64(ans) {
			t.Errorf("Wrong balance")
		}
	}
}
