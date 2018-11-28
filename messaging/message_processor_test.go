package messaging

import (
	"bytes"
	"testing"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/books"
)

func TestProcessMessagesCheckBalance(t *testing.T) {
	accounts := books.NewBookManager()
	accounts.CreateAccount([]byte("acc1"), 100)
	accounts.CreateAccount([]byte("acc2"), 100)
	vdf := []byte{1, 2, 3}
	accounts.AddValidVDFValue(vdf)
	trans := block.Transactions{}
	trans.Ts = make([]block.Transaction, 1)
	trans.Ts[0] = block.Transaction{From: []byte("acc1"), To: []byte("acc2"), Token: 5, Fee: 1, ValidVDFValue: vdf}
	accounts.ProcessTransactions(trans)

	requests := []Request{Request{Type: Balance, Addr: nil, PublicKey: []byte("acc2")}}
	responses := ProcessMessages(requests, accounts)
	r := responses.Responses[0].(*ResponseBalance)
	if r.Value != 104 {
		t.Errorf("response balance does not match expacted value, should be 104 got %d", r.Value)
	}

}

func TestProcessMessagesCheckValidVDFValue(t *testing.T) {
	accounts := books.NewBookManager()
	accounts.CreateAccount([]byte("acc1"), 100)
	accounts.CreateAccount([]byte("acc2"), 100)
	vdf := []byte{1, 2, 3}
	accounts.AddValidVDFValue(vdf)
	trans := block.Transactions{}
	trans.Ts = make([]block.Transaction, 1)
	trans.Ts[0] = block.Transaction{From: []byte("acc1"), To: []byte("acc2"), Token: 5, Fee: 1, ValidVDFValue: vdf}
	accounts.ProcessTransactions(trans)

	requests := []Request{Request{Type: ValidVDFValue, Addr: nil, PublicKey: []byte("acc2")}}
	responses := ProcessMessages(requests, accounts)
	lastVal := accounts.ValidVDFValue()
	r := responses.Responses[0].(*ResponseValidVDFValue)
	if !bytes.Equal(r.Value, lastVal) {
		t.Errorf("response ValidVDFValue does not match bank lastVal %v, %v", r.Value, lastVal)
	}
}

func TestProcessMessagesCheckTransactionsTotal(t *testing.T) {
	accounts := books.NewBookManager()
	accounts.CreateAccount([]byte("acc1"), 100)
	accounts.CreateAccount([]byte("acc2"), 100)
	vdf := []byte{1, 2, 3}
	accounts.AddValidVDFValue(vdf)
	trans := block.Transactions{}
	transactionsCount := uint64(5)
	trans.Ts = make([]block.Transaction, 5)
	for i := uint64(0); i < transactionsCount; i++ {
		sig := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		trans.Ts[i] = block.Transaction{From: []byte("acc1"), To: []byte("acc2"), Token: 5, Fee: 1, ValidVDFValue: vdf, Signature: sig[0:i]}
	}
	accounts.ProcessTransactions(trans)
	requests := []Request{Request{Type: TransactionsTotal, Addr: nil, PublicKey: []byte("acc2")}}
	responses := ProcessMessages(requests, accounts)
	r := responses.Responses[0].(*ResponseTransactionsTotal)
	if r.Value != transactionsCount {
		t.Errorf("response transaction count does not match expacted value, should be %d got %d", accounts.TransactionsTotal(), r.Value)
	}
}
