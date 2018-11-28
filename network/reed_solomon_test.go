package network

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/Ansiblock/Ansiblock/log"
	"github.com/klauspost/reedsolomon"
	"go.uber.org/zap"
	"golang.org/x/crypto/ed25519"
)

func random(min, max int) int {
	rand.Seed(int64(time.Now().Nanosecond()))
	return rand.Intn(max-min) + min
}

func fillRandom(data []byte) {
	for i := 0; i < len(data); i++ {
		data[i] = byte(random(0, 255))
	}
}

func randomDataCoding(dataLen, numData, numCoding int) ([][]byte, [][]byte) {
	var data [][]byte
	var coding [][]byte
	for i := 0; i < numData; i++ {
		randomData := make([]byte, dataLen)
		fillRandom(randomData)
		data = append(data, randomData)
	}
	for i := 0; i < numCoding; i++ {
		coding = append(coding, make([]byte, dataLen))
	}
	return data, coding
}

func TestGenerateCodingBlocks(t *testing.T) {
	data, coding := randomDataCoding(10, 10, 3)
	EncodeRSBlock(coding, data)
	shards := append(data, coding...)
	encoder, err := reedsolomon.New(len(data), len(coding))
	if err != nil {
		t.Error("Reed-Solomon initialization Error: ", zap.String("", err.Error()))
	}

	v, err := encoder.Verify(shards)
	if err != nil && v {
		t.Error("result of Reed-Solomon verify ", zap.String(strconv.FormatBool(v), err.Error()))
	}
}

func TestGenerateCodingBlocksWithZeroData(t *testing.T) {
	data, coding := randomDataCoding(10, 0, 3)
	codingClone := append([][]byte{}, coding...)
	err := EncodeRSBlock(coding, data)
	if err != nil && !reflect.DeepEqual(coding, codingClone) {
		t.Error("EncodeRSBlock failed zero data test")
	}
}

func TestGenerateCodingForInitError(t *testing.T) {
	data, coding := randomDataCoding(10, 10, 0)
	err := EncodeRSBlock(coding, data)
	if err != reedsolomon.ErrInvShardNum {
		t.Error("EncodeRS fails reedsolmon initialization")
	}
}

func TestGenerateCodingForEncodingError(t *testing.T) {
	data, coding := randomDataCoding(10, 10, 3)
	data[0] = append(data[0], []byte{1, 2, 3}...)
	err := EncodeRSBlock(coding, data)
	if err != reedsolomon.ErrShardSize {
		t.Error("EncodeRS fails encoding error check")
	}
}

func TestDecodeBlocksForInitError(t *testing.T) {
	data, coding := randomDataCoding(10, 10, 0)
	err := DecodeRSBlock(coding, data)
	if err != reedsolomon.ErrInvShardNum {
		t.Error("EncodeRS fails reedsolmon initialization")
	}
}

func TestDecodeBlocks(t *testing.T) {
	dataArr, codingArr := randomDataCoding(5, 5, 3)
	EncodeRSBlock(codingArr, dataArr)
	fmt.Println(dataArr, codingArr)

	// damage some data
	dataArr[0] = []byte{0, 0, 0, 0, 0}
	codingArr[1] = []byte{0, 0, 0, 0, 0}
	codingArr[2] = []byte{0, 0, 0, 0, 0}
	fmt.Println(dataArr, codingArr)

	var data [][]byte
	var coding [][]byte
	data = append(data, dataArr[0][:0])
	data = append(data, dataArr[1][:])
	data = append(data, dataArr[2][:])
	data = append(data, dataArr[3][:])
	data = append(data, dataArr[4][:])
	coding = append(coding, codingArr[0][:])
	coding = append(coding, codingArr[1][:0])
	coding = append(coding, codingArr[2][:0])
	fmt.Println(data, coding)

	err := DecodeRSBlock(coding, data)
	if err != nil {
		t.Error("Decoding Blocks failed")
	}
	fmt.Println(dataArr, codingArr)

	shards := append(dataArr, codingArr...)
	encoder, err := reedsolomon.New(len(dataArr), len(codingArr))
	v, err := encoder.Verify(shards)
	if !v || err != nil {
		t.Error("result of Reed-Solomon verify ", zap.String(strconv.FormatBool(v), err.Error()))
	}
}

func TestAddCodingBlobs(t *testing.T) {
	var blobs []*Blob
	length := uint64(50)
	for i := 0; i < 32; i++ {
		blob := new(Blob)
		fillRandom(blob.Data[:length])
		blobs = append(blobs, blob)
	}
	printFrame(blobs, length)
	blobs = AddCodingBlobs(blobs, 10)
	printFrame(blobs, length)
	if len(blobs) != 40 {
		fmt.Println(len(blobs))
		t.Error("Could not add blobs")
	}
}

func indexBlobs(dataBlobs []*Blob, from ed25519.PublicKey, startIndex uint64) {
	log.Info("Indexing blobs")
	for i, b := range dataBlobs {
		b.SetFrom(from)
		b.SetIndex(startIndex + uint64(i))
	}
}

func generateFrame(dataLen, offset, numBlobs uint64) ([]*Blob, int) {
	frameSize := 20
	var dataBlobs []*Blob
	for i := 0; i < int(numBlobs); i++ {
		b := Blob{Size: uint32(dataLen)}
		fillRandom(b.Data[:dataLen])
		// fmt.Println(b.Data[:dataLen])
		dataBlobs = append(dataBlobs, &b)
	}
	dataBlobs = AddCodingBlobs(dataBlobs, int(offset))
	blobLen := len(dataBlobs)
	frame := make([]*Blob, frameSize)
	for i := 0; i < frameSize; i++ {
		frame[i] = &Blob{Size: uint32(dataLen)}
	}

	from := make([]byte, 32)
	indexBlobs(dataBlobs, from, offset)
	for _, b := range dataBlobs {
		idx := b.Index()
		frame[idx] = b
		// fmt.Println(frame[idx].Data[:dataLen])
	}
	return frame, blobLen
}

func printFrame(frame []*Blob, length uint64) {
	fmt.Println("printFrame:")
	for i := range frame {
		fmt.Println(frame[i].Data[:length])
	}
	fmt.Println("")
}

func TestFrameRecovery(t *testing.T) {
	dataLen := uint64(150)
	offset := uint64(0)
	numBlobs := uint64(16)
	frame, blobLen := generateFrame(dataLen, offset, numBlobs)
	fmt.Println("bloblen", blobLen)
	fmt.Println("win length", len(frame))
	printFrame(frame, dataLen)
	EncodeRS(frame, offset, uint64(blobLen))
	printFrame(frame, dataLen)

	// damage data
	for _, j := range []int{18} {
		frame[j].Size = 0
		for i := uint64(0); i < dataLen; i++ {
			frame[j].Data[i] = byte(0)
		}
	}
	printFrame(frame, dataLen)
	err := DecodeRS(frame, offset, offset+uint64(blobLen))
	printFrame(frame, dataLen)

	var shards [][]byte
	for i := 0; i < NumCoded; i++ {
		shards = append(shards, frame[i].Data[DataOffset:dataLen])
	}

	encoder, err := reedsolomon.New(NumData, MaxMissing)
	v, err := encoder.Verify(shards)
	if err != nil {
		t.Error("result of Reed-Solomon verify ", err.Error())
	}

	if !v {
		t.Error("result of Reed-Solomon verify FAILED")
	}
}
