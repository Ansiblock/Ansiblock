package books

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"sync/atomic"

	"golang.org/x/crypto/ed25519"

	"go.uber.org/zap"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/log"
)

// ErrorCode values for transactions process
var (
	// ErrAccountNotFound is returned when `PublicKey` is not found.
	errAccountNotFound = errors.New("account not found")

	// ErrInsufficientFunds is returned when there is not enough amount.
	errInsufficientFunds = errors.New("insufficient funds")

	// ErrNegativeTokens is returned when requested a debit or credit of negative tokens.
	errNegativeTokens = errors.New("debit or credit of negative tokens")
)

// Accounts stores account balances and ledger info.
type Accounts struct {
	balances          map[string]int64
	lock              sync.Mutex
	ledger            *Ledger
	transactionsTotal uint64
	blocksTotal       uint64
}

// NewBookManager creates new Accounts object
func NewBookManager() *Accounts {
	bm := new(Accounts)
	bm.balances = make(map[string]int64)
	bm.ledger = newLedger()
	bm.transactionsTotal = 0
	bm.blocksTotal = 0
	return bm
}

func (bm *Accounts) applyTransactionWithdraw(tran *block.Transaction) error {
	if tran.Token < 0 {
		return errNegativeTokens
	}
	bm.lock.Lock()
	defer bm.lock.Unlock()
	fromBalance, ok := bm.balances[string(tran.From)]
	if !ok {
		return errAccountNotFound
	}
	err := bm.ledger.addSignature(tran.Signature, tran.ValidVDFValue)
	if err != nil {
		return err
	}
	if fromBalance < tran.Token {
		bm.ledger.removeSignature(tran.Signature, tran.ValidVDFValue)
		return errInsufficientFunds
	}
	bm.balances[string(tran.From)] -= tran.Token
	atomic.AddUint64(&bm.transactionsTotal, 1)
	return nil
}

func (bm *Accounts) applyTransactionDeposit(tran *block.Transaction) {
	bm.lock.Lock()
	defer bm.lock.Unlock()
	bm.balances[string(tran.To)] += tran.Token - tran.Fee
}

func (bm *Accounts) processTransactionsWithdraws(trans block.Transactions) block.Transactions {
	log.Info(fmt.Sprintf("AccountManager: apply withdraws %v transactions", len(trans.Ts)))
	res := make([]block.Transaction, 0, len(trans.Ts))
	for _, tran := range trans.Ts {
		err := bm.applyTransactionWithdraw(&tran)
		if err == nil {
			res = append(res, tran)
		} else {
			log.Error("failed to process transaction withdraw", zap.String(err.Error(), tran.String()))
		}
	}
	return block.Transactions{Ts: res}
}

func (bm *Accounts) processTransactionsDeposits(trans block.Transactions) {
	log.Info(fmt.Sprintf("AccountManager: apply deposits %v transactions", len(trans.Ts)))
	for _, tran := range trans.Ts {
		bm.applyTransactionDeposit(&tran)
	}
}

// ProcessTransactions process a batch of transactions.
func (bm *Accounts) ProcessTransactions(trans block.Transactions) block.Transactions {
	log.Info(fmt.Sprintf("AccountManager: process %v transactions", len(trans.Ts)))
	res := bm.processTransactionsWithdraws(trans)
	bm.processTransactionsDeposits(res)
	return res
}

// ProcessBlocks process a list of blocks.
// TODO: change copying transactions
func (bm *Accounts) ProcessBlocks(blocks []block.Block) (err error) {
	log.Info(fmt.Sprintf("AccountManager: process %v blocks", len(blocks)))
	for _, bl := range blocks {
		bm.UpdateLastBlock(&bl)
		bm.ProcessTransactions(*bl.Transactions)
	}
	return nil
}

// CreateAccount creates new account with initial balance.
// NOTE: Only for testing.
func (bm *Accounts) CreateAccount(publicKey ed25519.PublicKey, amount int64) {
	if bm.balances == nil {
		bm.balances = make(map[string]int64)
	}
	bm.balances[string(publicKey)] = amount
}

// Balance method returns balance of account with public key equals publicKey.
// NOTE: Only for testing.
func (bm *Accounts) Balance(publicKey ed25519.PublicKey) int64 {
	bm.lock.Lock()
	defer bm.lock.Unlock()
	return bm.balances[string(publicKey)]
}

// String method prints accounts balances.
// NOTE: Only for testing.
func (bm *Accounts) String() string {
	b := new(bytes.Buffer)
	for key, value := range bm.balances {
		fmt.Fprintf(b, "%v=\"%v\"\n", key, value)
	}
	fmt.Fprintf(b, "transactions total=%v\n", bm.transactionsTotal)
	return b.String()
}

// AddValidVDFValue adds last seen vdf value to the ledger
func (bm *Accounts) AddValidVDFValue(val block.VDFValue) {
	bm.ledger.AddValidVDFValue(val)
}

// ValidVDFValue returns the valid vdf value registered.
func (bm *Accounts) ValidVDFValue() []byte {
	return bm.ledger.validVDFValue()
}

