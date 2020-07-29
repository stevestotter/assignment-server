package event

//go:generate go run -mod=mod github.com/golang/mock/mockgen --source=event.go --destination=../mocks/event/event.go

type Publisher interface {
	Publish(message []byte) error
}

type Listener interface {
	Subscribe(topic string) (<-chan []byte, error)
}

type KafkaEventQueue struct {
	URL string
}

// TODO: Implement this
func (k KafkaEventQueue) Publish(message []byte) error {
	return nil
}
