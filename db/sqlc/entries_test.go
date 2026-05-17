package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateEntry(t *testing.T) {
	account := createRandomAccount(t)

	arg := CreateEntryParams{
		AccountID: account.ID,
		Amount:    500,
	}
	entry, err := testQueries.CreateEntry(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, entry)
	require.NotZero(t, entry.ID)
	require.Equal(t, arg.AccountID, entry.AccountID)
	require.Equal(t, arg.Amount, entry.Amount)
	require.NotZero(t, entry.CreatedAt)
}

func TestGetEntry(t *testing.T) {
	account := createRandomAccount(t)
	entry, err := testQueries.CreateEntry(context.Background(), CreateEntryParams{
		AccountID: account.ID,
		Amount:    1000,
	})
	require.NoError(t, err)

	getEntry, err := testQueries.GetEntry(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, entry.ID, getEntry.ID)
	require.Equal(t, entry.AccountID, getEntry.AccountID)
	require.Equal(t, entry.Amount, getEntry.Amount)
	require.WithinDuration(t, entry.CreatedAt.Time, getEntry.CreatedAt.Time(), time.Second)
}

func TestListEntries(t *testing.T) {
	account := createRandomAccount(t)

	for i := 0; i < 5; i++ {
		_, err := testQueries.CreateEntry(context.Background(), CreateEntryParams{
			AccountID: account.ID,
			Amount:    int64(i+1) * 100,
		})
		require.NoError(t, err)
	}

	entries, err := testQueries.ListEntries(context.Background(), ListEntriesParams{
		AccountID: account.ID,
		Limit:     5,
		Offset:    0,
	})
	require.NoError(t, err)
	require.NotEmpty(t, entries)
	require.Len(t, entries, 5)
}

func TestUpdateEntry(t *testing.T) {
	account := createRandomAccount(t)
	entry, err := testQueries.CreateEntry(context.Background(), CreateEntryParams{
		AccountID: account.ID,
		Amount:    1000,
	})
	require.NoError(t, err)

	newAmount := int64(2000)
	updatedEntry, err := testQueries.UpdateEntry(context.Background(), UpdateEntryParams{
		ID:     entry.ID,
		Amount: newAmount,
	})
	require.NoError(t, err)
	require.Equal(t, newAmount, updatedEntry.Amount)
	require.Equal(t, entry.ID, updatedEntry.ID)
}

func TestDeleteEntry(t *testing.T) {
	account := createRandomAccount(t)
	entry, err := testQueries.CreateEntry(context.Background(), CreateEntryParams{
		AccountID: account.ID,
		Amount:    1000,
	})
	require.NoError(t, err)

	err = testQueries.DeleteEntry(context.Background(), entry.ID)
	require.NoError(t, err)

	_, err = testQueries.GetEntry(context.Background(), entry.ID)
	require.Error(t, err)
}
