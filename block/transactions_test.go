package block

import (
	"bytes"
	"strings"
	"testing"
)

func TestSize2(t *testing.T) {
	trs := CreateRealTransactions(10)
	if trs.Size() != 10*176 {
		t.Errorf("Transactions size problem: %v != %v", trs.Size(), 10*176)
	}
}

func TestBlocksToBlobs(t *testing.T) {
	blocks := make([]Block, 10)
	trans := make([]Transaction, 0, 370)
	for i := 0; i < 10; i++ {
		trs := CreateRealTransactions(370)
		trans = append(trans, trs.Ts...)
		blocks[i].Transactions = &trs
		blocks[i].Count = 1
		blocks[i].Val = VDF([]byte{byte(i), 2, 3})
	}
	blobs := BlocksToBlobs(blocks)
	resBlocks := BlobsToBlocks(blobs)

	if len(blobs.Bs) != 10 {
		t.Errorf("BlocksToBlobs wrong blob count %v != %v", len(blobs.Bs), 20)
	}
	if len(resBlocks) != len(blocks) {
		t.Errorf("BlocksToBlobs wrong block count : %v instead of %v", len(resBlocks), len(blocks))
	}
	resTrans := make([]Transaction, 0)
	for i := range resBlocks {
		trs := resBlocks[i].Transactions
		resTrans = append(resTrans, trs.Ts...)
	}
	if !TransactionSetEqual(trans, resTrans) {
		t.Errorf("BlocksToBlobs something's wrong")

	}
}

func TestBlocksToBlobs2(t *testing.T) {
	blocks := make([]Block, 10)
	trans := make([]Transaction, 0, 1000)
	for i := 0; i < 10; i++ {
		trs := CreateRealTransactions(100)
		trans = append(trans, trs.Ts...)
		blocks[i].Transactions = &trs
		blocks[i].Count = 1
		blocks[i].Val = VDF([]byte{byte(i), 2, 3})
	}
	blobs := BlocksToBlobs(blocks)
	resBlocks := BlobsToBlocks(blobs)

	if len(blobs.Bs) != 4 {
		t.Errorf("BlocksToBlobs wrong blob count %v != %v", len(blobs.Bs), 4)
	}
	if len(resBlocks) != len(blocks) {
		t.Errorf("BlocksToBlobs wrong block count : %v instead of %v", len(resBlocks), len(blocks))
	}
	resTrans := make([]Transaction, 0)
	for i := range resBlocks {
		trs := resBlocks[i].Transactions
		resTrans = append(resTrans, trs.Ts...)
	}

	if !TransactionSetEqual(trans, resTrans) {
		t.Errorf("BlocksToBlobs something's wrong")
	}
}

func TestBlocksToBlobs3(t *testing.T) {
	blocks := make([]Block, 21)
	trans := make([]Transaction, 0, 1000)
	for i := 0; i < 10; i++ {
		trs := CreateRealTransactions(100)
		trans = append(trans, trs.Ts...)
		blocks[i].Transactions = &trs
		blocks[i].Count = 1
		blocks[i].Val = VDF([]byte{byte(i), 2, 3})
	}
	trs := CreateRealTransactions(370)
	trans = append(trans, trs.Ts...)
	blocks[10].Transactions = &trs
	blocks[10].Count = 1
	blocks[10].Val = VDF([]byte{byte(10), 2, 3})
	for i := 11; i < 21; i++ {
		trs := CreateRealTransactions(100)
		trans = append(trans, trs.Ts...)
		blocks[i].Transactions = &trs
		blocks[i].Count = 1
		blocks[i].Val = VDF([]byte{byte(i), 2, 3})
	}

	blobs := BlocksToBlobs(blocks)
	resBlocks := BlobsToBlocks(blobs)

	resTrans := make([]Transaction, 0)
	for i := range resBlocks {
		trs := resBlocks[i].Transactions
		resTrans = append(resTrans, trs.Ts...)
	}

	if !TransactionSetEqual(trans, resTrans) {
		t.Errorf("BlocksToBlobs something's wrong")
	}
}
func TestTransactionsEqual(t *testing.T) {
	trs1 := CreateDummyTransactions(10)
	trs2 := CreateDummyTransactions(10)
	trs3 := CreateDummyTransactions(20)

	if trs1.Equals(trs3) {
		t.Errorf("%v . Equals(%v) ", trs1, trs3)
	}
	if !trs1.Equals(trs2) {
		t.Errorf("! %v . Equals(%v) ", trs1, trs2)
	}
}

func TestTransactionsEqual2(t *testing.T) {
	trs1 := CreateDummyTransactions(10)
	trs2 := CreateDummyTransactions(11)
	trs2.Ts = trs2.Ts[1:]
	if trs1.Equals(trs2) {
		t.Errorf("! %v . Equals(%v) ", trs1, trs2)
	}
}
func TestTransactionsToBytes(t *testing.T) {
	tran1 := Transaction{Signature: []byte{1, 2, 3}}
	trans := Transactions{Ts: []Transaction{tran1}}
	res := trans.ToBytes()
	if !bytes.Equal(res, []byte{0, 1, 2, 3}) {
		t.Errorf("TransactionsToBytes(%v) = %v", trans, res)
	}
}

func TestTransactionsString(t *testing.T) {
	kp := NewKeyPair()
	tran := NewTransaction(&kp, []byte("2"), 3, 4, []byte{5, 6})
	trans := Transactions{Ts: []Transaction{tran}}
	res := trans.String()
	if !strings.Contains(res, "To: [50], \nAmount: 3, Fee: 4}") {
		t.Errorf("Transaction.String(%v) = %v Failed", tran, res)
	}
}

