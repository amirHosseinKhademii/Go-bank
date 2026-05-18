package repository

import (
	"bank/utils"
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
	user := createRandomUser(t)
	balance := int64(rand.Intn(10000))
	currency := "USD"

	arg := CreateAccountParams{
		Owner:    user.Username,
		Balance:  balance,
		Currency: currency,
	}
	account, err := testQueries.CreateAccount(context.Background(), arg)
	require.Equal(t, user.Username, account.Owner)
	require.Equal(t, balance, account.Balance)
	require.Equal(t, currency, account.Currency)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	return account
}

func createRandomUser(t *testing.T) CreateUserRow {
	hashedPassword, err := utils.HashPassword("secret")
	require.NoError(t, err)

	username := "random_user_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	fullName := "Random User"
	email := username + "@example.com"

	arg := CreateUserParams{
		Username:       username,
		HashedPassword: hashedPassword,
		FullName:       fullName,
		Email:          email,
	}
	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, username, user.Username)
	require.Equal(t, fullName, user.FullName)
	require.Equal(t, email, user.Email)
	require.True(t, user.PasswordChangedAt.Time.IsZero())

	return user
}
