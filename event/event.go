package event

type Publisher interface {
	Publish(message []byte) error
}

type Listener interface {
	Subscribe(topic string) (<-chan []byte, error)
}
