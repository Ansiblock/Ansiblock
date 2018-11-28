package api

import (
	"strconv"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/replication"
	"golang.org/x/crypto/ed25519"
)

type BlockchainApiMock struct {
	NodesMap             map[string]*replication.NodeData
	TransactionsTotalVal uint64
	BlocksToTalVal       uint64
	BalanceValues        []int64
	AccTransactions      *block.Transactions
	BlockTransactions    *block.Transactions
	BlocksList           []block.Block
	Block                *block.Block
	BlockHeightVal       uint64
	TPSVal               int64
	BlockTimeVal         int64
	QueryParams          map[string]string
}

func NewBlockchainAPIMock() *BlockchainApiMock {
	mock := new(BlockchainApiMock)
	mock.QueryParams = make(map[string]string)
	return mock
}

func (apiMock *BlockchainApiMock) Nodes() map[string]*replication.NodeData {
	return apiMock.NodesMap
}

func (apiMock *BlockchainApiMock) TransactionsTotal() uint64 {
	return apiMock.TransactionsTotalVal
}

func (apiMock *BlockchainApiMock) BlocksTotal() uint64 {
	return apiMock.BlocksToTalVal
}

func (apiMock *BlockchainApiMock) Balances(keys []string) []int64 {
	return apiMock.BalanceValues
}

func (apiMock *BlockchainApiMock) TransactionsFrom(keyBase64 string, offset, limit uint64) (*block.Transactions, uint64) {
	apiMock.QueryParams["from"] = keyBase64
	apiMock.QueryParams["offset"] = strconv.Itoa(int(offset))
	apiMock.QueryParams["limit"] = strconv.Itoa(int(limit))

	return apiMock.AccTransactions, 0
}

func (apiMock *BlockchainApiMock) TransactionsTo(keyBase64 string, offset, limit uint64) (*block.Transactions, uint64) {
	apiMock.QueryParams["to"] = keyBase64
	apiMock.QueryParams["offset"] = strconv.Itoa(int(offset))
	apiMock.QueryParams["limit"] = strconv.Itoa(int(limit))

	return apiMock.AccTransactions, 0
}

func (apiMock *BlockchainApiMock) AccountTransactions(keyBase64 string, offset, limit uint64) (*block.Transactions, uint64) {
	apiMock.QueryParams["accountKey"] = keyBase64
	apiMock.QueryParams["offset"] = strconv.Itoa(int(offset))
	apiMock.QueryParams["limit"] = strconv.Itoa(int(limit))

	return apiMock.AccTransactions, 0
}

func (apiMock *BlockchainApiMock) BlockTransactionsByHeight(blockHeight uint64, offset, limit uint64) (*block.Transactions, uint64) {
	apiMock.QueryParams["blockHeight"] = strconv.Itoa(int(blockHeight))
	apiMock.QueryParams["offset"] = strconv.Itoa(int(offset))
	apiMock.QueryParams["limit"] = strconv.Itoa(int(limit))

	return apiMock.BlockTransactions, 0
}

func (apiMock *BlockchainApiMock) Blocks(offset, limit uint64) ([]block.Block, uint64) {
	apiMock.QueryParams["offset"] = strconv.Itoa(int(offset))
	apiMock.QueryParams["limit"] = strconv.Itoa(int(limit))

	return apiMock.BlocksList, 0
}

func (apiMock *BlockchainApiMock) BlockByHeight(height uint64) *block.Block {
	apiMock.QueryParams["blockHeight"] = strconv.Itoa(int(height))

	return apiMock.Block
}

func (apiMock *BlockchainApiMock) BlockByHash(hashBase64 string) *block.Block {
	apiMock.QueryParams["blockHash"] = hashBase64
	return apiMock.Block
}

func (apiMock *BlockchainApiMock) BlockHeight() uint64 {
	return apiMock.BlockHeightVal
}

func (apiMock *BlockchainApiMock) TPS() int64 {
	return apiMock.TPSVal
}

func (apiMock *BlockchainApiMock) BlockTime() int64 {
	return apiMock.BlockTimeVal
}

func (apiMock *BlockchainApiMock) RandomKeys(uint64) []ed25519.PublicKey {
	return nil
}

func (apiMock *BlockchainApiMock) MintKey() ed25519.PublicKey {
	return nil
}
