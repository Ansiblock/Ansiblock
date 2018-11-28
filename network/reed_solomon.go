package network

import (
	"errors"
	"fmt"

	"github.com/Ansiblock/Ansiblock/log"
	"github.com/klauspost/reedsolomon"
	"go.uber.org/zap"
)

const (
	NumCoded   = 20
	MaxMissing = 4
	NumData    = NumCoded - MaxMissing
)

// EncodeRSBlock generates Reed-Solomon codes and writes to coding array
// coding and data elements must all be the same size
func EncodeRSBlock(coding [][]byte, data [][]byte) error {
	shards := append(data, coding...)
	encoder, err := reedsolomon.New(len(data), len(coding))
	if err != nil {
		log.Error("Reed-Solomon initialization Error: ", zap.String("", err.Error()))
		return err
	}

	err = encoder.Encode(shards)
	if err != nil {
		log.Error("Reed-Solomon encoding Error: ", zap.String("", err.Error()))
	}
	// if encoding was successfull error will be nil
	return err
}

// DecodeRSBlock receives data and coding blocks. Recovers data from coding blocks
// Damaged data or coding block should be set to zero-length or nil
// Data and Coding must contain slices or else data will be copied and recovered at another memory address
// After recovery data is verified. The unsuccessful recovery causes error.
func DecodeRSBlock(coding [][]byte, data [][]byte) error {
	encoder, err := reedsolomon.New(len(data), len(coding))
	if err != nil {
		log.Error("Reed-Solomon initialization Error: ", zap.String("", err.Error()))
		return err
	}
	shards := append(data, coding...)
	err = encoder.Reconstruct(shards)
	if err != nil {
		log.Error("Reed-Solomon decoding Error: ", zap.String("", err.Error()))
		return err
	}

	// verify if recovery was done right, only False value has non nil error;
	// when verify returns True error is always nil
	res, err := encoder.Verify(shards)
	if !res {
		if err != nil {
			log.Error("Reed-Solomon verify Error: ", zap.String("", err.Error()))
			return err
		}
		log.Error("Reed-Solomon verify False")
		return errors.New("Reed-Solomon verify False")
	}
	// if Verify returned true, err will be nil
	return err
}

func AddCodingBlobs2(blobs *Blobs, start int) {
	added := 0
	blobsLen := len(blobs.Bs)
	addedBlobs := (((start % NumCoded) + blobsLen) / NumData) * MaxMissing
	log.Info(fmt.Sprintf("add codding blobs start %d, blobsLen %d, addedBlobs %d ",
		start, blobsLen, addedBlobs))
	for i := start; i < start+blobsLen+addedBlobs; i++ {
		if i != 0 && ((i+MaxMissing)%NumCoded) == 0 {
			index := i - start
			log.Info(fmt.Sprintf("putting coding blobs at i %d, index %d, blobsLen %d ",
				i, index, blobsLen))
			var missingBlobSlice []Blob
			for j := 0; j < MaxMissing; j++ {
				codingBlob := new(Blob)
				codingBlob.Size = BlobDataSize
				missingBlobSlice = append(missingBlobSlice, *codingBlob)
			}
			blobs.Bs = append(blobs.Bs[:index], append(missingBlobSlice, blobs.Bs[index:]...)...)
			added += MaxMissing
		}
	}
	log.Info(fmt.Sprintf("add coding start %d, blobsLen %d, added %d",
		start, blobsLen, added))
}

// AddCodingBlobs adds dummy blobs for parity codes
func AddCodingBlobs(blobs []*Blob, start int) []*Blob {
	added := 0
	blobsLen := len(blobs)
	// addedBlobs := (blobsLen / NumData) * MaxMissing
	addedBlobs := (((start % NumCoded) + blobsLen) / NumData) * MaxMissing
	for i := start; i < start+blobsLen+addedBlobs; i++ {
		if i != 0 && ((i+MaxMissing)%NumCoded) == 0 {
			index := i - start
			log.Info("putting coding blobs at ", zap.Int("", index))
			// missingBlobSlice := make([]Blob, MaxMissing)
			var missingBlobSlice []*Blob
			for j := 0; j < MaxMissing; j++ {
				codingBlob := new(Blob)
				codingBlob.Size = BlobDataSize
				missingBlobSlice = append(missingBlobSlice, codingBlob)
			}
			blobs = append(blobs[:index], append(missingBlobSlice, blobs[index:]...)...)
			added += MaxMissing
		}
	}
	log.Info(fmt.Sprintf("add coding start %d, blobsLen %d, added %d",
		start, blobsLen, added))
	return blobs
}

