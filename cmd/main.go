package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/neptship/wbtech-orders/internal/api"
	"github.com/neptship/wbtech-orders/internal/config"
	"github.com/neptship/wbtech-orders/internal/service"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	svc, err := service.New(ctx, cfg)
	if err != nil {
		log.Fatalf("init service: %v", err)
	}
	svc.StartConsumer(ctx)

	h := api.NewHandler(svc, cfg)
	http.HandleFunc("/order_add", h.OrderAddHandler())
	http.HandleFunc("/order/", h.OrderGetHandler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })

	srv := &http.Server{Addr: ":" + cfg.HTTP.Port, Handler: nil}

	go func() {
		log.Printf("HTTP server on :%s", cfg.HTTP.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
	_ = svc.Producer.Close()
	_ = svc.DB.Close()
	log.Println("graceful shutdown complete")
}
