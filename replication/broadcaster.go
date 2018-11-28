package replication

import (
	"crypto/sha256"
	"fmt"
	"net"

	"github.com/Ansiblock/Ansiblock/log"
	"github.com/Ansiblock/Ansiblock/network"
)

func chunkUpBlobs(blobs *network.Blobs, size int) [][]*network.Blob {
	numChunks := len(blobs.Bs)/size + 1
	res := make([][]*network.Blob, numChunks)
	for i := 0; i*size < len(blobs.Bs); i++ {
		chunkSize := size
		if i+1 == numChunks {
			// the last chunk will be smaller then provided size
			chunkSize = len(blobs.Bs) - i*size
		}
		res[i] = make([]*network.Blob, chunkSize)
		for j := 0; j < size && i*size+j < len(blobs.Bs); j++ {
			res[i][j] = &blobs.Bs[i*size+j]
		}
	}
	return res
}

// Broadcaster thread is responsible for broadcasting blocks from producer to signers
func Broadcaster(sync *Sync, frame *network.Frame, outputConn net.PacketConn, input chan *network.Blobs) {
	go func() {
		index := uint64(0) // global blob index
		for {
			blobs := <-input
			nodes := sync.transitNodes()
			if len(nodes) < 1 {
				log.Info("No nodes to broadcast")
				continue
			}
			// codding blocks should be added before chunking,
			// or else batch will be greater then framesize
			// fmt.Printf("Broadcaster blobs len = %v\n", len(blobs.Bs))
			network.AddCodingBlobs2(blobs, int(index))
			// fmt.Printf("Broadcaster after AddCodingBlobs2 blobs len = %v\n", len(blobs.Bs))
			blobs.IndexBlobs(sync.MyCopy().Self, index)
			batches := chunkUpBlobs(blobs, network.FrameSize)
			// fmt.Printf("Broadcaster batches len = %v\n", len(batches))
			for i := 0; i < len(batches); i++ {
				for _, blob := range batches[i] {
					ind := blob.Index()
					frame.Blobs[ind%network.FrameSize] = blob
					fmt.Printf("Framed blob %v hash: %v, size %v\n", ind, sha256.Sum256(blob.Data[0:blob.Size]), blob.Size)

					// fmt.Println("Broadcaster before EncodeRS blob index size n", blob.Index(), blob.Size, block.ByteToInt32(blob.Data[network.DataOffset+16+32:], 0))
				}
				network.EncodeRS(frame.Blobs, index, uint64(len(batches[i])))
				for _, blob := range batches[i] {
					ind := blob.Index()
					fmt.Printf("2Framed blob %v hash: %v, size %v\n", ind, sha256.Sum256(frame.Blobs[ind%network.FrameSize].Data[0:frame.Blobs[ind%network.FrameSize].Size]), frame.Blobs[ind%network.FrameSize].Size)

					// fmt.Println("Broadcaster before EncodeRS blob index size n", blob.Index(), blob.Size, block.ByteToInt32(blob.Data[network.DataOffset+16+32:], 0))
				}

				index += uint64(len(batches[i]))
				// fmt.Printf("Index updated to %v\n", index)

				// broadcast
				for i, blob := range batches[i] {
					outputConn.WriteTo(blob.Data[:blob.Size], &nodes[i%len(nodes)].Addresses.Replication)
					// log.Debug(fmt.Sprintf("broadcast from %v blob %v to %v", outputConn.LocalAddr().String(), blob.Index(), nodes[i%len(nodes)]))
					// fmt.Println("*******broadcasted blob index size n", blob.Index(), blob.Size, block.ByteToInt32(blob.Data[network.DataOffset+16+32:], 0))

					// fmt.Printf("broadcast from %v blob %v size %v to %v\n", outputConn.LocalAddr().String(), blob.Index(), blob.Size, nodes[i%len(nodes)].Addresses.Replication)
					// time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}()
}
