package mint

import (
	"bytes"
	"testing"

	"github.com/Ansiblock/Ansiblock/block"
)

func TestNewMint(t *testing.T) {
	mint := NewMint(1000)

	if mint.PrivateKey == nil || mint.PublicKey == nil {
		t.Fatalf("Mint initialization failed. key == nil!")
	}
	if mint.Tokens != 1000 {
		t.Fatalf("Mint initialization failed. Incorrect token value!")
	}
}

func TestCreateTransactions(t *testing.T) {
	mint := NewMint(1000)
	transactions := mint.CreateTransactions()
	tran := block.NewTransaction(&mint.KeyPair, mint.PublicKey, mint.Tokens, 0, block.VDF(mint.KeyPair.Private))

	if len(transactions) != 1 || !tran.Equals(transactions[0]) {
		t.Fatalf("Mint CreateTransactions failed, wrong transaction data.")
	}
}

func TestCreateCreateBlocks(t *testing.T) {
	mint := NewMint(1000)
	blocks := mint.CreateBlocks()
	if len(blocks) != 2 {
		t.Errorf("Mint CreateBlocks failed, invalid number of blocks.")
	}
	b1 := block.NewEmpty(block.VDF(mint.KeyPair.Private), 0, 0)
	trans := block.Transactions{}
	tran := block.NewTransaction(&mint.KeyPair, mint.PublicKey, mint.Tokens, 0, block.VDF(mint.KeyPair.Private))
	trans.Ts = append(trans.Ts, tran)
	b2 := block.New(b1.Val, b1.Number, 0, &trans)

	if !bytes.Equal(b1.Val, blocks[0].Val) || !bytes.Equal(b2.Val, blocks[1].Val) || b2.Number != b1.Number+1 {
		t.Fatalf("Mint CreateBlocks failed, vdf value not equal.")
	}
}

func TestValidVDFValue(t *testing.T) {
	mint := NewMint(1000)
	blocks := mint.CreateBlocks()
	if !bytes.Equal(blocks[1].Val, mint.ValidVDFValue()) {
		t.Errorf("Error in ValidVDFValue")
	}
}
