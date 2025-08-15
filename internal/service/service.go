package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	"github.com/neptship/wbtech-orders/internal/cache"
	"github.com/neptship/wbtech-orders/internal/config"
	"github.com/neptship/wbtech-orders/internal/kafka"
	"github.com/neptship/wbtech-orders/internal/models"
	"github.com/neptship/wbtech-orders/internal/repository"
)

type Service struct {
	Producer *kafka.Producer
	DB       *sql.DB
	Cache    cache.Cache
	Repo     repository.OrderRepository
	cfg      config.Config
}

func New(ctx context.Context, cfg config.Config) (*Service, error) {
	db, err := sql.Open("postgres", config.GetDSN(cfg.Postgres))
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	producer, err := kafka.NewProducer(kafka.ProducerConfig{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.Topic,
	})
	if err != nil {
		return nil, fmt.Errorf("new producer: %w", err)
	}

	c := cache.NewCache(cfg.Cache.Size)

	repo := repository.NewPostgres(db)
	return &Service{Producer: producer, DB: db, Cache: c, Repo: repo, cfg: cfg}, nil
}

func (s *Service) StartConsumer(ctx context.Context) {
	go func() {
		consumer := kafka.NewConsumer(kafka.ConsumerConfig{
			Brokers: s.cfg.Kafka.Brokers,
			Topic:   s.cfg.Kafka.Topic,
			GroupID: s.cfg.Kafka.GroupID,
			DB:      s.DB,
		})
		consumer.OnSaved = func(o models.Order) { s.Cache.Set(o.OrderUID, o) }
		go consumer.Run(ctx)
		<-ctx.Done()
		_ = consumer.Close()
	}()
}

func (s *Service) PublishRawOrder(ctx context.Context, raw []byte) error {
	key := extractField(raw, "order_uid")
	if key == "" {
		key = "unknown"
	}
	return s.Producer.Publish(ctx, key, raw)
}

func (s *Service) GetOrder(ctx context.Context, id string) (models.Order, error) {
	if o, ok := s.Cache.Get(id); ok {
		return o, nil
	}
	o, err := s.Repo.Get(ctx, id)
	if err != nil {
		return models.Order{}, err
	}
	s.Cache.Set(o.OrderUID, o)
	return o, nil
}

func extractField(b []byte, field string) string {
	s := string(b)
	needle := `"` + field + `":"`
	idx := strings.Index(s, needle)
	if idx == -1 {
		return ""
	}
	rest := s[idx+len(needle):]
	end := strings.Index(rest, `"`)
	if end == -1 {
		return ""
	}
	return rest[:end]
}
