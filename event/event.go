package event

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

//go:generate go run -mod=mod github.com/golang/mock/mockgen --build_flags=-mod=mod --source=event.go --destination=../mocks/event/event.go

const (
	// TopicBuyerTrade is the queue topic for new buyer trades in the market
	TopicBuyerTrade string = "buyer-trade"
	// TopicSellerTrade is the queue topic for new seller trades in the market
	TopicSellerTrade string = "seller-trade"
	// TopicBuyerAssignment is the queue topic for new buyer assignments in the market
	TopicBuyerAssignment string = "buyer-assignment"
	// TopicSellerAssignment is the queue topic for new seller assignments in the market
	TopicSellerAssignment string = "seller-assignment"

	// GroupBuyer is the queue group for buyers in the market
	GroupBuyer string = "buyer"
	// GroupSeller is the queue group for sellers in the market
	GroupSeller string = "seller"
)

var (
	//ErrQueueWrite is an error thrown on write to the event queue
	ErrQueueWrite error = errors.New("Error on kafka write")
)

// Trade contains information about a trade in the market
type Trade struct {
	AssignmentID int    `json:"assignmentId"`
	Price        string `json:"price"`
	Quantity     string `json:"quantity"`
}

// ListenPublisher has the responsibility of both reading and publishing to
// an event queue
type ListenPublisher interface {
	Listener
	Publisher
}

// Publisher is able to send messages to an event queue
type Publisher interface {
	Publish(message []byte, topic string) error
}

// Listener is able to listen for messages on a topic on the event queue
// as part of a group. Being part of a group means two listeners of the
// same group don't both receive the same message, and instead consume
// messages on the topic as a team.
type Listener interface {
	Subscribe(topic string, group string) (<-chan []byte, error)
}

// KafkaQueue is a Kafka Event Queue that conforms to ListenPublisher
type KafkaQueue struct {
	URL string
}

// Publish sends a message to the kafka queue
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
		return ErrQueueWrite
	}

	return nil
}

// Subscribe listens for messages on the kafka queue
func (k *KafkaQueue) Subscribe(topic string, group string) (<-chan []byte, error) {
	mChan := make(chan []byte)

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{k.URL},
		GroupID:  group,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

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
