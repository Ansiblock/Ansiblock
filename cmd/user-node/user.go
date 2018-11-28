package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/replication"
	"github.com/Ansiblock/Ansiblock/user"
	isatty "github.com/mattn/go-isatty"

	"github.com/Ansiblock/Ansiblock/log"
	"github.com/Ansiblock/Ansiblock/mint"
	"go.uber.org/zap"
)

const startingAmount = 1000000
const numAccounts = 100000
const mintAmount = 1000000000
const threads = 1
const blockReaderAddr = "0.0.0.0:50053"

func newTransactions(from *block.KeyPair, tos []block.KeyPair, vdf block.VDFValue, reverse string, amount int64) []block.Transaction {
	log.Info(fmt.Sprintf("Creating %v Transactions...", reverse))
	start := time.Now()
	transactions := make([]block.Transaction, len(tos))
	counter := make(chan bool)
	for i := 0; i < len(tos); i++ {
		go func(i int) {
			if reverse == "" {
				transactions[i] = block.NewTransaction(from, tos[i].Public, amount, 0, vdf)

			} else {
				transactions[i] = block.NewTransaction(&tos[i], from.Public, amount, 0, vdf)
			}
			counter <- true
		}(i)
	}
	for range transactions {
		<-counter
	}
	duration := time.Since(start)
	log.Info(fmt.Sprintf("Created and Signed %v %v transactions in %v ", startingAmount, reverse, duration))
	return transactions
}

// user experiment
// 1. creates 1000000 random Keypairs
// 2. creates and signs 1000000 transactions from mint to random public key
// 3. creates `thread` number of gorutines and sends transactions to the producer
// 4. Samples processed transaction count in every second
// 5. Outputs highest TPS
func newExperiment(transactionAddrServerUDP *net.UDPAddr, MessagingAddrServerUDP *net.UDPAddr,
	transactionsConn net.PacketConn, messagingCon net.PacketConn) {
	log.Info("Create an experiment...")
	cl := user.NewUserAPI(MessagingAddrServerUDP, transactionAddrServerUDP, messagingCon, transactionsConn)
	log.Info("Get last VDF...")
	validVDFValue := cl.ValidVDFValue()
	mint := parseMint()
	mintBal, _ := cl.Balance(mint.PublicKey)
	log.Info("Mint Balance on Producer: ", zap.Int64("balance", mintBal))
	// log.Info("Tokens in Mint", zap.Int64("", mint.Tokens))
	log.Info(fmt.Sprintf("Create %v keypairs", numAccounts))
	keypairs := block.KeyPairs(numAccounts)
	transactions := newTransactions(&mint.KeyPair, keypairs, validVDFValue, "", 1)
	firstCount := cl.TransactionsTotal()
	log.Info("Initial count", zap.Uint64("", firstCount))
	log.Info(fmt.Sprintf("Transfering %v transactionsin %v batches", len(transactions), threads))
	for i := 0; i < threads; i++ {
		tranConn, err := net.ListenPacket("udp", "0.0.0.0:0")
		if err != nil {
			log.Error("can't create udp socket")
		}
		reqConn, err := net.ListenPacket("udp", "0.0.0.0:0")
		if err != nil {
			log.Error("can't create udp socket")
		}
		iUser := user.NewUserAPI(MessagingAddrServerUDP, transactionAddrServerUDP, reqConn, tranConn)
		go func(index int, iUser *user.API) {
			// time.Sleep(time.Millisecond * time.Duration(index*10))
			for j := 0; j < len(transactions)/threads; j++ {
				go iUser.TransferTransaction(transactions[threads*j+index])
			}
		}(i, iUser)
	}

	log.Info("Sampling tsp every second....")

	start := time.Now()
	maxTPS := 0.0
	for i := 0; i < 1000; i++ {
		trCount := cl.TransactionsTotal()
		duration := time.Since(start)
		start = time.Now()
		count := trCount - firstCount
		firstCount = trCount
		log.Info(fmt.Sprintf("Transactions processed %v", count))
		log.Info(fmt.Sprintf("==Total Transactions processed %v ==", trCount))
		tps := float64(count) / duration.Seconds()
		if maxTPS < tps {
			maxTPS = tps
		}
		log.Info("TPS: ", zap.Float64("", tps))
		if trCount == numAccounts || (count == 0 && i > 100) {
			break
		}
		time.Sleep(1 * time.Second)
	}

	log.Info(fmt.Sprintf("Highest TPS: %v", maxTPS))
}

