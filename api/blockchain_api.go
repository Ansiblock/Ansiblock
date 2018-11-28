package api

import (
	"encoding/base64"
	synchro "sync"
	"time"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/books"
	"github.com/Ansiblock/Ansiblock/mint"
	"github.com/Ansiblock/Ansiblock/replication"
	"golang.org/x/crypto/ed25519"
)

const (
	statsCount     = 100
	maxBlockTime   = 10000
	blockTimeDelay = 1 * time.Second
	tpsDelay       = 1 * time.Second
)

// BlockchainAPI gives access to information about blocks, transactions, nodes etc. part of information is extracted from db.
type BlockchainAPI interface {
	Nodes() map[string]*replication.NodeData
	TransactionsTotal() uint64
	BlocksTotal() uint64
	Balances(keys []string) []int64
	TransactionsFrom(keyBase64 string, offset, limit uint64) (*block.Transactions, uint64)
	TransactionsTo(keyBase64 string, offset, limit uint64) (*block.Transactions, uint64)
	AccountTransactions(keyBase64 string, offset, limit uint64) (*block.Transactions, uint64)
	BlockTransactionsByHeight(blockHeight uint64, offset, limit uint64) (*block.Transactions, uint64)
	Blocks(offset, limit uint64) ([]block.Block, uint64)
	BlockByHeight(height uint64) *block.Block
	BlockByHash(hashBase64 string) *block.Block
	BlockHeight() uint64
	TPS() int64
	BlockTime() int64
	RandomKeys(uint64) []ed25519.PublicKey
	MintKey() ed25519.PublicKey
}

// API struct stores all data source objects for blockchain api methods
type API struct {
	bm    *books.Accounts
	db    DataBase
	mint  *mint.Mint
	sync  *replication.Sync
	stats *Stats
}

// Stats struct stores global statistics
type Stats struct {
	tps      [statsCount]int64
	maxTPS   int64
	tpsMutex *synchro.Mutex
	onceTPS  synchro.Once

	blockTime      [statsCount]int64
	minBlockTime   int64
	blockTimeMutex *synchro.Mutex
	onceBlockTime  synchro.Once
}

// New returns new instance of blockchain api
func New(bm *books.Accounts, db DataBase, sync *replication.Sync, m *mint.Mint) *API {
	api := new(API)
	api.bm = bm
	api.db = db
	api.mint = m
	api.sync = sync
	api.stats = new(Stats)
	api.stats.maxTPS = 0
	api.stats.tpsMutex = &synchro.Mutex{}
	api.stats.minBlockTime = maxBlockTime
	api.stats.blockTimeMutex = &synchro.Mutex{}
	return api
}

// Nodes returns remote nodes connected to us
func (api *API) Nodes() map[string]*replication.NodeData {
	return api.sync.RemoteTableCopy()
}

// TransactionsTotal returns total transactions count
func (api *API) TransactionsTotal() uint64 {
	return api.bm.TransactionsTotal()
}

// BlocksTotal returns number of processed blocks from books
func (api *API) BlocksTotal() uint64 {
	return api.bm.BlocksTotal()
}

// Balances takes array of account public keys in base64 format and returns balances
func (api *API) Balances(keys []string) []int64 {
	balances := make([]int64, len(keys))
	for i := 0; i < len(keys); i++ {
		key, _ := base64.StdEncoding.DecodeString(keys[i])
		balances[i] = api.bm.Balance(key)
	}
	return balances
}

// TransactionsFrom returns transactions from 'keyBase64' account
func (api *API) TransactionsFrom(keyBase64 string, offset, limit uint64) (*block.Transactions, uint64) {
	key, _ := base64.StdEncoding.DecodeString(keyBase64)
	return api.db.GetTransactionsFrom(key, offset, limit)
}

// TransactionsTo returns transactions to 'keyBase64' account
func (api *API) TransactionsTo(keyBase64 string, offset, limit uint64) (*block.Transactions, uint64) {
	key, _ := base64.StdEncoding.DecodeString(keyBase64)
	return api.db.GetTransactionsTo(key, offset, limit)
}

