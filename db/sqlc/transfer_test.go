package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateTransfer(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	arg := CreateTransferParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        1000,
	}
	transfer, err := testQueries.CreateTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)
	require.NotZero(t, transfer.ID)
	require.Equal(t, arg.FromAccountID, transfer.FromAccountID)
	require.Equal(t, arg.ToAccountID, transfer.ToAccountID)
	require.Equal(t, arg.Amount, transfer.Amount)
	require.NotZero(t, transfer.CreatedAt)
}

func TestGetTransfer(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	transfer, err := testQueries.CreateTransfer(context.Background(), CreateTransferParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        2000,
	})
	require.NoError(t, err)

	getTransfer, err := testQueries.GetTransfer(context.Background(), transfer.ID)
	require.NoError(t, err)
	require.Equal(t, transfer.ID, getTransfer.ID)
	require.Equal(t, transfer.FromAccountID, getTransfer.FromAccountID)
	require.Equal(t, transfer.ToAccountID, getTransfer.ToAccountID)
	require.Equal(t, transfer.Amount, getTransfer.Amount)
	require.WithinDuration(t, transfer.CreatedAt.Time, getTransfer.CreatedAt.Time(), time.Second)
}

func TestListTransfers(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	for i := 0; i < 5; i++ {
		_, err := testQueries.CreateTransfer(context.Background(), CreateTransferParams{
			FromAccountID: account1.ID,
			ToAccountID:   account2.ID,
			Amount:        int64(i+1) * 100,
		})
		require.NoError(t, err)
	}

	transfers, err := testQueries.ListTransfers(context.Background(), ListTransfersParams{
		FromAccountID: account1.ID,
		Limit:         5,
		Offset:        0,
	})
	require.NoError(t, err)
	require.NotEmpty(t, transfers)
	require.Len(t, transfers, 5)
}

func TestUpdateTransfer(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	transfer, err := testQueries.CreateTransfer(context.Background(), CreateTransferParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        1000,
	})
	require.NoError(t, err)

	newAmount := int64(5000)
	updatedTransfer, err := testQueries.UpdateTransfer(context.Background(), UpdateTransferParams{
		ID:            transfer.ID,
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        newAmount,
	})
	require.NoError(t, err)
	require.Equal(t, newAmount, updatedTransfer.Amount)
	require.Equal(t, transfer.ID, updatedTransfer.ID)
}

func TestDeleteTransfer(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	transfer, err := testQueries.CreateTransfer(context.Background(), CreateTransferParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        1000,
	})
	require.NoError(t, err)

	err = testQueries.DeleteTransfer(context.Background(), transfer.ID)
	require.NoError(t, err)

	_, err = testQueries.GetTransfer(context.Background(), transfer.ID)
	require.Error(t, err)
}
