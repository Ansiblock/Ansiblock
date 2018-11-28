package pipelines_test

import (
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/Ansiblock/Ansiblock/mint"
	"github.com/Ansiblock/Ansiblock/user"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/books"
	"github.com/Ansiblock/Ansiblock/log"
	"github.com/Ansiblock/Ansiblock/network"
	"github.com/Ansiblock/Ansiblock/pipelines"
	"github.com/Ansiblock/Ansiblock/replication"
)

// creates socket mock from which you can read blobs passed as argument
func createSignerSocketMock(blobs *network.Blobs) *network.SocketMock {
	sm := network.NewSocketMock(nil, nil, nil)
	for _, b := range blobs.Bs {
		sm.AddToReadBuff(b.Data[:b.Size])
	}

	return sm
}

// runs Synchronization pipelines for producer node and numNodes signer nodes
func runSynchronizationPipelines(numNodes int) ([]replication.Node, []*replication.Sync) {
	nodeList := make([]replication.Node, numNodes)
	syncList := make([]*replication.Sync, numNodes)
	producer := replication.NewNode("producer", "test1")
	producer.Data.Producer = producer.Data.Self
	syncL, _ := replication.NewSync(producer.Data)
	go pipelines.Synchronization(syncL, producer.Sockets.Sync, producer.Sockets.SyncSend)
	for i := 0; i < numNodes; i++ {
		node := replication.NewNode("signer", "test"+strconv.Itoa(i))
		node.Data.Producer = producer.Data.Self
		syncV, _ := replication.NewSync(node.Data)
		syncV.Insert(producer.Data)
		go pipelines.Synchronization(syncV, node.Sockets.Sync, node.Sockets.SyncSend)
		nodeList[i] = node
		syncList[i] = syncV
	}

	return nodeList, syncList
}

