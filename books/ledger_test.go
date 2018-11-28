package books

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/Ansiblock/Ansiblock/block"
)

func TestNewLedger(t *testing.T) {
	l := newLedger()
	if l.index != 0 || l.full || l.mutex == nil ||
		l.signatures == nil || l.values == nil {
		t.Error("Wrong Ledger")
	}
}

func TestAddValidVDFValue(t *testing.T) {
	l := newLedger()
	for i := 0; i < 10000; i++ {
		vdf := []byte(strconv.Itoa(i))
		l.AddValidVDFValue(vdf)
	}
	last := l.validVDFValue()
	arr := []byte(strconv.Itoa(9999))
	if !bytes.Equal(last, arr) {
		t.Errorf("Wrong last vdf value: was %v should be %v", last, arr)
	}
	for i := 0; i < 10000-maxSize; i++ {
		val := string([]byte(strconv.Itoa(i)))
		if _, ok := l.signatures[val]; ok {
			t.Error("Old vdf value in ledger")
		}
	}
}

func TestAddSignature(t *testing.T) {
	l := newLedger()
	for i := 0; i < 1000; i++ {
		vdf := []byte(strconv.Itoa(i))
		l.AddValidVDFValue(vdf)
	}
	for i := 0; i < 1000; i++ {
		sig := []byte(strconv.Itoa(i))
		vdf := []byte(strconv.Itoa(i))
		err := l.addSignature(sig, vdf)
		if err != nil {
			t.Error("Can't add signature in ledger")
		}
	}
	for i := 0; i < 100; i++ {
		sig := []byte(strconv.Itoa(i))
		vdf := []byte(strconv.Itoa(i))
		l.removeSignature(sig, vdf)
	}
	for i := 0; i < 100; i++ {
		sig := []byte(strconv.Itoa(i))
		vdf := []byte(strconv.Itoa(i))
		if signs, ok := l.signatures[string(vdf)]; !ok {
			t.Error("VDF value not in ledger")
		} else {
			if _, ok := signs[string(sig)]; ok {
				t.Error("Wrong signature in ledger")
			}

		}
	}

	for i := 100; i < 1000; i++ {
		sig := []byte(strconv.Itoa(i))
		vdf := []byte(strconv.Itoa(i))
		if signs, ok := l.signatures[string(vdf)]; !ok {
			t.Error("VDF value not in ledger")
		} else {
			if _, ok := signs[string(sig)]; !ok {
				t.Error("Signature not in ledger")
			}

		}
	}
}

func TestCloneEquals(t *testing.T) {
	l1 := newLedger()
	for i := 0; i < 10000; i++ {
		vdf := []byte(strconv.Itoa(i))
		l1.AddValidVDFValue(vdf)
	}
	l2 := l1.Clone()
	if !l1.Equals(l2) {
		t.Errorf("Test Clone or Equals failed")
	}
}
func TestUpdateLastBlock(t *testing.T) {
	l1 := newLedger()
	bl := block.NewEmpty([]byte{1, 2, 3}, 1, 1)
	l1.UpdateLastBlock(&bl)
	if l1.lastBlock.Number != 2 {
		t.Errorf("UpdateLastBlock failed")
	}
	if l1.LastBlock().Number != 2 {
		t.Errorf("UpdateLastBlock failed")
	}
}
