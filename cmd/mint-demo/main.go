package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Ansiblock/Ansiblock/mint"
)

// Creates a new mint and whites to stdout in JSON format
// Output of this program should be fed to genesis-demo program
func main() {
	if len(os.Args) != 2 {
		fmt.Println("you need to pass number of tokens as an argument")
		os.Exit(1)
	}

	numTokens, err := strconv.Atoi(os.Args[1])
	check(err)
	newmint := mint.NewMint(int64(numTokens))

	data, err := json.Marshal(newmint)
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}
	fmt.Printf("%s", data)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