func TestTransactionsVerify(t *testing.T) {
	tran1 := Transaction{Token: 5, Fee: 1}
	tran2 := Transaction{Token: 5, Fee: 10}

	trans1 := Transactions{Ts: []Transaction{tran1}}
	trans2 := Transactions{Ts: []Transaction{tran2}}
	if !trans1.Verify() {
		t.Errorf("Transaction Verified: %v", tran1)
	}
	if trans2.Verify() {
		t.Errorf("Transaction Verified: %v", tran2)

	}
}
func TestVerifyThreads(t *testing.T) {
	tran1 := Transaction{Token: 5, Fee: 1}
	tran2 := Transaction{Token: 5, Fee: 10}

	trans1 := Transactions{Ts: []Transaction{tran1}}
	trans2 := Transactions{Ts: []Transaction{tran2}}
	if !trans1.Verify() {
		t.Errorf("Transaction Verified: %v", tran1)
	}
	if trans2.Verify() {
		t.Errorf("Transaction Verified: %v", tran2)
	}

	if !trans1.verifyThreads() {
		t.Errorf("Transaction Verified: %v", tran1)
	}
	if trans2.verifyThreads() {
		t.Errorf("Transaction Verified: %v", tran2)
	}

	if !trans1.verifySerial() {
		t.Errorf("Transaction Verified: %v", tran1)
	}
	if trans2.verifySerial() {
		t.Errorf("Transaction Verified: %v", tran2)
	}
	tran3 := CreateRealTransactions(1001)
	if !tran3.Verify() {
		t.Errorf("Transaction Verified: %v", tran1)
	}

}

func TestCreateRealTransactionsFrom(t *testing.T) {
	from := NewKeyPair()
	trans := CreateRealTransactionsFrom(10, from.Public)
	for i := 0; i < len(trans.Ts); i++ {
		if trans.Ts[i].Token != int64(i) || !bytes.Equal(trans.Ts[i].From, from.Public) {
			t.Errorf("CreateRealTransactionsFrom error")
		}
	}
}

func TestTransactionSetEqual(t *testing.T) {
	trans := CreateDummyTransactions(10)
	trans2 := CreateDummyTransactions(11)
	if TransactionSetEqual(trans.Ts, trans2.Ts) {
		t.Errorf("TransactionSetEqual error")
	}
	trans = CreateRealTransactions(10)
	trans2 = CreateRealTransactions(10)
	if TransactionSetEqual(trans.Ts, trans2.Ts) {
		t.Errorf("TransactionSetEqual error")
	}

	tran1 := CreateRealTransaction(10)
	tran2 := CreateRealTransaction(11)
	ts1 := make([]Transaction, 2)
	ts1[0] = tran1
	ts1[1] = tran1

	ts2 := make([]Transaction, 2)
	ts2[0] = tran1
	ts2[1] = tran2
	if TransactionSetEqual(ts1, ts2) {
		t.Errorf("TransactionSetEqual error")
	}
}

func TestFromPackets(t *testing.T) {
	transactions := CreateRealTransactions(100)
	packets := transactions.ToPackets(nil)
	resTran := new(Transactions)
	resTran.Ts = make([]Transaction, len(packets.Ps))

	ch := make(chan bool, 1)
	fromPackets(resTran, 0, len(packets.Ps), packets, ch)
	if !TransactionSetEqual(transactions.Ts, resTran.Ts) || !<-ch {
		t.Errorf("fromPackets error")
	}

	transactions2 := CreateRealTransactions(100)
	packets2 := transactions2.ToPackets(nil)
	resTran2 := new(Transactions)
	resTran2.Ts = make([]Transaction, len(packets2.Ps))
	ch2 := make(chan bool, 1)

	packets2.Ps[0].Size = 0
	packets2.Ps[50].Size = 0
	fromPackets(resTran2, 0, len(packets2.Ps), packets2, ch2)
	if resTran2.Ts[0].Fee != -1 || resTran2.Ts[50].Fee != -1 {
		t.Errorf("fromPackets error: zero sized packets %v %v", transactions.Ts[0].Fee, transactions.Ts[50].Fee)

	}
}

func TestToPackets(t *testing.T) {
	transactions := CreateRealTransactions(100)
	packets := transactions.ToPackets(nil)
	resTran := new(Transactions)

	resTran.FromPackets(packets)
	if !TransactionSetEqual(transactions.Ts, resTran.Ts) {
		t.Errorf("FromPackets error")
	}
}

func BenchmarkVerify(b *testing.B) {
	transactions := CreateRealTransactions(2000)
	for n := 0; n < b.N; n++ {
		transactions.Verify()
	}
}

// func BenchmarkFromPacket3(b *testing.B) {
// 	transactions := CreateRealTransactions(8192)
// 	pack := transactions.ToPackets(nil)
// 	for n := 0; n < b.N; n++ {
// 		transactions.FromPackets3(pack)
// 	}
// }
// func BenchmarkFromPacket4(b *testing.B) {
// 	transactions := CreateRealTransactions(8192)
// 	pack := transactions.ToPackets(nil)
// 	for n := 0; n < b.N; n++ {
// 		transactions.FromPackets4(pack)
// 	}
// }
func BenchmarkFromPacket(b *testing.B) {
	transactions := CreateRealTransactions(8192)
	pack := transactions.ToPackets(nil)
	for n := 0; n < b.N; n++ {
		transactions.FromPackets(pack)
	}
}

// func BenchmarkFromPacket2(b *testing.B) {
// 	transactions := CreateRealTransactions(8192)
// 	pack := transactions.ToPackets(nil)
// 	for n := 0; n < b.N; n++ {
// 		transactions.FromPackets2(pack)
// 	}
// }
