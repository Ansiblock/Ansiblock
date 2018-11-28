package block

import (
	"bytes"
	"fmt"
	"testing"
)

func TestNewEmpty(t *testing.T) {
	val := VDF([]byte("hello"))
	var count uint64
	count = 10
	block := NewEmpty(val, 0, count)
	blockValS := fmt.Sprintf("%x", block.Val)
	valS := fmt.Sprintf("%x", val)
	if blockValS != valS {
		t.Fatalf("NewEmpty Function: NewEmpty(%v, %v) = %v want {%v, %v, %v}", val, count, block, val, count, nil)
	}
}

// TODO test Transactions
func TestNew(t *testing.T) {
	val0 := VDF([]byte("hello"))
	val1 := VDF(val0)
	transactions := new(Transactions)
	block := New(val0, 10, 0, transactions)
	if block.Count != 0 || !bytes.Equal(block.Val, val0) || block.Number != 11 {
		t.Errorf("New Function1: New(%v, %v, %v) = %v want {%v, %v, %v}", val0, 0, transactions, block, val0, 0, transactions)
	}

	block = New(VDF(val0), 0, 1, transactions)
	if block.Count != 1 || !bytes.Equal(block.Val, val1) || block.Number != 1 {
		t.Errorf("New Function2: New(%v, %v, %v) = %v want {%v, %v, %v}", val0, 1, transactions, block, val1, 1, transactions)
	}
}

func TestVerify(t *testing.T) {
	val0 := VDF([]byte("hello"))
	transactions := new(Transactions)
	block := New(VDF(val0), 0, 1, transactions)
	if !block.Verify(val0) {
		t.Fatalf("Verify Function: can't Verify(%v, %v)", block, val0)
	}
}

func TestVerifyFalse(t *testing.T) {
	val0 := VDF([]byte("hello"))
	keypair1 := NewKeyPair()
	keypair2 := NewKeyPair()
	tran1 := NewTransaction(&keypair1, keypair2.Public, 1, 5, []byte{6, 7, 8})
	transactions := Transactions{Ts: []Transaction{tran1}}
	block := New(VDF(val0), 0, 1, &transactions)
	if block.Verify(val0) {
		t.Fatalf("Verify Function: can't Verify(%v, %v)", block, val0)
	}
}
func TestBlockString(t *testing.T) {
	keypair1 := NewKeyPair()
	keypair2 := NewKeyPair()
	tran := NewTransaction(&keypair1, keypair2.Public, 3, 4, []byte{5, 6})
	trans := &Transactions{Ts: []Transaction{tran}}
	// b := Block{count: 1, val: []byte{1, 2}, transactions: trans}
	b := New([]byte{1, 2}, 0, 1, trans)
	if b.String() != trans.String() {
		t.Errorf("Block.String(%v) = %v Failed wanted %v", b, b.String(), trans.String())
	}
}

func TestNextBlock(t *testing.T) {
	val0 := VDF([]byte("hello"))
	val1 := VDF(val0)
	transactions := new(Transactions)
	bl := NextBlock(val0, 0, 1, transactions)
	if bl.Count != 1 || !bytes.Equal(bl.Val, val1) || bl.Number != 1 {
		t.Errorf("NextBlock Error")
	}
}

func BenchmarkBlock1000(b *testing.B) {
	val0 := VDF([]byte("hello"))
	transactions := CreateDummyTransactions(100)
	for n := 0; n < b.N; n++ {
		New(val0, 0, 1000, &transactions)
	}
}

func BenchmarkBlock1000WithLoop(b *testing.B) {
	val := VDF([]byte("hello"))
	transactions := CreateDummyTransactions(100)
	block := New(val, 0, 1000, &transactions)
	for n := 0; n < b.N; n++ {
		for i := 0; i < 1000; i++ {
			val = block.Val
			block = New(val, 0, 1, &transactions)
		}
	}
}
func TestSize(t *testing.T) {
	val := VDF([]byte("hello"))
	transactions := CreateDummyTransactions(100)
	block := New(val, 0, 1000, &transactions)
	size := 52 + transactions.Size()
	if size != block.Size() {
		t.Errorf("Size: wring size")
	}
}
