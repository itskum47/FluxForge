package streaming

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

type LogPublisher struct {
	logger *log.Logger
}

func NewLogPublisher() *LogPublisher {
	return &LogPublisher{
		logger: log.Default(),
	}
}

func (p *LogPublisher) Publish(ctx context.Context, topic string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	event := Event{
		ID:        "log-id-stub", // In real impl, generate UUID
		Topic:     topic,
		Payload:   data,
		Timestamp: time.Now(),
		Source:    "control-plane",
	}

	eventBytes, _ := json.Marshal(event)
	p.logger.Printf("[STREAMING] PUBLISH %s: %s", topic, string(eventBytes))
	return nil
}

func (p *LogPublisher) Close() error {
	p.logger.Println("[STREAMING] Closed LogPublisher")
	return nil
}
