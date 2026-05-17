// package repository provides database operations for the bank service.
// It implements transactional operations with deadlock prevention strategies.
//
// Transaction and Locking Strategy:
// - All money transfers are executed within a single database transaction to ensure atomicity.
// - To prevent deadlocks when updating multiple accounts, we always update accounts in a consistent order:
//   we compare account IDs and update the account with the smaller ID first.
// - The TransferTx function orchestrates a transfer by:
//   1. Creating a transfer record
//   2. Creating two entry records (debit and credit)
//   3. Updating both account balances in a deadlock-safe order
//
// This approach ensures that concurrent transfers between the same pair of accounts
// will not deadlock because they always attempt to acquire locks in the same order.
package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Stor struct {
	*Queries
	db *pgxpool.Pool // underlying connection pool
}

func NewStore(db *pgxpool.Pool) *Stor {
	return &Stor{
		Queries: New(db),
		db:      db,
	}
}

// execTx runs a function within a database transaction and handles commit or rollback.
// It ensures that all operations in fn are executed atomically.
func (s *Stor) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}

// TransferTx performs a money transfer from one account to another.
// It creates a transfer record, two entry records (one for each account), and updates the balances of both accounts.
// To avoid deadlock when updating multiple accounts, the accounts are updated in a consistent order (by account ID).
// This function is executed within a transaction to ensure atomicity.
func (s *Stor) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := s.execTx(ctx, func(q *Queries) error {
		var err error

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		// check the id of account to see which one is smaller to avoid deadlock
		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
		} else {
			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
		}
		return err
	})

	return result, err
}

// addMoney updates the balances of two accounts within a single transaction.
// It updates account1 then account2 in the order provided by the caller.
// The TransferTx function orders the account IDs (by comparing FromAccountID and ToAccountID)
// before calling addMoney to ensure a consistent lock order and prevent deadlocks.
func addMoney(ctx context.Context, q *Queries, accountID1 int64, amount1 int64, accountID2 int64, amount2 int64) (account1 Account, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}
	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})
	if err != nil {
		return
	}
	return
}

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
}

var txKey = struct{}{}
