package mint

import (
	"github.com/Ansiblock/Ansiblock/block"
	"golang.org/x/crypto/ed25519"
)

const privateKeySize = 1024

// Mint structure stores initial amount.
type Mint struct {
	KeyPair    block.KeyPair
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
	Tokens     int64

	// publicKey  string
	// publicKey  rsa.PublicKey
}

// NewMint returns Mint with random private and public keys, and initial amount(tokens).
func NewMint(tokens int64) Mint {
	keyPair := block.NewKeyPair()
	privateKey := keyPair.Private
	publicKey := keyPair.Public

	return Mint{keyPair, privateKey, publicKey, tokens}
}

// CreateTransactions returns transaction from mint account to mint account
func (m *Mint) CreateTransactions() []block.Transaction {
	trans := []block.Transaction{block.NewTransaction(&m.KeyPair, m.PublicKey, m.Tokens, 0, block.VDF(m.KeyPair.Private))}
	return trans
}

// CreateBlocks returns two blocks, empty block and block with single transaction
func (m *Mint) CreateBlocks() []block.Block {
	emptyBlock := block.NewEmpty(block.VDF(m.KeyPair.Private), 0, 0)
	trans := block.Transactions{}
	trans.Ts = m.CreateTransactions()
	selfTransactionBlock := block.New(emptyBlock.Val, 1, 0, &trans)
	return []block.Block{emptyBlock, selfTransactionBlock}
}

// ValidVDFValue returns last vdf of first block
func (m *Mint) ValidVDFValue() block.VDFValue {
	bls := m.CreateBlocks()
	return bls[1].Val
}

// publicKeyToString converts rsa.PublicKey object to string
// func publicKeyToString(publicKey *rsa.PublicKey) string {
// 	publicKeyDer, err := x509.MarshalPKIXPublicKey(publicKey)
// 	if err != nil {
// 		if err != nil {
// 			log.Error("failed to convert public ket to byte[]", zap.String("error", err.Error()))
// 			os.Exit(1)
// 		}
// 	}

// 	publicKeyBlock := pem.Block{
// 		Type:    "PUBLIC KEY",
// 		Headers: nil,
// 		Bytes:   publicKeyDer,
// 	}
// 	publicKeyPem := string(pem.EncodeToMemory(&publicKeyBlock))
// 	return publicKeyPem
// }
