package block

import (
	"bytes"
)

// Block represents main struct of our blockchain
// It uses Verifiable Delayed Function for computational timestamping.
// VDF can be constructed using incrementally verifiable computation (IVC),
// in which a proof of correctness for a computation of length t can be
// computed in parallel to the computation with only polylog(t) processors.
// Block uses SHA256 as an Incremental VDF.
// val1 = VDF(val0)
// val2 = VDF(val1)
// ...
// valN = VDF(valN-1)
// `count` field in Block represents the number of iterations in Incremental VDF
// `val` field represents the final value after `count` iterations
// `transactions` is the slice of transactions incorporated in the Block
// if `transactions` is not empty then Block performes one additional iteration
// and adds transactions' data into the generated VDF value.
// This way we can timestamp the transactions.
// Note: application of VDFs is fragile as it requires precise bounds on the
// attacker’s computation speed.
type Block struct {
	Number       uint64
	Count        uint64
	Val          VDFValue
	Transactions *Transactions
}

// * Transactions provides an interface for objects to store in the Block
// type * Transactions interface {
// 	Count() int32
// 	Verify() bool
// 	String() string
// 	ToBytes() []byte
// 	Size() int
// }

// NewEmpty returns an empty Block
func NewEmpty(previousValue VDFValue, previousBlockNumber uint64, count uint64) Block {
	return Block{Number: previousBlockNumber + 1, Count: count, Val: previousValue}
}

// New returns a Block with transactions
func New(previousValue VDFValue, previousBlockNumber uint64, count uint64, trans *Transactions) Block {
	vdfValue, c := nextVDFForNew(previousValue, trans)
	return Block{Number: previousBlockNumber + 1, Count: c + count, Val: vdfValue, Transactions: trans}
}

// NextBlock returns Block with increased VDFValue
func NextBlock(previousValue VDFValue, previousBlockNumber uint64, count uint64, trans *Transactions) Block {
	nextVal, _ := nextVDF(previousValue, count, trans)
	return Block{Number: previousBlockNumber + 1, Count: count, Val: nextVal, Transactions: trans}
}

// Verify method verifies that block is legal
// TODO should be added transaction verification
func (b *Block) Verify(previousValue VDFValue) bool {
	if b.Transactions != nil && !b.Transactions.Verify() {
		return false
	}
	vdfValue, _ := nextVDF(previousValue, b.Count, b.Transactions)
	return bytes.Equal(b.Val, vdfValue)
}

// Size method returns size of the block
func (b *Block) Size() int {
	byteCount := 0
	byteCount += 8           //block.number
	byteCount += 8 + VDFSize // block.Count + block.VDFValue
	byteCount += 4           //len(block.transactions)
	if b.Transactions != nil {
		byteCount += b.Transactions.Size()
	}
	return byteCount
}

// String method of Block struct
func (b *Block) String() string {
	return b.Transactions.String()
}

// nextVDF returns next value of vdf and count number.
// nextVDF uses VDF function count number of times and once more
// if transactions is not empty.
func nextVDF(previousValue VDFValue, count uint64, transactions *Transactions) (VDFValue, uint64) {
	value := previousValue
	returnCount := count
	for i := uint64(1); i < count; i++ {
		value = VDF(value)
	}
	value, c := nextVDFForNew(value, transactions)
	if c == 0 && count != 0 {
		value = VDF(value)
	}
	return value, returnCount
}

func nextVDFForNew(previousValue VDFValue, transactions *Transactions) (VDFValue, uint64) {
	value := previousValue
	returnCount := uint64(0)
	if transactions != nil && transactions.Count() != 0 {
		transactionData := transactions.ToBytes()
		value = ExtendedVDF(transactionData, value)
		returnCount++
	}
	return value, returnCount
}
