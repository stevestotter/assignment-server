package main

import (
	"context"
	"fmt"

	"stevestotter/assignment-server/api"
	"stevestotter/assignment-server/config"

	"github.com/google/uuid"
	kafka "github.com/segmentio/kafka-go"
)

const topic string = "buyer-trade"

func main() {
	fmt.Println("Started")

	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Printf("Error processing env config: %s", cfg)
	}

	// TODO: Frontend to add assignment
	// TODO: Tests for frontend

	// TODO: API for publishing assignments
	go api.Start()

	// TODO: Listen to buy & sell topics
	// TODO: Publish assignments based on listens

	go produce(cfg.Kafka)

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{cfg.Kafka.URL},
		GroupID:  "buyer",
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	r.SetOffset(0)

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			fmt.Printf("Error on kafka read: %s\n", err)
			break
		}
		fmt.Printf("message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
	}

	r.Close()
}

func produce(cfg config.Kafka) {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{cfg.URL},
		Topic:    topic,
		Balancer: &kafka.Hash{},
	})

	err := w.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(uuid.New().String()),
			Value: []byte("One!"),
		},
		kafka.Message{
			Key:   []byte(uuid.New().String()),
			Value: []byte("Two!"),
		},
		kafka.Message{
			Key:   []byte(uuid.New().String()),
			Value: []byte("Three!"),
		},
	)

	if err != nil {
		fmt.Printf("Error on kafka write: %s", err)
	} else {
		fmt.Println("Finished publishing")
	}

	w.Close()
}
