package replication

import (
	"bytes"
	"math/rand"
	"testing"
	"time"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/books"
	"github.com/Ansiblock/Ansiblock/network"
)

// func TestReplicator(t *testing.T) {
// 	rand.Seed(0)
// 	bm, keyPairs := RandomAccounts(100)
// 	bmClone := bm.Clone()

// 	blocks := RandomTransactionsBlocks(bm, 200, 20, keyPairs)
// 	blobs := block.BlocksToBlobs(blocks)
// 	blobsReceiver := make(chan *network.Blobs, 1)
// 	blobsReceiver <- blobs

// 	bm.ProcessBlocks(blocks)
// 	ReplicateRequests(bmClone, blobsReceiver)
// 	if !bm.Equals(bmClone) && bm.TransactionsTotal() == 3727 {
// 		t.Errorf("Replicate Requests failed! ")
// 	}
// }

func TestReplicatorNew(t *testing.T) {
	rand.Seed(0)
	bm, keyPairs := books.RandomAccounts(100)

	blocks := books.RandomTransactionsBlocks(bm, 100, 10, keyPairs)
	bmClone := bm.Clone()
	blobs := block.BlocksToBlobs(blocks)
	blobsReceiver := make(chan *network.Blobs, 1)
	blobsReceiver <- blobs

	bm.ProcessBlocks(blocks)
	// var exit uint64
	New(bmClone, blobsReceiver)

	//wait while replicator thread is finished
	time.Sleep(2 * time.Second)
	// atomic.AddUint64(&exit, 1)

	if !bm.Equals(bmClone) || bm.TransactionsTotal() != 936 {
		t.Errorf("Replicate new failed! transaction %v =? 936 ", bm.TransactionsTotal())
	}
}

func TestTransport(t *testing.T) {
	blob := network.NewBlobs().Bs[0]
	blob.Data[0] = 1
	blob.Data[1] = 2
	blob.Data[2] = 3
	blob.Size = 3
	producer := NewNode("producer", "test")
	producer.Data.Producer = producer.Data.Self
	syncL, _ := NewSync(producer.Data)
	syncL.Insert(producer.Data)

	conn := network.NewSocketMock(nil, nil, nil)
	transport(syncL, &blob, conn)
	if conn.WriteBuffSize() != 0 {
		t.Error("error")
	}
}

func TestTransport2(t *testing.T) {
	blob := network.NewBlobs().Bs[0]
	blob.Data[0] = 1
	blob.Data[1] = 2
	blob.Data[2] = 3
	blob.Size = 3
	producer := NewNode("producer", "test1")
	producer.Data.Producer = producer.Data.Self
	syncL, _ := NewSync(producer.Data)
	val := NewNode("signer", "test2")
	val.Data.Producer = producer.Data.Self
	syncL.Insert(val.Data)

	conn := network.NewSocketMock(nil, nil, nil)
	// conn.WriteToSize = int(10)
	transport(syncL, &blob, conn)
	writeBuff := conn.EmptyWriteBuff()
	if !bytes.Equal(writeBuff, blob.Data[:blob.Size]) ||
		conn.Addr.String() != val.Sockets.Replicate.LocalAddr().String() {
		t.Error("error")
	}
}
