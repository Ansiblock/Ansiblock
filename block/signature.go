package block

import (
	"log"

	"golang.org/x/crypto/ed25519"
)

// KeyPair struct represents ed25519 PublicKey and PrivateKey
type KeyPair struct {
	Public  ed25519.PublicKey
	Private ed25519.PrivateKey
}

// NewKeyPair generates Public and Private Keypair
func NewKeyPair() KeyPair {
	pub, pr, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Panic("Can not generate KeyPair")
	}
	return KeyPair{Public: pub, Private: pr}
}

// KeyPairs generates n Keypair
func KeyPairs(n int) []KeyPair {
	keypairs := make([]KeyPair, n)
	counter := make(chan bool)
	for i := range keypairs {
		go func(i int) {
			keypairs[i] = NewKeyPair()
			counter <- true
		}(i)
	}
	for range keypairs {
		<-counter
	}
	return keypairs
}
