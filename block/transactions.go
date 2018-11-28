package block

import (
	"crypto/sha256"
	"net"
	"runtime"

	"github.com/Ansiblock/Ansiblock/network"
	"golang.org/x/crypto/ed25519"
)

// Transactions represents slice of Transaction-s
type Transactions struct {
	Ts []Transaction
}

// String method of Transactions struct
func (ts *Transactions) String() string {
	res := ""
	for _, tran := range ts.Ts {
		res += tran.String()
	}
	return res
}

// ToBytes method converts slice of transactions to the slice of bytes
func (ts *Transactions) ToBytes() []byte {
	transactionData := make([]byte, 0, (ed25519.SignatureSize+1)*len(ts.Ts))
	for _, tran := range ts.Ts {
		transactionData = append(transactionData, 0)
		transactionData = append(transactionData, tran.Signature...)
	}
	return transactionData
}

// Verify method verifies each transaction
func (ts *Transactions) Verify() bool {
	if len(ts.Ts) > 1000 {
		return ts.verifyThreads()
	}
	return ts.verifySerial()
}

func helper(ts *Transactions, start int, finish int, res chan bool) {
	for j := start; j < finish; j++ {
		if !ts.Ts[j].Verify() {
			res <- false
			return
		}
	}
	res <- true
}

func (ts *Transactions) verifyThreads() bool {
	threads := runtime.NumCPU()
	res := make(chan bool, threads)
	l := len(ts.Ts)
	for i := 0; i < threads-1; i++ {
		go helper(ts, i*l/threads, (i+1)*l/threads, res)
	}
	go helper(ts, (threads-1)*l/threads, l, res)
	for i := 0; i < threads; i++ {
		if !<-res {
			return false
		}
	}
	return true
}

func (ts *Transactions) verifySerial() bool {
	for _, trans := range ts.Ts {
		if !trans.Verify() {
			return false
		}
	}
	return true
}

// Count method returns number of transactions
func (ts *Transactions) Count() int32 {
	return int32(len(ts.Ts))
}

// Equals comapres two slices of transactions
func (ts *Transactions) Equals(trs Transactions) bool {
	if len(ts.Ts) != len(trs.Ts) {
		return false
	}
	for i, tr := range ts.Ts {
		if !tr.Equals(trs.Ts[i]) {
			return false
		}
	}
	return true
}

func fromPackets(ts *Transactions, start int, finish int, packets *network.Packets, res chan bool) {
	for i := start; i < finish; i++ {
		if packets.Ps[i].Size != 0 {
			ts.Ts[i].Deserialize(&packets.Ps[i].Data)
		} else {
			ts.Ts[i] = Transaction{Fee: -1}
		}
	}
	res <- true
}

// FromPackets method creates Transactions from Packets
func (ts *Transactions) FromPackets(packets *network.Packets) {
	ts.Ts = make([]Transaction, len(packets.Ps))
	threads := runtime.NumCPU()
	res := make(chan bool, threads)
	l := len(ts.Ts)

	for i := 0; i < threads-1; i++ {
		go fromPackets(ts, i*l/threads, (i+1)*l/threads, packets, res)
	}

	go fromPackets(ts, (threads-1)*l/threads, l, packets, res)
	for i := 0; i < threads; i++ {
		<-res
	}
}

// ToPackets method creates Packets from Transactions
func (ts *Transactions) ToPackets(addr net.Addr) *network.Packets {
	packets := new(network.Packets)
	packets.Ps = make([]network.Packet, len(ts.Ts))
	counter := make(chan bool, 100)
	for i := range ts.Ts {
		go func(i int) {
			packets.Ps[i].Data = *ts.Ts[i].Serialize()
			packets.Ps[i].Size = uint8(TransactionSize())
			packets.Ps[i].Addr = addr
			counter <- true
		}(i)
	}
	for range ts.Ts {
		<-counter
	}
	return packets
}

// Size method returns size of transactions in bytes
func (ts *Transactions) Size() int {
	return len(ts.Ts) * TransactionSize()
}

// countNumberOfBlocksInBlob determines how many blocks will be in a single block
// starting from the start block. It returns index of the end block( including)
func countNumberOfBlocksInBlob(blocks []Block, start int) int {
	byteCount := 0
	for end := start; end < len(blocks); end++ {
		byteCount += blocks[end].Size()

		if byteCount >= network.BlobDataSize-network.DataOffset {
			return end - 1
		}
	}
	return len(blocks) - 1
}

