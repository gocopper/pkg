package cpubsub

import "context"

// Handler defines a function that is called when a payload is published on the subscribed topic
type Handler func(ctx context.Context, payload interface{}) error

// PubSub defines the methods for publishing payloads to a topic and subscribing to topics with a handler
type PubSub interface {
	Subscribe(ctx context.Context, topic string, handler Handler) error
	Publish(ctx context.Context, topic string, payload interface{}) error
}
