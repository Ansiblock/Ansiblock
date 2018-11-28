package books

import (
	"math/rand"
	"testing"
	"time"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/network"
)

func transactionsInList(tr block.Transactions, trs []block.Transactions) bool {
	for _, trans := range trs {
		if trans.Equals(tr) {
			return true
		}
	}
	return false
}

func transactionsSetEqual(trsSet1, trsSet2 []block.Transactions) bool {
	if len(trsSet1) != len(trsSet2) {
		return false
	}
	for _, trSet := range trsSet1 {
		if !transactionsInList(trSet, trsSet2) {
			return false
		}
	}
	for _, trSet := range trsSet2 {
		if !transactionsInList(trSet, trsSet1) {
			return false
		}
	}
	return true
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func TestTransactionGeneratorSingleThread2(t *testing.T) {
	var in = make(chan *network.Packets)
	accounts := NewBookManager()
	accounts.CreateAccount(block.VDF([]byte("acc1")), 100)
	accounts.CreateAccount(block.VDF([]byte("acc2")), 100)
	vdf := block.VDF([]byte{1, 2, 3})
	accounts.AddValidVDFValue(vdf)
	signature := make([]byte, 64)

	out := TransactionGenerator(accounts, in)
	for i := int64(0); i < 10; i++ {
		trans := block.Transactions{}
		trans.Ts = make([]block.Transaction, 1)
		signature[0]++
		trans.Ts[0] = block.Transaction{From: block.VDF([]byte("acc1")), To: block.VDF([]byte("acc2")), Token: 5, Fee: 1, ValidVDFValue: vdf, Signature: signature}

		packets := trans.ToPackets(nil)
		in <- packets
		trs2 := <-out
		x := *trs2
		if !block.TransactionSetEqual(trans.Ts, x.Ts) {
			t.Errorf("TransactionGenerator not correct transaction list %v and %v", trans, trs2)
		}
	}
}
