package reconstruction

import (
	"bytes"
	"encoding/gob"

	"github.com/Ansiblock/Ansiblock/log"
	"github.com/Ansiblock/Ansiblock/network"
	"github.com/Ansiblock/Ansiblock/replication"
	"go.uber.org/zap"
)

// Request is a struct to request missing blobs from other nodes.
type Request struct {
	Index uint64
	From  *replication.NodeData
}

// Serialize serializes the request structure into the blob
func (r *Request) Serialize() *network.Blob {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(r)
	if err != nil {
		log.Error("encode error:", zap.Error(err))
	}
	blob := new(network.Blob)
	blob.Size = uint32(b.Len())
	copy(blob.Data[0:blob.Size], b.Bytes())
	return blob
}

// Deserialize deserializes request structure from blob
func (r *Request) Deserialize(blob *network.Blob) {
	b := bytes.NewBuffer(blob.Data[0:blob.Size])
	dec := gob.NewDecoder(b)
	err := dec.Decode(r)
	if err != nil {
		log.Error("decode error:", zap.Error(err))
	}
}

// Equals compares to Request struct
func (r *Request) Equals(r2 *Request) bool {
	if r.Index == r2.Index && r.From.Equals(r2.From) {
		return true
	}
	return false
}
