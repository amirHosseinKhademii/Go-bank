package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	user := createRandomUser(t)

	require.NotEmpty(t, user)
	require.NotZero(t, user.Username)
	require.NotZero(t, user.CreatedAt)
}

func TestGetUser(t *testing.T) {
	user := createRandomUser(t)

	getUser, err := testQueries.GetUser(context.Background(), user.Username)
	require.NoError(t, err)
	require.NotEmpty(t, getUser)
	require.Equal(t, user.Username, getUser.Username)
	require.Equal(t, user.FullName, getUser.FullName)
	require.Equal(t, user.Email, getUser.Email)
	require.WithinDuration(t, user.CreatedAt.Time, getUser.CreatedAt.Time, time.Second)
	require.WithinDuration(t, user.PasswordChangedAt.Time, getUser.PasswordChangedAt.Time, time.Second)
}
