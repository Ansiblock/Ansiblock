package api

import (
	"encoding/base64"
	"math/rand"
	"reflect"
	"testing"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/books"
	"github.com/Ansiblock/Ansiblock/network"
	"github.com/Ansiblock/Ansiblock/replication"
)

func Nodes(t *testing.T) {
	self := block.NewKeyPair().Public
	node := replication.NewNodeData(self, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	sync, _ := replication.NewSync(node)
	pk2 := block.NewKeyPair().Public
	node2 := replication.NewNodeData(pk2, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	sync.Insert(node2)
	pk3 := block.NewKeyPair().Public
	node3 := replication.NewNodeData(pk3, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	sync.Insert(node3)

	blockchainAPI := New(nil, nil, sync, nil)
	nodes := blockchainAPI.Nodes()
	if len(nodes) != 0 || !reflect.DeepEqual(nodes, sync.RemoteTableCopy()) {
		t.Errorf("Nodes method expected %v was %v", 3, len(nodes))
	}
}

func TestTransactionsTotal1(t *testing.T) {
	bm := books.NewBookManager()

	blockchainAPI := New(bm, nil, nil, nil)
	if blockchainAPI.TransactionsTotal() != bm.TransactionsTotal() {
		t.Errorf("Transaction count should be %v!", bm.TransactionsTotal())
	}
}

func TestTransactionsTotal2(t *testing.T) {
	rand.Seed(0)
	bm, keyPairs := books.RandomAccounts(100)
	blocks := books.RandomTransactionsBlocks(bm, 100, 10, keyPairs)
	bm.ProcessBlocks(blocks)

	blockchainAPI := New(bm, nil, nil, nil)
	if blockchainAPI.TransactionsTotal() != bm.TransactionsTotal() || bm.TransactionsTotal() != 936 {
		t.Errorf("Transaction count should be %v!", bm.TransactionsTotal())
	}
}

func TestBlocksTotal(t *testing.T) {
	bm, keyPairs := books.RandomAccounts(100)
	blocks := books.RandomTransactionsBlocks(bm, 100, 10, keyPairs)
	bm.ProcessBlocks(blocks)

	blockchainAPI := New(bm, nil, nil, nil)
	if blockchainAPI.BlocksTotal() != 10 {
		t.Errorf("Transaction count should be %v!", 10)
	}
}

func TestBalances(t *testing.T) {
	bm, keyPairs := books.RandomAccounts(100)
	blocks := books.RandomTransactionsBlocks(bm, 100, 10, keyPairs)
	bm.ProcessBlocks(blocks)

	blockchainAPI := New(bm, nil, nil, nil)
	keys := make([]string, 0, 10)
	keys = append(keys, base64.StdEncoding.EncodeToString(keyPairs[0].Public))

	balances := blockchainAPI.Balances(keys)
	if len(balances) != 1 || balances[0] != bm.Balance(keyPairs[0].Public) {
		t.Errorf("Account Balance should be %v!", bm.Balance(keyPairs[0].Public))
	}

	//add two more accounts
	keys = append(keys, base64.StdEncoding.EncodeToString(keyPairs[1].Public), base64.StdEncoding.EncodeToString(keyPairs[2].Public))

	balances = blockchainAPI.Balances(keys)
	if len(balances) != 3 || balances[0] != bm.Balance(keyPairs[0].Public) ||
		balances[1] != bm.Balance(keyPairs[1].Public) || balances[2] != bm.Balance(keyPairs[2].Public) {
		t.Errorf("One of account Balance mismatch!")
	}
}

func createDBMock() *DBMock {
	db := new(DBMock)
	transactions := block.CreateDummyTransactions(10)
	db.Trans = &transactions
	block := block.New(block.VDF([]byte("hello")), 1, 100, &transactions)
	db.Block = &block

	db.From = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 12, 11, 12, 13}
	db.To = []byte{13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	return db
}

func TestTransactionsFrom(t *testing.T) {
	db := createDBMock()
	blockchainAPI := New(nil, db, nil, nil)

	trans, _ := blockchainAPI.TransactionsFrom(base64.StdEncoding.EncodeToString(db.From), 0, 30)
	if !reflect.DeepEqual(trans.Ts, db.Trans.Ts) {
		t.Errorf("Transactions not equal!")
	}

	trans, _ = blockchainAPI.TransactionsFrom(base64.StdEncoding.EncodeToString(db.To), 0, 30)
	if len(trans.Ts) != 0 {
		t.Errorf("Result should be empty! %v", len(trans.Ts))
	}
}

func TestTransactionsTo(t *testing.T) {
	db := createDBMock()
	blockchainAPI := New(nil, db, nil, nil)

	trans, _ := blockchainAPI.TransactionsTo(base64.StdEncoding.EncodeToString(db.To), 0, 30)
	if !reflect.DeepEqual(trans.Ts, db.Trans.Ts) {
		t.Errorf("Transactions not equal!")
	}

	trans, _ = blockchainAPI.TransactionsTo(base64.StdEncoding.EncodeToString(db.From), 0, 30)
	if len(trans.Ts) != 0 {
		t.Errorf("Result should be empty! %v", len(trans.Ts))
	}
}

func TestTransactionsByBlockHeight(t *testing.T) {
	db := createDBMock()
	blockchainAPI := New(nil, db, nil, nil)
	trans, _ := blockchainAPI.BlockTransactionsByHeight(2, 0, 30)
	if !reflect.DeepEqual(trans.Ts, db.Trans.Ts) {
		t.Errorf("Transactions not equal!")
	}
}

func TestTransactionsByBlockHeight2(t *testing.T) {
	db := new(DBMock)
	db.Block = nil
	blockchainAPI := New(nil, db, nil, nil)
	trans, _ := blockchainAPI.BlockTransactionsByHeight(2, 0, 30)
	if trans != nil {
		t.Errorf("Transactions not nil!")
	}
}

func TestBlocks(t *testing.T) {
	db := new(DBMock)
	db.Blocks = make([]*block.Block, 0, 20)
	accounts := books.NewBookManager()
	accounts.AddValidVDFValue([]byte{1})
	accounts.CreateAccount([]byte("me"), 1000)
	accounts.CreateAccount([]byte("you"), 0)

	blockchainAPI := New(accounts, db, nil, nil)
	for i := 0; i < 10; i++ {
		transactions := block.CreateDummyTransactions(10)
		bl := block.New(block.VDF([]byte("hello")), uint64(i), 100, &transactions)
		db.Blocks = append(db.Blocks, &bl)
		accounts.ProcessBlocks([]block.Block{bl})
	}
	blocks, _ := blockchainAPI.Blocks(100, 30)
	if len(blocks) != 10 {
		t.Errorf("Blocks wrong len: expected %v, was %v", 10, len(blocks))
	}
}

func TestBlockByHeight(t *testing.T) {
	db := createDBMock()
	blockchainAPI := New(nil, db, nil, nil)
	block := blockchainAPI.BlockByHeight(2)
	if !reflect.DeepEqual(block, db.Block) {
		t.Errorf("BlockWithNumber not equal!")
	}
}

func TestBlockByHeight2(t *testing.T) {
	db := new(DBMock)
	db.Block = nil
	blockchainAPI := New(nil, db, nil, nil)
	block := blockchainAPI.BlockByHeight(2)
	if block != nil {
		t.Errorf("BlockWithNumber not nil!")
	}
}

func TestBlockByHash(t *testing.T) {
	db := createDBMock()
	blockchainAPI := New(nil, db, nil, nil)
	block := blockchainAPI.BlockByHash("Hello")
	if !reflect.DeepEqual(block, db.Block) {
		t.Errorf("BlockWithHash not equal!")
	}
}
func TestBlockByHash2(t *testing.T) {
	db := new(DBMock)
	db.Block = nil
	blockchainAPI := New(nil, db, nil, nil)
	block := blockchainAPI.BlockByHash("Hello")
	if !reflect.DeepEqual(block, db.Block) {
		t.Errorf("BlockWithHash not nil!")
	}
}

// func TestCalculateTPS(t *testing.T) {
// 	db := new(DBMock)
// 	db.Blocks = make([]*block.Block, 0, 20)
// 	accounts := books.NewBookManager()
// 	accounts.AddValidVDFValue(block.VDF([]byte{1}))

// 	accounts.CreateAccount([]byte("me"), 1000)
// 	accounts.CreateAccount([]byte("you"), 0)

// 	blockchainAPI := New(accounts, db, nil, nil)
// 	stop := true
// 	go blockchainAPI.calculateTPS(&stop)
// 	time.Sleep(1 * time.Second)
// 	for i := 0; i < 10; i++ {
// 		transactions := block.CreateRealTransactionsFrom(10, []byte("me"))
// 		bl := block.New(block.VDF([]byte("hello")), uint64(i), 100, &transactions)
// 		db.Blocks = append(db.Blocks, &bl)
// 		accounts.ProcessBlocks([]block.Block{bl})
// 	}
// 	time.Sleep(1 * time.Second)
// 	stop = false
// 	if blockchainAPI.stats.maxTPS < 80 {
// 		t.Errorf("calculateTPS expected %v was %v!", 80, blockchainAPI.stats.maxTPS)
// 	}
// 	fmt.Println(blockchainAPI.stats)
// }

// func TestTPS(t *testing.T) {
// 	db := new(DBMock)
// 	db.Blocks = make([]*block.Block, 0, 20)
// 	accounts := books.NewBookManager()
// 	accounts.AddValidVDFValue(block.VDF([]byte{1}))

// 	accounts.CreateAccount([]byte("me"), 100000000)
// 	accounts.CreateAccount([]byte("you"), 0)

// 	blockchainAPI := New(accounts, db, nil, nil)
// 	tps1 := blockchainAPI.TPS()
// 	// time.Sleep(1 * time.Second)

// 	for i := 0; i < 10; i++ {
// 		transactions := block.CreateRealTransactionsFrom(10, []byte("me"))
// 		bl := block.New(block.VDF([]byte("hello")), uint64(i), 100, &transactions)
// 		db.Blocks = append(db.Blocks, &bl)
// 		accounts.ProcessBlocks([]block.Block{bl})
// 	}
// 	time.Sleep(1 * time.Second)
// 	tps2 := blockchainAPI.TPS()

// 	for i := 0; i < 10; i++ {
// 		transactions := block.CreateRealTransactionsFrom(100, []byte("me"))
// 		bl := block.New(block.VDF([]byte("hello")), uint64(i), 100, &transactions)
// 		db.Blocks = append(db.Blocks, &bl)
// 		accounts.ProcessBlocks([]block.Block{bl})
// 	}
// 	time.Sleep(1 * time.Second)
// 	tps3 := blockchainAPI.TPS()

// 	if tps1 != 0 || tps2 < 60 || tps3 < 900 {
// 		t.Errorf("TPS expected %v.%v.%v was %v.%v.%v!", 0, 60, 900, tps1, tps2, tps3)
// 	}
// }

// func TestCalculateBlockTime(t *testing.T) {
// 	db := new(DBMock)
// 	db.Blocks = make([]*block.Block, 0, 20)
// 	accounts := books.NewBookManager()
// 	accounts.AddValidVDFValue(block.VDF([]byte{1}))

// 	accounts.CreateAccount([]byte("me"), 1000)
// 	accounts.CreateAccount([]byte("you"), 0)

// 	blockchainAPI := New(accounts, db, nil, nil)
// 	stop := true
// 	go blockchainAPI.calculateBlockTime(&stop)
// 	time.Sleep(blockTimeDelay)
// 	for i := 0; i < 10; i++ {
// 		transactions := block.CreateRealTransactionsFrom(10, []byte("me"))
// 		bl := block.New(block.VDF([]byte("hello")), uint64(i), 100, &transactions)
// 		bl.Number = uint64(i)
// 		db.Blocks = append(db.Blocks, &bl)
// 		accounts.ProcessBlocks([]block.Block{bl})
// 	}
// 	time.Sleep(blockTimeDelay)
// 	stop = false
// 	if blockchainAPI.stats.minBlockTime >= 10000 {
// 		t.Errorf("calculateTPS expected %v was %v!", 10000, blockchainAPI.stats.minBlockTime)
// 	}
// 	fmt.Println(blockchainAPI.stats)
// }

// func TestBlockTime(t *testing.T) {
// 	db := new(DBMock)
// 	db.Blocks = make([]*block.Block, 0, 20)
// 	accounts := books.NewBookManager()
// 	accounts.AddValidVDFValue(block.VDF([]byte{1}))

// 	accounts.CreateAccount([]byte("me"), 1000)
// 	accounts.CreateAccount([]byte("you"), 0)

// 	blockchainAPI := New(accounts, db, nil, nil)
// 	blTime1 := blockchainAPI.BlockTime()
// 	// time.Sleep(1 * time.Second)

// 	for i := 0; i < 10; i++ {
// 		transactions := block.CreateRealTransactionsFrom(10, []byte("me"))
// 		bl := block.New(block.VDF([]byte("hello")), uint64(i), 100, &transactions)
// 		db.Blocks = append(db.Blocks, &bl)
// 		accounts.ProcessBlocks([]block.Block{bl})
// 	}
// 	time.Sleep(blockTimeDelay)
// 	blTime2 := blockchainAPI.BlockTime()

// 	for i := 0; i < 20; i++ {
// 		transactions := block.CreateRealTransactionsFrom(100, []byte("me"))
// 		bl := block.New(block.VDF([]byte("hello")), uint64(i), 100, &transactions)
// 		db.Blocks = append(db.Blocks, &bl)
// 		accounts.ProcessBlocks([]block.Block{bl})
// 	}
// 	time.Sleep(blockTimeDelay)
// 	blTime3 := blockchainAPI.BlockTime()

// 	if blTime1 != 0 || blTime2 >= 10000 || blTime3 >= 10000 {
// 		t.Errorf("TPS expected %v.%v.%v was %v.%v.%v!", 0, 9, 19, blTime1, blTime2, blTime3)
// 	}
// }
