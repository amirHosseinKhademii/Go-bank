package repository

import (
	"context"
	"log"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

var dbSource = os.Getenv("DB_SOURCE")

var testQueries *Queries

func TestMain(m *testing.M) {
	conn, err := pgxpool.New(context.Background(), dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	defer conn.Close()

	testQueries = New(conn)

	m.Run()
}

func createRandomAccount(t *testing.T) Account {
	owner := "random_user_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	balance := int64(rand.Intn(10000))
	currency := "USD"

	arg := CreateAccountParams{
		Owner:    owner,
		Balance:  balance,
		Currency: currency,
	}
	account, err := testQueries.CreateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	return account
}
