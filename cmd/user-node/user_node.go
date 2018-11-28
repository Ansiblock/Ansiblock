package main

/*
import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	isatty "github.com/mattn/go-isatty"
	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/user"

	"github.com/Ansiblock/Ansiblock/books"
	"github.com/Ansiblock/Ansiblock/log"
	"github.com/Ansiblock/Ansiblock/mint"
	"go.uber.org/zap"
)

const startingAmount = 1000000
const numAccounts = 1000000
const mintAmount = 1000000000
const threads = 1
const blockReaderAddr = "0.0.0.0:50053"

func parseMintJSON() mint.Mint {
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

	var m mint.Mint
	err := json.Unmarshal([]byte(data), &m)
	if err != nil {
		log.Fatal(err.Error())
	}

	return m
}

// startTransactionsSender receives transactions from channel and sends them to producer node
func startTransactionsSender(transactionReceiver <-chan block.Transactions,
	transactionAddrServerUDP *net.UDPAddr) {

	transactionsConn, err := net.ListenPacket("udp", "0.0.0.0")
	if err != nil {
		log.Error("can't create udp socket")
	}

	go func(transactionReceiver <-chan block.Transactions, conn net.PacketConn) {
		for {
			transactions, ok := <-transactionReceiver
			if !ok {
				log.Error("user-node > transaction sender's receiver failed, closing channel")
				return
			}
			log.Info(fmt.Sprintf("user-node > transaction sender: %v transactions received from chanel", len(transactions.Ts)))

			//TODO: send packets to server
			packets := transactions.ToPackets(transactionAddrServerUDP)
			packets.WriteTo(transactionsConn)
			log.Info(fmt.Sprintf("user-node > %v transactions has been successfully sent to server", len(transactions.Ts)))
		}
	}(transactionReceiver, transactionsConn)
}

// startBlocksReader receives blocks from producer node and prints
func startBlocksReader() {
	blockConn, err := net.ListenPacket("udp", blockReaderAddr)
	if err != nil {
		log.Error("can't create udp socket")
	}

	for {
		var b [1024]byte
		fmt.Println("Try to read block")
		n, _, err := blockConn.ReadFrom(b[:])
		fmt.Println("block read")
		if err != nil {
			log.Error("user-node > error while reading blocks")
		}
		log.Info("user-node > read", zap.Int("bytes", n))
		var block block.Block
		err = json.Unmarshal(b[:], &block)
		if err != nil {
			log.Error("user-node > can not unmarshal block")

		}
	}
}

func createTransactions(from *block.KeyPair, tos []block.KeyPair, vdf block.VDFValue) []block.Transaction {
	log.Info("Creating Transactions...")
	start := time.Now()

	transactions := make([]block.Transaction, len(tos))
	counter := make(chan bool)
	for i := 0; i < len(tos); i++ {
		go func(i int) {
			transactions[i] = block.NewTransaction(from, tos[i].Public, 1, 0, vdf)
			counter <- true
		}(i)
	}
	for range transactions {
		<-counter
	}
	duration := time.Since(start)
	log.Info(fmt.Sprintf("Created and Signed %v transactions in %v ", startingAmount, duration))
	return transactions
}

func createReverseTransactions(from *block.KeyPair, tos []block.KeyPair, vdf block.VDFValue) []block.Transaction {
	log.Info("Creating Reverse Transactions...")
	start := time.Now()
	transactions := make([]block.Transaction, len(tos))
	counter := make(chan bool)
	for i := 0; i < len(tos); i++ {
		go func(i int) {
			transactions[i] = block.NewTransaction(&tos[i], from.Public, 1, 0, vdf)
			counter <- true
		}(i)
	}
	for range transactions {
		<-counter
	}
	duration := time.Since(start)
	log.Info(fmt.Sprintf("Created and Signed %v reverse transactions in %v ", startingAmount, duration))
	return transactions
}

// user experiment
// 1. creates 1000000 random Keypairs
// 2. creates and signs 1000000 transactions from mint to random public key
// 3. creates `thread` number of gorutines and sends transactions to the producer
// 4. Samples processed transaction count in every second
// 5. Outputs highest TPS
func experiment(transactionAddrServerUDP *net.UDPAddr, MessagingAddrServerUDP *net.UDPAddr,
	transactionsConn net.PacketConn, messagingCon net.PacketConn) {
	log.Info("Create an experiment...")
	cl := user.NewUserAPI(MessagingAddrServerUDP, transactionAddrServerUDP, messagingCon, transactionsConn)
	log.Info("Get last VDF...")
	validVDFValue := cl.ValidVDFValue()
	mint := parseMintJSON()
	mintBal, _ := cl.Balance(mint.PublicKey)
	log.Info("Mint Balance on Producer: ", zap.Int64("balance", mintBal))
	// log.Info("Tokens in Mint", zap.Int64("", mint.Tokens))
	log.Info(fmt.Sprintf("Create %v keypairs", numAccounts))
	keypairs := block.KeyPairs(numAccounts)
	transactions := createTransactions(&mint.KeyPair, keypairs, validVDFValue)
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
		if trCount == numAccounts || (count == 0 && i > 10) {
			break
		}
		time.Sleep(1 * time.Second)
	}

	log.Info(fmt.Sprintf("Highest TPS: %v", maxTPS))
}

func experiment2(transactionAddrServerUDP *net.UDPAddr, MessagingAddrServerUDP *net.UDPAddr,
	transactionsConn net.PacketConn, messagingCon net.PacketConn) {
	mint := parseMintJSON()
	cl := user.NewUserAPI(MessagingAddrServerUDP, transactionAddrServerUDP, messagingCon, transactionsConn)
	validVDFValue := cl.ValidVDFValue()
	fmt.Printf("validVDFValue %v\n", validVDFValue)

	mintBal, _ := cl.Balance(mint.PublicKey)
	log.Info("Mint: ", zap.Int64("balance", mintBal))

	start := time.Now()

	transactions := make([]block.Transaction, startingAmount)
	for i := 0; i < startingAmount; i++ {
		randKey := block.NewKeyPair().Public
		transactions[i] = block.NewTransaction(&mint.KeyPair, randKey, 1, 0, validVDFValue)
	}
	duration := time.Since(start)
	log.Info(fmt.Sprintf("Created and Signed %v transactions in %v ", startingAmount, duration))

	start = time.Now()
	go func() {
		for _, tran := range transactions {
			cl.TransferTransaction(tran)
		}
	}()
	prevBal := int64(0)
	for {
		bal, _ := cl.Balance(mint.PublicKey)
		if bal == 0 || bal == prevBal {
			log.Info("Last", zap.Int64("Balance", bal))
			break
		}
		prevBal = bal
		log.Info("Current", zap.Int64("Balance", bal))
		time.Sleep(500 * time.Millisecond)
	}
	duration = time.Since(start)
	tps := float64(startingAmount-prevBal) / duration.Seconds()

	log.Info("Final", zap.Float64("tps", tps))
	log.Info("Summary time", zap.Float64("duration", duration.Seconds()))
	trCount := cl.TransactionsTotal()

	log.Info(fmt.Sprintf("==Total Transactions processed %v ==", trCount))

}

func infiniteExperiment(transactionAddrServerUDP *net.UDPAddr, MessagingAddrServerUDP *net.UDPAddr,
	transactionsConn net.PacketConn, messagingCon net.PacketConn) {
	log.Info("Create an Infinite experiment...")
	cl := user.NewUserAPI(MessagingAddrServerUDP, transactionAddrServerUDP, messagingCon, transactionsConn)
	log.Info("Get last VDF...")
	mint := parseMintJSON()
	mintBal, _ := cl.Balance(mint.PublicKey)
	log.Info("Mint Balance on Producer: ", zap.Int64("balance", mintBal))
	validVDFValue := cl.ValidVDFValue()
	fmt.Printf("validVDFValue : %v\n", validVDFValue)
	validVDFValue = cl.ValidVDFValue()
	fmt.Printf("validVDFValue : %v\n", validVDFValue)
	validVDFValue = cl.ValidVDFValue()
	fmt.Printf("validVDFValue : %v\n", validVDFValue)

	firstCount := cl.TransactionsTotal()
	log.Info("Initial count", zap.Uint64("", firstCount))
	tranConn, err := net.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		log.Error("can't create udp socket")
	}
	reqConn, err := net.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		log.Error("can't create udp socket")
	}
	iUser := user.NewUserAPI(MessagingAddrServerUDP, transactionAddrServerUDP, reqConn, tranConn)
	go func(iUser *user.API) {
		// time.Sleep(time.Millisecond * time.Duration(index*10))
		for {
			log.Info(fmt.Sprintf("Create %v keypairs", 100000))
			keypairs := block.KeyPairs(100000)
			validVDFValue := cl.ValidVDFValue()
			transactions := createTransactions(&mint.KeyPair, keypairs, validVDFValue)
			log.Info(fmt.Sprintf("Transfering %v transactions in %v batches", len(transactions), threads))

			for j := 0; j < len(transactions); j++ {
				go iUser.TransferTransaction(transactions[j])
			}
			// time.Sleep(1 * time.Second)
			validVDFValue = cl.ValidVDFValue()
			transactions = createReverseTransactions(&mint.KeyPair, keypairs, validVDFValue)
			log.Info(fmt.Sprintf("Transfering %v reverse transactions in %v batches", len(transactions), threads))
			for j := 0; j < len(transactions)/threads; j++ {
				go iUser.TransferTransaction(transactions[j])
			}
			// time.Sleep(1 * time.Second)

		}
	}(iUser)

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
		// if trCount == numAccounts || (count == 0 && i > 10) {
		// 	break
		// }
		time.Sleep(5 * time.Second)
	}

	log.Info(fmt.Sprintf("Highest TPS: %v", maxTPS))
}

func main() {
	log.Init()

	if len(os.Args) != 4 {
		log.Fatal("Usage: ./client server_IP < mint.json")
	}

	ip := net.ParseIP(os.Args[1])
	if ip == nil {
		log.Fatal("Illegal format of IP address")
	}
	tranP, _ := strconv.Atoi(os.Args[2])
	reqP, _ := strconv.Atoi(os.Args[3])

	transactionAddrServerUDP := net.UDPAddr{Port: tranP, IP: ip}
	MessagingAddrServerUDP := net.UDPAddr{Port: reqP, IP: ip}

	transactionsConn, err := net.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		log.Error("can't create udp socket")
	}

	messagingCon, err := net.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		log.Error("can't create udp socket")
	}

	experiment(&transactionAddrServerUDP, &MessagingAddrServerUDP, transactionsConn, messagingCon)
	// infiniteExperiment(&transactionAddrServerUDP, &MessagingAddrServerUDP, transactionsConn, messagingCon)

}
*/
