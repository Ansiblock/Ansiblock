package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/network"
	"github.com/Ansiblock/Ansiblock/replication"

	"github.com/gin-gonic/gin"
	"github.com/Ansiblock/Ansiblock/books"
)

func TestStats(t *testing.T) {
	apiMock := NewBlockchainAPIMock()
	apiMock.BlockHeightVal = 1092
	apiMock.TPSVal = 100001
	apiMock.BlockTimeVal = 879

	blockchainAPI = apiMock

	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.GET("/api/stats", stats)
	request, err := http.NewRequest(http.MethodGet, "/api/stats", nil)
	if err != nil {
		t.Fatalf("Couldn’t create request: %v\n", err)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("/api/stats failed with error code %v.", response.Code)
	}

	var stats StatsModel
	json.Unmarshal(response.Body.Bytes(), &stats)
	if stats.BlockHeight != 1092 || stats.TPS != 100001 || stats.BlockTime != 879 {
		t.Errorf("/api/stats returned wrong data: %v.", response.Body.String())
	}
}

func TestAccounts(t *testing.T) {
	apiMock := NewBlockchainAPIMock()
	numAccounts := 10
	apiMock.BalanceValues = make([]int64, numAccounts)
	accountKeys := make([]string, numAccounts)
	for i := 0; i < numAccounts; i++ {
		accountKeys[i] = strconv.Itoa(i)
		apiMock.BalanceValues[i] = int64(1000 * i)
	}

	blockchainAPI = apiMock

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/api/accounts", accounts)
	keysJSON, err := json.Marshal(accountKeys)
	body := bytes.NewBuffer(keysJSON)
	request, err := http.NewRequest(http.MethodGet, "/api/accounts", body)
	if err != nil {
		t.Fatalf("Couldn’t create request: %v\n", err)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("/api/accounts failed with error code %v.", response.Code)
	}

	var accountsList []AccountModel
	json.Unmarshal(response.Body.Bytes(), &accountsList)
	for i, acc := range accountsList {
		if acc.Balance != int64(i*1000) || acc.PublicKey != strconv.Itoa(i) {
			t.Errorf("/api/accounts returned wrong data. balance=%v, key=%v", acc.Balance, acc.PublicKey)
			return
		}
	}
}

func TestRestAPIBlocks(t *testing.T) {
	apiMock := NewBlockchainAPIMock()
	numBlocks := 98
	am, keyPairs := books.RandomAccounts(10)
	apiMock.BlocksList = books.RandomTransactionsBlocks(am, 10, numBlocks, keyPairs)

	blockchainAPI = apiMock

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/api/blocks", blocks)
	request, err := http.NewRequest(http.MethodGet, "/api/blocks?offset=1&limit=10", nil)
	if err != nil {
		t.Fatalf("Couldn’t create request: %v\n", err)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("/api/blocks failed with error code %v.", response.Code)
	}

	if apiMock.QueryParams["offset"] != "1" || apiMock.QueryParams["limit"] != "10" {
		t.Errorf("/api/blocks failed reading parameters, offset: %v, len: %v", apiMock.QueryParams["offset"], apiMock.QueryParams["limit"])
	}

	var blockList BlockListModel
	json.Unmarshal(response.Body.Bytes(), &blockList)
	blocks := blockList.Blocks
	if len(blocks) != len(apiMock.BlocksList) {
		t.Errorf("/api/blocks returned wrong data. block list size is %v, instead %v.", len(blocks), len(apiMock.BlocksList))
	}
	for i, b := range blocks {
		if b.BlockHeight != apiMock.BlocksList[i].Number || b.Size != apiMock.BlocksList[i].Size() || b.VDF != base64.StdEncoding.EncodeToString(apiMock.BlocksList[i].Val) {
			t.Errorf("/api/blocks returned wrong data. number=%v, size=%v, vdf=%v", b.BlockHeight, b.Size, b.VDF)
			return
		}
	}
}

