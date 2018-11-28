package books

import (
	"fmt"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/log"
	"github.com/Ansiblock/Ansiblock/network"
	"go.uber.org/zap"
)

// Filter out transactions with incorrect/insufficient Fee or Token value
func filter(transactions block.Transactions) block.Transactions {
	res := make([]block.Transaction, 0, len(transactions.Ts))
	count := 0
	for _, tran := range transactions.Ts {
		if tran.Verify() {
			res = append(res, tran)
			count++
		} else {
			log.Warn("transaction did not passed verification", zap.String("transaction", tran.String()))
		}
	}
	return block.Transactions{Ts: res[:count]}
}

// TransactionGenerator is responsible for creating Transactions.
// It should receive raw data through transactionReceiver channel and
// deserialize it(implement deserialization later)
// For now TransactionGenerator receives Transactions.
func TransactionGenerator(bm *Accounts, packetReceiver <-chan *network.Packets) <-chan *block.Transactions {
	out := make(chan *block.Transactions, cap(packetReceiver))
	go func(out chan<- *block.Transactions, packetReceiver <-chan *network.Packets, bm *Accounts) {
		for {
			packets, ok := <-packetReceiver
			if !ok {
				log.Error("transaction generator's packet receiver failed, closing channel")
				close(out)
				return
			}
			log.Info(fmt.Sprintf("TransactionGenerator: %v packets received", len(packets.Ps)))
			var transactions block.Transactions
			transactions.FromPackets(packets)
			transactions = bm.ProcessTransactions(filter(transactions))
			out <- &transactions
			log.Info(fmt.Sprintf("TransactionGenerator: %v transactions has been successfully processed", len(transactions.Ts)))
		}
	}(out, packetReceiver, bm)
	return out
}
