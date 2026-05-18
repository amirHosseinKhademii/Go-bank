package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "bank/db/mocks"
	repository "bank/db/sqlc"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func randomAccount(owner string) repository.Account {
	return repository.Account{
		ID:       int64(time.Now().UnixNano()),
		Owner:    owner,
		Balance:  1000,
		Currency: "USD",
		CreatedAt: pgtype.Timestamp{
			Time:  time.Now(),
			Valid: true,
		},
	}
}

type mockStore struct {
	*mockdb.MockQuerier
	transferTxErr error
}

func (m *mockStore) TransferTx(ctx context.Context, arg repository.TransferTxParams) (repository.TransferTxResult, error) {
	if m.transferTxErr != nil {
		return repository.TransferTxResult{}, m.transferTxErr
	}
	return repository.TransferTxResult{}, nil
}

func TestCreateAccountAPI(t *testing.T) {
	account := randomAccount("test_user")

	testCases := []struct {
		name          string
		body          CreateAccountRequest
		buildStubs    func(store *mockdb.MockQuerier)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: CreateAccountRequest{
				Owner:    account.Owner,
				Currency: account.Currency,
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				arg := repository.CreateAccountParams{
					Owner:    account.Owner,
					Currency: account.Currency,
					Balance:  0,
				}
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name: "MissingOwner",
			body: CreateAccountRequest{
				Currency: account.Currency,
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "MissingCurrency",
			body: CreateAccountRequest{
				Owner: account.Owner,
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidCurrency",
			body: CreateAccountRequest{
				Owner:    account.Owner,
				Currency: "INVALID",
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "DuplicateAccount",
			body: CreateAccountRequest{
				Owner:    account.Owner,
				Currency: account.Currency,
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				arg := repository.CreateAccountParams{
					Owner:    account.Owner,
					Currency: account.Currency,
					Balance:  0,
				}
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(repository.Account{}, mockPgUniqueViolationError())
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusConflict, recorder.Code)
			},
		},
		{
			name: "InvalidOwnerForeignKeyViolation",
			body: CreateAccountRequest{
				Owner:    "non_existent_user",
				Currency: account.Currency,
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				arg := repository.CreateAccountParams{
					Owner:    "non_existent_user",
					Currency: account.Currency,
					Balance:  0,
				}
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(repository.Account{}, mockPgForeignKeyError())
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockQuerier := mockdb.NewMockQuerier(ctrl)
			tc.buildStubs(mockQuerier)

			server := NewServer(&mockStore{MockQuerier: mockQuerier})
			recorder := httptest.NewRecorder()

			body, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(body))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetAccountAPI(t *testing.T) {
	account := randomAccount("test_user")

	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockQuerier)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(repository.Account{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockQuerier := mockdb.NewMockQuerier(ctrl)
			tc.buildStubs(mockQuerier)

			server := NewServer(&mockStore{MockQuerier: mockQuerier})
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListAccountsAPI(t *testing.T) {
	owner := "test_user"
	n := int32(5)
	accounts := make([]repository.Account, n)
	for i := int32(0); i < n; i++ {
		accounts[i] = randomAccount(owner)
	}

	type Query struct {
		pageID   int32
		pageSize int32
	}

	testCases := []struct {
		name          string
		query         Query
		buildStubs    func(store *mockdb.MockQuerier)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				pageID:   1,
				pageSize: 5,
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				arg := repository.ListAccountsParams{
					Limit:  5,
					Offset: 0,
				}
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccounts(t, recorder.Body, accounts)
			},
		},
		{
			name: "InvalidPageID",
			query: Query{
				pageID:   0,
				pageSize: 5,
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidPageSize",
			query: Query{
				pageID:   1,
				pageSize: 3,
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockQuerier := mockdb.NewMockQuerier(ctrl)
			tc.buildStubs(mockQuerier)

			server := NewServer(&mockStore{MockQuerier: mockQuerier})
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts?page_id=%d&page_size=%d",
				tc.query.pageID, tc.query.pageSize)

			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account repository.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccount repository.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account.ID, gotAccount.ID)
	require.Equal(t, account.Owner, gotAccount.Owner)
	require.Equal(t, account.Balance, gotAccount.Balance)
	require.Equal(t, account.Currency, gotAccount.Currency)
}

func requireBodyMatchAccounts(t *testing.T, body *bytes.Buffer, accounts []repository.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccounts []repository.Account
	err = json.Unmarshal(data, &gotAccounts)
	require.NoError(t, err)
	require.Len(t, gotAccounts, len(accounts))
	for i := range gotAccounts {
		require.Equal(t, accounts[i].ID, gotAccounts[i].ID)
		require.Equal(t, accounts[i].Owner, gotAccounts[i].Owner)
		require.Equal(t, accounts[i].Balance, gotAccounts[i].Balance)
		require.Equal(t, accounts[i].Currency, gotAccounts[i].Currency)
	}
}

func mockPgUniqueViolationError() error {
	return &pgconn.PgError{Code: "23505"}
}

func mockPgForeignKeyError() error {
	return &pgconn.PgError{Code: "23503"}
}
