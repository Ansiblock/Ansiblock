package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Ansiblock/Ansiblock/pipelines"
	"github.com/Ansiblock/Ansiblock/replication"
	"go.uber.org/zap"

	"github.com/Ansiblock/Ansiblock/log"
)

func main() {
	// CPU profiling by default
	// defer profile.Start().Stop()
	log.Init()
	runProducerNode()
}

func outputProducer(producer replication.Node) {
	lData, err := json.Marshal(producer)
	if err != nil {
		log.Error(fmt.Sprintf("Can't marshal block %v", producer))
	}
	f, err := os.Create("producer.json")
	if err != nil {
		log.Error("Error while opening file: ", zap.Error(err))
	}

	defer f.Close()

	if _, err = f.Write(lData); err != nil {
		log.Error("Error while writing to file: ", zap.Error(err))
	}
}

func runProducerNode() {
	producer := replication.NewProducerNode("producer", "Zeus")
	producer.Data.Producer = producer.Data.Self
	outputProducer(producer)
	pipelines.ProducerNode(producer)
}
