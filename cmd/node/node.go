package main

import (
	"fmt"

	"github.com/Ansiblock/Ansiblock/pipelines"
	"github.com/Ansiblock/Ansiblock/replication"

	"github.com/Ansiblock/Ansiblock/log"
)

func main() {
	// CPU profiling by default
	// defer profile.Start().Stop()
	log.Init()
	runNode()
}

func runNode() {
	producer := replication.NewProducerNode("producer", "Zeus")
	producer.Data.Producer = producer.Data.Self
	log.Info(fmt.Sprintf("Producer transactions: %v\nProducer messages: %v\n", producer.Sockets.Transaction.LocalAddr().String(), producer.Sockets.Messages.LocalAddr().String()))
	go pipelines.ProducerNode(producer)
	go pipelines.SignerNode(producer, "Hera")
	go pipelines.ServerNode(producer, "Server")
	go pipelines.SignerNode(producer, "Athena")
	pipelines.SignerNode(producer, "Demeter")
}