func TestRestAPIBlocks2(t *testing.T) {
	apiMock := NewBlockchainAPIMock()
	numBlocks := 98
	am, keyPairs := books.RandomAccounts(10)
	apiMock.BlocksList = books.RandomTransactionsBlocks(am, 10, numBlocks, keyPairs)

	blockchainAPI = apiMock

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/api/blocks", blocks)
	request, err := http.NewRequest(http.MethodGet, "/api/blocks", nil)
	if err != nil {
		t.Fatalf("Couldn’t create request: %v\n", err)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("/api/blocks failed with error code %v.", response.Code)
	}

	if apiMock.QueryParams["offset"] != "9223372036854775807" || apiMock.QueryParams["limit"] != "30" {
		t.Errorf("/api/blocks failed reading parameters, offset: %v, len: %v", apiMock.QueryParams["offset"], apiMock.QueryParams["limit"])
	}

	var blockList BlockListModel
	json.Unmarshal(response.Body.Bytes(), &blockList)
	blocks := blockList.Blocks
	if len(blocks) != len(apiMock.BlocksList) {
		t.Errorf("/api/blocks returned wrong data. block list size is %v, instead %v.", len(blocks), len(apiMock.BlocksList))
	}
	for i, b := range blocks {
		if b.BlockHeight != apiMock.BlocksList[i].Number || b.Size != apiMock.BlocksList[i].Size() || b.VDF != base64.StdEncoding.EncodeToString(apiMock.BlocksList[i].Val) {
			t.Errorf("/api/blocks returned wrong data. number=%v, size=%v, vdf=%v", b.BlockHeight, b.Size, b.VDF)
			return
		}
	}
}

func TestBlockTransactions(t *testing.T) {
	apiMock := NewBlockchainAPIMock()
	trans := block.CreateRealTransactions(198)
	apiMock.BlockTransactions = &trans

	blockchainAPI = apiMock

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/api/blockTransactions", blockTransactions)
	request, err := http.NewRequest(http.MethodGet, "/api/blockTransactions?blockHeight=10&offset=10&limit=11", nil)
	if err != nil {
		t.Fatalf("Couldn’t create request: %v\n", err)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("/api/blockTransactions failed with error code %v.", response.Code)
	}

	if apiMock.QueryParams["offset"] != "10" || apiMock.QueryParams["limit"] != "11" || apiMock.QueryParams["blockHeight"] != "10" {
		t.Errorf("/api/blocks failed reading parameters, offset: %v, len: %v, height %v", apiMock.QueryParams["offset"], apiMock.QueryParams["limit"], apiMock.QueryParams["blockHeight"])
	}

	var blockTransactions []TransactionModel
	json.Unmarshal(response.Body.Bytes(), &blockTransactions)
	for i, tr := range blockTransactions {
		if tr.From != base64.StdEncoding.EncodeToString(trans.Ts[i].From) || tr.To != base64.StdEncoding.EncodeToString(trans.Ts[i].To) ||
			tr.Token != trans.Ts[i].Token || tr.Signature != base64.StdEncoding.EncodeToString(trans.Ts[i].Signature) {
			t.Errorf("/api/blockTransactions returned wrong data. number=%v, size=%v, vdf=%v", tr.From, tr.To, tr.Token)
			return
		}
	}
}

