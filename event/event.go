package event

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

//go:generate go run -mod=mod github.com/golang/mock/mockgen --build_flags=-mod=mod --source=event.go --destination=../mocks/event/event.go

const (
	TopicBuyerTrade       string = "buyer-trade"
	TopicSellerTrade      string = "seller-trade"
	TopicBuyerAssignment  string = "buyer-assignment"
	TopicSellerAssignment string = "seller-assignment"

	GroupBuyer  string = "buyer"
	GroupSeller string = "seller"
)

type Trade struct {
	AssignmentID int    `json:"assignmentId"`
	Price        string `json:"price"`
	Quantity     string `json:"quantity"`
}

type ListenPublisher interface {
	Listener
	Publisher
}

type Publisher interface {
	Publish(message []byte, topic string) error
}

type Listener interface {
	Subscribe(topic string, group string) (<-chan []byte, error)
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

func (k *KafkaQueue) Subscribe(topic string, group string) (<-chan []byte, error) {
	mChan := make(chan []byte)

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{k.URL},
		GroupID:  group,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	r.SetOffset(0)

	go func() {
		defer func() {
			r.Close()
			close(mChan)
		}()

		for {
			m, err := r.ReadMessage(context.Background())
			if err != nil {
				// TODO: Change logger
				log.Printf("Error on kafka read: %s\n", err)
				continue
			}
			mChan <- m.Value
		}
	}()

	return mChan, nil

}
