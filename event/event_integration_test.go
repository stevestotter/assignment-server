package event

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

var kafkaAddress = os.Getenv("KAFKA_URL")

func readMessages(topic string, numMessages int) (chan []byte, chan bool) {
	mChan := make(chan []byte)
	done := make(chan bool)

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaAddress},
		GroupID:  "buyer",
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	r.SetOffset(0)

	go func() {
		for i := 0; i < numMessages; i++ {
			m, err := r.ReadMessage(context.Background())
			if err != nil {
				fmt.Printf("Error on kafka read: %s\n", err)
			}
			fmt.Printf("message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
			mChan <- m.Value
		}
		r.Close()
		close(mChan)
		done <- true
		return
	}()

	return mChan, done
}

func TestIntegrationKafkaQueuePublishSendsMessageToKafka(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	expectedMessage := "hello"

	kq := &KafkaQueue{URL: kafkaAddress}
	err := kq.Publish([]byte(expectedMessage), TopicBuyerAssignment)

	assert.NoError(t, err)

	messageChan, done := readMessages(TopicBuyerAssignment, 1)
	m := <-messageChan

	assert.Equal(t, expectedMessage, string(m))

	// Wait for kafka close before finishing test
	<-done
}
