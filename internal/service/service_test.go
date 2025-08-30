package service_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"github.com/neptship/wbtech-orders/internal/cache"
	"github.com/neptship/wbtech-orders/internal/models"
	"github.com/neptship/wbtech-orders/internal/repository"
	"github.com/neptship/wbtech-orders/internal/service"
)

func TestService_GetOrder_Integration(t *testing.T) {
	t.Parallel()

	dsn := "postgres://postgres:postgres@localhost:5432/orders_data?sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		t.Skipf("postgres not available: %v", err)
	}
	ctx := context.Background()
	orderUID := "integ-test-order-" + time.Now().Format("20060102150405")

	repo := repository.NewPostgres(db)
	c := cache.NewCache(0)

	svc := &service.Service{Cache: c, Repo: repo}

	want := models.Order{
		OrderUID:    orderUID,
		TrackNumber: "TRACK-1",
		Entry:       "ENTRY",
		Delivery:    models.Delivery{Name: "X"},
		Payment:     models.Payment{Amount: 100},
		Items:       []models.Item{{ChrtID: 1, Name: "item1", Price: 100}},
		DateCreated: time.Now().UTC().Format(time.RFC3339),
	}

	if err := repo.Save(ctx, want); err != nil {
		t.Fatalf("repo.Save: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM orders WHERE order_uid=$1", orderUID)
	})

	got, err := svc.GetOrder(ctx, orderUID)
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if got.OrderUID != want.OrderUID {
		t.Fatalf("unexpected order uid: %s != %s", got.OrderUID, want.OrderUID)
	}

	got2, err := svc.GetOrder(ctx, orderUID)
	if err != nil {
		t.Fatalf("GetOrder second: %v", err)
	}
	if got2.OrderUID != want.OrderUID {
		t.Fatalf("unexpected order uid second: %s != %s", got2.OrderUID, want.OrderUID)
	}
}
