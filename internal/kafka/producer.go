package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type ProducerConfig struct {
	Brokers []string
	Topic   string
}

type Producer struct {
	w *kafkago.Writer
}

func NewProducer(cfg ProducerConfig) (*Producer, error) {
	if len(cfg.Brokers) == 0 || cfg.Topic == "" {
		return nil, errors.New("invalid producer config: brokers and topic are required")
	}
	w := &kafkago.Writer{
		Addr:     kafkago.TCP(cfg.Brokers...),
		Topic:    cfg.Topic,
		Balancer: &kafkago.LeastBytes{},
	}
	return &Producer{w: w}, nil
}

func (p *Producer) Publish(ctx context.Context, key string, value []byte) error {
	msg := kafkago.Message{
		Key:   []byte(key),
		Value: value,
		Time:  time.Now(),
	}
	if ctx == nil {
		ctx = context.Background()
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
