package main

import (
	"fmt"
	"log"

	"stevestotter/assignment-server/api"
	"stevestotter/assignment-server/assignment"
	"stevestotter/assignment-server/config"
	"stevestotter/assignment-server/event"
)

func main() {
	fmt.Println("Started")

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error processing env config: %s", err)
	}

	queue := &event.KafkaQueue{URL: cfg.Kafka.URL}

	generator := assignment.Generator{
		MessageQueue:        queue,
		PercentageChangeMin: cfg.Generator.PercentageChangeMin,
		PercentageChangeMax: cfg.Generator.PercentageChangeMax,
	}

	a := api.API{Port: cfg.API.Port, AssignmentSubmitter: &generator}

	err = a.Start()
	if err != nil {
		log.Fatalf("Couldn't start API server: %s", err)
	}

	err = generator.GenerateFromTrades()
	if err != nil {
		log.Fatalf("Couldn't start generator for trades: %s", err)
	}
}
