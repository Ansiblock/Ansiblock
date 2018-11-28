package block

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/Ansiblock/Ansiblock/network"
	"golang.org/x/crypto/ed25519"

	"github.com/Ansiblock/Ansiblock/log"
	"go.uber.org/zap"
)

const transactionSize = ed25519.PublicKeySize*2 + ed25519.SignatureSize + 16 + sha256.Size
const maxTransactionsInBlock = int32(network.BlobRealDataSize/transactionSize - 2)

// generatorHelper holds valid VDF value and count of VDF iterations since last block
type generatorHelper struct {
	validVDFValue VDFValue
	count         uint64
	number        uint64
}

// Generator accepts three arguments, channel for reading transaction slices, last value of
// VDF and a starting number. This function creates a goroutine which creates blocks from
// transactions read from the input channel. The newly created blocks are sent through output
// channel. The output channel is created and returned to user by Generator.
func Generator(transactionsReceiver <-chan *Transactions, previousValue VDFValue, startNumber uint64) <-chan Block {
	out := make(chan Block, cap(transactionsReceiver))
	generator := generatorHelper{validVDFValue: previousValue, count: 0, number: startNumber}
	log.Info("block.Generator: create block generator goroutine with", zap.Int32("Max transactions", maxTransactionsInBlock))
	go func(out chan<- Block, transactionsReceiver <-chan *Transactions, generator generatorHelper) {
		for {
			transactions, ok := <-transactionsReceiver
			if !ok {
				log.Error("block generator's receiver failed, closing channel")
				close(out)
				return
			}

			blocksNum := transactions.Count()/(maxTransactionsInBlock+1) + 1
			blocks := make([]Block, 0, blocksNum)

			var nb Block
			if transactions.Count() == 0 {
				nb = NewEmpty(generator.validVDFValue, generator.number, generator.count)
				nb.Transactions = transactions
				blocks = append(blocks, nb)
				log.Debug("block.Generator: received zero transactions, creating empty block", zap.Uint64("Height", nb.Number))
			} else {
				var start int32
				for start < transactions.Count() {
					transactionsSlice := new(Transactions)
					end := start + maxTransactionsInBlock
					if end > transactions.Count() {
						end = transactions.Count()
					}
					transactionsSlice.Ts = transactions.Ts[start:end]
					nb = New(generator.validVDFValue, generator.number, generator.count, transactionsSlice)
					blocks = append(blocks, nb)
					start = end
					log.Info("block.Generator: create new block", zap.Uint64("Height", nb.Number), zap.Int32("TranCount", nb.Transactions.Count()))
					generator.validVDFValue = nb.Val
					generator.count = 0
					generator.number = nb.Number
				}
			}

			for i := range blocks {
				out <- blocks[i]
			}
		}
	}(out, transactionsReceiver, generator)
	return out
}

// GeneratorWithTick accept an input channel to read transactions and hash of previous
// block. It creates an output channel and returns to the user. The transactions slices are sent
// through the input channel and newly creates blocks are returned through the out channel.
// GeneratorWithTick creates new blocks from transactions, also it creates timestamping
// blocks at very 'tick'. The tick duration is defined by the user. The "tick block" is just the
// next block in blockchain without transactions.
func GeneratorWithTick(transactionsReceiver <-chan *Transactions, previousValue VDFValue, startNumber uint64, tickDuration time.Duration) <-chan Block {
	out := make(chan Block, cap(transactionsReceiver))
	generator := generatorHelper{validVDFValue: previousValue, count: 0, number: startNumber}
	log.Debug("create block generator with tick goroutine")
	go func(transactionsReceiver <-chan *Transactions, generator generatorHelper, tickDuration time.Duration) {
		ticker := time.NewTicker(tickDuration)
		defer ticker.Stop()

		var nb Block
		for {
			select {
			case <-ticker.C:
				nb = NewEmpty(generator.validVDFValue, generator.number, generator.count)
				generator.count = 0
				generator.validVDFValue = nb.Val
				generator.number = nb.Number
				log.Debug("block generator with tick > tick received, create empty block")
				out <- nb
			case transactions, ok := <-transactionsReceiver:
				if !ok {
					log.Error("block generator's receiver failed, closing channel")
					close(out)
					return
				}

				if transactions.Count() == 0 {
					nb = NewEmpty(generator.validVDFValue, generator.number, generator.count)
					log.Debug("block generator with tick > received zero transactions, create empty block")
				} else {
					nb = New(generator.validVDFValue, generator.number, generator.count, transactions)
					log.Info("block generator with tick > create new block")
				}
				generator.count = 0
				generator.validVDFValue = nb.Val
				generator.number = nb.Number
				out <- nb
			default:
				generator.validVDFValue = VDF(generator.validVDFValue)
				generator.count++
				log.Debug("block generator with tick > update VDF value")
			}
		}
	}(transactionsReceiver, generator, tickDuration)
	return out
}

