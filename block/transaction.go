package block

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strconv"

	"golang.org/x/crypto/ed25519"
)

// Transaction represents user transaction information
// For now we treat transaction as `token` transfer `from`
// user `to` another user. This transfer may have a `fee`.
type Transaction struct {
	From          ed25519.PublicKey
	To            ed25519.PublicKey
	Token         int64
	Fee           int64
	ValidVDFValue VDFValue
	Signature     []byte
}

// NewTransaction will create new Transaction object
func NewTransaction(from *KeyPair, to ed25519.PublicKey, token int64, fee int64, validVDFValue VDFValue) Transaction {
	tr := Transaction{From: from.Public, To: to, Token: token, Fee: fee, ValidVDFValue: validVDFValue}
	tr.Sign(from)
	return tr
}

// Serialize is responsible converting Transaction to byte array
func (t *Transaction) Serialize() *[256]byte {
	var res [256]byte
	start := 0
	for i := 0; i < ed25519.SignatureSize; i++ {
		res[start+i] = t.Signature[i]
	}
	start += ed25519.SignatureSize
	for i := 0; i < len(t.From); i++ {
		res[start+i] = t.From[i]
	}
	start += len(t.From)
	for i := 0; i < len(t.To); i++ {
		res[start+i] = t.To[i]
	}
	start += len(t.To)
	res[start] = byte(t.Token >> 56)
	res[start+1] = byte(t.Token >> 48)
	res[start+2] = byte(t.Token >> 40)
	res[start+3] = byte(t.Token >> 32)
	res[start+4] = byte(t.Token >> 24)
	res[start+5] = byte(t.Token >> 16)
	res[start+6] = byte(t.Token >> 8)
	res[start+7] = byte(t.Token)
	start += 8
	res[start] = byte(t.Fee >> 56)
	res[start+1] = byte(t.Fee >> 48)
	res[start+2] = byte(t.Fee >> 40)
	res[start+3] = byte(t.Fee >> 32)
	res[start+4] = byte(t.Fee >> 24)
	res[start+5] = byte(t.Fee >> 16)
	res[start+6] = byte(t.Fee >> 8)
	res[start+7] = byte(t.Fee)
	start += 8
	for i := 0; i < sha256.Size; i++ {
		res[start+i] = t.ValidVDFValue[i]
	}
	return &res
}

// TODO check on littleendian system
func ByteToInt64(b []byte, st int) int64 {
	return int64(uint64(b[st+7]) | uint64(b[st+6])<<8 | uint64(b[st+5])<<16 | uint64(b[st+4])<<24 |
		uint64(b[st+3])<<32 | uint64(b[st+2])<<40 | uint64(b[st+1])<<48 | uint64(b[st])<<56)
}
func ByteToInt32(b []byte, st int) int32 {
	return int32(uint32(b[st+3]) | uint32(b[st+2])<<8 | uint32(b[st+1])<<16 | uint32(b[st])<<24)
}

// DeserializeFromSlice is responsible for creating Transaction object from byte slice
func (t *Transaction) DeserializeFromSlice(b []byte) {
	t.From = make([]byte, ed25519.PublicKeySize)
	t.To = make([]byte, ed25519.PublicKeySize)
	t.ValidVDFValue = make([]byte, sha256.Size)
	t.Signature = make([]byte, ed25519.SignatureSize)
	start := 0
	for i := 0; i < ed25519.SignatureSize; i++ {
		t.Signature[i] = b[start+i]
	}
	start += ed25519.SignatureSize
	for i := 0; i < ed25519.PublicKeySize; i++ {
		t.From[i] = b[start+i]
	}
	start += ed25519.PublicKeySize
	for i := 0; i < ed25519.PublicKeySize; i++ {
		t.To[i] = b[start+i]
	}
	start += ed25519.PublicKeySize
	t.Token = ByteToInt64(b[:], start)
	start += 8
	t.Fee = ByteToInt64(b[:], start)
	start += 8
	for i := 0; i < sha256.Size; i++ {
		t.ValidVDFValue[i] = b[start+i]
	}
}

