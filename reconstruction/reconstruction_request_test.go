package reconstruction

import (
	"math/rand"
	"testing"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/network"
	"github.com/Ansiblock/Ansiblock/replication"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n uint64) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
func TestSerialize(t *testing.T) {
	randomIndex := uint64(1)
	for i := 0; i < 1000; i++ {
		randomIndex *= 71
	}
	req := new(Request)
	req.Index = randomIndex
	self := block.NewKeyPair().Public
	str1 := RandStringBytes(randomIndex % 1000)
	str2 := RandStringBytes(randomIndex % 999)
	req.From = replication.NewNodeData(self, str1, str2, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	serial := req.Serialize()
	res := new(Request)
	res.Deserialize(serial)
	if !res.Equals(req) {
		t.Errorf("Request wrong Serialization")
	}
}
