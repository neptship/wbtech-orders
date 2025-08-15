package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/neptship/wbtech-orders/internal/config"
	"github.com/neptship/wbtech-orders/internal/repository"
	"github.com/neptship/wbtech-orders/internal/service"
	"github.com/neptship/wbtech-orders/internal/validation"
)

type Handler struct {
	svc   *service.Service
	cfg   config.Config
	limit int64
}

func NewHandler(svc *service.Service, cfg config.Config) *Handler {
	return &Handler{svc: svc, cfg: cfg, limit: 1 << 20}
}

func (h *Handler) OrderAddHandler() http.HandlerFunc { return h.handleOrderAdd }
func (h *Handler) OrderGetHandler() http.HandlerFunc { return h.handleOrderGet }

func (h *Handler) handleOrderAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(r.Body, h.limit))
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}
	var meta struct {
		OrderUID    string `json:"order_uid"`
		TrackNumber string `json:"track_number"`
	}
	if err := json.Unmarshal(raw, &meta); err != nil || meta.OrderUID == "" {
		http.Error(w, "invalid json: order_uid required", http.StatusBadRequest)
		return
	}
	if err := validation.Basic(meta.OrderUID, meta.TrackNumber); err != nil {
		http.Error(w, "validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.svc.PublishRawOrder(r.Context(), raw); err != nil {
		http.Error(w, "kafka error", http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"status":"queued"}`))
}

func (h *Handler) handleOrderGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	orderUID := strings.TrimPrefix(r.URL.Path, "/order/")
	if orderUID == "" {
		http.Error(w, "order_uid required", 400)
		return
	}
	o, err := h.svc.GetOrder(r.Context(), orderUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(o); err != nil {
		http.Error(w, "encode error", http.StatusInternalServerError)
	}
}
