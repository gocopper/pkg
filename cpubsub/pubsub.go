package cpubsub

// Handler defines a function that is called when a payload is published on the subscribed topic
type Handler func(payload interface{})

// PubSub defines the methods for publishing payloads to a topic and subscribing to topics with a handler
type PubSub interface {
	Subscribe(topic string, handler Handler) error
	Publish(topic string, payload interface{}) error
}
