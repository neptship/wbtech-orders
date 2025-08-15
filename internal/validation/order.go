package validation

import (
	"errors"
	"strings"

	"github.com/neptship/wbtech-orders/internal/models"
)

var (
	ErrEmptyOrderUID  = errors.New("order_uid empty")
	ErrOrderUIDFormat = errors.New("order_uid has invalid characters")
	ErrOrderUIDLen    = errors.New("order_uid too long")
	ErrEmptyTrack     = errors.New("track_number empty")
	ErrNoItems        = errors.New("items empty")
	ErrNegativeAmount = errors.New("payment amount negative")
	ErrInvalidEmail   = errors.New("delivery email invalid")
)

func Basic(o any, track ...string) error {
	switch v := o.(type) {
	case models.Order:
		uid := strings.TrimSpace(v.OrderUID)
		if uid == "" {
			return ErrEmptyOrderUID
		}
		if len(uid) > 100 {
			return ErrOrderUIDLen
		}
		if !isSafeID(uid) {
			return ErrOrderUIDFormat
		}
		if strings.TrimSpace(v.TrackNumber) == "" {
			return ErrEmptyTrack
		}
		if len(v.Items) == 0 {
			return ErrNoItems
		}
		if v.Payment.Amount < 0 {
			return ErrNegativeAmount
		}
		if e := strings.TrimSpace(v.Delivery.Email); e != "" && !strings.Contains(e, "@") {
			return ErrInvalidEmail
		}
	case string:
		if strings.TrimSpace(v) == "" {
			return ErrEmptyOrderUID
		}
		if len(track) > 0 && strings.TrimSpace(track[0]) == "" {
			return ErrEmptyTrack
		}
	default:
	}
	return nil
}

func isSafeID(s string) bool {
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		return false
	}
	return true
}
