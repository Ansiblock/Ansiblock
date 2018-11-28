package books

// to run test on cuda use following command:
// go test -v -tags cuda -timeout 30s github.com/Ansiblock/Ansiblock/books -run ^TestSigVerifyThread$

import (
	"testing"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/network"
)

// func TestSignatureVerificationSingleThread(t *testing.T) {
// 	var in = make(chan *network.Packets)
// 	out := SignatureVerification(in)
// 	for i := int64(0); i < 10; i++ {
// 		trs1 := CreateRealTransactions(10 + i)
// 		p := trs1.ToPackets(nil)
// 		in <- p
// 		p2 := <-out
// 		var trs2 Transactions
// 		trs2.FromPackets(p2)
// 		if !trs1.Equals(trs2) {
// 			t.Errorf("TransactionGenerator not correct transaction list %v\n and\n %v\n", trs1, trs2)
// 		}
// 	}
// }

func TestSigVerifyThread(t *testing.T) {
	var in = make(chan *network.Packets)
	out := SignatureVerification(in)
	tr := block.CreateRealTransaction(2)
	trs1 := block.Transactions{Ts: []block.Transaction{tr}}
	p := trs1.ToPackets(nil)
	in <- p
	p2 := <-out
	var trs2 block.Transactions
	trs2.FromPackets(p2)
	if !trs1.Equals(trs2) {
		t.Errorf("TransactionGenerator not correct transaction list %v\n and\n %v\n", trs1, trs2)
	}
}

func TestSigVerifyWithTwoTransactions(t *testing.T) {
	var in = make(chan *network.Packets)
	out := SignatureVerification(in)
	tr1 := block.CreateRealTransaction(2)
	tr2 := block.CreateRealTransaction(2)
	trs1 := block.Transactions{Ts: []block.Transaction{tr1, tr2}}
	p := trs1.ToPackets(nil)
	in <- p
	p2 := <-out
	var trs2 block.Transactions
	trs2.FromPackets(p2)
	if !trs1.Equals(trs2) {
		t.Errorf("TransactionGenerator not correct transaction list %v\n and\n %v\n", trs1, trs2)
	}
}

func TestSigVerifyWithDemagedTransactions(t *testing.T) {
	var in = make(chan *network.Packets)
	out := SignatureVerification(in)
	tr1 := block.CreateRealTransaction(2)
	tr2 := block.CreateRealTransaction(2)
	tr2.Signature[0] ^= 1
	trs1 := block.Transactions{Ts: []block.Transaction{tr1, tr2}}
	p := trs1.ToPackets(nil)
	in <- p
	pOut := <-out
	var trsOut block.Transactions
	trsOut.FromPackets(pOut)
	trsOut = filter(trsOut)
	// fmt.Println(trsOut)
	if len(trsOut.Ts) != 1 {
		t.Errorf("demaged transaction was not filtered out %v", trsOut)
	}

	if !trsOut.Ts[0].Equals(tr1) {
		t.Errorf("TransactionGenerator not correct transaction list %v\n and\n %v\n", trsOut.Ts[0], tr1)
	}

}

func TestSigVerifyWithTwoPackets(t *testing.T) {
	var in = make(chan *network.Packets)
	out := SignatureVerification(in)
	tr1 := block.CreateRealTransaction(2)
	tr2 := block.CreateRealTransaction(2)
	trs1 := block.Transactions{Ts: []block.Transaction{tr1}}
	trs2 := block.Transactions{Ts: []block.Transaction{tr2}}
	p1 := trs1.ToPackets(nil)
	p2 := trs2.ToPackets(nil)
	in <- p1
	in <- p2
	p1out := <-out
	p2out := <-out
	var trs1out block.Transactions
	var trs2out block.Transactions
	trs1out.FromPackets(p1out)
	trs2out.FromPackets(p2out)
	if !trs1.Equals(trs1out) {
		t.Errorf("TransactionGenerator not correct transaction list %v\n and\n %v\n", trs1, trs2)
	}
	if !trs2.Equals(trs2out) {
		t.Errorf("TransactionGenerator not correct transaction list %v\n and\n %v\n", trs1, trs2)
	}

}

// func TestSignatureVerificationSingleThread2(t *testing.T) {
// 	var in = make(chan *network.Packets)
// 	out := SignatureVerification(in)
// 	for i := int64(0); i < 10; i++ {
// 		trs1 := CreateRealTransactions(10 + i)
// 		trs1.Ts = append(trs1.Ts, CreateDummyTransaction(1))
// 		p := trs1.ToPackets(nil)
// 		in <- p
// 		p2 := <-out
// 		var trs2 Transactions
// 		trs2.FromPackets(p2)
// 		trs1.Ts = trs1.Ts[:len(trs1.Ts)-1]
// 		if !trs1.Equals(trs2) {
// 			t.Errorf("TransactionGenerator not correct transaction list %v and %v", trs1, trs2)
// 		}
// 	}
// }
