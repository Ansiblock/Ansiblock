package replication

import (
	"bytes"
	"encoding/gob"

	"go.uber.org/zap"

	"github.com/Ansiblock/Ansiblock/log"
	"github.com/Ansiblock/Ansiblock/network"
	"golang.org/x/crypto/ed25519"
)

const (
	getUpdatesType byte = iota
	updatesType
)

// GetUpdates is struct for requesting updates for Sync
type GetUpdates struct {
	LastUpdateIndex uint64
	MyInfo          *NodeData
}

// Updates is a struct for receiving updates for Sync
type Updates struct {
	From            ed25519.PublicKey
	LastUpdateIndex uint64
	Updates         []*NodeData
}

// Serialize method is for turning GetUpdates into blobs
func (req *GetUpdates) Serialize() *network.Blob {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(req)
	if err != nil {
		log.Error("encode error:", zap.Error(err))
	}
	blob := new(network.Blob)
	blob.Size = uint32(b.Len() + 1)
	copy(blob.Data[1:blob.Size], b.Bytes())
	blob.Data[0] = getUpdatesType
	return blob
}

// Deserialize method is for turning Blobs into GetUpdates
func (req *GetUpdates) Deserialize(blob *network.Blob) {
	b := bytes.NewBuffer(blob.Data[1:blob.Size])
	dec := gob.NewDecoder(b)
	err := dec.Decode(req)
	if err != nil {
		log.Error("decode error:", zap.Error(err))
	}
}

// Serialize method is for turning Updates into blobs
func (rec *Updates) Serialize() *network.Blob {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(rec)
	if err != nil {
		log.Error("encode error:", zap.Error(err))
	}
	blob := new(network.Blob)
	blob.Size = uint32(b.Len() + 1)
	copy(blob.Data[1:blob.Size], b.Bytes())
	blob.Data[0] = updatesType
	return blob
}

// Deserialize method is for turning Blobs into Updates
func (rec *Updates) Deserialize(blob *network.Blob) {
	b := bytes.NewBuffer(blob.Data[1:blob.Size])
	dec := gob.NewDecoder(b)
	err := dec.Decode(rec)
	if err != nil {
		log.Error("decode error:", zap.Error(err))
	}
}
