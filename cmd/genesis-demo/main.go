// genesis-demo accepts mint json file from stdin and generates numAccounts
// new accounts. It to each account it transfers tokensPerUser amount
// go run cmd/genesis-demo/main.go < mint-demo.json > genesis-demo.log

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	isatty "github.com/mattn/go-isatty"
	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/mint"
)

// type MintDemo struct {
// 	Mint        mint.Mint
// 	NumAccounts int64
// }

const maxBlockVals int = 4096
const numAccounts = 10000
const tokensPerUser = 1000

func main() {
	if isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		log.Fatal("Expected json file as stdin")
	}
	input := bufio.NewScanner(os.Stdin)
	var data string

	for input.Scan() {
		data = data + input.Text()
	}

	data = strings.TrimSpace(data)
	if len(data) == 0 {
		fmt.Println("Expected json file, got nothing")
		os.Exit(1)
	}

	var demoMint mint.Mint
	err := json.Unmarshal([]byte(data), &demoMint)
	check(err)

	keyPairs := block.KeyPairs(numAccounts)
	mintKeyPair := demoMint.KeyPair
	// mint initial blocks
	blks := demoMint.CreateBlocks()
	validVDFValue := blks[1].Val

	for _, b := range blks {
		j, err := json.Marshal(b)
		check(err)
		fmt.Printf(string(j))
	}

	var validVDFValues []block.VDFValue
	lastNumber := uint64(0)
	for i := 0; i < maxBlockVals; i++ {
		b := block.NextBlock(validVDFValue, lastNumber, 1, &block.Transactions{})
		lastNumber++
		validVDFValue = b.Val
		validVDFValues = append(validVDFValues, validVDFValue)
		serializeAndPrint(b)
	}

	var transactions block.Transactions
	for i := 0; i < numAccounts; i++ {
		tx := block.NewTransaction(&mintKeyPair, keyPairs[i].Public, tokensPerUser, 0, validVDFValue)
		transactions.Ts = append(transactions.Ts, tx)
	}

	b := block.New(validVDFValue, lastNumber, 0, &transactions)
	serializeAndPrint(b)
}

func serializeAndPrint(b interface{}) {
	j, err := json.Marshal(b)
	check(err)
	fmt.Println(string(j))
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