// waits to give Synchronization pipelines some time for sync
func waitForNodeSync(sync *replication.Sync, numNodes uint64) {
	for i := 0; i < 30; i++ {
		num := sync.ConnectedNodes()
		fmt.Println(num)
		if num == numNodes {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// func TestSigner(t *testing.T) {
// 	rand.Seed(0)
// 	numSignerNodes := 2
// 	bm, keyPairs := RandomAccounts(100)
// 	blocks := RandomTransactionsBlocks(bm, 2, 5, keyPairs)
// 	blobsOriginal := block.BlocksToBlobs(blocks)

// 	replicateConn := createSignerSocketMock(blobsOriginal)

// 	nodeList, syncList := runSyncPipelines(numSignerNodes)
// 	fmt.Printf("Repl Addr1: %v \nRepl Addr2: %v \n", nodeList[0].Data.Addresses.Replication, nodeList[1].Data.Addresses.Replication)

// 	//wait until node sync info between each other
// 	time.Sleep(2 * time.Second)
// 	transportConn, err := net.ListenPacket("udp", "0.0.0.0:0")
// 	if err != nil {
// 		t.Errorf("Failed open transport connection \n")
// 	}
// 	waitForNodeSync(syncList[0], uint64(numSignerNodes))

// 	bmList := make([]*books.Accounts, numSignerNodes)
// 	for i := 0; i < numSignerNodes; i++ {
// 		bmList[i] = bm.Clone()
// 	}

// 	for i := 1; i < numSignerNodes; i++ {
// 		transportConn, _ := net.ListenPacket("udp", "0.0.0.0:0")
// 		if err != nil {
// 			t.Errorf("Failed open transport connection \n")
// 		}
// 		go pipelines.Signer(bmList[i], syncList[i], transportConn, nodeList[i].Sockets.Replicate)
// 	}
// 	time.Sleep(1 * time.Second) //give some time to other nodes, to run signer pipelines
// 	go pipelines.Signer(bmList[0], syncList[0], transportConn, replicateConn)

// 	bm.ProcessBlocks(blocks)
// 	time.Sleep(5 * time.Second) //give some time to retrnasmiter
// 	for i := 0; i < numSignerNodes; i++ {
// 		if !bm.Equals(bmList[i]) {
// 			t.Errorf("Failed signer or retramsmition \n")
// 		}
// 	}
// }

func ProducerNodeForTest(producer replication.Node, bm *books.Accounts) {
	producer.Data.Producer = producer.Data.Self
	sync, _ := replication.NewSync(producer.Data)
	go pipelines.Messaging(bm, producer.Sockets.Messages, producer.Sockets.Respond)
	go pipelines.Synchronization(sync, producer.Sockets.Sync, producer.Sockets.SyncSend)
	go pipelines.BlockGeneration(bm, sync, producer.Sockets.Transaction, producer.Sockets.Replicate, nil, 2, nil)
	fmt.Printf("Producer Node: %v", producer.Data.Addresses)
	fmt.Printf("Producer Node:\n Transaction: %v\n Messages: %v\n Replicate: %v\n Transport: %v\n",
		producer.Sockets.Transaction.LocalAddr().String(),
		producer.Sockets.Messages.LocalAddr().String(),
		producer.Sockets.Replicate.LocalAddr().String(),
		producer.Sockets.Transport.LocalAddr().String())
	ch := make(chan bool)
	<-ch
}

func SignerNodeForTest(producer replication.Node, bm *books.Accounts) {
	node := replication.NewNode("signer", "test")
	fmt.Printf("signer message: %v\n", node.Sockets.Messages.LocalAddr().String())
	node.Data.Producer = producer.Data.Self
	sync, _ := replication.NewSync(node.Data)
	sync.Insert(producer.Data)
	go pipelines.Messaging(bm, node.Sockets.Messages, node.Sockets.Respond)
	go pipelines.Synchronization(sync, node.Sockets.Sync, node.Sockets.SyncSend)
	go pipelines.BlockSigner(bm, sync, node.Sockets.Replicate, node.Sockets.Repair, node.Sockets.Transport, nil)
	fmt.Printf("Signer Node: %v", node.Data.Addresses)
	fmt.Printf("Producer Node:\n Transaction: %v\n Messages: %v\n Replicate: %v\n Transport: %v\n",
		node.Sockets.Transaction.LocalAddr().String(),
		node.Sockets.Messages.LocalAddr().String(),
		node.Sockets.Replicate.LocalAddr().String(),
		node.Sockets.Transport.LocalAddr().String())
	ch := make(chan bool)
	<-ch
}

func createTransactions(from *block.KeyPair, tos []block.KeyPair, vdf block.VDFValue) []block.Transaction {
	transactions := make([]block.Transaction, numAccounts)
	for i := 0; i < len(tos); i++ {
		transactions[i] = block.NewTransaction(from, tos[i].Public, 1, 0, vdf)
	}
	return transactions
}

func experiment(transactionAddrServerUDP *net.UDPAddr, MessagingAddrServerUDP *net.UDPAddr,
	transactionsConn net.PacketConn, messagingCon net.PacketConn, mint *mint.Mint) {
	us := user.NewUserAPI(MessagingAddrServerUDP, transactionAddrServerUDP, messagingCon, transactionsConn)
	validVDFValue := us.ValidVDFValue()
	keypairs := block.KeyPairs(numAccounts)
	transactions := createTransactions(&mint.KeyPair, keypairs, validVDFValue)
	// time.Sleep(time.Millisecond * time.Duration(index*10))
	for j := 0; j < len(transactions); j++ {
		us.TransferTransaction(transactions[j])
	}

}

var numAccounts = 50

func TestSigner2(t *testing.T) {
	bmL, m := processMintAndCreateAccounts()
	bmV1, _ := processMintAndCreateAccounts()
	bmV2, _ := processMintAndCreateAccounts()

	producer := replication.NewNode("signer", "test")
	producer.Data.Producer = producer.Data.Self
	fmt.Printf("Producer transactions: %v\nProducer messages: %v\n", producer.Sockets.Transaction.LocalAddr().String(), producer.Sockets.Messages.LocalAddr().String())
	go ProducerNodeForTest(producer, bmL)
	go SignerNodeForTest(producer, bmV1)
	go SignerNodeForTest(producer, bmV2)
	transactionAddrServerUDP := net.UDPAddr{Port: producer.Data.Addresses.Transaction.Port, IP: producer.Data.Addresses.Transaction.IP}
	MessagingAddrServerUDP := net.UDPAddr{Port: producer.Data.Addresses.Message.Port, IP: producer.Data.Addresses.Message.IP}

	transactionsConn, err := net.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		log.Error("can't create udp socket")
	}

	messagingCon, err := net.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		log.Error("can't create udp socket")
	}
	time.Sleep(5 * time.Second)
	go experiment(&transactionAddrServerUDP, &MessagingAddrServerUDP, transactionsConn, messagingCon, &m)
	fmt.Printf("Transactions Total bml=%v, bmV1=%v, bmV2=%v \n", bmL.TransactionsTotal(), bmV1.TransactionsTotal(), bmV2.TransactionsTotal())
	fmt.Printf("Block Total bml=%v, bmV1=%v, bmV2=%v\n", bmL.BlocksTotal(), bmV1.BlocksTotal(), bmV2.BlocksTotal())
	for i := 0; i < 10; i++ {
		fmt.Printf("Transaction Count <%v> : \n", i)
		fmt.Printf("Transactions Total bml=%v, bmV1=%v, bmV2=%v \n", bmL.TransactionsTotal(), bmV1.TransactionsTotal(), bmV2.TransactionsTotal())
		fmt.Printf("Block Total bml=%v, bmV1=%v, bmV2=%v\n", bmL.BlocksTotal(), bmV1.BlocksTotal(), bmV2.BlocksTotal())
		time.Sleep(1 * time.Second)
	}

	// if !bmL.Equals(bmV1) || !bmL.Equals(bmV2) {
	// 	t.Error("error")
	// }
}
