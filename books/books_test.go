package books

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Ansiblock/Ansiblock/block"
)

func TestProcessTransactionsSingleTransaction(t *testing.T) {
	bm := NewBookManager()
	bm.CreateAccount([]byte("acc1"), 100)
	bm.CreateAccount([]byte("acc2"), 100)
	vdf := []byte{1, 2, 3}
	bm.AddValidVDFValue(vdf)
	trans := block.Transactions{}
	trans.Ts = make([]block.Transaction, 1)
	trans.Ts[0] = block.Transaction{From: []byte("acc1"), To: []byte("acc2"), Token: 5, Fee: 1, ValidVDFValue: vdf}

	res := bm.ProcessTransactions(trans)

	failed := bm.Balance([]byte("acc1")) != 95 || bm.Balance([]byte("acc2")) != 104
	failed = failed || !trans.Equals(res) || bm.TransactionsTotal() != 1
	if failed {
		t.Errorf("ProcessTransactions Function SingleTransaction Error! \n%v", bm.String())
	}
}

func TestProcessTransactionsMultipleTransactions(t *testing.T) {
	bm := NewBookManager()
	bm.CreateAccount([]byte("acc1"), 10)
	bm.CreateAccount([]byte("acc2"), 10)
	bm.CreateAccount([]byte("acc3"), 10)
	bm.CreateAccount([]byte("acc4"), 10)
	vdf := []byte{1, 2, 3}
	bm.AddValidVDFValue(vdf)

	trans := block.Transactions{}
	trans.Ts = make([]block.Transaction, 3)
	trans.Ts[0] = block.Transaction{From: []byte("acc1"), To: []byte("acc2"), Token: 10, Fee: 1, ValidVDFValue: vdf, Signature: []byte{1}}
	trans.Ts[1] = block.Transaction{From: []byte("acc2"), To: []byte("acc3"), Token: 9, Fee: 2, ValidVDFValue: vdf, Signature: []byte{2}}
	trans.Ts[2] = block.Transaction{From: []byte("acc3"), To: []byte("acc4"), Token: 7, Fee: 3, ValidVDFValue: vdf, Signature: []byte{3}}

	res := bm.ProcessTransactions(trans)
	failed := bm.Balance([]byte("acc1")) != 0 || bm.Balance([]byte("acc2")) != 10 ||
		bm.Balance([]byte("acc3")) != 10 || bm.Balance([]byte("acc4")) != 14
	failed = failed || !trans.Equals(res) || bm.TransactionsTotal() != 3

	if failed {
		t.Errorf("ProcessTransactions Function MultipleTransactions Error! \n%v", bm.String())
	}
}

func TestProcessTransactionAccountNotFound(t *testing.T) {
	bm := NewBookManager()
	bm.CreateAccount([]byte("acc1"), 100)
	bm.CreateAccount([]byte("acc2"), 100)
	vdf := []byte{1, 2, 3}
	bm.AddValidVDFValue(vdf)
	trans := block.Transactions{}
	trans.Ts = make([]block.Transaction, 1)
	trans.Ts[0] = block.Transaction{From: []byte("acc?"), To: []byte("acc2"), Token: 5, Fee: 1, ValidVDFValue: vdf}

	err := bm.applyTransactionWithdraw(&trans.Ts[0])

	if err != errAccountNotFound || bm.TransactionsTotal() != 0 {
		t.Errorf("ProcessTransactions Function: Should be AccountNotFound error")
	}
}

func TestProcessTransactionInsufficientFunds(t *testing.T) {
	bm := NewBookManager()
	bm.CreateAccount([]byte("acc1"), 10)
	bm.CreateAccount([]byte("acc2"), 100)
	vdf := []byte{1, 2, 3}
	bm.AddValidVDFValue(vdf)
	trans := block.Transactions{}
	trans.Ts = make([]block.Transaction, 1)
	trans.Ts[0] = block.Transaction{From: []byte("acc1"), To: []byte("acc2"), Token: 11, Fee: 1, ValidVDFValue: vdf}

	err := bm.applyTransactionWithdraw(&trans.Ts[0])

	if err != errInsufficientFunds || bm.TransactionsTotal() != 0 {
		t.Errorf("ProcessTransactions Function: Should be InsufficientFunds error")
	}
}

