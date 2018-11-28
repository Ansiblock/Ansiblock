package api

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/Ansiblock/Ansiblock/block"
)

var blockchainAPI BlockchainAPI

// StatsModel is the data model of the Ansiblock blockchain stats. It is passed to the front end to display
type StatsModel struct {
	BlockHeight uint64
	TPS         uint64
	NodeCount   uint64
	BlockTime   uint64
}

// TransactionModel is the data model of the Ansiblock blockchain transaction.
// It is passed to the front end to display each transaction
type TransactionModel struct {
	From          string
	To            string
	Token         int64
	Fee           int64
	ValidVDFValue string
	Signature     string
}

// AccountModel is the data model of the Ansiblock blockchain Accounts.
// It is passed to the front end to display each account public key with its balance
type AccountModel struct {
	PublicKey string
	Balance   int64
}

// BlockModel is the data model of the Ansiblock blockchain blocks.
// It is passed to the front end to display each block
type BlockModel struct {
	Size         int
	VDF          string
	BlockHeight  uint64
	Transactions int32
}

// BlockDetailsModel is the data model of the Ansiblock blockchain detailed blocks.
// It is passed to the front end to display detailed info of the block
type BlockDetailsModel struct {
	Count         uint64
	VDF           string
	BlockHeight   uint64
	Taransactions []TransactionModel
}

// NodeDataModel is the data model of the Ansiblock blockchain nodes.
// It is passed to the front end to display info of the nodes in the blockchain
type NodeDataModel struct {
	Version  uint64
	Address  string
	NodeType string
	Name     string
}

// TransactinListModel stores transaction list and offset value, for gaping, to get next part of transactions
type TransactinListModel struct {
	Ts     []TransactionModel
	Offset uint64
}

// BlockListModel stores block list and offset value, for paging, to get next part of blocks
type BlockListModel struct {
	Blocks []BlockModel
	Offset uint64
}

//TODO: there is no proper error handlin for services

func extractOffsetAndLimit(offsetStr string, limitStr string) (uint64, uint64) {
	offset, err := strconv.ParseUint(offsetStr, 10, 64)
	if err != nil {
		offset = ^uint64(0) / 2
	}

	limit, err := strconv.ParseUint(limitStr, 10, 64)
	if err != nil {
		limit = 30
	}

	return offset, limit
}

func stats(c *gin.Context) {
	blockHeight := blockchainAPI.BlockHeight()
	tps := uint64(blockchainAPI.TPS())
	blockTime := uint64(blockchainAPI.BlockTime())
	nodeCount := uint64(len(blockchainAPI.Nodes()))
	stat := StatsModel{BlockHeight: blockHeight, TPS: tps, NodeCount: nodeCount, BlockTime: blockTime}
	c.JSON(http.StatusOK, stat)
}

func accounts(c *gin.Context) {
	if c.Request.Body == nil {
		c.JSON(http.StatusBadRequest, "")
		return
	}
	bodyBytes, _ := ioutil.ReadAll(c.Request.Body)
	var accountKeys []string
	json.Unmarshal(bodyBytes, &accountKeys)
	//if empty list is passed, return random accounts balances
	if len(accountKeys) == 0 {
		keys := blockchainAPI.RandomKeys(4)
		accountKeys = make([]string, len(keys)+1)
		accountKeys[0] = base64.StdEncoding.EncodeToString(blockchainAPI.MintKey())
		for i := range keys {
			accountKeys[i+1] = base64.StdEncoding.EncodeToString(keys[i])
		}
	}

	balances := blockchainAPI.Balances(accountKeys)

	resp := make([]AccountModel, len(balances))
	for i := 0; i < len(balances); i++ {
		resp[i].PublicKey = accountKeys[i]
		resp[i].Balance = balances[i]
	}

	c.JSON(http.StatusOK, resp)
}

func blocks(c *gin.Context) {
	offset, limit := extractOffsetAndLimit(c.Query("offset"), c.Query("limit"))
	blocks, resOffset := blockchainAPI.Blocks(offset, limit)
	blocksRes := make([]BlockModel, len(blocks))
	for i, b := range blocks {
		blocksRes[i].BlockHeight = b.Number
		blocksRes[i].Size = b.Size()
		blocksRes[i].VDF = base64.StdEncoding.EncodeToString(b.Val)
		blocksRes[i].Transactions = b.Transactions.Count()
	}

	resp := new(BlockListModel)
	resp.Blocks = blocksRes
	resp.Offset = resOffset
	c.JSON(http.StatusOK, resp)
}

func blockTransactions(c *gin.Context) {
	blockHeight := c.Query("blockHeight")
	height, err := strconv.ParseUint(blockHeight, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid parameter",
		})
		return
	}
	offset, limit := extractOffsetAndLimit(c.Query("offset"), c.Query("limit"))
	trans, resOffset := blockchainAPI.BlockTransactionsByHeight(height, offset, limit)
	resTransactions := make([]TransactionModel, len(trans.Ts))
	for i, tr := range trans.Ts {
		resTransactions[i].Fee = tr.Fee
		resTransactions[i].From = base64.StdEncoding.EncodeToString(tr.From)
		resTransactions[i].Signature = base64.StdEncoding.EncodeToString(tr.Signature)
		resTransactions[i].To = base64.StdEncoding.EncodeToString(tr.To)
		resTransactions[i].Token = tr.Token
		resTransactions[i].ValidVDFValue = base64.StdEncoding.EncodeToString(tr.ValidVDFValue)
	}

	resp := new(TransactinListModel)
	resp.Ts = resTransactions
	resp.Offset = resOffset
	c.JSON(http.StatusOK, resp)
}

