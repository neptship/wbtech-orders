package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type ProducerConfig struct {
	Brokers              []string
	Topic                string
	AllowAutoTopicCreate bool
	RequiredAcks         kafkago.RequiredAcks
	BatchBytes           int64
	BatchTimeout         time.Duration
	Async                bool
}

type Producer struct {
	w *kafkago.Writer
}

func NewProducer(cfg ProducerConfig) (*Producer, error) {
	if len(cfg.Brokers) == 0 || cfg.Topic == "" {
		return nil, errors.New("invalid producer config: brokers and topic are required")
	}
	if cfg.RequiredAcks == 0 {
		cfg.RequiredAcks = kafkago.RequireAll
	}
	if cfg.BatchTimeout == 0 {
		cfg.BatchTimeout = 10 * time.Millisecond
	}
	if cfg.BatchBytes == 0 {
		cfg.BatchBytes = 1 << 20
	}

	w := &kafkago.Writer{
		Addr:                   kafkago.TCP(cfg.Brokers...),
		Topic:                  cfg.Topic,
		Balancer:               &kafkago.LeastBytes{},
		RequiredAcks:           cfg.RequiredAcks,
		BatchTimeout:           cfg.BatchTimeout,
		BatchBytes:             cfg.BatchBytes,
		AllowAutoTopicCreation: cfg.AllowAutoTopicCreate,
		Async:                  cfg.Async,
	}
	return &Producer{w: w}, nil
}

func (p *Producer) Publish(ctx context.Context, key string, value []byte) error {
	msg := kafkago.Message{
		Key:   []byte(key),
		Value: value,
		Time:  time.Now(),
	}
	return p.w.WriteMessages(ctx, msg)
}

func (p *Producer) PublishJSON(ctx context.Context, key string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return p.Publish(ctx, key, b)
}

func (p *Producer) Close() error {
	return p.w.Close()
}