func TestProcessTransactionNegativeTokens(t *testing.T) {
	bm := NewBookManager()
	bm.CreateAccount([]byte("acc1"), 200)
	bm.CreateAccount([]byte("acc2"), 100)
	vdf := []byte{1, 2, 3}
	bm.AddValidVDFValue(vdf)
	trans := block.Transactions{}
	trans.Ts = make([]block.Transaction, 1)
	trans.Ts[0] = block.Transaction{From: []byte("acc1"), To: []byte("acc2"), Token: -1, Fee: 1, ValidVDFValue: vdf}

	err := bm.applyTransactionWithdraw(&trans.Ts[0])

	if err != errNegativeTokens || bm.TransactionsTotal() != 0 {
		t.Errorf("ProcessTransactions Function: Should be NegativeTokens error")
	}
}

func TestProcessTransactionsWithErrors(t *testing.T) {
	bm := NewBookManager()
	bm.CreateAccount([]byte("acc1"), 10)
	bm.CreateAccount([]byte("acc2"), 10)
	bm.CreateAccount([]byte("acc3"), 10)
	bm.CreateAccount([]byte("acc4"), 10)
	bm.CreateAccount([]byte("acc5"), 10)

	vdf := []byte{1, 2, 3}
	bm.AddValidVDFValue(vdf)

	trans := block.Transactions{}
	trans.Ts = make([]block.Transaction, 7)
	trans.Ts[0] = block.Transaction{From: []byte("acc1"), To: []byte("acc2"), Token: 9, Fee: 1, ValidVDFValue: vdf, Signature: []byte{1}}
	//InsufficientFunds
	trans.Ts[1] = block.Transaction{From: []byte("acc1"), To: []byte("acc2"), Token: 11, Fee: 1, ValidVDFValue: vdf, Signature: []byte{2}}
	trans.Ts[2] = block.Transaction{From: []byte("acc2"), To: []byte("acc3"), Token: 10, Fee: 4, ValidVDFValue: vdf, Signature: []byte{3}}
	//AccountNotFound
	trans.Ts[3] = block.Transaction{From: []byte("acc?"), To: []byte("acc2"), Token: 1, Fee: 1, ValidVDFValue: vdf, Signature: []byte{4}}
	//NegativeTokens
	trans.Ts[4] = block.Transaction{From: []byte("acc1"), To: []byte("acc2"), Token: -1, Fee: 1, ValidVDFValue: vdf, Signature: []byte{5}}
	trans.Ts[5] = block.Transaction{From: []byte("acc2"), To: []byte("acc1"), Token: 5, Fee: 1, ValidVDFValue: vdf, Signature: []byte{6}}
	trans.Ts[6] = block.Transaction{From: []byte("acc4"), To: []byte("acc3"), Token: 8, Fee: 2, ValidVDFValue: vdf, Signature: []byte{7}}

	res := bm.ProcessTransactions(trans)
	failed := bm.Balance([]byte("acc1")) != 1 || bm.Balance([]byte("acc2")) != 8 ||
		bm.Balance([]byte("acc3")) != 22 || bm.Balance([]byte("acc4")) != 2 ||
		bm.Balance([]byte("acc5")) != 10 || bm.TransactionsTotal() != 3
	trans.Ts = append(trans.Ts[:1], trans.Ts[2:]...)
	trans.Ts = append(trans.Ts[:2], trans.Ts[5:]...)
	failed = failed || !trans.Equals(res)

	if failed {
		t.Errorf("ProcessTransactions Function with errors Failed! \n%v", bm.String())
	}
}
func TestProcessBlocks(t *testing.T) {
	var blocks []block.Block
	for i := 0; i < 10; i++ {
		val := block.VDF([]byte("hello"))
		transactions := block.CreateDummyTransactions(100)
		bl := block.New(val, uint64(i), 1000, &transactions)
		blocks = append(blocks, bl)
	}

	bm := NewBookManager()
	bm.AddValidVDFValue([]byte{1})
	bm.CreateAccount([]byte("me"), 1000)
	bm.CreateAccount([]byte("you"), 0)
	bm.ProcessBlocks(blocks)
	if bm.Balance([]byte("me")) != 10 || bm.Balance([]byte("you")) != 990 {
		t.Errorf("ProcessBlocks Failed! \n%v", bm.String())
	}
	if bm.LastBlock().Number != 10 {
		t.Errorf("ProcessBlocks last block Failed! \n%v", bm.String())
	}
}

