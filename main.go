package main

import (
	"fmt"
	"log"

	"stevestotter/assignment-server/api"
	"stevestotter/assignment-server/config"
	"stevestotter/assignment-server/event"
)

const topic string = "buyer-trade"

func main() {
	fmt.Println("Started")

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error processing env config: %s", cfg)
	}

	// TODO: Frontend to add assignment
	// TODO: Tests for frontend

	queue := &event.KafkaQueue{URL: cfg.Kafka.URL}

	a := api.API{Port: cfg.API.Port, MessageQueue: queue}

	err = a.Start()
	if err != nil {
		log.Fatalf("Couldn't start API server: %s", err)
	}

	// TODO: remove when blocking call is made
	for {
	}

	// TODO: Listen to buy & sell topics
	// TODO: Publish assignments based on listens
}