// AccountTransactions returns transactions from and to 'keyBase64' account
func (api *API) AccountTransactions(keyBase64 string, offset, limit uint64) (*block.Transactions, uint64) {
	key, _ := base64.StdEncoding.DecodeString(keyBase64)
	return api.db.GetAccountTransactions(key, offset, limit)
}

//BlockTransactionsByHeight looks for block according heigh and returns it's transactions
func (api *API) BlockTransactionsByHeight(blockHeight uint64, offset, limit uint64) (*block.Transactions, uint64) {
	return api.db.GetTxFromBlockByHeight(blockHeight, offset, limit)
}

// Blocks returns block according offset
func (api *API) Blocks(offset, limit uint64) ([]block.Block, uint64) {
	blocks, offset := api.db.GetBlocksBeforeHeight(offset, limit)
	return blocks, offset
}

// BlockByHeight returns meta info of block with specific height value
func (api *API) BlockByHeight(height uint64) *block.Block {
	return api.db.GetBlockByHeight(height)
}

// BlockByHash returns block with hash
func (api *API) BlockByHash(hashBase64 string) *block.Block {
	hash, _ := base64.StdEncoding.DecodeString(hashBase64)
	return api.db.GetBlockByHash(hash)
}

// BlockHeight returns last block's height, registered in the ledger
func (api *API) BlockHeight() uint64 {
	return api.bm.LastBlock().Number
}

// TPS returns latest TPS
func (api *API) TPS() int64 {
	stop := true
	api.stats.onceTPS.Do(func() {
		go api.calculateTPS(&stop)
	})
	api.stats.tpsMutex.Lock()
	defer api.stats.tpsMutex.Unlock()
	return api.stats.tps[0]
}

func (api *API) calculateTPS(stop *bool) {
	start := time.Now()
	firstCount := api.bm.TransactionsTotal()
	for *stop {
		time.Sleep(tpsDelay)

		trCount := api.bm.TransactionsTotal()
		duration := time.Since(start)
		start = time.Now()
		count := trCount - firstCount
		firstCount = trCount
		tps := int64(float64(count) / duration.Seconds())
		api.stats.tpsMutex.Lock()
		if api.stats.maxTPS < tps {
			api.stats.maxTPS = tps
		}
		if tps == 0 {
			tps = api.stats.tps[0]
		}
		for i := statsCount - 2; i >= 0; i-- {
			api.stats.tps[i+1] = api.stats.tps[i]
		}
		api.stats.tps[0] = tps

		api.stats.tpsMutex.Unlock()
	}
}

// BlockTime returns time for creating single block in milliseconds
func (api *API) BlockTime() int64 {
	stop := true
	api.stats.onceBlockTime.Do(func() {
		go api.calculateBlockTime(&stop)
	})
	api.stats.blockTimeMutex.Lock()
	defer api.stats.blockTimeMutex.Unlock()
	return api.stats.blockTime[0]
}

func (api *API) calculateBlockTime(stop *bool) {
	start := time.Now()
	firstCount := api.bm.BlocksTotal()
	for i := range api.stats.blockTime {
		api.stats.blockTime[i] = int64(maxBlockTime)
	}
	for *stop {
		time.Sleep(blockTimeDelay)

		blCount := api.bm.BlocksTotal()
		// fmt.Printf("blCount = %v\n", blCount)
		duration := time.Since(start)
		start = time.Now()
		count := blCount - firstCount
		firstCount = blCount

		blockTime := api.stats.blockTime[0]
		if count != 0 {
			blockTime = int64(duration.Seconds() * 1000 / float64(count))
		}
		api.stats.blockTimeMutex.Lock()
		if api.stats.minBlockTime > blockTime {
			api.stats.minBlockTime = blockTime
		}
		for i := statsCount - 2; i >= 0; i-- {
			api.stats.blockTime[i+1] = api.stats.blockTime[i]
		}
		api.stats.blockTime[0] = blockTime
		api.stats.blockTimeMutex.Unlock()
	}
}

// RandomKeys returns account keys from book manager, to monitor them on web site
func (api *API) RandomKeys(num uint64) []ed25519.PublicKey {
	return api.bm.RandomKeys(num)
}

// MintKey returns mint key, to monitor it's balance on web site
func (api *API) MintKey() ed25519.PublicKey {
	return api.mint.PublicKey
}