// TransactionsTotal returns count of transactions
func (bm *Accounts) TransactionsTotal() uint64 {
	return atomic.LoadUint64(&bm.transactionsTotal)
}

// BlocksTotal returns count of processed blocks
func (bm *Accounts) BlocksTotal() uint64 {
	return atomic.LoadUint64(&bm.blocksTotal)
}

// IncreaseBlocksTotal increases Block Count
func (bm *Accounts) IncreaseBlocksTotal() {
	atomic.AddUint64(&bm.blocksTotal, 1)
}

// LastBlock returns last block registered in legder
func (bm *Accounts) LastBlock() *block.Block {
	// fmt.Printf("bm = %v\n", bm)
	if bm.ledger != nil {
		return bm.ledger.LastBlock()
	}
	return nil
}

// UpdateLastBlock updates last block in ledger and counts number of blocks
func (bm *Accounts) UpdateLastBlock(bl *block.Block) {
	bm.ledger.UpdateLastBlock(bl)
	atomic.AddUint64(&bm.blocksTotal, 1)
	bm.ledger.AddValidVDFValue(bl.Val)
}

// Clone method returns clone of Accounts struct
func (bm *Accounts) Clone() *Accounts {
	clone := new(Accounts)
	clone.balances = make(map[string]int64)
	for k, v := range bm.balances {
		clone.balances[k] = v
	}
	clone.ledger = bm.ledger.Clone()
	clone.transactionsTotal = bm.transactionsTotal
	clone.blocksTotal = bm.blocksTotal
	return clone
}

// Equals compares two Accounts and returns true if they have same accounts with same balances
func (bm *Accounts) Equals(bm2 *Accounts) bool {
	return bm.transactionsTotal == bm2.transactionsTotal && bm.blocksTotal == bm2.blocksTotal &&
		bm.ledger.Equals(bm2.ledger) && reflect.DeepEqual(bm.balances, bm2.balances)
}

// RandomKeys returns account keys, which are used as default account list on web monitoring tool
func (bm *Accounts) RandomKeys(num uint64) []ed25519.PublicKey {
	list := make([]ed25519.PublicKey, 0, num)
	for k := range bm.balances {
		list = append(list, []byte(k))
		num--
		if num == 0 {
			break
		}
	}
	return list
}

// RandomAccounts creates bank with random accounts, for testing
func RandomAccounts(numAccounts int) (*Accounts, []block.KeyPair) {
	bm := NewBookManager()
	keyPairs := make([]block.KeyPair, numAccounts)
	for i := 0; i < numAccounts; i++ {
		kp := block.NewKeyPair()
		keyPairs[i] = kp
		accBalance := rand.Int63n(1000)
		bm.CreateAccount(kp.Public, accBalance)
	}
	return bm, keyPairs
}

// RandomTransactions returns random transactions and number of transactions that will be processed successfully. Only for testing!
func RandomTransactions(bm *Accounts, numTransactions int, keyPairs []block.KeyPair, vdfValue []byte) (*block.Transactions, uint64) {
	bmClone := bm.Clone()
	numAccounts := len(keyPairs)
	trans := new(block.Transactions)
	trans.Ts = make([]block.Transaction, numTransactions)
	var numValidTransactions uint64
	for j := 0; j < numTransactions; j++ {
		fromIndex := rand.Intn(numAccounts)
		toIndex := rand.Intn(numAccounts)
		amount := rand.Int63n(100) + 1
		trans.Ts[j] = block.NewTransaction(&keyPairs[fromIndex], keyPairs[toIndex].Public, amount, 0, vdfValue)
		if bmClone.Balance(keyPairs[fromIndex].Public) >= amount {
			numValidTransactions++
			bmClone.balances[string(trans.Ts[j].From)] -= amount
		}
	}

	return trans, numValidTransactions
}

// RandomTransactionsBlocks generates blocks with random transactions between real accounts, for testing
func RandomTransactionsBlocks(bm *Accounts, numTransactionsPerBlock int, numBlocks int, keyPairs []block.KeyPair) []block.Block {
	blocks := make([]block.Block, numBlocks)
	numAccounts := len(keyPairs)
	for i := 0; i < numBlocks; i++ {
		trans := new(block.Transactions)
		trans.Ts = make([]block.Transaction, numTransactionsPerBlock)
		var vdfValue = block.VDF([]byte{byte(i), 2, 3})
		bm.AddValidVDFValue(vdfValue)

		for j := 0; j < numTransactionsPerBlock; j++ {
			fromIndex := rand.Intn(numAccounts)
			toIndex := rand.Intn(numAccounts)
			amount := rand.Int63n(100) + 1
			trans.Ts[j] = block.NewTransaction(&keyPairs[fromIndex], keyPairs[toIndex].Public, amount, 0, vdfValue)
		}
		blocks[i].Transactions = trans
		blocks[i].Count = 1
		blocks[i].Val = vdfValue
	}
	return blocks
}
