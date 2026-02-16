package streaming

import (
	"context"
	"time"
)

type Event struct {
	ID        string    `json:"id"`
	Topic     string    `json:"topic"`
	Payload   []byte    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

type Publisher interface {
	Publish(ctx context.Context, topic string, payload interface{}) error
	Close() error
}

type Subscriber interface {
	Subscribe(topic string, handler func(event Event)) (Subscription, error)
}

type Subscription interface {
	Unsubscribe() error
}