func infiniteExperiment4(transactionAddrServerUDP *net.UDPAddr, MessagingAddrServerUDP *net.UDPAddr,
	transactionsConn net.PacketConn, messagingCon net.PacketConn) {
	log.Info("Create an Infinite experiment...")
	cl := user.NewUserAPI(MessagingAddrServerUDP, transactionAddrServerUDP, messagingCon, transactionsConn)
	log.Info("Get last VDF...")
	mint := parseMint()
	mintBal, _ := cl.Balance(mint.PublicKey)
	log.Info("Mint Balance on Producer: ", zap.Int64("balance", mintBal))
	validVDFValue := cl.ValidVDFValue()
	fmt.Printf("validVDFValue : %v\n", validVDFValue)

	tranConn, err := net.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		log.Error("can't create udp socket")
	}
	reqConn, err := net.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		log.Error("can't create udp socket")
	}
	iUser := user.NewUserAPI(MessagingAddrServerUDP, transactionAddrServerUDP, reqConn, tranConn)
	// time.Sleep(time.Millisecond * time.Duration(index*10))

	for {

		log.Info(fmt.Sprintf("Create %v keypairs", 64*1024))
		keypairs := block.KeyPairs(64 * 1024)
		// validVDFValue := cl.ValidVDFValue()

		transactions := newTransactions(&mint.KeyPair, keypairs, validVDFValue, "", 1)
		log.Info(fmt.Sprintf("Transfering %v transactions in %v batches", len(transactions), threads))
		post := make(chan bool, 10)
		for j := 0; j < len(transactions); j++ {
			go func(j int) {
				iUser.TransferTransaction(transactions[j])
				post <- true
			}(j)
		}
		for j := 0; j < len(transactions); j++ {
			<-post
		}
		// time.Sleep(1 * time.Second)
		// validVDFValue = cl.ValidVDFValue()
		// transactions = newTransactions(&mint.KeyPair, keypairs, validVDFValue, "reverse")
		// log.Info(fmt.Sprintf("Transfering %v reverse transactions in %v batches", len(transactions), threads))
		// for j := 0; j < len(transactions); j++ {
		// 	go func(j int) {
		// 		iUser.TransferTransaction(transactions[j])
		// 		post <- true
		// 	}(j)
		// }
		// for j := 0; j < len(transactions); j++ {
		// 	<-post
		// }
		// time.Sleep(1 * time.Second)
	}
}

func infiniteExperiment2(transactionAddrServerUDP *net.UDPAddr, MessagingAddrServerUDP *net.UDPAddr,
	transactionsConn net.PacketConn, messagingCon net.PacketConn) {
	log.Info("Create an Infinite experiment...")
	cl := user.NewUserAPI(MessagingAddrServerUDP, transactionAddrServerUDP, messagingCon, transactionsConn)
	log.Info("Get last VDF...")
	mint := parseMint()
	mintBal, _ := cl.Balance(mint.PublicKey)
	log.Info("Mint Balance on Producer: ", zap.Int64("balance", mintBal))
	validVDFValue := cl.ValidVDFValue()
	fmt.Printf("validVDFValue : %v\n", validVDFValue)

	tranConn, err := net.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		log.Error("can't create udp socket")
	}
	reqConn, err := net.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		log.Error("can't create udp socket")
	}
	iUser := user.NewUserAPI(MessagingAddrServerUDP, transactionAddrServerUDP, reqConn, tranConn)
	// time.Sleep(time.Millisecond * time.Duration(index*10))
	log.Info(fmt.Sprintf("Create %v keypairs", numAccounts))
	keypairs := block.KeyPairs(numAccounts)

	transactions := newTransactions(&mint.KeyPair, keypairs, validVDFValue, "", numAccounts)
	log.Info(fmt.Sprintf("Transfering %v transactions in %v batches", len(transactions), threads))
	post := make(chan bool, 10)
	for j := 0; j < len(transactions); j++ {
		go func(j int) {
			iUser.TransferTransaction(transactions[j])
			post <- true
		}(j)
	}
	for j := 0; j < len(transactions); j++ {
		<-post
	}

	for {
		// validVDFValue := cl.ValidVDFValue()

		transactions := newTransactions(&mint.KeyPair, keypairs, validVDFValue, "", 1)

		log.Info(fmt.Sprintf("Transfering %v transactions in %v batches", len(transactions), threads))
		post := make(chan bool, 10)
		for j := 0; j < len(transactions); j++ {
			go func(j int) {
				iUser.TransferTransaction(transactions[j])
				post <- true
			}(j)
		}
		for j := 0; j < len(transactions); j++ {
			<-post
		}
		// time.Sleep(1 * time.Second)
		// validVDFValue = cl.ValidVDFValue()
		transactions = newTransactions(&mint.KeyPair, keypairs, validVDFValue, "reverse", 1)
		log.Info(fmt.Sprintf("Transfering %v reverse transactions in %v batches", len(transactions), threads))
		for j := 0; j < len(transactions); j++ {
			go func(j int) {
				iUser.TransferTransaction(transactions[j])
				post <- true
			}(j)
		}
		for j := 0; j < len(transactions); j++ {
			<-post
		}
		// time.Sleep(1 * time.Second)
	}
}

