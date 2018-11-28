package pipelines

import (
	"encoding/json"
	"fmt"

	"github.com/Ansiblock/Ansiblock/api"
	"github.com/Ansiblock/Ansiblock/books"
	"github.com/Ansiblock/Ansiblock/log"
	"github.com/Ansiblock/Ansiblock/mint"
	"github.com/Ansiblock/Ansiblock/replication"
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
func processMintAndCreateAccounts() (*books.Accounts, mint.Mint, uint64) {
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
	return bm, m, 2
}

func producerNodeHelper(producer replication.Node, db api.DataBase) (*books.Accounts, *replication.Sync, mint.Mint) {
	bm, m, startingBlocksTotal := processMintAndCreateAccounts()
	producer.Data.Producer = producer.Data.Self
	sync, _ := replication.NewSync(producer.Data)
	go Messaging(bm, producer.Sockets.Messages, producer.Sockets.Respond)
	go Synchronization(sync, producer.Sockets.Sync, producer.Sockets.SyncSend)
	go BlockGenerationFaster(bm, sync, producer.Sockets.Transaction, producer.Sockets.Replicate, producer.Sockets.Repair, startingBlocksTotal, db)
	// log.Debug(fmt.Sprintf("Producer Node: %v", producer.Data.Addresses))
	log.Debug(fmt.Sprintf("Producer Node:\n blocks: %v\n messages: %v\n Replicate: %v\n Transport: %v\n",
		producer.Sockets.Transaction.LocalAddr().String(),
		producer.Sockets.Messages.LocalAddr().String(),
		producer.Sockets.Replicate.LocalAddr().String(),
		producer.Sockets.Transport.LocalAddr().String()))

	return bm, sync, m
}

// ProducerNodeWithServer is responsible createing producer node and run server on it
func ProducerNodeWithServer(producer replication.Node) {
	db := api.NewDBConnection(api.DBFilename)
	bm, sync, mint := producerNodeHelper(producer, db)
	blockchainAPI := api.New(bm, db, sync, &mint)
	api.RunRestAPI(blockchainAPI)
	ch := make(chan bool)
	<-ch
}

// ProducerNode is responsible creating producer node
func ProducerNode(producer replication.Node) {
	producerNodeHelper(producer, nil)
	ch := make(chan bool)
	<-ch
}

// SignerNode is responsible creating signer
func SignerNode(producer replication.Node, name string) {
	bm, _, _ := processMintAndCreateAccounts()
	node := replication.NewNode("signer", name)
	log.Info(fmt.Sprintf(" ==== Signer %v: %v ==== \n", name, node.Sockets.Messages.LocalAddr().String()))
	node.Data.Producer = producer.Data.Self
	sync, _ := replication.NewSync(node.Data)
	sync.Insert(producer.Data)
	go Messaging(bm, node.Sockets.Messages, node.Sockets.Respond)
	go Synchronization(sync, node.Sockets.Sync, node.Sockets.SyncSend)
	go BlockSigner(bm, sync, node.Sockets.Replicate, node.Sockets.Repair, node.Sockets.Transport, nil)
	// log.Debug(fmt.Sprintf("Signer Node: %v", node.Data.Addresses))
	log.Debug(fmt.Sprintf("Signer Node:\n blocks: %v\n messages: %v\n Replicate: %v\n Transport: %v\n",
		node.Sockets.Transaction.LocalAddr().String(),
		node.Sockets.Messages.LocalAddr().String(),
		node.Sockets.Replicate.LocalAddr().String(),
		node.Sockets.Transport.LocalAddr().String()))
	ch := make(chan bool)
	<-ch
}

// ServerNode is responsible creating server node
func ServerNode(producer replication.Node, name string) {
	bm, mint, _ := processMintAndCreateAccounts()
	node := replication.NewNode("server", name)
	node.Data.Producer = producer.Data.Self
	sync, _ := replication.NewSync(node.Data)
	sync.Insert(producer.Data)

	db := api.NewDBConnection(api.DBFilename)

	go Synchronization(sync, node.Sockets.Sync, node.Sockets.SyncSend)
	go BlockSigner(bm, sync, node.Sockets.Replicate, node.Sockets.Repair, node.Sockets.Transport, db)

	blockchainAPI := api.New(bm, db, sync, &mint)
	api.RunRestAPI(blockchainAPI)
	ch := make(chan bool)
	<-ch
}
