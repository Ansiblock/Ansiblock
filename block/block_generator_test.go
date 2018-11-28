package block

import (
	"bytes"
	"math/rand"
	"testing"
	"time"
)

func TestBlockGeneratorCloseChannel(t *testing.T) {
	transactionsReceiver := make(chan *Transactions)
	previousValue := VDF([]byte("hello"))
	out := Generator(transactionsReceiver, previousValue, 0)
	var transactions Transactions
	transactionsReceiver <- &transactions
	block := <-out
	if !block.Verify(previousValue) || block.Number != 1 {
		t.Errorf("can't verify block %v", block)
	}

	close(transactionsReceiver)
	block, ok := <-out
	if ok {
		t.Fatalf("BlockGenerator did not close out channel")
	}
}

func TestBlockGeneratorWithoutTransactions(t *testing.T) {
	transactionsReceiver := make(chan *Transactions)
	previousValue := VDF([]byte("hello"))
	out := Generator(transactionsReceiver, previousValue, 10)
	var transactions Transactions
	transactionsReceiver <- &transactions
	block := <-out
	if !block.Verify(previousValue) || block.Number != 11 {
		t.Fatalf("can't verify block %v", block)
	}
}

func TestBlockGeneratorWithTransactions(t *testing.T) {
	transactionsReceiver := make(chan *Transactions)
	previousValue := VDF([]byte("hello"))
	out := Generator(transactionsReceiver, previousValue, 0)
	trans := CreateDummyTransactions(10)
	transactionsReceiver <- &trans
	block := <-out
	if !block.Verify(previousValue) || block.Number != 1 {
		t.Fatalf("can't verify block %v", block)
	}
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}
func TestBlockGeneratorWithManyBlocks(t *testing.T) {
	transactionsReceiver := make(chan *Transactions)
	previousValue := VDF([]byte("hello"))
	out := Generator(transactionsReceiver, previousValue, 0)

	for i := 0; i < 10; i++ {
		n := random(0, 100)
		trans := CreateDummyTransactions(int64(n))
		transactionsReceiver <- &trans
		block := <-out
		if !block.Verify(previousValue) || block.Number != uint64(i)+1 {
			t.Fatalf("can't verify block %v", block)
		}
		previousValue = block.Val
	}
}

func TestBlockGeneratorWithLargeNumberOfTransactions(t *testing.T) {
	transactionsReceiver := make(chan *Transactions)
	previousValue := VDF([]byte("hello"))
	out := Generator(transactionsReceiver, previousValue, 0)

	trans := CreateDummyTransactions(int64(maxTransactionsInBlock*2 + 1))
	transactionsReceiver <- &trans
	for i := 0; i < 2; i++ {
		block := <-out
		if !block.Verify(previousValue) || block.Number != uint64(i)+1 || block.Transactions.Count() != maxTransactionsInBlock || block.Size() > 65536 {
			t.Fatalf("can't verify block %v", block)
		}
		previousValue = block.Val
	}
	block := <-out
	if !block.Verify(previousValue) || block.Number != 3 || block.Transactions.Count() != 1 {
		t.Fatalf("can't verify block %v", block)
	}
}

func TestBlockGeneratorWithTickCloseChannel(t *testing.T) {
	transactionsReceiver := make(chan *Transactions)
	previousValue := VDF([]byte("hello"))
	out := GeneratorWithTick(transactionsReceiver, previousValue, 0, time.Second)
	var transactions Transactions
	transactionsReceiver <- &transactions

	block := <-out
	if !block.Verify(previousValue) || block.Number != 1 {
		t.Errorf("can't verify block %v", block)
	}

	close(transactionsReceiver)
	block, ok := <-out
	if ok {
		t.Fatalf("GeneratorWithTicker did not close out channel")
	}
}

func TestBlockGeneratorWithTickWithoutTransactions(t *testing.T) {
	transactionsReceiver := make(chan *Transactions)
	previousValue := VDF([]byte("hello"))
	out := GeneratorWithTick(transactionsReceiver, previousValue, 10, time.Microsecond)
	var transactions Transactions
	transactionsReceiver <- &transactions
	block := <-out
	if !block.Verify(previousValue) || block.Number != 11 {
		t.Fatalf("can't verify block %v", block)
	}
}