func TestProcessBlocks2(t *testing.T) {
	blocks := make([]block.Block, 1)
	bm, keyPairs := RandomAccounts(20)

	var vdfValue = block.VDF([]byte{0, 2, 3})
	bm.AddValidVDFValue(vdfValue)

	var numValidTransactions uint64
	blocks[0].Transactions, numValidTransactions = RandomTransactions(bm, 200, keyPairs, vdfValue)
	blocks[0].Count = 200
	blocks[0].Val = vdfValue

	bm.ProcessBlocks(blocks)
	if bm.TransactionsTotal() != numValidTransactions {
		t.Errorf("Processed transactions number is differenc! %v != %v\n", bm.TransactionsTotal(), numValidTransactions)
	}
}

func TestValidVDFValue(t *testing.T) {
	bm := NewBookManager()
	if !bytes.Equal(bm.ValidVDFValue(), []byte("hello")) {
		t.Errorf("ValidVDFValue Failed! %v\n", bm.ValidVDFValue())
	}
}

func TestString(t *testing.T) {
	bm := NewBookManager()
	bm.CreateAccount([]byte("acc1"), 10)
	if bm.String() != "acc1=\"10\"\ntransactions total=0\n" {
		t.Errorf("String Failed! %v.String() = %v\n", bm.balances, bm.String())
	}

}

func TestClone(t *testing.T) {
	var blocks []block.Block
	for i := 0; i < 10; i++ {
		val := block.VDF([]byte("hello"))
		transactions := block.CreateDummyTransactions(100)
		bl := block.New(val, uint64(i), 1000, &transactions)
		blocks = append(blocks, bl)
	}

	bm := NewBookManager()
	bm.AddValidVDFValue([]byte{1})
	bm.CreateAccount([]byte("me"), 1000)
	bm.CreateAccount([]byte("you"), 0)
	bm.ProcessBlocks(blocks)

	clone := bm.Clone()

	if bm.Balance([]byte("me")) != 10 || bm.Balance([]byte("you")) != 990 || !bm.Equals(clone) {
		t.Errorf("ProcessBlocks Failed! \n%v", bm.String())
	}
}

