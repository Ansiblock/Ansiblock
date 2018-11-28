package main

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"

	"github.com/Ansiblock/Ansiblock/pipelines"
	"github.com/Ansiblock/Ansiblock/replication"
	isatty "github.com/mattn/go-isatty"

	"github.com/Ansiblock/Ansiblock/log"
)

func main() {
	// CPU profiling by default
	// defer profile.Start().Stop()
	log.Init()
	runSignerNode()
}

func parseProducerJSON() replication.Node {
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
		log.Fatal("Empty file, expected json")
	}
	var objMap map[string]*json.RawMessage
	err := json.Unmarshal([]byte(data), &objMap)
	checkErr(err)

	var producer replication.Node

	err = json.Unmarshal(*objMap["Data"], &producer.Data)
	checkErr(err)

	return producer
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}
func runSignerNode() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: cat producer.json | go run signer.go name")
	}
	producer := parseProducerJSON()
	name := os.Args[1]
	if name == "" {
		log.Fatal("Illegal format of name ")
	}
	pipelines.SignerNode(producer, name)
}