func TestBlockGeneratorWithTickWithTransactions(t *testing.T) {
	transactionsReceiver := make(chan *Transactions)
	previousValue := VDF([]byte("hello"))
	out := GeneratorWithTick(transactionsReceiver, previousValue, 0, 1*time.Second)
	trans := CreateDummyTransactions(10)
	transactionsReceiver <- &trans
	block := <-out
	if !block.Verify(previousValue) || block.Number != 1 {
		t.Fatalf("can't verify block %v", block)
	}
}

func TestBlockGeneratorWithTickWithDelay(t *testing.T) {
	transactionsReceiver := make(chan *Transactions)
	previousValue := VDF([]byte("hello"))
	out := GeneratorWithTick(transactionsReceiver, previousValue, 0, 1*time.Second)
	// wait for a while to increase VDFValue
	time.Sleep(time.Microsecond)
	trans := CreateDummyTransactions(10)
	transactionsReceiver <- &trans
	block := <-out
	if !block.Verify(previousValue) || block.Number != 1 {
		t.Fatalf("can't verify block %v", block)
	}

}

func TestBlockGeneratorWithTickWithManyBlocks(t *testing.T) {
	transactionsReceiver := make(chan *Transactions, 1)
	previousValue := VDF([]byte("hello"))
	out := GeneratorWithTick(transactionsReceiver, previousValue, 0, time.Microsecond)

	go func(transactionsReceiver chan *Transactions) {
		for i := 0; i < 100; i++ {
			n := random(0, 100)
			trans := CreateDummyTransactions(int64(n))
			transactionsReceiver <- &trans
		}
	}(transactionsReceiver)

	for i := 0; i < 200; i++ {
		block := <-out
		if !block.Verify(previousValue) || block.Number != uint64(i)+1 {
			t.Errorf("can't verify block %d\n %v", i, block)
		}
		previousValue = block.Val
	}
}

func TestBlockGeneratorWithTickBlock(t *testing.T) {
	transactionsReceiver := make(chan *Transactions)
	previousValue := VDF([]byte("hello"))
	out := GeneratorWithTick(transactionsReceiver, previousValue, 0, time.Microsecond)
	// without transactions there should eventually happen tick block
	block := <-out
	if !block.Verify(previousValue) || block.Number != 1 {
		t.Fatalf("can't verify block %v", block)
	}

	previousValue = block.Val
	block = <-out
	if !block.Verify(previousValue) {
		t.Fatalf("can't verify block %v", block)
	}
}

func TestSaver(t *testing.T) {
	blocks := make([]Block, 10)
	input := make(chan Block, 10)
	for i := 0; i < 10; i++ {
		blocks[i].Val = VDF([]byte{byte(i), 2, 3})
		blocks[i].Count = 0
		trans := CreateRealTransactions(100)
		blocks[i].Transactions = &trans
		input <- blocks[i]
	}

	out := Saver(input, nil)
	bls := <-out
	if len(bls) != len(blocks) {
		t.Errorf("Saver wrong batch size %v != %v", len(bls), len(blocks))
	}
	for i := 0; i < 10; i++ {
		if bls[i].Count != blocks[i].Count || !bytes.Equal(bls[i].Val, blocks[i].Val) ||
			!bytes.Equal(bls[i].Transactions.ToBytes(), blocks[i].Transactions.ToBytes()) {
			t.Errorf("Saver %v != %v", bls[i], blocks[i])

		}
	}

}

func TestBatcher(t *testing.T) {
	blocks := make([]Block, 100)
	bch := make(chan Block, 100)
	for _, bl := range blocks {
		bch <- bl
	}
	resChan := Batcher(bch)
	res := <-resChan
	if len(res) != len(blocks) {
		t.Errorf("Batcher error: %v!=%v\n", len(res), len(blocks))
	}
}
