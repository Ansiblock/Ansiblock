package network

import (
	"net"
)

// BlobGenerator thread is responsible for reading blobs from socket
// and sending to the output channel
func BlobGenerator(reader net.PacketConn, capacity int) <-chan *Blobs {
	// packetCount := 0
	out := make(chan *Blobs, capacity)
	go func(reader net.PacketConn, p chan<- *Blobs) {
		for {
			blobs := NewBlobs()
			n := blobs.ReadFrom(reader)
			// fmt.Printf("BlobGenerator: %v blobs read from %v\n", n, reader.LocalAddr())
			// for i := 0; i < n; i++ {
			// 	fmt.Printf("%v-th blob is %v\n", i, blobs.Bs[i].Index())
			// }
			if n > 0 {
				// log.Info("BlobGenerator: ", zap.Int("Total Blobs", n))
				// fmt.Println(n)
				out <- blobs
			}
		}
	}(reader, out)
	return out
}

// BlobSender thread is responsible for getting blobs from input
// channel and sending to the socket
func BlobSender(writer net.PacketConn, input <-chan *Blobs) {
	go func() {
		for {
			blobs := <-input
			blobs.WriteTo(writer)
			// fmt.Printf("BlobSender From %v -> %v \n", writer.LocalAddr(), blobs.Bs[0].Addr)
		}
	}()
}
