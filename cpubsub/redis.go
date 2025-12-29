package cpubsub

import (
	"context"
	"encoding/json"

	"github.com/gocopper/copper/cerrors"
	"github.com/gocopper/copper/clifecycle"
	"github.com/gocopper/copper/clogger"
	"github.com/redis/go-redis/v9"
)

// RedisPubSub is an implementation of PubSub that uses Redis for
// cross-instance pub/sub messaging.
type RedisPubSub struct {
	app    *clifecycle.Lifecycle
	client *redis.Client
	logger clogger.Logger
}

// NewRedisPubSub creates a new RedisPubSub
func NewRedisPubSub(app *clifecycle.Lifecycle, client *redis.Client, logger clogger.Logger) PubSub {
	return &RedisPubSub{
		app:    app,
		client: client,
		logger: logger,
	}
}

// Subscribe registers the handler to the topic. When a payload is published on the topic, the handler will be called.
func (r *RedisPubSub) Subscribe(ctx context.Context, topic string, handler Handler) error {
	pubsub := r.client.Subscribe(ctx, topic)

	r.app.Go(func(ctx context.Context) {
		ch := pubsub.Channel()

		for {
			select {
			case <-ctx.Done():
				_ = pubsub.Close()
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}

				var payload interface{}
				if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
					r.logger.WithTags(map[string]interface{}{
						"topic": topic,
					}).Error("[pubsub/redis] Failed to unmarshal payload", err)
					continue
				}

				if err := handler(ctx, payload); err != nil {
					r.logger.WithTags(map[string]interface{}{
						"topic": topic,
					}).Error("[pubsub/redis] Failed to run handler", err)
				}
			}
		}
	})

	return nil
}

// Publish publishes the payload on the given topic. All handlers subscribed to the topic will be called with the
// payload.
func (r *RedisPubSub) Publish(ctx context.Context, topic string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return cerrors.New(err, "failed to marshal payload", nil)
	}

	if err := r.client.Publish(ctx, topic, data).Err(); err != nil {
		return cerrors.New(err, "failed to publish to redis", map[string]interface{}{
			"topic": topic,
		})
	}

	return nil
}
