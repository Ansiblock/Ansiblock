package reconstruction

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"net"

	"github.com/Ansiblock/Ansiblock/log"
	"github.com/Ansiblock/Ansiblock/network"
	"github.com/Ansiblock/Ansiblock/replication"
	"go.uber.org/zap"
)

func handleReconstruct(sync *replication.Sync, frame *network.Frame, blob *network.Blob) *network.Blob {
	req := new(Request)
	req.Deserialize(blob)
	sync.Insert(req.From)
	me := sync.MyCopy()
	if bytes.Equal(req.From.Self, me.Self) {
		log.Debug("Ignore Reconstruct request from self")
		return nil
	}
	fmt.Printf("requested blob: %v\n", req.Index)
	if frame.Blobs[req.Index%network.FrameSize] != nil && frame.Blobs[req.Index%network.FrameSize].Index() == req.Index {
		//found the blob
		newBlob := new(network.Blob)
		newBlob.Size = frame.Blobs[req.Index%network.FrameSize].Size
		copy(newBlob.Data[0:newBlob.Size], frame.Blobs[req.Index%network.FrameSize].Data[0:newBlob.Size])
		addr := req.From.Addresses.Replication
		newBlob.Addr = &addr
		// TODO logarithmic transport if I'm producer
		newBlob.SetFrom(req.From.Self)
		fmt.Printf("Found blob %v hash: %v size %v\n", newBlob.Index(), sha256.Sum256(newBlob.Data[0:newBlob.Size]), newBlob.Size)
		return newBlob
	}
	return nil
}

// Listener listens to the incoming requests and sends missing blobs to the nodes
func Listener(sync *replication.Sync, frame *network.Frame, input <-chan *network.Blobs) <-chan *network.Blobs {
	out := make(chan *network.Blobs, 10)
	go func(out chan<- *network.Blobs) {
		for {
			blobs := <-input
			log.Debug("Listener", zap.Int("len(blobs)", len(blobs.Bs)))
			// fmt.Println("============here=======================")
			responses := make([]network.Blob, 0, len(blobs.Bs))
			for _, blob := range blobs.Bs {
				resp := handleReconstruct(sync, frame, &blob)
				if resp != nil {
					responses = append(responses, *resp)
				}
			}
			if len(responses) > 0 {
				// for _, r := range responses {
				// 	fmt.Printf(" blob size %v\n", r.Size)
				// }
				out <- &network.Blobs{Bs: responses}
			}
		}
	}(out)
	return out
}

// Reconstruct runs the reconstruction fo the missing blobs
func Reconstruct(frame *network.Frame, input <-chan *network.Blobs, sync *replication.Sync, outputConn net.PacketConn) chan *network.Blobs {
	doneBlobs, missingIndexes := network.FrameGenerator(frame, input)
	go func() {
		for indexes := range missingIndexes {
			for _, i := range indexes {
				req := new(Request)
				req.Index = i
				req.From = sync.MyCopy()
				reqBlob := req.Serialize()
				randomNode, _ := sync.RandomNode()
				log.Debug(fmt.Sprintf("Reconstruct sending blob %v from %v to %v", i, req.From.NodeName, randomNode.NodeName))
				fmt.Printf("sent blob %v to %v\n", i, randomNode.Addresses.Repair)
				outputConn.WriteTo(reqBlob.Data[0:reqBlob.Size], &randomNode.Addresses.Repair)
			}
		}
	}()
	return doneBlobs
}