// Deserialize is responsible for creating Transaction object from byte array
func (t *Transaction) Deserialize(b *[256]byte) {
	t.DeserializeFromSlice(b[:])
}

// Verify method verifies that transaction fee is correct
func (t *Transaction) Verify() bool {
	return 0 <= t.Fee && t.Fee <= t.Token
}

// Equals compares two transactions
func (t *Transaction) Equals(tran Transaction) bool {

	if t.Fee != tran.Fee || !bytes.Equal(t.From, tran.From) ||
		!bytes.Equal(t.To, tran.To) || t.Token != tran.Token ||
		!bytes.Equal(t.Signature, tran.Signature) {
		return false
	}
	return true
}

// String method of Transaction struct
func (t *Transaction) String() string {
	return fmt.Sprintf("{ \nFrom: %v, \nTo: %v, \nAmount: %v, Fee: %v}", t.From, t.To, t.Token, t.Fee)
}

func (t *Transaction) signData() []byte {
	sData := make([]byte, 0, 112)
	sData = append(sData, t.From...)
	sData = append(sData, t.To...)
	// sData = append(sData, int64toBytes(t.Token)...)
	// sData = append(sData, int64toBytes(t.Fee)...)

	sData = append(sData, byte(t.Token>>56))
	sData = append(sData, byte(t.Token>>48))
	sData = append(sData, byte(t.Token>>40))
	sData = append(sData, byte(t.Token>>32))
	sData = append(sData, byte(t.Token>>24))
	sData = append(sData, byte(t.Token>>16))
	sData = append(sData, byte(t.Token>>8))
	sData = append(sData, byte(t.Token))
	sData = append(sData, byte(t.Fee>>56))
	sData = append(sData, byte(t.Fee>>48))
	sData = append(sData, byte(t.Fee>>40))
	sData = append(sData, byte(t.Fee>>32))
	sData = append(sData, byte(t.Fee>>24))
	sData = append(sData, byte(t.Fee>>16))
	sData = append(sData, byte(t.Fee>>8))
	sData = append(sData, byte(t.Fee))

	sData = append(sData, t.ValidVDFValue...)
	return sData
}

// Sign method signs Transaction with ed25519
func (t *Transaction) Sign(keypair *KeyPair) {
	sigData := t.signData()
	t.Signature = ed25519.Sign(keypair.Private, sigData)
}

// VerifySignature method verifies Transaction signature
func (t *Transaction) VerifySignature() bool {
	return len(t.From) == ed25519.PublicKeySize &&
		ed25519.Verify(t.From, t.signData(), t.Signature)
}

// TransactionSize returns size of transaction in bytes
func TransactionSize() int {
	return ed25519.PublicKeySize*2 + ed25519.SignatureSize + 16 + sha256.Size
}

// CreateDummyTransaction responsible for creating dummy Transaction object
// NOTE: Mainly for testing
func CreateDummyTransaction(tok int64) Transaction {
	tr := Transaction{Fee: 0, From: []byte("me"), To: []byte("you"), Token: tok, ValidVDFValue: []byte{1}, Signature: []byte(strconv.Itoa(int(tok)))}
	return tr
}

// CreateRealTransaction responsible for creating real Transaction object
// NOTE: Mainly for testing
func CreateRealTransaction(tok int64) Transaction {
	kp := NewKeyPair()
	// tr := NewTransaction(&kp, kp.Public, tok, 0, []byte{1})
	tr := NewTransaction(&kp, kp.Public, tok, 0, VDF([]byte{1}))
	return tr
}

// CreateRealTransactionFrom responsible for creating real Transaction object
// NOTE: Mainly for testing
func CreateRealTransactionFrom(tok int64, from []byte) Transaction {
	kp := NewKeyPair()
	// tr := NewTransaction(&kp, kp.Public, tok, 0, []byte{1})
	kp.Public = from
	tr := NewTransaction(&kp, from, tok, 0, VDF([]byte{1}))
	return tr
}
