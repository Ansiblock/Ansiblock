package block

import "testing"

func TestNewKeyPair(t *testing.T) {
	NewKeyPair()
}

func TestKeyPairs(t *testing.T) {
	kps := KeyPairs(5)
	if len(kps) != 5 {
		t.Errorf("KeyPairs: wanted %v got %v", 5, len(kps))
	}
}