func nodes(c *gin.Context) {
	nodesMap := blockchainAPI.Nodes()
	resp := make([]NodeDataModel, len(nodesMap))

	counter := 0
	for _, v := range nodesMap {
		resp[counter].Address = v.Addresses.Replication.IP.String()
		resp[counter].Version = v.Version
		resp[counter].NodeType = v.NodeType
		resp[counter].Name = v.NodeName
		counter++
	}

	c.JSON(http.StatusOK, resp)
}

func transactions(c *gin.Context) {
	offset, limit := extractOffsetAndLimit(c.Query("offset"), c.Query("limit"))
	key := c.Query("accountKey")
	trans, resOffset := blockchainAPI.AccountTransactions(key, offset, limit)
	resTransactions := make([]TransactionModel, len(trans.Ts))
	for i, tr := range trans.Ts {
		resTransactions[i].Fee = tr.Fee
		resTransactions[i].From = base64.StdEncoding.EncodeToString(tr.From)
		resTransactions[i].Signature = base64.StdEncoding.EncodeToString(tr.Signature)
		resTransactions[i].To = base64.StdEncoding.EncodeToString(tr.To)
		resTransactions[i].Token = tr.Token
		resTransactions[i].ValidVDFValue = base64.StdEncoding.EncodeToString(tr.ValidVDFValue)
	}

	resp := new(TransactinListModel)
	resp.Ts = resTransactions
	resp.Offset = resOffset
	c.JSON(http.StatusOK, resp)
}

func findBlock(c *gin.Context) {
	hash := strings.Replace(c.Query("blockHash"), " ", "+", -1)
	height := c.Query("blockHeight")
	var block *block.Block
	if hash != "" {
		block = blockchainAPI.BlockByHash(hash)
	} else if height != "" {
		h, err := strconv.ParseUint(height, 10, 64)
		if err == nil {
			block = blockchainAPI.BlockByHeight(h)
		}
	}

	if block == nil {
		c.JSON(http.StatusNotFound, "")
		return
	}
	model := new(BlockModel)
	model.BlockHeight = block.Number
	model.Size = block.Size()
	model.VDF = base64.StdEncoding.EncodeToString(block.Val)
	model.Transactions = block.Transactions.Count()
	c.JSON(http.StatusOK, model)
}

func findTransactions(c *gin.Context) {
	from := strings.Replace(c.Query("from"), " ", "+", -1)
	to := strings.Replace(c.Query("to"), " ", "+", -1)
	offset, limit := extractOffsetAndLimit(c.Query("offset"), c.Query("limit"))
	var trans *block.Transactions
	var resOffset uint64
	if from != "" {
		trans, resOffset = blockchainAPI.TransactionsFrom(from, offset, limit)
	} else if to != "" {
		trans, resOffset = blockchainAPI.TransactionsTo(to, offset, limit)
	}
	if trans == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid parameter",
		})
		return
	}

	resTransactions := make([]TransactionModel, len(trans.Ts))
	for i, tr := range trans.Ts {
		resTransactions[i].Fee = tr.Fee
		resTransactions[i].From = base64.StdEncoding.EncodeToString(tr.From)
		resTransactions[i].Signature = base64.StdEncoding.EncodeToString(tr.Signature)
		resTransactions[i].To = base64.StdEncoding.EncodeToString(tr.To)
		resTransactions[i].Token = tr.Token
		resTransactions[i].ValidVDFValue = base64.StdEncoding.EncodeToString(tr.ValidVDFValue)
	}
	resp := new(TransactinListModel)
	resp.Ts = resTransactions
	resp.Offset = resOffset

	c.JSON(http.StatusOK, resp)
}

func index(c *gin.Context) {
	c.HTML(
		// Set the HTTP status to 200 (OK)
		http.StatusOK,
		// Use the index.html template
		"index.html",
		// Pass the data that the page uses (in this case, 'title')
		gin.H{},
	)
}

func setupRouter() {
	router := gin.Default()

	router.Use(static.Serve("/", static.LocalFile("./views", true)))
	router.LoadHTMLGlob("views/index.html")

	router.GET("/api/stats", stats)
	router.POST("/api/accounts", accounts)
	router.GET("/api/blocks", blocks)
	router.GET("/api/blockTransactions", blockTransactions)
	router.GET("/api/nodes", nodes)
	router.GET("/api/transactions", transactions)
	router.GET("/api/findBlock", findBlock)
	router.GET("/api/findTransactions", findTransactions)

	router.GET("/", index)

	router.Run(":8080")
}

// RunRestAPI registers router and runs web server
func RunRestAPI(api BlockchainAPI) {
	blockchainAPI = api
	setupRouter()
}
