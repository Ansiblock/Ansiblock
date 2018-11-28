package api

import (
	"crypto/rand"
	"os"
	"reflect"
	"testing"

	"github.com/Ansiblock/Ansiblock/block"
)

func createBlock() block.Block {
	keypair1 := block.NewKeyPair()
	keypair2 := block.NewKeyPair()
	c := 10
	b := make([]byte, c)
	rand.Read(b)
	vdf0 := block.VDF(b)
	tr1 := block.NewTransaction(&keypair1, keypair2.Public, 3, 4, vdf0)
	tr2 := block.NewTransaction(&keypair2, keypair1.Public, 10, 11, vdf0)
	trans := &block.Transactions{Ts: []block.Transaction{tr1, tr2}}
	vdf1 := block.VDF(vdf0)
	return block.Block{Count: 1, Val: vdf1, Transactions: trans}
}

func TestNewDBConnection(t *testing.T) {
	os.Remove(DBFilename)
	db := NewDBConnection(DBFilename)
	err := db.conn.Ping()
	if err != nil {
		t.Error("Error connecting DB", err.Error())
	}
}

func TestSaveBlock(t *testing.T) {
	os.Remove(DBFilename)
	db := NewDBConnection(DBFilename)
	for i := 0; i < 10; i++ {
		b := createBlock()
		b.Number = uint64(i)
		err := db.SaveBlock(b)
		if err != nil {
			t.Error("Error saving block to DB", err.Error())
		}
	}
}

func TestGetBlockByHash(t *testing.T) {
	os.Remove(DBFilename)
	db := NewDBConnection(DBFilename)
	b := createBlock()
	db.SaveBlock(b)
	blk := db.GetBlockByHash(b.Val)
	numTrans := b.Transactions.Count()
	b.Transactions = &block.Transactions{Ts: make([]block.Transaction, numTrans)}
	if !reflect.DeepEqual(*blk, b) {
		t.Error("Error getting block")
	}
}
func TestGetBlockByHeight(t *testing.T) {
	os.Remove(DBFilename)
	db := NewDBConnection(DBFilename)
	b := createBlock()
	db.SaveBlock(b)
	blk := db.GetBlockByHeight(0)
	numTrans := b.Transactions.Count()
	b.Transactions = &block.Transactions{Ts: make([]block.Transaction, numTrans)}
	if !reflect.DeepEqual(*blk, b) {
		t.Error("Error getting block")
	}
}

func TestGetBlocksStartingAtHeight(t *testing.T) {
	os.Remove(DBFilename)
	db := NewDBConnection(DBFilename)
	for i := 0; i < 10; i++ {
		b := createBlock()
		b.Number = uint64(i)
		db.SaveBlock(b)
	}
	result, _ := db.GetBlocksAfterHeight(7, 10)
	if len(result) != 3 {
		t.Error("Error getting new blocks starting at given height")
	}
}

func reverseSlice(a []block.Transaction) {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
}
func TestGetTxFromBlockByHeight(t *testing.T) {
	os.Remove(DBFilename)
	db := NewDBConnection(DBFilename)
	b := createBlock()
	db.SaveBlock(b)
	transPtr, _ := db.GetTxFromBlockByHeight(0, MaxOffset, 100)
	reverseSlice(transPtr.Ts)
	if !reflect.DeepEqual(transPtr.Ts, b.Transactions.Ts) {
		t.Error("Error getting transactions", b.Transactions, transPtr.Ts)
	}
}

func TestGetTransactionsFrom(t *testing.T) {
	os.Remove(DBFilename)
	db := NewDBConnection(DBFilename)
	b := createBlock()
	db.SaveBlock(b)
	transactions := b.Transactions
	from := transactions.Ts[0].From
	transPtr, _ := db.GetTransactionsFrom(from, MaxOffset, 100)
	for _, tr := range transPtr.Ts {
		if !reflect.DeepEqual(tr.From, from) {
			t.Error("Error getting transactions from: ", from)
		}
	}
}

func TestGetTransactionsTo(t *testing.T) {
	os.Remove(DBFilename)
	db := NewDBConnection(DBFilename)
	b := createBlock()
	db.SaveBlock(b)
	transactions := b.Transactions
	to := transactions.Ts[0].To
	transPtr, _ := db.GetTransactionsTo(to, MaxOffset, 100)
	for _, tr := range transPtr.Ts {
		if !reflect.DeepEqual(tr.To, to) {
			t.Error("Error getting transactions to: ", to)
		}
	}
}

func TestGetAccountTransactions(t *testing.T) {
	os.Remove(DBFilename)
	db := NewDBConnection(DBFilename)
	b := createBlock()
	db.SaveBlock(b)
	transactions := b.Transactions
	account := transactions.Ts[0].From
	transPtr, _ := db.GetAccountTransactions(account, MaxOffset, 100)
	for _, tr := range transPtr.Ts {
		if !reflect.DeepEqual(tr.To, account) && !reflect.DeepEqual(tr.From, account) {
			t.Error("Error getting transactions for account: ", account)
		}
	}
}
