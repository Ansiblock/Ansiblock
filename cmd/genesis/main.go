package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	isatty "github.com/mattn/go-isatty"
	"github.com/Ansiblock/Ansiblock/mint"
)

// This program take mint block as json file at stdin.
// It generates genesis blocks and writes to stdout
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

	var m mint.Mint
	err := json.Unmarshal([]byte(data), &m)
	check(err)

	for _, b := range m.CreateBlocks() {
		j, err := json.Marshal(b)
		check(err)
		fmt.Println(j)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
