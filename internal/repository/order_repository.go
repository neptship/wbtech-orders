package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/neptship/wbtech-orders/internal/models"
)

type OrderRepository interface {
	Save(ctx context.Context, o models.Order) error
	Get(ctx context.Context, id string) (models.Order, error)
}

var ErrNotFound = errors.New("order not found")

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) *PostgresOrderRepository { return &PostgresOrderRepository{db: db} }

func (r *PostgresOrderRepository) Save(ctx context.Context, o models.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	deliveryJSON, err := json.Marshal(o.Delivery)
	if err != nil {
		return err
	}
	paymentJSON, err := json.Marshal(o.Payment)
	if err != nil {
		return err
	}
	itemsJSON, err := json.Marshal(o.Items)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
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
		o.OrderUID, o.TrackNumber, o.Entry,
		deliveryJSON, paymentJSON, itemsJSON,
		o.Locale, o.InternalSig, o.CustomerID, o.DeliverySvc,
		o.ShardKey, o.SmID, o.DateCreated, o.OofShard,
	)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (r *PostgresOrderRepository) Get(ctx context.Context, id string) (models.Order, error) {
	var (
		o         models.Order
		deliveryB []byte
		paymentB  []byte
		itemsB    []byte
	)
	err := r.db.QueryRowContext(ctx, `SELECT order_uid, track_number, entry, delivery, payment, items, locale,
        internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders WHERE order_uid=$1`, id).
		Scan(&o.OrderUID, &o.TrackNumber, &o.Entry, &deliveryB, &paymentB, &itemsB, &o.Locale, &o.InternalSig, &o.CustomerID, &o.DeliverySvc, &o.ShardKey, &o.SmID, &o.DateCreated, &o.OofShard)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Order{}, ErrNotFound
		}
		return models.Order{}, err
	}
	_ = json.Unmarshal(deliveryB, &o.Delivery)
	_ = json.Unmarshal(paymentB, &o.Payment)
	_ = json.Unmarshal(itemsB, &o.Items)
	return o, nil
}
