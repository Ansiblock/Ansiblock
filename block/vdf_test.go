package block

import (
	"fmt"
	"testing"
	// . "github.com/Ansiblock/Ansiblock/block"
)

func TestVDF(t *testing.T) {
	input := "hello"
	output := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	s := fmt.Sprintf("%x", VDF([]byte(input)))
	if s != output {
		t.Fatalf("VDF function: VDF(%s) = %s want %s", input, s, output)
	}
}

func TestExtendedVDF(t *testing.T) {
	input := "hello"
	vdfValue := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	output := "ae3967e4aa96162198fb8b1f706cc8565388a7fe0a09d9ef9bae38fd6e77d34d"
	// var arr [32]byte
	// copy(arr[:], []byte(vdfValue))
	s := fmt.Sprintf("%x", ExtendedVDF([]byte(input), []byte(vdfValue)))
	if s != output {
		t.Fatalf("ExtendedVDF function: ExtendedVDF(%s, %s) = %s want %s", input, vdfValue, s, output)
	}
}

func BenchmarkVDF(b *testing.B) {
	// run the VDF chain of length b.N
	input := []byte("start of something beautiful")
	// var data [32]byte
	// copy(data[:], []byte(input))
	for n := 0; n < b.N; n++ {
		input = VDF(input)
	}
}

func BenchmarkExtendedVDF(b *testing.B) {
	// run the edtendedVDF chain of length b.N
	data := []byte("start of something beautiful")
	hash := data
	// var hash [32]byte
	// copy(hash[:], data)
	for n := 0; n < b.N; n++ {
		hash = ExtendedVDF(data, hash)
	}
}
