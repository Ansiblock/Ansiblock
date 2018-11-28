package api

import (
	"reflect"

	"github.com/Ansiblock/Ansiblock/block"
)

type DBMock struct {
	Trans  *block.Transactions
	Block  *block.Block
	Blocks []*block.Block
	From   []byte
	To     []byte
}

func (db *DBMock) SaveBlock(blk block.Block) error {
	return nil
}

func (db *DBMock) GetBlockByHash(hash []byte) *block.Block {
	return db.Block
}

func (db *DBMock) GetBlockByHeight(height uint64) *block.Block {
	return db.Block
}

func (db *DBMock) GetTxFromBlockByHeight(height uint64, offset, limit uint64) (*block.Transactions, uint64) {
	return db.Trans, 0
}

func (db *DBMock) GetBlocksAfterHeight(offset, limit uint64) ([]block.Block, uint64) {
	res := make([]block.Block, 0)
	for _, bl := range db.Blocks {
		if bl.Number > offset {
			blk := block.Block{Number: bl.Number, Val: bl.Val}
			res = append(res, blk)
		}
	}
	return res, offset + uint64(len(res))
}

func (db *DBMock) GetBlocksBeforeHeight(offset, limit uint64) ([]block.Block, uint64) {
	res := make([]block.Block, 0)
	for _, bl := range db.Blocks {
		if bl.Number < offset {
			blk := block.Block{Number: bl.Number, Val: bl.Val}
			res = append(res, blk)
		}
	}
	return res, offset + uint64(len(res))
}

func (db *DBMock) GetTransactionsFrom(from []byte, offset, limit uint64) (*block.Transactions, uint64) {
	if !reflect.DeepEqual(from, db.From) {
		trans := new(block.Transactions)
		return trans, 0
	}
	return db.Trans, 0
}

func (db *DBMock) GetTransactionsTo(to []byte, offset, limit uint64) (*block.Transactions, uint64) {
	if !reflect.DeepEqual(to, db.To) {
		trans := new(block.Transactions)
		return trans, 0
	}
	return db.Trans, 0
}

func (db *DBMock) GetAccountTransactions(account []byte, offset, limit uint64) (*block.Transactions, uint64) {
	return db.Trans, 0
}
