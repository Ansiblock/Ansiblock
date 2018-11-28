package api

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	"github.com/Ansiblock/Ansiblock/block"

	// initialize sqlite driver
	"github.com/Ansiblock/Ansiblock/log"
	// to initialize sqlite driver
	_ "github.com/mattn/go-sqlite3"
)

const (
	// DBFilename is a path of the database file
	DBFilename = "./api.db"

	// maximum number of transactions to get from database per request
	maxTransactions = uint64(100)

	// MaxOffset is maximum number of transactions in database
	// https://github.com/golang/go/issues/9373 issue prevents to use max uint64 value
	MaxOffset = ^uint64(0) / 2

	// maximum number of blocks in database
	maxNumberOfBlocks = 1000

	// check number of blocks in database every time counter reaches checkinterval
	checkInterval = 100
)

// DB is database object, created by NewDBConnection method
type DB struct {
	// database connection handler
	conn    *sql.DB
	mutex   *sync.RWMutex
	counter uint64
}

// DataBase interface
type DataBase interface {
	SaveBlock(blk block.Block) error
	GetBlockByHash(hash []byte) *block.Block
	GetBlockByHeight(Height uint64) *block.Block
	GetBlocksAfterHeight(offset, limit uint64) ([]block.Block, uint64)
	GetBlocksBeforeHeight(offset, limit uint64) ([]block.Block, uint64)
	GetTxFromBlockByHeight(height uint64, offset, limit uint64) (*block.Transactions, uint64)
	GetTransactionsFrom(from []byte, offset, limit uint64) (*block.Transactions, uint64)
	GetTransactionsTo(to []byte, offset, limit uint64) (*block.Transactions, uint64)
	GetAccountTransactions(account []byte, offset, limit uint64) (*block.Transactions, uint64)
}

// NewDBConnection creates a connection to database
// Database will be initialized with tables and indexes if
// there is no file or it does not contain necessary schema
func NewDBConnection(dbFileName string) *DB {
	os.Remove(DBFilename)
	connection, err := sql.Open("sqlite3", DBFilename)
	checkErr(err)
	err = connection.Ping()
	checkErr(err)
	db := DB{conn: connection, mutex: &sync.RWMutex{}}
	db.createTablesIfNotExist()
	db.createIndexesIfNotExist()
	return &db
}

// helper function to create tables
func (db *DB) createTablesIfNotExist() {
	statement, err := db.conn.Prepare("CREATE TABLE IF NOT EXISTS blocks " +
		"(Height INTEGER PRIMARY KEY, Count INTEGER, Val BLOB, numTrans INTEGER)")
	checkErr(err)
	_, err = statement.Exec()
	checkErr(err)

	stmt := "CREATE TABLE IF NOT EXISTS transactions (id INTEGER PRIMARY KEY, Height INTEGER, [From] BLOB, [To] BLOB, " +
		"Token INTEGER, Fee INTEGER, ValidVDFValue BLOB, Signature BLOB, FOREIGN KEY (Height) REFERENCES blocks(Height))"
	statement, err = db.conn.Prepare(stmt)
	checkErr(err)
	_, err = statement.Exec()
	checkErr(err)
}

// helper function to create indexes
func (db *DB) createIndexesIfNotExist() {
	statements := []string{"CREATE INDEX IF NOT EXISTS blocks_vdf_index ON blocks (Val)",
		"CREATE INDEX IF NOT EXISTS transactions_height_index ON transactions (Height)",
		"CREATE INDEX IF NOT EXISTS transactions_from_index ON transactions ([From])",
		"CREATE INDEX IF NOT EXISTS transactions_to_index ON transactions ([To])"}

	for _, stmt := range statements {
		statement, err := db.conn.Prepare(stmt)
		checkErr(err)
		_, err = statement.Exec()
		checkErr(err)
	}
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err.Error())
		log.Fatal(err.Error())
	}
}

// SaveBlock saves given block to database
func (db *DB) SaveBlock(blk block.Block) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	statement, err := db.conn.Prepare("INSERT INTO blocks (Height, Count, Val, numTrans) VALUES(?, ?, ?, ?)")
	checkErr(err)
	defer statement.Close()
	_, err = statement.Exec(blk.Number, blk.Count, blk.Val, blk.Transactions.Count())
	checkErr(err)
	transactions := blk.Transactions
	tx, err := db.conn.Begin()
	checkErr(err)
	stmt, err := tx.Prepare("INSERT INTO transactions (Height, [From], [To], Token, Fee, " +
		"ValidVDFValue, Signature) VALUES(?, ?, ?, ?, ?, ?, ?)")
	checkErr(err)
	defer stmt.Close()
	for _, t := range transactions.Ts {
		_, err = stmt.Exec(blk.Number, t.From, t.To, t.Token, t.Fee, t.ValidVDFValue, t.Signature)
		checkErr(err)
	}
	// commit transaction
	tx.Commit()

	db.counter++
	db.checkAndDeleteOldData()
	return nil
}

func (db *DB) checkAndDeleteOldData() {
	if db.counter >= checkInterval {
		var numberOfBlocks uint64
		err := db.conn.QueryRow("SELECT Count(*) FROM blocks").Scan(&numberOfBlocks)
		checkErr(err)

		if numberOfBlocks > maxNumberOfBlocks {
			db.deleteOldData(numberOfBlocks - maxNumberOfBlocks)
		}

		db.counter = 0
	}
}

