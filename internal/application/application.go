package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	_ "github.com/lib/pq"
	"github.com/neptship/wbtech-orders/internal/kafka"
	"github.com/neptship/wbtech-orders/internal/model"
)

type Application struct {
	producer *kafka.Producer
	db       *sql.DB
	cache    map[string]model.Order
}

func New() *Application {
	dsn := "postgres://postgres:postgres@localhost:5432/orders_data?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	brokers := strings.Split("localhost:9092", ",")
	producer, err := kafka.NewProducer(kafka.ProducerConfig{
		Brokers:              brokers,
		Topic:                "orders",
		AllowAutoTopicCreate: true,
	})
	if err != nil {
		log.Fatalf("kafka producer: %v", err)
	}

	app := &Application{
		producer: producer,
		db:       db,
		cache:    make(map[string]model.Order),
	}

	app.warmUpCache()

	go func() {
		consumer := kafka.NewConsumer(kafka.ConsumerConfig{
			Brokers: brokers,
			Topic:   "orders",
			GroupID: "order-consumer-group",
			DB:      db,
		})
		consumer.OnSaved = app.cacheSet
		consumer.Run(context.Background())
	}()

	return app
}

func (a *Application) RunServer() error {
	http.HandleFunc("/order_add", a.handleOrderAdd)
	http.HandleFunc("/order/", a.handleOrderGet)
	http.Handle("/", http.FileServer(http.Dir("web")))
	addr := "8081"
	log.Printf("HTTP server on :%s", addr)
	return http.ListenAndServe(":"+addr, nil)
}

func (a *Application) handleOrderAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	var meta struct {
		OrderUID string `json:"order_uid"`
	}
	if err := json.Unmarshal(raw, &meta); err != nil || meta.OrderUID == "" {
		http.Error(w, "invalid json: order_uid required", 400)
		return
	}
	if err := a.producer.Publish(r.Context(), meta.OrderUID, raw); err != nil {
		http.Error(w, "kafka error", http.StatusBadGateway)
		return
	}
	w.WriteHeader(202)
	w.Write([]byte(`{"status":"queued"}`))
}

func (a *Application) handleOrderGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	orderUID := strings.TrimPrefix(r.URL.Path, "/order/")
	if orderUID == "" {
		http.Error(w, "order_uid required", 400)
		return
	}
	if o, ok := a.cacheGet(orderUID); ok {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(o)
		return
	}

	var o model.Order
	var delivery, payment, items []byte
	err := a.db.QueryRow(`SELECT order_uid, track_number, entry, delivery, payment, items, locale,
		internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders WHERE order_uid=$1`, orderUID).
		Scan(&o.OrderUID, &o.TrackNumber, &o.Entry, &delivery, &payment, &items,
			&o.Locale, &o.InternalSig, &o.CustomerID, &o.DeliverySvc, &o.ShardKey, &o.SmID, &o.DateCreated, &o.OofShard)
	if err == sql.ErrNoRows {
		http.Error(w, "not found", 404)
		return
	}
	if err != nil {
		http.Error(w, "db error", 500)
		return
	}
	_ = json.Unmarshal(delivery, &o.Delivery)
	_ = json.Unmarshal(payment, &o.Payment)
	_ = json.Unmarshal(items, &o.Items)
	a.cacheSet(o)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(o)
}

func (a *Application) warmUpCache() {
	a.cache = make(map[string]model.Order)
	rows, err := a.db.Query(`SELECT order_uid, track_number, entry, delivery, payment, items, locale,
		internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders`)
	if err != nil {
		log.Printf("cache warmup error: %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var o model.Order
		var delivery, payment, items []byte
		if err := rows.Scan(&o.OrderUID, &o.TrackNumber, &o.Entry, &delivery, &payment, &items,
			&o.Locale, &o.InternalSig, &o.CustomerID, &o.DeliverySvc, &o.ShardKey, &o.SmID, &o.DateCreated, &o.OofShard); err != nil {
			continue
		}
		_ = json.Unmarshal(delivery, &o.Delivery)
		_ = json.Unmarshal(payment, &o.Payment)
		_ = json.Unmarshal(items, &o.Items)
		a.cacheSet(o)
	}
}

func (a *Application) cacheSet(o model.Order) {
	a.cache[o.OrderUID] = o
}

func (a *Application) cacheGet(id string) (model.Order, bool) {
	o, ok := a.cache[id]
	return o, ok
}