func convertToBlob(blocks []Block, start int, end int) *network.Blob {
	res := new(network.Blob)
	st := network.DataOffset
	tranSize := TransactionSize()
	for i := start; i <= end; i++ {
		res.Data[st+0] = byte(blocks[i].Number >> 56)
		res.Data[st+1] = byte(blocks[i].Number >> 48)
		res.Data[st+2] = byte(blocks[i].Number >> 40)
		res.Data[st+3] = byte(blocks[i].Number >> 32)
		res.Data[st+4] = byte(blocks[i].Number >> 24)
		res.Data[st+5] = byte(blocks[i].Number >> 16)
		res.Data[st+6] = byte(blocks[i].Number >> 8)
		res.Data[st+7] = byte(blocks[i].Number)
		st += 8
		res.Data[st+0] = byte(blocks[i].Count >> 56)
		res.Data[st+1] = byte(blocks[i].Count >> 48)
		res.Data[st+2] = byte(blocks[i].Count >> 40)
		res.Data[st+3] = byte(blocks[i].Count >> 32)
		res.Data[st+4] = byte(blocks[i].Count >> 24)
		res.Data[st+5] = byte(blocks[i].Count >> 16)
		res.Data[st+6] = byte(blocks[i].Count >> 8)
		res.Data[st+7] = byte(blocks[i].Count)
		for j := 0; j < VDFSize; j++ {
			res.Data[st+8+j] = blocks[i].Val[j]
		}
		count := blocks[i].Transactions.Count()
		res.Data[st+40] = byte(count >> 24)
		res.Data[st+41] = byte(count >> 16)
		res.Data[st+42] = byte(count >> 8)
		res.Data[st+43] = byte(count)
		trans := blocks[i].Transactions
		for j := 0; j < int(count); j++ {
			tranData := trans.Ts[j].Serialize()
			for k := 0; k < tranSize; k++ {
				res.Data[st+44+j*tranSize+k] = tranData[k]
			}
		}
		st += 44 + tranSize*int(count)
	}
	res.Size = uint32(st)
	// fmt.Println("++++++++convertToBlob blob index size n", res.Index(), res.Size, ByteToInt32(res.Data[network.DataOffset+16+32:], 0))
	return res
}

// BlocksToBlobs converts blocks to blobs. Assumes that block is no large then blob size.
func BlocksToBlobs(blocks []Block) *network.Blobs {
	res := new(network.Blobs)
	res.Bs = make([]network.Blob, len(blocks))
	start := 0
	end := 0
	blobIndex := 0
	for start < len(blocks) {
		end = countNumberOfBlocksInBlob(blocks, start)
		if end < start {
			panic("something's wrong")
		}
		res.Bs[blobIndex] = *convertToBlob(blocks, start, end)
		blobIndex++
		start = end + 1

	}
	res.Bs = res.Bs[0:blobIndex]
	return res
}

// BlobsToBlocks will deserialize slice of blobs into slice of Blocks
// TODO should be optimized!!!
func BlobsToBlocks(blobs *network.Blobs) []Block {
	blockMap := make(map[string]Block)
	blockIndexesMap := make(map[string]int)

	blockIndex := 0
	for _, bl := range blobs.Bs {
		start := network.DataOffset
		for start < int(bl.Size) {
			number := ByteToInt64(bl.Data[start:], 0)
			start += 8
			count := ByteToInt64(bl.Data[start:], 0)
			vdf := make([]byte, sha256.Size)
			start += 8
			for j := 0; j < sha256.Size; j++ {
				vdf[j] = bl.Data[start+j]
			}
			start += sha256.Size
			n := ByteToInt32(bl.Data[start:], 0)
			start += 4
			transactions := make([]Transaction, n)
			for j := int32(0); j < n; j++ {
				// fmt.Printf("bl.Data = %v\n", len(bl.Data))
				// fmt.Printf("start = %v\n", start)
				transactions[j].DeserializeFromSlice(bl.Data[start:])
				start += 176
			}
			if bl, ok := blockMap[string(vdf)]; ok {
				trans := bl.Transactions
				trans.Ts = append(trans.Ts, transactions...)
			} else {
				b := new(Block)
				b.Count = uint64(count)
				b.Val = make([]byte, sha256.Size)
				b.Number = uint64(number)
				for i := range vdf {
					b.Val[i] = vdf[i]
				}
				trans := new(Transactions)
				for i := range transactions {
					trans.Ts = append(trans.Ts, transactions[i])
				}
				b.Transactions = trans
				blockMap[string(vdf)] = *b
				blockIndexesMap[string(vdf)] = blockIndex
				blockIndex++
			}
		}
	}
	blocks := make([]Block, len(blockMap))
	for k, v := range blockMap {
		blocks[blockIndexesMap[k]] = v
	}
	return blocks
}

// CreateDummyTransactions responsible for creating dummy Transactions object
// NOTE: Mainly for testing
func CreateDummyTransactions(n int64) Transactions {
	trs := make([]Transaction, n)
	for i := int64(0); i < n; i++ {
		trs[i] = CreateDummyTransaction(i)
	}
	return Transactions{Ts: trs}
}

// CreateRealTransactions responsible for creating real Transactions object
// NOTE: Mainly for testing
func CreateRealTransactions(n int64) Transactions {
	trs := make([]Transaction, n)
	for i := int64(0); i < n; i++ {
		trs[i] = CreateRealTransaction(i)
	}
	return Transactions{Ts: trs}
}

// CreateRealTransactionsFrom responsible for creating real Transactions object
// NOTE: Mainly for testing
func CreateRealTransactionsFrom(n int64, from []byte) Transactions {
	trs := make([]Transaction, n)
	for i := int64(0); i < n; i++ {
		trs[i] = CreateRealTransactionFrom(i, from)
	}
	return Transactions{Ts: trs}
}

func transactionInList(tr Transaction, trs []Transaction) bool {
	for _, tran := range trs {
		if tran.Equals(tr) {
			return true
		}
	}
	return false
}

// TransactionSetEqual is responsible for checking equality between transactions
// Note: Mainly for testing
func TransactionSetEqual(trSet1, trSet2 []Transaction) bool {
	if len(trSet1) != len(trSet2) {
		return false
	}
	for _, tr := range trSet1 {
		if !transactionInList(tr, trSet2) {
			return false
		}
	}
	for _, tr := range trSet2 {
		if !transactionInList(tr, trSet1) {
			return false
		}
	}
	return true
}