// trims database size to "limit" number of blocks
func (db *DB) deleteOldData(limit uint64) {
	// Get heights of old transactions
	query := "SELECT Height FROM blocks ORDER BY Height ASC LIMIT ?"
	rows, err := db.conn.Query(query, limit)
	checkErr(err)
	defer rows.Close()
	var blockHeights []uint64
	var height uint64
	for rows.Next() {
		err := rows.Scan(&height)
		checkErr(err)
		blockHeights = append(blockHeights, height)
	}
	err = rows.Err()
	checkErr(err)

	// delete old data from blocks and transactions tables in on sql transaction
	tx, err := db.conn.Begin()
	checkErr(err)
	blkStmt, err := tx.Prepare("DELETE FROM blocks WHERE Height = ?")
	checkErr(err)
	defer blkStmt.Close()
	transStmt, err := tx.Prepare("DELETE FROM transactions WHERE Height = ?")
	checkErr(err)
	defer transStmt.Close()
	for _, h := range blockHeights {
		_, err = blkStmt.Exec(h)
		checkErr(err)
		_, err = transStmt.Exec(h)
		checkErr(err)
	}
	// commit transaction
	tx.Commit()
}

// getBlock gets only one block from database with provided query
func (db *DB) getBlock(query string, param interface{}) *block.Block {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	var blk block.Block
	var numTrans int64
	err := db.conn.QueryRow(query, param).Scan(&blk.Number, &blk.Count, &blk.Val, &numTrans)
	if err == sql.ErrNoRows {
		return nil
	}
	blk.Transactions = &block.Transactions{Ts: make([]block.Transaction, numTrans)}
	return &blk
}

// GetBlockByHash gets block from database with VDF value
// returns pointer to block or nil if not found
func (db *DB) GetBlockByHash(hash []byte) *block.Block {
	query := "SELECT * FROM blocks WHERE Val = ?"
	return db.getBlock(query, hash)
}

// GetBlockByHeight gets block from database searching by height
// returns pointer to block or nil if not found
func (db *DB) GetBlockByHeight(height uint64) *block.Block {
	query := "SELECT * FROM blocks WHERE Height = ?"
	return db.getBlock(query, height)
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

// getBlocks get array of blocks from database with provided query
func (db *DB) getBlocks(query string, offset, limit uint64) ([]block.Block, uint64) {
	limit = min(maxTransactions, limit)
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	rows, err := db.conn.Query(query, offset, limit)
	checkErr(err)
	defer rows.Close()

	var blkArray []block.Block
	var blk block.Block
	var numTrans uint64
	for rows.Next() {
		err := rows.Scan(&blk.Number, &blk.Count, &blk.Val, &numTrans)
		checkErr(err)
		blk.Transactions = &block.Transactions{Ts: make([]block.Transaction, numTrans)}
		blkArray = append(blkArray, blk)
	}
	err = rows.Err()
	checkErr(err)
	if len(blkArray) > 0 {
		offset = blkArray[len(blkArray)-1].Number
	}
	return blkArray, offset
}

// GetBlocksAfterHeight gets unseen new blocks from database by height,
// including block with given height
func (db *DB) GetBlocksAfterHeight(offset, limit uint64) ([]block.Block, uint64) {
	query := "SELECT * FROM blocks WHERE Height >= ? ORDER BY Height ASC LIMIT ?"
	return db.getBlocks(query, offset, limit)
}

// GetBlocksBeforeHeight gets old blocks from database by height,
// including block with given height
func (db *DB) GetBlocksBeforeHeight(offset, limit uint64) ([]block.Block, uint64) {
	query := "SELECT * FROM blocks WHERE Height <= ? ORDER BY Height DESC LIMIT ?"
	return db.getBlocks(query, offset, limit)
}

// getTransactions get array of transactions from database with provided query
func (db *DB) getTransactions(query string, params interface{}, offset, limit uint64) (*block.Transactions, uint64) {
	limit = min(maxTransactions, limit)
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	rows, err := db.conn.Query(query, params, offset, limit)
	checkErr(err)
	defer rows.Close()
	var transactions block.Transactions
	var tr block.Transaction
	var id uint64
	var height int
	for rows.Next() {
		err := rows.Scan(&id, &height, &tr.From, &tr.To, &tr.Token, &tr.Fee, &tr.ValidVDFValue, &tr.Signature)
		checkErr(err)
		transactions.Ts = append(transactions.Ts, tr)
	}
	err = rows.Err()
	checkErr(err)
	return &transactions, id
}

// GetTxFromBlockByHeight returns transactions associated with block given by height
// including transaction with given offset
func (db *DB) GetTxFromBlockByHeight(height uint64, offset, limit uint64) (*block.Transactions, uint64) {
	query := "SELECT * FROM transactions WHERE Height = ? AND id <= ? ORDER BY id DESC LIMIT ?"
	return db.getTransactions(query, height, offset, limit)
}

// GetTransactionsFrom gets transactions from DB sent from given address
// including transaction with given offset
func (db *DB) GetTransactionsFrom(from []byte, offset, limit uint64) (*block.Transactions, uint64) {
	query := "SELECT * FROM transactions WHERE [From] = ? AND id <= ? ORDER BY id DESC LIMIT ?"
	return db.getTransactions(query, from, offset, limit)
}

// GetTransactionsTo gets transactions from DB received by given address
// including transaction with given offset
func (db *DB) GetTransactionsTo(to []byte, offset, limit uint64) (*block.Transactions, uint64) {
	query := "SELECT * FROM transactions WHERE [To] = ? AND id <= ? ORDER BY id DESC LIMIT ?"
	return db.getTransactions(query, to, offset, limit)
}

// GetAccountTransactions gets abstract of given account from database
// including transaction with given offset
func (db *DB) GetAccountTransactions(account []byte, offset, limit uint64) (*block.Transactions, uint64) {
	query := "SELECT * FROM transactions WHERE ? IN ([From], [To]) AND id <= ? ORDER BY id DESC LIMIT ?"
	return db.getTransactions(query, account, offset, limit)
}