func TestTransactions(t *testing.T) {
	apiMock := NewBlockchainAPIMock()
	trans := block.CreateRealTransactions(194)
	apiMock.AccTransactions = &trans

	blockchainAPI = apiMock

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/api/transactions", transactions)
	request, err := http.NewRequest(http.MethodGet, "/api/transactions?accountKey=aAb&offset=10&limit=11", nil)
	if err != nil {
		t.Fatalf("Couldn’t create request: %v\n", err)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("/api/transactions failed with error code %v.", response.Code)
	}

	var transactionList TransactinListModel
	json.Unmarshal(response.Body.Bytes(), &transactionList)
	accTransactions := transactionList.Ts
	if len(accTransactions) != len(trans.Ts) {
		t.Errorf("/api/transactions returned wrong data. transactions number should be %v, instead %v", len(trans.Ts), len(accTransactions))
	}

	if apiMock.QueryParams["offset"] != "10" || apiMock.QueryParams["limit"] != "11" || apiMock.QueryParams["accountKey"] != "aAb" {
		t.Errorf("/api/blocks failed reading parameters, offset: %v, len: %v, height %v", apiMock.QueryParams["offset"], apiMock.QueryParams["limit"], apiMock.QueryParams["blockHeight"])
	}

	for i, tr := range accTransactions {
		if tr.From != base64.StdEncoding.EncodeToString(trans.Ts[i].From) || tr.To != base64.StdEncoding.EncodeToString(trans.Ts[i].To) ||
			tr.Token != trans.Ts[i].Token || tr.Signature != base64.StdEncoding.EncodeToString(trans.Ts[i].Signature) {
			t.Errorf("/api/transactions returned wrong data. number=%v, size=%v, vdf=%v", tr.From, tr.To, tr.Token)
			return
		}
	}
}

func TestNodes(t *testing.T) {
	apiMock := NewBlockchainAPIMock()

	self := block.NewKeyPair().Public
	testnode := replication.NewNodeData(self, "producer", "test", network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP, network.BlockAddrUserUDP)
	testnodes := make(map[string]*replication.NodeData)
	testnodes["test1"] = testnode
	apiMock.NodesMap = testnodes
	blockchainAPI = apiMock

	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.GET("/api/nodes", nodes)
	request, err := http.NewRequest(http.MethodGet, "/api/nodes", nil)
	if err != nil {
		t.Fatalf("Couldn’t create request: %v\n", err)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("/api/nodes failed with error code %v.", response.Code)
	}

	var resNodes []NodeDataModel
	json.Unmarshal(response.Body.Bytes(), &resNodes)
	if len(resNodes) != 1 || resNodes[0].NodeType != "producer" || resNodes[0].Name != "test" ||
		resNodes[0].Address != network.BlockAddrUserUDP.IP.String() || resNodes[0].Version != 0 {
		t.Errorf("/api/nodes returned wrong data: %v.", response.Body.String())
	}
}

func TestFindBlockByHash(t *testing.T) {
	apiMock := NewBlockchainAPIMock()
	am, keyPairs := books.RandomAccounts(10)
	apiMock.BlocksList = books.RandomTransactionsBlocks(am, 10, 1, keyPairs)
	apiMock.Block = &apiMock.BlocksList[0]

	blockchainAPI = apiMock

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/api/findBlock", findBlock)
	request, err := http.NewRequest(http.MethodGet, "/api/findBlock?blockHash=Qm9V6rlU6DLE7EQAg3zQ6pPLJX4gWCWunBiB6+c+p1U=", nil)
	if err != nil {
		t.Fatalf("Couldn’t create request: %v\n", err)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("/api/findBlock failed with error code %v.", response.Code)
	}

	var resBlock BlockModel
	json.Unmarshal(response.Body.Bytes(), &resBlock)
	if apiMock.QueryParams["blockHash"] != "Qm9V6rlU6DLE7EQAg3zQ6pPLJX4gWCWunBiB6+c+p1U=" || resBlock.BlockHeight != apiMock.Block.Number ||
		resBlock.Size != apiMock.Block.Size() || resBlock.VDF != base64.StdEncoding.EncodeToString(apiMock.Block.Val) {
		t.Errorf("/api/findBlock returned wrong data: %v.", response.Body.String())
	}
}