func TestBookEquals(t *testing.T) {
	vdf := []byte{1, 2, 3}
	//book manager 2
	bm1 := NewBookManager()
	bm1.CreateAccount([]byte("acc1"), 10)
	bm1.CreateAccount([]byte("acc2"), 10)
	bm1.CreateAccount([]byte("acc3"), 10)
	bm1.CreateAccount([]byte("acc4"), 10)
	bm1.AddValidVDFValue(vdf)
	//book manager 2
	bm2 := NewBookManager()
	bm2.CreateAccount([]byte("acc1"), 10)
	bm2.CreateAccount([]byte("acc2"), 10)
	bm2.CreateAccount([]byte("acc3"), 10)
	bm2.CreateAccount([]byte("acc4"), 10)
	bm2.AddValidVDFValue(vdf)

	trans := block.Transactions{}
	trans.Ts = make([]block.Transaction, 3)
	trans.Ts[0] = block.Transaction{From: []byte("acc1"), To: []byte("acc2"), Token: 10, Fee: 1, ValidVDFValue: vdf, Signature: []byte{1}}
	trans.Ts[1] = block.Transaction{From: []byte("acc2"), To: []byte("acc3"), Token: 9, Fee: 2, ValidVDFValue: vdf, Signature: []byte{2}}
	trans.Ts[2] = block.Transaction{From: []byte("acc3"), To: []byte("acc4"), Token: 7, Fee: 3, ValidVDFValue: vdf, Signature: []byte{3}}

	bm1.ProcessTransactions(trans)
	bm2.ProcessTransactions(trans)

	if !bm1.Equals(bm2) {
		t.Errorf("Books should be equal!")
	}

	bm1.CreateAccount([]byte("acc5"), 10)
	if bm1.Equals(bm2) {
		t.Errorf("Books should be different, because of acc5!")
	}

	bm2.CreateAccount([]byte("acc5"), 10)
	if !bm1.Equals(bm2) {
		t.Errorf("Books should be equal again!")
	}

	tr := block.Transaction{From: []byte("acc4"), To: []byte("acc2"), Token: 10, Fee: 1, ValidVDFValue: vdf, Signature: []byte{100}}
	err := bm1.applyTransactionWithdraw(&tr)
	if err != nil || bm1.Equals(bm2) {
		t.Errorf("Books should be different, because of balances!  err: %v", err)
	}
	bm2.applyTransactionWithdraw(&tr)
	if !bm1.Equals(bm2) {
		t.Errorf("Books should be equal again! 2")
	}
}

func TestRandomKeys(t *testing.T) {
	bm := NewBookManager()
	bm.CreateAccount([]byte("acc1"), 11)
	bm.CreateAccount([]byte("acc2"), 12)
	bm.CreateAccount([]byte("acc3"), 13)
	bm.CreateAccount([]byte("acc4"), 14)

	allKeys := ":acc1:acc2:acc3:acc4:"
	keys := bm.RandomKeys(3)
	if len(keys) != 3 || !strings.Contains(allKeys, string(keys[0])) ||
		!strings.Contains(allKeys, string(keys[1])) || !strings.Contains(allKeys, string(keys[2])) {
		t.Errorf("RandomKeys Failed!")
	}
}

func TestCreateAccount(t *testing.T) {
	bm := new(Accounts)
	kp := block.NewKeyPair()
	bm.CreateAccount(kp.Public, 10)
	if bm.balances[string(kp.Public)] != 10 {
		t.Errorf("CreateAccount Failed!")
	}
}

func TestBlocksTotal(t *testing.T) {
	bm := new(Accounts)
	bm.blocksTotal = 10
	if bm.BlocksTotal() != 10 {
		t.Errorf("BlocksTotal Failed!")
	}
	for i := uint64(0); i < 100; i++ {
		bm.IncreaseBlocksTotal()
		if bm.BlocksTotal() != 11+i {
			t.Errorf("BlocksTotal Failed!")
		}
	}
}

func TestLastBlock(t *testing.T) {
	bm := new(Accounts)
	if bm.LastBlock() != nil {
		t.Errorf("LastBlock Failed!")
	}
	l := newLedger()
	block := new(block.Block)
	l.lastBlock = block
	bm.ledger = l
	if bm.LastBlock() != block {
		t.Errorf("LastBlock Failed!")
	}
}

func TestRandomTransactionsBlocks(t *testing.T) {
	mb := NewBookManager()
	pks := block.KeyPairs(100)
	res := RandomTransactionsBlocks(mb, 10, 100, pks)
	if len(res) != 100 {
		t.Errorf("RandomTransactionsBlocks Failed!")
	}
	for i := 0; i < 100; i++ {
		if len(res[i].Transactions.Ts) != 10 {
			t.Errorf("RandomTransactionsBlocks Failed!")
		}
	}
}
