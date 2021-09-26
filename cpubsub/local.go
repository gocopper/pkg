package cpubsub

// LocalPubSub is an implementation of PubSub that only allows publishing
// of payloads to 'local' subscribers i.e. subscribers on the same instance
// as the publisher.
type LocalPubSub struct {
	subscriptions map[string][]Handler
}

// NewLocalPubSub creates a new LocalPubSub
func NewLocalPubSub() PubSub {
	return &LocalPubSub{
		subscriptions: make(map[string][]Handler),
	}
}

// Subscribe register the handler to the topic. When a payload is published on the topic, the handler will be called.
func (l *LocalPubSub) Subscribe(topic string, handler Handler) error {
	handlers, ok := l.subscriptions[topic]
	if !ok {
		handlers = make([]Handler, 0)
	}

	l.subscriptions[topic] = append(handlers, handler)

	return nil
}

// Publish publishes the payload on the given topic. All handlers subscribed to the topic will be called with the
// payload.
func (l *LocalPubSub) Publish(topic string, payload interface{}) error {
	handlers, ok := l.subscriptions[topic]
	if !ok {
		return nil
	}

	for _, h := range handlers {
		go h(payload)
	}

	return nil
}
