package books

import (
	"bytes"
	"errors"
	"reflect"
	"sync"

	"github.com/Ansiblock/Ansiblock/block"
	"golang.org/x/crypto/ed25519"
)

const maxSize = 1024 * 4 //TODO to be determined
var (
	errDuplicateSignatures = errors.New("duplicate signatures")
	errVDFValueNotFound    = errors.New("vdf value not found")
)

// Ledger struct is responsible to save last vdf values
// For each VDFValue we store a set of Signatures.
// This way we can determine which signature came with which VDFValue
type Ledger struct {
	// key: VDFValue, value: map with key - Signature and value bool
	signatures map[string]map[string]bool
	values     []block.VDFValue
	index      int
	mutex      *sync.RWMutex
	blockMutex *sync.Mutex
	full       bool
	lastBlock  *block.Block
}

// NewLedger creates ledger object
func newLedger() *Ledger {
	l := new(Ledger)
	l.values = make([]block.VDFValue, maxSize)
	first := []byte("hello")
	l.values[0] = first
	l.signatures = make(map[string]map[string]bool)
	l.signatures[string(first)] = make(map[string]bool)
	l.index = 0
	l.full = false
	l.mutex = &sync.RWMutex{}
	l.blockMutex = &sync.Mutex{}
	l.lastBlock = nil
	return l
}

// validVDFValue returns last VDF value in the ledger
func (l *Ledger) validVDFValue() block.VDFValue {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.values[l.index]
}

// removeSignature removes signature from ledger for a specific vdf value
func (l *Ledger) removeSignature(signature ed25519.PublicKey, vdfVal block.VDFValue) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if signs, ok := l.signatures[string(vdfVal)]; ok {
		if _, ok := signs[string(signature)]; ok {
			delete(signs, string(signature))
		}
	}
}

// addSignature adds signature for a specific vdf value
func (l *Ledger) addSignature(signature ed25519.PublicKey, vdfVal block.VDFValue) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	signs, ok := l.signatures[string(vdfVal)]
	if !ok {
		return errVDFValueNotFound
	}
	if _, ok := signs[string(signature)]; ok {
		return errDuplicateSignatures
	}
	signs[string(signature)] = true

	return nil
}

func (l *Ledger) contains(vdfVal block.VDFValue) bool {
	for _, val := range l.values {
		if bytes.Equal(vdfVal, val) {
			return true
		}
	}
	return false
}

// AddValidVDFValue registers new vdf value.
// typically used when producer is sending blocks to the network.
func (l *Ledger) AddValidVDFValue(vdfVal block.VDFValue) {
	// fmt.Printf("add valid vdf %v\n", vdfVal)
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.contains(vdfVal) {
		return
	}

	if l.full || l.index == maxSize-1 {
		l.full = true
		l.index++
		l.index = l.index % maxSize
		delete(l.signatures, string(l.values[l.index]))
	} else {
		l.index++
	}
	l.signatures[string(vdfVal)] = make(map[string]bool)
	l.values[l.index] = vdfVal
}

// UpdateLastBlock updates last block in ledger.
// last block means a block with the highest number(height).
func (l *Ledger) UpdateLastBlock(bl *block.Block) {
	l.blockMutex.Lock()
	defer l.blockMutex.Unlock()
	if l.lastBlock == nil || bl.Number > l.lastBlock.Number {
		l.lastBlock = bl
	}
}

// LastBlock returns last block registered in the ledger
func (l *Ledger) LastBlock() *block.Block {
	l.blockMutex.Lock()
	defer l.blockMutex.Unlock()
	return l.lastBlock
}

// Clone returns clone of Ledger object
func (l *Ledger) Clone() *Ledger {
	clone := new(Ledger)
	clone.values = make([]block.VDFValue, maxSize)
	for i := 0; i < cap(l.values); i++ {
		clone.values[i] = l.values[i]
	}
	clone.signatures = make(map[string]map[string]bool)
	for k, v := range l.signatures {
		clone.signatures[k] = make(map[string]bool)
		for sign, val := range v {
			clone.signatures[k][sign] = val
		}
	}

	clone.index = l.index
	clone.full = l.full
	clone.mutex = &sync.RWMutex{}
	clone.lastBlock = l.lastBlock
	clone.blockMutex = &sync.Mutex{}
	return clone
}

// Equals compares two Ledger objects, for testing
func (l *Ledger) Equals(l2 *Ledger) bool {
	return reflect.DeepEqual(l.values, l2.values) && reflect.DeepEqual(l.signatures, l2.signatures) &&
		l.index == l2.index && l.full == l2.full &&
		((l.lastBlock == nil && l2.lastBlock == nil) || (l.lastBlock.Number == l2.lastBlock.Number))
}
