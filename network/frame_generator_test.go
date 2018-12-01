package network

import (
	"fmt"
	"testing"
)

func TestNewFrame(t *testing.T) {
	frame := NewFrame()
	if len(frame.Blobs) != FrameSize {
		t.Errorf("Wring number of blobs in frame %v\n", len(frame.Blobs))
	}
}

func TestMissingBlobIndexes(t *testing.T) {
	frame := NewFrame()
	indexes := frame.missingBlobIndexes(10000, 10010)
	for i := 0; i < len(indexes); i++ {
		if indexes[i] != uint64(i)+10000 {
			t.Errorf("MissingBlobIndexes failed %v!=%v\n", indexes[i], i-1)
		}
	}
	frame.Blobs[10].SetIndex(10)
	frame.Blobs[20].SetIndex(20)
	indexes = frame.missingBlobIndexes(10, 10010)
	if len(indexes) != NumCoded-1 {
		t.Errorf("MissingBlobIndexes failed %v\n", indexes)
	}
	for i := 0; i < 9; i++ {
		if indexes[i] != uint64(i)+11 {
			t.Errorf("MissingBlobIndexes failed %v!=%v\n", indexes[i], i+11)
		}
	}
	for i := 0; i < 10; i++ {
		if indexes[i+9] != uint64(i)+21 {
			t.Errorf("MissingBlobIndexes failed %v!=%v\n", indexes[i+9], i+21)
		}
	}
}

func TestFillFrame(t *testing.T) {
	frame := NewFrame()
	blobs := NewNumBlobs(20)
	from := make([]byte, 32)
	blobs.IndexBlobs(from, 100)
	end := fillFrame(frame, 0, 100, blobs)
	if end != 119 {
		t.Errorf("fillFrame failed %v!=%v\n", end, 119)
	}
	for i := uint64(100); i < 120; i++ {
		if frame.Blobs[i].Index() != i {
			t.Errorf("fillFrame failed index %v!=%v\n", frame.Blobs[i].Index(), i)
		}
	}
}

func TestFillFrame2(t *testing.T) {
	frame := NewFrame()
	blobs := NewNumBlobs(40)
	from := make([]byte, 32)
	blobs.IndexBlobs(from, 100)
	end := fillFrame(frame, 110, 1000, blobs)
	if end != 1000 {
		t.Errorf("fillFrame failed %v!=%v\n", end, 139)
	}
	for i := uint64(110); i < 140; i++ {
		if frame.Blobs[i].Index() != i {
			t.Errorf("fillFrame failed index %v!=%v\n", frame.Blobs[i].Index(), i)
		}
	}
}

func TestRecoverBlobs(t *testing.T) {
	frame := NewFrame()
	doneBlobs := make(chan *Blobs, 1)
	start := recoverBlobs(frame, 10, 100, doneBlobs)
	if start != 10 {
		t.Errorf("recoverBlobs failed %v!=%v\n", start, 10)
	}
}

func TestRecoverBlobs2(t *testing.T) {
	frame := NewFrame()
	blobs := NewNumBlobs(40)
	doneBlobs := make(chan *Blobs, 1)
	from := make([]byte, 32)
	blobs.IndexBlobs(from, 100)
	fillFrame(frame, 100, 1000, blobs)
	start := recoverBlobs(frame, 100, 1000, doneBlobs)
	if start != 120 {
		t.Errorf("recoverBlobs failed %v!=%v\n", start, 120)
	}
	bl := <-doneBlobs
	for i := range bl.Bs {
		if bl.Bs[i].Index() != 100+uint64(i) {
			t.Errorf("recoverBlobs failed done blobs index %v!=%v\n", bl.Bs[i].Index(), 100+i)
		}
	}
}

func TestRecoverBlobs3(t *testing.T) {
	frame := NewFrame()
	blobs := NewNumBlobs(100)
	doneBlobs := make(chan *Blobs, 1)
	from := make([]byte, 32)
	blobs.IndexBlobs(from, 100)
	fillFrame(frame, 0, 1000, blobs)
	start := recoverBlobs(frame, 120, 1000, doneBlobs)
	if start != 140 {
		t.Errorf("recoverBlobs failed %v!=%v\n", start, 140)
	}
	bl := <-doneBlobs
	for i := range bl.Bs {
		if bl.Bs[i].Index() != 120+uint64(i) {
			t.Errorf("recoverBlobs failed done blobs index %v!=%v\n", bl.Bs[i].Index(), 120+i)
		}
	}
}

func TestProcessBlobs(t *testing.T) {
	frame := NewFrame()
	blobs := NewNumBlobs(1000)
	doneBlobs := make(chan *Blobs, 1)
	from := make([]byte, 32)
	blobs.IndexBlobs(from, 100)
	missingIndexes := make(chan []uint64, 1)
	start, end := processBlobs(frame, 100, 1000, blobs, doneBlobs, missingIndexes)
	if start != 120 && end != 1099 {
		t.Errorf("processBlobs error %v.%v\n", start, end)
	}
	indexes := <-missingIndexes
	for i := range indexes {
		if indexes[i] != uint64(i)+120 {
			t.Errorf("wrong indexes: %v!=%v\n", indexes[i], i+120)
		}
	}
}

func TestFrameGenerator(t *testing.T) {
	frame := NewFrame()
	blobs := make(chan *Blobs, 1)
	_, missing := FrameGenerator(frame, blobs)
	bs := NewNumBlobs(20)
	from := make([]byte, 32)
	bs.IndexBlobs(from, 100)
	blobs <- bs
	close(blobs)
	fmt.Println(missing)
	indexes := <-missing

	for i := range indexes {
		if indexes[i] != uint64(i)+1 {
			t.Errorf("wrong indexes: %v!=%v\n", indexes[i], i+1)
		}
	}
}
