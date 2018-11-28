package block

import (
	"bytes"
	"strings"
	"testing"

	"golang.org/x/crypto/ed25519"
)

func TestTransactionGetters(t *testing.T) {
	tran := Transaction{From: []byte("me"), To: []byte("you"), Token: 5, Fee: 1}
	from := tran.From
	to := tran.To
	token := tran.Token
	fee := tran.Fee
	if !bytes.Equal(from, []byte("me")) || !bytes.Equal(to, []byte("you")) ||
		token != 5 || fee != 1 {
		t.Fatalf("Transaction Getters: %v = {%v, %v, %v, %v}", tran, from, to, token, fee)
	}
}

func TestVerify2(t *testing.T) {
	tran1 := Transaction{Token: 5, Fee: 1}
	tran2 := Transaction{Token: 5, Fee: 10}

	if !tran1.Verify() {
		t.Errorf("Transaction Verified: %v", tran1)
	}
	if tran2.Verify() {
		t.Errorf("Transaction Verified: %v", tran2)

	}
}

func TestEquals(t *testing.T) {
	tr1 := CreateDummyTransaction(10)
	tr2 := CreateDummyTransaction(10)
	tr3 := CreateDummyTransaction(20)

	if tr1.Equals(tr3) {
		t.Errorf("%v . Equals(%v) ", tr1, tr3)
	}
	if !tr1.Equals(tr2) {
		t.Errorf("! %v . Equals(%v) ", tr1, tr2)
	}
}

func TestSignature(t *testing.T) {
	tran1 := Transaction{Signature: []byte{1, 2, 3}}
	if !bytes.Equal(tran1.Signature, []byte{1, 2, 3}) {
		t.Errorf("Signature Fail")
	}
}

func TestNew2(t *testing.T) {
	kp := NewKeyPair()
	tran := NewTransaction(&kp, []byte("2"), 3, 4, []byte{5, 6})
	if !bytes.Equal(tran.From, kp.Public) || string(tran.To) != "2" || tran.Token != 3 ||
		tran.Fee != 4 || !tran.VerifySignature() {
		t.Errorf("New Failed")
	}

}

func TestTransactionString(t *testing.T) {
	kp := NewKeyPair()
	tran := NewTransaction(&kp, []byte("2"), 3, 4, []byte{5, 6})
	res := tran.String()
	if !strings.Contains(res, "To: [50], \nAmount: 3, Fee: 4}") {
		t.Errorf("Transaction.String(%v) = %v Failed", tran, res)
	}
}

func TestSerialize(t *testing.T) {
	for i := int64(0); i < 10; i++ {
		tr := CreateRealTransaction(i)
		a := Transaction{}
		a.Deserialize(tr.Serialize())
		if !a.Equals(tr) {
			t.Errorf("Serialize/deserialize problem: %v != %v", a, tr)
		}
	}
}

func TestSigVerify(t *testing.T) {
	for i := int64(0); i < 10; i++ {
		tran := CreateRealTransaction(i)
		if !tran.VerifySignature() {
			t.Errorf("SigVerify problem %v: %v ", i, tran)
		}
	}
}
func TestSigVerify2(t *testing.T) {
	tran := CreateRealTransaction(1)
	s := tran.Serialize()
	var tr Transaction
	tr.Deserialize(s)
	if !tr.Equals(tran) {
		t.Errorf("!!SigVerify problem: %v\n != %v\n ", tran, tr)
	}
	if !tr.VerifySignature() {
		t.Errorf("SigVerify problem: %v \n %v\n", tran, tr)
	}
}

func TestSigVerify3(t *testing.T) {
	tran := CreateRealTransaction(1)
	s := tran.Serialize()

	if !bytes.Equal(s[64:176], tran.signData()) {
		t.Errorf("!!SigVerify problem: %v\n !=\n %v\n ", s[64:], tran.signData())
	}
	if !ed25519.Verify(s[64:96], s[64:176], s[0:64]) {
		t.Errorf("!!SigVerify problem: %v\n %v\n %v\n !=\n %v\n ", s[64:96], s[64:176], s[0:64], tran)
	}

}

func TestCreateRealTransactionFrom(t *testing.T) {
	from := NewKeyPair()
	tran := CreateRealTransactionFrom(10, from.Public)
	if tran.Token != 10 || !bytes.Equal(tran.From, from.Public) {
		t.Errorf("CreateRealTransactionFrom error %v!=%v\n%v\n!=\n%v\n", tran.Token, 10, tran.From, from.Public)
	}
}
