package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/neptship/wbtech-orders/internal/model"
	"github.com/segmentio/kafka-go"
)

type ConsumerConfig struct {
	Brokers []string
	Topic   string
	GroupID string
	DB      *sql.DB
}

type Consumer struct {
	reader  *kafka.Reader
	db      *sql.DB
	OnSaved func(order model.Order)
}

func NewConsumer(cfg ConsumerConfig) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		GroupID:  cfg.GroupID,
		MinBytes: 1,
		MaxBytes: 1 << 20,
	})
	return &Consumer{
		reader: reader,
		db:     cfg.DB,
	}
}

func (c *Consumer) Run(ctx context.Context) {
	log.Printf("Kafka consumer started for topic %s", c.reader.Config().Topic)
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("kafka read error: %v", err)
			time.Sleep(time.Second)
			continue
		}
		var order model.Order
		if err := json.Unmarshal(m.Value, &order); err != nil {
			log.Printf("invalid order json: %v", err)
			continue
		}
		if order.OrderUID == "" {
			log.Printf("skip: empty order_uid")
			continue
		}
		deliveryJSON, _ := json.Marshal(order.Delivery)
		paymentJSON, _ := json.Marshal(order.Payment)
		itemsJSON, _ := json.Marshal(order.Items)
		_, err = c.db.ExecContext(ctx, `
			INSERT INTO orders (
				order_uid, track_number, entry, delivery, payment, items, locale,
				internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
			ON CONFLICT (order_uid) DO UPDATE SET
				track_number=EXCLUDED.track_number,
				entry=EXCLUDED.entry,
				delivery=EXCLUDED.delivery,
				payment=EXCLUDED.payment,
				items=EXCLUDED.items,
				locale=EXCLUDED.locale,
				internal_signature=EXCLUDED.internal_signature,
				customer_id=EXCLUDED.customer_id,
				delivery_service=EXCLUDED.delivery_service,
				shardkey=EXCLUDED.shardkey,
				sm_id=EXCLUDED.sm_id,
				date_created=EXCLUDED.date_created,
				oof_shard=EXCLUDED.oof_shard
		`,
			order.OrderUID, order.TrackNumber, order.Entry,
			deliveryJSON, paymentJSON, itemsJSON,
			order.Locale, order.InternalSig, order.CustomerID, order.DeliverySvc,
			order.ShardKey, order.SmID, order.DateCreated, order.OofShard,
		)
		if err != nil {
			log.Printf("db insert error: %v", err)
			continue
		}
		log.Printf("order %s saved", order.OrderUID)
		if c.OnSaved != nil {
			c.OnSaved(order)
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