// EncodeRS generates coding blocks from frame
func EncodeRS(frame []*Blob, start, numBlobs uint64) error {
	log.Info(fmt.Sprintf("EncodeRS start %d, numBlobs %d", start, numBlobs))

	blobStart := start - (start % NumCoded)
	for i := start; i < start+numBlobs; i++ {
		if (i % NumCoded) == (NumCoded - 1) {
			var data, coding [][]byte

			blobEnd := blobStart + NumData
			// Find out maximum data length from blobs. Every blob will be padded to maximum size
			maxDataSize := getMaxDataSize(frame, blobStart, blobEnd)

			// add data blobs to array for encoding, skip meta information in data to keep it immutable
			for j := blobStart; j < blobEnd; j++ {
				data = append(data, frame[j%FrameSize].Data[DataOffset:maxDataSize])
			}

			codingStart := blobStart + NumData
			codingEnd := blobStart + NumCoded
			for j := codingStart; j < codingEnd; j++ {
				coding = append(coding, frame[j%FrameSize].Data[DataOffset:maxDataSize])
				// set coding blob size
				frame[j%FrameSize].Size = maxDataSize
			}

			err := EncodeRSBlock(coding, data)
			if err != nil {
				return err
			}
			log.Info(fmt.Sprintf("EncodeRS successful for blobs from: %d to: %d", blobStart, codingEnd))
			blobStart += NumCoded
		}
	}
	return nil
}

// func nullify(frame []*Blob, start, end uint) {
// 	for i := start; i <= end; i++ {
// 		if frame[i%FrameSize].Index() != uint64(i) {
// 			for j := 0; j < len(frame[i%FrameSize].Data); j++ {
// 				frame[i%FrameSize].Data[j] = 0
// 			}
// 		}
// 	}
// }

func getMaxDataSize(frame []*Blob, blobStart, blobEnd uint64) uint32 {
	maxDataSize := uint32(0)
	for i := blobStart; i < blobEnd; i++ {
		index := i % FrameSize
		if frame[index].Size > maxDataSize {
			maxDataSize = frame[index].Size
		}
	}
	return maxDataSize
}

func countMissing(frame []*Blob, blockStart uint64) (uint64, uint64) {
	dataMissing := uint64(0)
	codedMissing := uint64(0)
	codingStart := blockStart + NumData
	codingEnd := blockStart + NumCoded
	for i := blockStart; i < codingEnd; i++ {
		index := int(i) % FrameSize
		// blob is assumed demaged if it's Size attribute is zero
		if frame[index].Size == 0 || frame[index].Index() != uint64(i) {
			if i >= codingStart {
				codedMissing++
			} else {
				dataMissing++
			}
		}
	}
	return dataMissing, codedMissing
}

func prepairForDecoding(frame []*Blob, data [][]byte, maxDataSize uint32, start, end uint64) {
	di := 0
	for i := start; i < end; i++ {
		index := int(i) % FrameSize
		// if size is zero, so slice is zero-length and will be recovered
		if frame[index].Size == 0 || frame[index].Index() != uint64(i) {
			data[di] = frame[index].Data[DataOffset:][:0]
		} else {
			data[di] = frame[index].Data[DataOffset:maxDataSize]
		}
		di++
	}
}

func recoverIndex(frame []*Blob, start uint64) {
	end := start + NumCoded
	for i := start; i < end; i++ {
		if frame[i%FrameSize].Index() != uint64(i) {
			log.Info(fmt.Sprintf("Recovered index %v\n", i))
			frame[i%FrameSize].SetIndex(uint64(i))
		}
	}
}

// DecodeRS receives frame and recovers damaged data
// for efficiency it recovers only data,
// if coded blobs are demaged they will not be recovered
func DecodeRS(frame []*Blob, start, end uint64) error {
	if end <= start {
		log.Debug("wrong: ", zap.Int("start", int(start)), zap.Int("end", int(end)))
		return errors.New("Wrong indexes, end <= start")
	}
	blockStart := start - (start % NumCoded)
	if end-blockStart < NumCoded {
		// return if not enough blobs
		return errors.New("not enough number of blobs")
	}
	// trim end if needed, to decode only one block
	if blockStart+NumCoded < end {
		end = blockStart + NumCoded
	}
	// nullify(frame, start, end)
	codingStart := blockStart + NumData
	codingEnd := blockStart + NumCoded
	log.Info(fmt.Sprintf("DecodeRS: block_start: %v coding_start: %v coding_end: %v",
		blockStart, codingStart, codingEnd))
	// find how many blobs/codes are demaged in total
	dataMissing, codedMissing := countMissing(frame, blockStart)
	if (dataMissing + codedMissing) > MaxMissing {
		return errors.New("can not decode, too many blobs are missing")
	}
	log.Debug(fmt.Sprintf("DecodeRS: dataMissing: %d codedMissing: %d", dataMissing, codedMissing))
	// Find out maximum data length from blobs. Every blob will be padded to maximum size
	maxDataSize := getMaxDataSize(frame, blockStart, codingStart)
	// allocate NumCoded slots, preallocated memory will accelerate DecodeRSBlock
	var data = make([][]byte, NumData, NumCoded)
	var coding = make([][]byte, MaxMissing)
	prepairForDecoding(frame, data, maxDataSize, blockStart, codingStart)
	prepairForDecoding(frame, coding, maxDataSize, codingStart, codingEnd)
	err := DecodeRSBlock(coding, data)
	if err == nil {
		log.Info(fmt.Sprintf("DecodeRS successful for blobs from: %d to: %d", blockStart, codingEnd))
		// recover demaged indexes
		recoverIndex(frame, blockStart)
	}
	return err
}