func infiniteExperiment3(transactionAddrServerUDP *net.UDPAddr, MessagingAddrServerUDP *net.UDPAddr,
	transactionsConn net.PacketConn, messagingCon net.PacketConn) {
	log.Info("Create an Infinite experiment...")
	cl := user.NewUserAPI(MessagingAddrServerUDP, transactionAddrServerUDP, messagingCon, transactionsConn)
	log.Info("Get last VDF...")
	mint := parseMint()
	mintBal, _ := cl.Balance(mint.PublicKey)
	log.Info("Mint Balance on Producer: ", zap.Int64("balance", mintBal))
	validVDFValue := cl.ValidVDFValue()
	fmt.Printf("validVDFValue : %v\n", validVDFValue)

	tranConn, err := net.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		log.Error("can't create udp socket")
	}
	reqConn, err := net.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		log.Error("can't create udp socket")
	}
	iUser := user.NewUserAPI(MessagingAddrServerUDP, transactionAddrServerUDP, reqConn, tranConn)
	log.Info(fmt.Sprintf("Create %v keypairs", numAccounts))
	keypairs := block.KeyPairs(numAccounts)

	vdfs := make(chan []byte, 1)
	go func() {
		for {
			validVDFValue := cl.ValidVDFValue()
			vdfs <- validVDFValue
		}
	}()
	oldVDF := []byte{1}
	post := make(chan bool, 10)

	for {
		for {
			vdf := <-vdfs
			if !bytes.Equal(oldVDF, vdf) {
				oldVDF = vdf
				fmt.Println("New Batch")
				break
			}
		}
		for _, pair := range keypairs {
			go func(pair block.KeyPair) {
				transaction := block.NewTransaction(&mint.KeyPair, pair.Public, 1, 0, oldVDF)
				iUser.TransferTransaction(transaction)
				post <- true
			}(pair)
		}
		for j := 0; j < len(keypairs); j++ {
			<-post
		}
	}

}

func parseProducerJSON() replication.Node {
	if isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		log.Fatal("Expected json file as stdin")
	}
	input := bufio.NewScanner(os.Stdin)
	var data string
	for input.Scan() {
		data = data + input.Text()
	}

	data = strings.TrimSpace(data)
	if len(data) == 0 {
		log.Fatal("Empty file, expected json")
	}
	var objMap map[string]*json.RawMessage
	err := json.Unmarshal([]byte(data), &objMap)
	checkErr(err)

	var producer replication.Node

	err = json.Unmarshal(*objMap["Data"], &producer.Data)
	checkErr(err)

	return producer
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	log.Init()

	if len(os.Args) != 1 {
		log.Fatal("Usage: cat producer.json | go run server.go")
	}
	producer := parseProducerJSON()

	transactionAddrServerUDP := producer.Data.Addresses.Transaction
	MessagingAddrServerUDP := producer.Data.Addresses.Message

	transactionsConn, err := net.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		log.Error("can't create udp socket")
	}

	messagingCon, err := net.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		log.Error("can't create udp socket")
	}

	infiniteExperiment3(&transactionAddrServerUDP, &MessagingAddrServerUDP, transactionsConn, messagingCon)

}

