package replication

import (
	"bytes"
	"fmt"
	"net"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/books"
	"github.com/Ansiblock/Ansiblock/log"
	"github.com/Ansiblock/Ansiblock/network"
	"go.uber.org/zap"
)

// ReplicateRequests replicates blocks for clients
// func ReplicateRequests(bm *books.Accounts, blobsReceiver <-chan *network.Blobs) error {
// 	for {
// 		blobs, ok := <-blobsReceiver
// 		if !ok {
// 			log.Error("Replicate request failed!")
// 			return errors.New("Error while reading from blob channel")
// 		}
// 		blocks := books.BlobsToBlocks(blobs)
// 		err := bm.ProcessBlocks(blocks)
// 		if err != nil {
// 			log.Error("Process blocks failed! ", zap.Int("blobs num", len(blobs.Bs)), zap.Error(err))
// 			return err
// 		}

// 	}

// }

// New runs goroutine which in infinite loop replicates blocks
func New(bm *books.Accounts, blobsReceiver <-chan *network.Blobs) {
	go func(bm *books.Accounts, blobsReceiver <-chan *network.Blobs) {
		for {
			blobs, ok := <-blobsReceiver
			fmt.Printf("Replicate! got blobs %v\n", len(blobs.Bs))
			if !ok {
				log.Error("Replicate request failed!")
			}
			fmt.Print("blobs :[")
			for i := range blobs.Bs {
				fmt.Printf("%v ", blobs.Bs[i].Index())
			}
			fmt.Println("]")
			blocks := block.BlobsToBlocks(blobs)
			fmt.Printf("Replicate block books : [")
			for _, bl := range blocks {
				fmt.Printf("%v ", bl.Number)
			}
			fmt.Println("]")
			err := bm.ProcessBlocks(blocks)
			if err != nil {
				log.Error("Process blocks failed! ", zap.Int("blobs num", len(blobs.Bs)), zap.Error(err))
				break
			}
		}
	}(bm, blobsReceiver)
}

// Transporter thread is responsible for transporting producer blocks to other signers
func Transporter(sync *Sync, input <-chan *network.Blobs, outputConn net.PacketConn) {
	go func(sync *Sync, input <-chan *network.Blobs, outputConn net.PacketConn) {
		for {
			blobs := <-input
			log.Debug(fmt.Sprintf("Transport %v blobs", len(blobs.Bs)))
			fmt.Printf("Transport %v blobs\n", len(blobs.Bs))
			for _, blob := range blobs.Bs {
				if bytes.Equal(blob.From(), sync.ProducerNodeData().Self) {
					// log.Debug(fmt.Sprintf("Blob %v sent from producer to all nodes", blob.Index()))
					// fmt.Printf("Blob %v sent from producer to all nodes\n", blob.Index())
					transport(sync, &blob, outputConn)
				}
			}
		}
	}(sync, input, outputConn)
}

func transport(sync *Sync, blob *network.Blob, outputConn net.PacketConn) {
	blob.SetFrom(sync.MyCopy().Self)
	nodes := sync.transitNodes()
	res := make(chan bool)
	for _, node := range nodes {
		log.Debug("Transporting", zap.Uint64("Blob", blob.Index()), zap.String("Node", node.NodeName))
		go func(node *NodeData) {
			outputConn.WriteTo(blob.Data[:blob.Size], &node.Addresses.Replication)
			res <- true
		}(node)
	}
	for range nodes {
		<-res
	}
}
