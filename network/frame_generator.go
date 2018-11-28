package network

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/Ansiblock/Ansiblock/log"
)

// FrameSize represents blob count in frame
const FrameSize = 2 * 1024

// Frame struct saves blobs. We use it to encode and decode blobs
type Frame struct {
	Blobs []*Blob
}

// NewFrame returns brand new frame with empty blobs
func NewFrame() *Frame {
	frame := new(Frame)
	frame.Blobs = make([]*Blob, FrameSize)
	for i := range frame.Blobs {
		frame.Blobs[i] = new(Blob)
	}
	return frame
}

// missingBlobIndexes goes though the frame and returns missing blobs
func (f *Frame) missingBlobIndexes(start, end uint64) []uint64 {
	res := make([]uint64, 0)
	if start+NumCoded < end {
		end = start + NumCoded
	}
	for i := start; i <= end; i++ {
		if f.Blobs[i%FrameSize] == nil || f.Blobs[i%FrameSize].Index() != i {
			res = append(res, i)
		}
	}
	return res
}

func fillFrame(frame *Frame, start, end uint64, blobs *Blobs) uint64 {
	for i := range blobs.Bs {
		index := blobs.Bs[i].Index()
		if end < index {
			end = index
		}
		if index < start {
			log.Debug("Received blob with old index", zap.Uint64("index", index))
			continue
		}
		frame.Blobs[index%FrameSize] = &blobs.Bs[i]
	}
	return end
}

func outputFrame(frame *Frame, start, end uint64, header string) {
	fmt.Printf("======%v=========\n", header)
	for i := start; i <= end; i++ {
		fmt.Printf("%v. index %v\n", i, frame.Blobs[i%FrameSize].Index())
	}
	fmt.Printf("======%v=========\n", header)
}

func recoverBlobs(frame *Frame, start, end uint64, doneBlobs chan *Blobs) uint64 {
	res := make([]Blob, NumData)
	recoveredBlobs := &Blobs{Bs: res}
	recoveredBlobsIndex := 0
	for {
		index := start % FrameSize
		if frame.Blobs[index] == nil || frame.Blobs[index].Index() != uint64(start) {
			break
		}
		start++
		recoveredBlobs.Bs[recoveredBlobsIndex] = *frame.Blobs[index] //TODO refactor copying
		recoveredBlobsIndex++
		if recoveredBlobsIndex == NumData {
			start += MaxMissing
			break
		}
	}
	if recoveredBlobsIndex == NumData {
		doneBlobs <- recoveredBlobs
	} else {
		log.Debug("Not enough data to recover Blobs")
	}
	return start
}

func processBlobs(frame *Frame, start, end uint64, blobs *Blobs, doneBlobs chan *Blobs, missingIndexes chan []uint64) (uint64, uint64) {
	end = fillFrame(frame, start, end, blobs)
	outputFrame(frame, start, end, "before")
	for start+NumCoded < end {
		err := DecodeRS(frame.Blobs, start, end)
		if err == nil {
			start = recoverBlobs(frame, start, end, doneBlobs)
		} else {
			log.Error("reed-solomon error!", zap.Error(err))
			// TODO exponential
			indexes := frame.missingBlobIndexes(start, end)
			fmt.Printf("########## Missing index = %v\n", indexes)
			missingIndexes <- indexes
			break
		}
	}
	outputFrame(frame, start, end, "after")

	return start, end
}

// FrameGenerator thread is responsible for generating recovered chunks of blobs
// form the Frame. Other nodes send Encoded blobs, so we have to make sure blobs
// can be decoded properly.
func FrameGenerator(frame *Frame, input <-chan *Blobs) (chan *Blobs, chan []uint64) {
	start := uint64(0) //frame start
	end := uint64(0)   // frame end
	doneBlobs := make(chan *Blobs, 10)
	missingIndexes := make(chan []uint64, 10)
	go func() {
		for {
			log.Debug("FrameGenerator", zap.Uint64("Start", start), zap.Uint64("End", end))
			blobs, ok := <-input
			if !ok {
				log.Error("Frame request failed!")
				return
			}
			start, end = processBlobs(frame, start, end, blobs, doneBlobs, missingIndexes)
		}
	}()
	return doneBlobs, missingIndexes
}
