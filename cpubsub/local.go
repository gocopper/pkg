package cpubsub

import (
	"context"

	"github.com/gocopper/copper/clogger"
)

// LocalPubSub is an implementation of PubSub that only allows publishing
// of payloads to 'local' subscribers i.e. subscribers on the same instance
// as the publisher.
type LocalPubSub struct {
	subscriptions map[string][]Handler
	logger        clogger.Logger
}

// NewLocalPubSub creates a new LocalPubSub
func NewLocalPubSub(logger clogger.Logger) PubSub {
	return &LocalPubSub{
		subscriptions: make(map[string][]Handler),
		logger:        logger,
	}
}

// Subscribe register the handler to the topic. When a payload is published on the topic, the handler will be called.
func (l *LocalPubSub) Subscribe(ctx context.Context, topic string, handler Handler) error {
	handlers, ok := l.subscriptions[topic]
	if !ok {
		handlers = make([]Handler, 0)
	}

	l.subscriptions[topic] = append(handlers, handler)

	return nil
}

// Publish publishes the payload on the given topic. All handlers subscribed to the topic will be called with the
// payload.
func (l *LocalPubSub) Publish(ctx context.Context, topic string, payload interface{}) error {
	handlers, ok := l.subscriptions[topic]
	if !ok {
		return nil
	}

	for i := range handlers {
		go func(i int) {
			err := handlers[i](ctx, payload)
			if err != nil {
				l.logger.WithTags(map[string]interface{}{
					"topic": topic,
				}).Error("Failed to run pubsub handler", err)
			}
		}(i)
	}

	return nil
}
