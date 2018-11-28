package network

import (
	"net"
	"time"

	"golang.org/x/crypto/ed25519"

	"go.uber.org/zap"

	"github.com/Ansiblock/Ansiblock/log"
)

const (
	// BlobDataSize represents size of data in blob
	BlobDataSize = 1024 * 64
	// BlobRealDataSize represents real data in blob
	BlobRealDataSize = BlobDataSize - ed25519.PublicKeySize - 8 - 4
	// NumBlobs represents number of packets in Packets struct
	NumBlobs = NumPackets * packetDataSize / BlobDataSize

	fromOffset   = 8
	flagsOffset  = fromOffset + ed25519.PublicKeySize
	flagIsCoding = 1
	// DataOffset identifies where blob data starts in blob.Data[] array
	DataOffset = flagsOffset + 4
)

// Blob stores block data
// Note: for now blob is just the copy of the Packet type with the exception of the byte array size
type Blob struct {
	Data [BlobDataSize]byte
	Addr net.Addr //UDP Address
	Size uint32
}

// Blobs stores slice of Blobs
type Blobs struct {
	Bs []Blob
}

// NewNumBlobs returns Packets object with num packets
func NewNumBlobs(num uint64) *Blobs {
	res := make([]Blob, num)
	return &Blobs{Bs: res}
}

// NewBlobs returns Packets object
func NewBlobs() *Blobs {
	res := make([]Blob, NumBlobs)
	return &Blobs{Bs: res}
}

// ReadFrom reads from the UDP socket and return number of bytes read.
func (bs *Blobs) ReadFrom(reader net.PacketConn) int {
	for i := range bs.Bs {
		n, addr, err := reader.ReadFrom(bs.Bs[i].Data[:])
		// fmt.Printf("ReadFrom: %v bytes read from %v on %v\n", n, addr, reader.LocalAddr())

		bs.Bs[i].Size = uint32(n)
		bs.Bs[i].Addr = addr
		if err != nil || n == 0 {
			bs.Bs = bs.Bs[:i]
			reader.SetReadDeadline(time.Now().Add(1 * time.Hour))
			return i
		}
		if i == 0 {
			reader.SetReadDeadline(time.Now().Add(readTimeout))
		}
	}
	reader.SetReadDeadline(time.Now().Add(1 * time.Hour))
	return len(bs.Bs)
}

// WriteTo writes all blobs into the socket
func (bs *Blobs) WriteTo(writer net.PacketConn) int {
	res := 0
	for i := range bs.Bs {
		n, err := writer.WriteTo(bs.Bs[i].Data[:bs.Bs[i].Size], bs.Bs[i].Addr)
		if err != nil {
			log.Error("Packets WriteTo error", zap.Error(err))
		}
		res += n
	}
	return res
}

// Index method gets index from Data array
func (b *Blob) Index() uint64 {
	return uint64(uint64(b.Data[7]) | uint64(b.Data[6])<<8 | uint64(b.Data[5])<<16 | uint64(b.Data[4])<<24 |
		uint64(b.Data[3])<<32 | uint64(b.Data[2])<<40 | uint64(b.Data[1])<<48 | uint64(b.Data[0])<<56)
}

// SetIndex method sets blob index in front of Data array
func (b *Blob) SetIndex(index uint64) {
	b.Data[0] = byte(index >> 56)
	b.Data[1] = byte(index >> 48)
	b.Data[2] = byte(index >> 40)
	b.Data[3] = byte(index >> 32)
	b.Data[4] = byte(index >> 24)
	b.Data[5] = byte(index >> 16)
	b.Data[6] = byte(index >> 8)
	b.Data[7] = byte(index)
}

// From method gets PublicKey from blob
func (b *Blob) From() ed25519.PublicKey {
	res := make([]byte, ed25519.PublicKeySize)
	for i := 0; i < ed25519.PublicKeySize; i++ {
		res[i] = b.Data[fromOffset+i]
	}
	return res
}

// SetFrom method sets PublicKey to blob
func (b *Blob) SetFrom(from ed25519.PublicKey) {
	for i := 0; i < ed25519.PublicKeySize; i++ {
		b.Data[fromOffset+i] = from[i]
	}
}

// Flags method gets flags from blob
func (b *Blob) Flags() uint32 {
	return uint32(uint32(b.Data[flagsOffset+3]) | uint32(b.Data[flagsOffset+2])<<8 |
		uint32(b.Data[flagsOffset+1])<<16 | uint32(b.Data[flagsOffset])<<24)
}

// SetFlags method sets flags to blob
func (b *Blob) SetFlags(flags uint32) {
	b.Data[flagsOffset] = byte(flags >> 24)
	b.Data[flagsOffset+1] = byte(flags >> 16)
	b.Data[flagsOffset+2] = byte(flags >> 8)
	b.Data[flagsOffset+3] = byte(flags)
}

// IsCoding returns true if blob's coding flag is set
func (b *Blob) IsCoding() bool {
	return b.Flags()&flagIsCoding != 0
}

// SetCoding sets blob's coding flag
func (b *Blob) SetCoding() {
	flags := b.Flags()
	b.SetFlags(flags | flagIsCoding)
}

// IndexBlobs method will index blobs and set From
func (bs *Blobs) IndexBlobs(from ed25519.PublicKey, startIndex uint64) {
	log.Info("Indexing blobs")
	for i := range bs.Bs {
		bs.Bs[i].SetFrom(from)
		bs.Bs[i].SetIndex(startIndex + uint64(i))
	}
}
