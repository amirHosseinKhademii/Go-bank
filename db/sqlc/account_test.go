package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateAccount(t *testing.T) {
	account := createRandomAccount(t)

	require.NotEmpty(t, account)
	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)
}

func TestGetAccount(t *testing.T) {
	account := createRandomAccount(t)

	getAccount, err := testQueries.GetAccount(context.Background(), account.ID)
	require.NoError(t, err)
	require.NotEmpty(t, getAccount)
	require.Equal(t, account.ID, getAccount.ID)
	require.Equal(t, account.Owner, getAccount.Owner)
	require.Equal(t, account.Balance, getAccount.Balance)
	require.Equal(t, account.Currency, getAccount.Currency)
}

func TestListAccounts(t *testing.T) {
	for i := 0; i < 5; i++ {
		createRandomAccount(t)
	}

	accounts, err := testQueries.ListAccounts(context.Background(), ListAccountsParams{
		Limit:  5,
		Offset: 0,
	})
	require.NoError(t, err)
	require.NotEmpty(t, accounts)
	require.Len(t, accounts, 5)
}

func TestUpdateAccount(t *testing.T) {
	account := createRandomAccount(t)

	newBalance := account.Balance + 500
	updatedAccount, err := testQueries.UpdateAccount(context.Background(), UpdateAccountParams{
		ID:      account.ID,
		Balance: newBalance,
	})
	require.NoError(t, err)
	require.Equal(t, newBalance, updatedAccount.Balance)
	require.Equal(t, account.ID, updatedAccount.ID)
	require.Equal(t, account.Owner, updatedAccount.Owner)
	require.Equal(t, account.Currency, updatedAccount.Currency)
}

func TestDeleteAccount(t *testing.T) {
	account := createRandomAccount(t)

	err := testQueries.DeleteAccount(context.Background(), account.ID)
	require.NoError(t, err)

	_, err = testQueries.GetAccount(context.Background(), account.ID)
	require.Error(t, err)
}