func TestFindBlockByHeight(t *testing.T) {
	apiMock := NewBlockchainAPIMock()
	am, keyPairs := books.RandomAccounts(10)
	apiMock.BlocksList = books.RandomTransactionsBlocks(am, 10, 1, keyPairs)
	apiMock.Block = &apiMock.BlocksList[0]

	blockchainAPI = apiMock

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/api/findBlock", findBlock)
	request, err := http.NewRequest(http.MethodGet, "/api/findBlock?blockHeight=1098", nil)
	if err != nil {
		t.Fatalf("Couldn’t create request: %v\n", err)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("/api/findBlock failed with error code %v.", response.Code)
	}

	var resBlock BlockModel
	json.Unmarshal(response.Body.Bytes(), &resBlock)
	if apiMock.QueryParams["blockHeight"] != "1098" || resBlock.BlockHeight != apiMock.Block.Number ||
		resBlock.Size != apiMock.Block.Size() || resBlock.VDF != base64.StdEncoding.EncodeToString(apiMock.Block.Val) {
		t.Errorf("/api/findBlock returned wrong data: %v.", response.Body.String())
	}
}

func TestFindTransactionsFrom(t *testing.T) {
	apiMock := NewBlockchainAPIMock()
	trans := block.CreateRealTransactions(194)
	apiMock.AccTransactions = &trans

	blockchainAPI = apiMock

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/api/findTransactions", findTransactions)
	request, err := http.NewRequest(http.MethodGet, "/api/findTransactions?from=fromSomeAccountKey&offset=101&limit=90", nil)
	if err != nil {
		t.Fatalf("Couldn’t create request: %v\n", err)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("/api/findTransactions failed with error code %v.", response.Code)
	}

	var transactionList TransactinListModel
	json.Unmarshal(response.Body.Bytes(), &transactionList)
	if apiMock.QueryParams["from"] != "fromSomeAccountKey" || apiMock.QueryParams["offset"] != "101" || apiMock.QueryParams["limit"] != "90" {
		t.Errorf("/api/findTransactions parameter parsing error!")
	}

	accTransactions := transactionList.Ts
	if len(accTransactions) != len(trans.Ts) {
		t.Errorf("/api/transactions returned wrong data. transactions number should be %v, instead %v", len(trans.Ts), len(accTransactions))
	}

	for i, tr := range accTransactions {
		if tr.From != base64.StdEncoding.EncodeToString(trans.Ts[i].From) || tr.To != base64.StdEncoding.EncodeToString(trans.Ts[i].To) ||
			tr.Token != trans.Ts[i].Token || tr.Signature != base64.StdEncoding.EncodeToString(trans.Ts[i].Signature) {
			t.Errorf("/api/transactions returned wrong data. number=%v, size=%v, vdf=%v", tr.From, tr.To, tr.Token)
			return
		}
	}
}

func TestFindTransactionsTo(t *testing.T) {
	apiMock := NewBlockchainAPIMock()
	trans := block.CreateRealTransactions(194)
	apiMock.AccTransactions = &trans

	blockchainAPI = apiMock

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/api/findTransactions", findTransactions)
	request, err := http.NewRequest(http.MethodGet, "/api/findTransactions?to=toSomeAccountKey&offset=101&limit=90", nil)
	if err != nil {
		t.Fatalf("Couldn’t create request: %v\n", err)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Errorf("/api/findTransactions failed with error code %v.", response.Code)
	}

	var transactionList TransactinListModel
	json.Unmarshal(response.Body.Bytes(), &transactionList)
	if apiMock.QueryParams["to"] != "toSomeAccountKey" || apiMock.QueryParams["offset"] != "101" || apiMock.QueryParams["limit"] != "90" {
		t.Errorf("/api/findTransactions parameter parsing error!")
	}

	accTransactions := transactionList.Ts
	if len(accTransactions) != len(trans.Ts) {
		t.Errorf("/api/transactions returned wrong data. transactions number should be %v, instead %v", len(trans.Ts), len(accTransactions))
	}

	for i, tr := range accTransactions {
		if tr.From != base64.StdEncoding.EncodeToString(trans.Ts[i].From) || tr.To != base64.StdEncoding.EncodeToString(trans.Ts[i].To) ||
			tr.Token != trans.Ts[i].Token || tr.Signature != base64.StdEncoding.EncodeToString(trans.Ts[i].Signature) {
			t.Errorf("/api/transactions returned wrong data. number=%v, size=%v, vdf=%v", tr.From, tr.To, tr.Token)
			return
		}
	}
}