type BlockSaver interface {
	SaveBlock(blk Block) error
}

func saveBlockInDB(bl Block, db BlockSaver) {
	log.Debug("Saving block", zap.Uint64("Height", bl.Number))
	err := db.SaveBlock(bl)
	if err != nil {
		log.Error(fmt.Sprintf("Can't save block %v", bl))
	}

}

// func saveBlockInFile(bl Block) {
// 	b, err := json.Marshal(bl)
// 	if err != nil {
// 		log.Error(fmt.Sprintf("Can't marshal block %v", bl))
// 	}

// 	f, err := os.OpenFile("blocks.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
// 	if err != nil {
// 		log.Error("Error while opening file: ", zap.Error(err))
// 	}

// 	defer f.Close()

// 	if _, err = f.Write(b); err != nil {
// 		log.Error("Error while writing to file: ", zap.Error(err))
// 	}
// }

func saveBlock(bl Block, db BlockSaver) {
	if db != nil {
		saveBlockInDB(bl, db)
	}
	// saveBlockInFile(bl)
}

// Saver thread saves blocks in the file and creates slice of blocks
// to convert to blobs and send to the socket
func Saver(blocks <-chan Block, db BlockSaver) <-chan []Block {
	out := make(chan []Block, 10)
	go func() {
		// index := 0
		for {
			// fmt.Printf("in for %v\n", index)
			// index++
			bls := make([]Block, 0, 10)
		LoopForBlocks:
			for {
				var block Block
				select {
				case block = <-blocks:
					saveBlock(block, db)
					bls = append(bls, block)
					// fmt.Printf("blocks size1 = %v\n", len(bls))
				case <-time.After(1 * time.Second):
					// fmt.Println("break1")
					break LoopForBlocks
				}
				for {
					select {
					case block = <-blocks:
						saveBlock(block, db)
						bls = append(bls, block)
						// fmt.Printf("blocks size2 = %v\n", len(bls))
					default:
						// fmt.Println("break2")
						break LoopForBlocks
					}
				}
			}
			if len(bls) != 0 {
				// fmt.Printf("size3 = %v\n", len(bls))
				out <- bls
			}
		}
	}()
	return out
}

// Batcher thread converts blocks to batches
// to convert to blobs and send to the socket
func Batcher(blocks <-chan Block) <-chan []Block {
	out := make(chan []Block, 10)
	go func() {
		for {
			bls := make([]Block, 0, 10)
			timeout := time.After(1 * time.Second)
		LoopForBlocks:
			for {
				var block Block
				select {
				case block = <-blocks:
					bls = append(bls, block)

				case <-timeout:
					break LoopForBlocks

				}
			}
			if len(bls) != 0 {
				out <- bls
			}
		}
	}()
	return out
}

// BatchSaver saves given batch to database.
// The database is used for api access
func BatchSaver(batch []Block, db BlockSaver) {
	for _, b := range batch {
		saveBlock(b, db)
	}
}
