package event

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

//go:generate go run -mod=mod github.com/golang/mock/mockgen --source=event.go --destination=../mocks/event/event.go

const (
	TopicBuyerTrade       string = "buyer-trade"
	TopicSellerTrade      string = "seller-trade"
	TopicBuyerAssignment  string = "buyer-assignment"
	TopicSellerAssignment string = "seller-assignment"
)

type Publisher interface {
	Publish(message []byte, topic string) error
}

type Listener interface {
	Subscribe(topic string) (<-chan []byte, error)
}

type KafkaQueue struct {
	URL string
}

func (k *KafkaQueue) Publish(message []byte, topic string) error {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{k.URL},
		Topic:    topic,
		Balancer: &kafka.Hash{},
	})
	defer w.Close()

	err := w.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(uuid.New().String()),
			Value: message,
		},
	)

	if err != nil {
		// TODO: Change logger
		log.Printf("Error on kafka write: %s", err)
		return err
	}

	return nil
}
