package repository

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

var dbSource = os.Getenv("DB_TEST")

var testQueries *Queries

var testDb *pgxpool.Pool

func migrateDbDownAndUp() {
	var err error

	cmd := exec.Command("make", "-C", "../..", "migrateTestDbdown")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatal("cannot run migrateTestDbdown:", err)
	}

	cmd = exec.Command("make", "-C", "../..", "migrateTestDbup")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatal("cannot run migrateTestDbup:", err)
	}
}

func TestMain(m *testing.M) {
	var err error

	migrateDbDownAndUp()

	testDb, err = pgxpool.New(context.Background(), dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	defer testDb.Close()

	testQueries = New(testDb)

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
