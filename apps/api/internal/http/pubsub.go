package http

import (
	"context"
	"errors"
	"os"

	redis "github.com/redis/go-redis/v9"
)

// PubSub defines the minimal interface used by SSE for subscription.
type PubSub interface {
	Subscribe(ctx context.Context, channel string) (<-chan []byte, func() error, error)
}

// NewDefaultPubSubFromEnv creates a Redis-backed PubSub if REDIS_HOST and REDIS_PORT are set.
func NewDefaultPubSubFromEnv() (PubSub, error) {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		return nil, errors.New("REDIS_HOST not set")
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	addr := host + ":" + port
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	return &redisPubSub{rdb: rdb}, nil
}

type redisPubSub struct {
	rdb *redis.Client
}

func (r *redisPubSub) Subscribe(ctx context.Context, channel string) (<-chan []byte, func() error, error) {
	sub := r.rdb.Subscribe(ctx, channel)
	ch := make(chan []byte)

	// Start a goroutine to forward messages
	go func() {
		defer close(ch)
		for msg := range sub.Channel() {
			select {
			case ch <- []byte(msg.Payload):
			case <-ctx.Done():
				return
			}
		}
	}()

	unsubscribe := func() error { return sub.Unsubscribe(context.Background(), channel) }
	return ch, unsubscribe, nil
}
