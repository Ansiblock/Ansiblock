package pipelines

import (
	"fmt"
	"net"
	"time"

	"github.com/Ansiblock/Ansiblock/api"
	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/books"
	"github.com/Ansiblock/Ansiblock/log"
	"github.com/Ansiblock/Ansiblock/messaging"
	"github.com/Ansiblock/Ansiblock/network"
	"github.com/Ansiblock/Ansiblock/reconstruction"
	"github.com/Ansiblock/Ansiblock/replication"
)

const (
	blockGenerationChannelCapacity = 1
	messageChannelCapacity         = 10
	synchronizationChannelCapacity = 10
	signerChannelCapacity          = 10
	synchronizationTimeoutDuration = 1000 * time.Millisecond
)

// BlockGeneration is run on the producer node and is responsible for transaction processing and generating blocks
func BlockGeneration(bm *books.Accounts, sync *replication.Sync, inputConn net.PacketConn, outputConn net.PacketConn, reconstructionConn net.PacketConn, startingBlocksTotal uint64, db api.DataBase) {
	packets := network.PacketGenerator(inputConn, blockGenerationChannelCapacity)
	filteredPackets := books.SignatureVerification(packets)
	transactions := books.TransactionGenerator(bm, filteredPackets)
	blocks := block.Generator(transactions, bm.ValidVDFValue(), startingBlocksTotal)
	batch := block.Saver(blocks, db)
	blobs := make(chan *network.Blobs, cap(batch))
	index := int32(0)
	frame := network.NewFrame()
	go replication.Broadcaster(sync, frame, outputConn, blobs)
	if reconstructionConn != nil {
		go RequestBlobs(sync, frame, reconstructionConn, outputConn)
	}
	// go func() {
	for b := range batch {
		num := int32(0)
		for _, block := range b {
			fmt.Printf("block %v\n", block.Number)
			bm.UpdateLastBlock(&block)
			num += block.Transactions.Count()
			index += block.Transactions.Count()
		}
		fmt.Printf("broadcasting %v transactions. sum = %v\n", num, index)
		blobs <- block.BlocksToBlobs(b)
	}
	// }()
}

// BlockGeneration is run on the producer node and is responsible for transaction processing and generating blocks
func BlockGenerationFaster(bm *books.Accounts, sync *replication.Sync, inputConn net.PacketConn, outputConn net.PacketConn, reconstructionConn net.PacketConn, startingBlocksTotal uint64, db api.DataBase) {
	packets := network.PacketGenerator(inputConn, blockGenerationChannelCapacity)
	filteredPackets := books.SignatureVerification(packets)
	transactions := books.TransactionGenerator(bm, filteredPackets)
	blocks := block.Generator(transactions, bm.ValidVDFValue(), startingBlocksTotal)
	batch := block.Batcher(blocks)
	blobs := make(chan *network.Blobs, cap(batch))
	frame := network.NewFrame()
	go replication.Broadcaster(sync, frame, outputConn, blobs)
	go RequestBlobs(sync, frame, reconstructionConn, outputConn)
	// temp := 0
	// go func() {
	for b := range batch {
		for _, block := range b {
			go bm.UpdateLastBlock(&block)
		}
		go block.BatchSaver(b, db)
		b1 := block.BlocksToBlobs(b)
		// for _, b2 := range b1.Bs {
		// 	fmt.Printf("Generated blob %v... %v \n", temp, block.ByteToInt32(b2.Data[network.DataOffset+16+32:], 0))
		// 	temp++
		// }
		blobs <- b1
	}
	// }()
}

// Messaging is run on the producer or signer node and is responsible for Message processing
func Messaging(bm *books.Accounts, inputConn net.PacketConn, outputConn net.PacketConn) {
	packets := network.PacketGenerator(inputConn, messageChannelCapacity)
	responses := messaging.ResponseGenerator(packets, bm)
	messaging.ResponseSender(outputConn, responses)
}

// Synchronization is run on every node and is responsible for vital data replication
func Synchronization(sync *replication.Sync, inputConn net.PacketConn, outputConn net.PacketConn) {
	// fmt.Printf("Synchronization from Listening on %v, sending %v\n", inputConn.LocalAddr(), outputConn.LocalAddr())
	netBlobs := network.BlobGenerator(inputConn, synchronizationChannelCapacity)
	listenBlobs := replication.SyncListener(sync, netBlobs)
	mixedBlobs := make(chan *network.Blobs)
	network.BlobSender(outputConn, mixedBlobs)
	syncBlobs := replication.SyncGenerator(sync, synchronizationTimeoutDuration)
	go func() {
		for syncBlob := range syncBlobs {
			mixedBlobs <- &network.Blobs{Bs: []network.Blob{*syncBlob}}
		}
	}()
	for listenBlob := range listenBlobs {
		mixedBlobs <- listenBlob

	}
}

// BlockSigner is run on every node and is responsible for block signer
func BlockSigner(bm *books.Accounts, sync *replication.Sync, replicationConn net.PacketConn, reconstructionConn net.PacketConn, outputConn net.PacketConn, db api.DataBase) {
	blobsReceiver := network.BlobGenerator(replicationConn, signerChannelCapacity)
	frame := network.NewFrame()
	go RequestBlobs(sync, frame, reconstructionConn, outputConn)
	replicationBlobs := make(chan *network.Blobs, cap(blobsReceiver))
	transportBlobs := make(chan *network.Blobs, cap(blobsReceiver))
	go func() {
		for blobs := range blobsReceiver {
			log.Debug(fmt.Sprintf("Got blobs on replication socket %v", replicationConn.LocalAddr().String()))
			fmt.Printf("Got %v blobs on replicate socket %v\n", len(blobs.Bs), replicationConn.LocalAddr().String())
			replicationBlobs <- blobs
			transportBlobs <- blobs
			if db != nil {
				blocks := block.BlobsToBlocks(blobs)
				for _, bl := range blocks {
					fmt.Printf("Saving block %v\n", bl.Number)
					db.SaveBlock(bl)
				}
			}
		}
	}()
	reconBlobs := reconstruction.Reconstruct(frame, replicationBlobs, sync, outputConn)

	replication.New(bm, reconBlobs)
	replication.Transporter(sync, transportBlobs, outputConn)

}

func RequestBlobs(sync *replication.Sync, frame *network.Frame, reconstructionConn net.PacketConn, outputConn net.PacketConn) {
	fmt.Printf("@@@@@@@@@ Listening on %v @@@@@@@@@@\n", reconstructionConn.LocalAddr().String())
	requestBlobs := network.BlobGenerator(reconstructionConn, signerChannelCapacity)
	requests := reconstruction.Listener(sync, frame, requestBlobs)
	network.BlobSender(outputConn, requests)
}