func parseMint() mint.Mint {

	// dat := []byte{123, 34, 75, 101, 121, 80, 97, 105, 114, 34, 58, 123, 34, 80, 117, 98, 108, 105, 99, 34, 58, 34, 66, 87, 84, 85, 71, 48, 55, 89, 112, 65, 67, 50, 66, 73, 43, 98, 97, 122, 74, 69, 69, 97, 115, 104, 71, 72, 67, 66, 54, 120, 108, 53, 66, 105, 74, 108, 81, 118, 66, 51, 88, 75, 107, 61, 34, 44, 34, 80, 114, 105, 118, 97, 116, 101, 34, 58, 34, 69, 106, 49, 86, 120, 117, 110, 55, 81, 85, 66, 107, 90, 76, 90, 120, 54, 113, 86, 115, 109, 114, 51, 114, 101, 110, 119, 87, 109, 70, 84, 78, 87, 87, 48, 66, 110, 106, 89, 55, 107, 57, 89, 70, 90, 78, 81, 98, 84, 116, 105, 107, 65, 76, 89, 69, 106, 53, 116, 114, 77, 107, 81, 82, 113, 121, 69, 89, 99, 73, 72, 114, 71, 88, 107, 71, 73, 109, 86, 67, 56, 72, 100, 99, 113, 81, 61, 61, 34, 125, 44, 34, 80, 114, 105, 118, 97, 116, 101, 75, 101, 121, 34, 58, 34, 69, 106, 49, 86, 120, 117, 110, 55, 81, 85, 66, 107, 90, 76, 90, 120, 54, 113, 86, 115, 109, 114, 51, 114, 101, 110, 119, 87, 109, 70, 84, 78, 87, 87, 48, 66, 110, 106, 89, 55, 107, 57, 89, 70, 90, 78, 81, 98, 84, 116, 105, 107, 65, 76, 89, 69, 106, 53, 116, 114, 77, 107, 81, 82, 113, 121, 69, 89, 99, 73, 72, 114, 71, 88, 107, 71, 73, 109, 86, 67, 56, 72, 100, 99, 113, 81, 61, 61, 34, 44, 34, 80, 117, 98, 108, 105, 99, 75, 101, 121, 34, 58, 34, 66, 87, 84, 85, 71, 48, 55, 89, 112, 65, 67, 50, 66, 73, 43, 98, 97, 122, 74, 69, 69, 97, 115, 104, 71, 72, 67, 66, 54, 120, 108, 53, 66, 105, 74, 108, 81, 118, 66, 51, 88, 75, 107, 61, 34, 44, 34, 84, 111, 107, 101, 110, 115, 34, 58, 49, 48, 48, 48, 48, 48, 48, 125}
	dat := []byte{123, 34, 75, 101, 121, 80, 97, 105, 114, 34, 58, 123, 34, 80, 117, 98, 108, 105,
		99, 34, 58, 34, 56, 109, 112, 55, 120, 83, 120, 97, 79, 55, 88, 53, 55, 81, 79, 69, 108, 103,
		75, 77, 83, 119, 80, 71, 110, 118, 70, 80, 57, 118, 100, 49, 104, 100, 107, 121, 88, 102, 120,
		109, 65, 103, 56, 61, 34, 44, 34, 80, 114, 105, 118, 97, 116, 101, 34, 58, 34, 67, 75, 103, 118,
		105, 121, 66, 120, 54, 112, 99, 104, 78, 107, 111, 121, 118, 49, 87, 104, 107, 70, 71, 119, 43,
		99, 71, 75, 112, 69, 99, 68, 74, 87, 56, 88, 65, 53, 120, 84, 65, 119, 84, 121, 97, 110, 118, 70,
		76, 70, 111, 55, 116, 102, 110, 116, 65, 52, 83, 87, 65, 111, 120, 76, 65, 56, 97, 101, 56, 85,
		47, 50, 57, 51, 87, 70, 50, 84, 74, 100, 47, 71, 89, 67, 68, 119, 61, 61, 34, 125, 44, 34, 80,
		114, 105, 118, 97, 116, 101, 75, 101, 121, 34, 58, 34, 67, 75, 103, 118, 105, 121, 66, 120, 54,
		112, 99, 104, 78, 107, 111, 121, 118, 49, 87, 104, 107, 70, 71, 119, 43, 99, 71, 75, 112, 69, 99,
		68, 74, 87, 56, 88, 65, 53, 120, 84, 65, 119, 84, 121, 97, 110, 118, 70, 76, 70, 111, 55, 116,
		102, 110, 116, 65, 52, 83, 87, 65, 111, 120, 76, 65, 56, 97, 101, 56, 85, 47, 50, 57, 51, 87,
		70, 50, 84, 74, 100, 47, 71, 89, 67, 68, 119, 61, 61, 34, 44, 34, 80, 117, 98, 108, 105, 99,
		75, 101, 121, 34, 58, 34, 56, 109, 112, 55, 120, 83, 120, 97, 79, 55, 88, 53, 55, 81, 79, 69,
		108, 103, 75, 77, 83, 119, 80, 71, 110, 118, 70, 80, 57, 118, 100, 49, 104, 100, 107, 121, 88,
		102, 120, 109, 65, 103, 56, 61, 34, 44, 34, 84, 111, 107, 101, 110, 115, 34, 58, 49, 48, 48, 48,
		48, 48, 48, 48, 48, 48, 48, 48, 48, 125}
	var m mint.Mint
	err := json.Unmarshal([]byte(dat), &m)

	if err != nil {
		log.Fatal(err.Error())
	}
	return m
}
