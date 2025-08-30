package mocks_test

import (
	"context"
	"testing"

	"github.com/neptship/wbtech-orders/internal/models"
	"github.com/neptship/wbtech-orders/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestOrderRepositoryMock_Basic(t *testing.T) {
	repo := mocks.NewOrderRepositoryMock(t)
	ctx := context.Background()

	order := models.Order{OrderUID: "uid-1"}

	repo.EXPECT().Save(mock.Anything, order).Return(nil)
	repo.EXPECT().Get(mock.Anything, "uid-1").Return(order, nil)

	require.NoError(t, repo.Save(ctx, order))
	got, err := repo.Get(ctx, "uid-1")
	require.NoError(t, err)
	require.Equal(t, order, got)
}
