package mocks_test

import (
	"testing"

	"github.com/neptship/wbtech-orders/internal/models"
	"github.com/neptship/wbtech-orders/mocks"
	"github.com/stretchr/testify/require"
)

func TestCacheMock_Basic(t *testing.T) {
	c := mocks.NewCacheMock(t)

	order := models.Order{OrderUID: "uid-1"}

	c.EXPECT().Set("uid-1", order).Return()
	c.EXPECT().Get("uid-1").Return(order, true)

	c.Set("uid-1", order)
	got, ok := c.Get("uid-1")

	require.True(t, ok)
	require.Equal(t, order, got)
}
