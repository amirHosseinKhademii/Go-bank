package api

import (
	"bytes"
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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func randomTransfer(fromID, toID int64) repository.Transfer {
	return repository.Transfer{
		ID:            int64(time.Now().UnixNano()),
		FromAccountID: fromID,
		ToAccountID:   toID,
		Amount:        100,
		CreatedAt: pgtype.Timestamp{
			Time:  time.Now(),
			Valid: true,
		},
	}
}

func TestCreateTransferAPI(t *testing.T) {
	fromAccount := randomAccount("user1")
	toAccount := randomAccount("user2")

	testCases := []struct {
		name          string
		body          CreateTransferRequest
		buildStubs    func(store *mockdb.MockQuerier)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: CreateTransferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        100,
				Currency:      "USD",
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(1).
					Return(toAccount, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "MissingFromAccountID",
			body: CreateTransferRequest{
				ToAccountID: toAccount.ID,
				Amount:      100,
				Currency:    "USD",
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "MissingToAccountID",
			body: CreateTransferRequest{
				FromAccountID: fromAccount.ID,
				Amount:        100,
				Currency:      "USD",
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "MissingAmount",
			body: CreateTransferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Currency:      "USD",
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidAmount",
			body: CreateTransferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        0,
				Currency:      "USD",
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "MissingCurrency",
			body: CreateTransferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        100,
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidCurrency",
			body: CreateTransferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        100,
				Currency:      "INVALID",
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "FromAccountNotFound",
			body: CreateTransferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        100,
				Currency:      "USD",
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(repository.Account{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "ToAccountNotFound",
			body: CreateTransferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        100,
				Currency:      "USD",
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(1).
					Return(repository.Account{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "CurrencyMismatchFromAccount",
			body: CreateTransferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        100,
				Currency:      "EUR",
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BothAccountsValidSameCurrency",
			body: CreateTransferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        100,
				Currency:      "USD",
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(1).
					Return(toAccount, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
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

			request, err := http.NewRequest(http.MethodPost, "/transfers", bytes.NewReader(body))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetTransferAPI(t *testing.T) {
	transfer := randomTransfer(1, 2)

	testCases := []struct {
		name          string
		transferID    int64
		buildStubs    func(store *mockdb.MockQuerier)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:       "OK",
			transferID: transfer.ID,
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).
					Times(1).
					Return(transfer, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTransfer(t, recorder.Body, transfer)
			},
		},
		{
			name:       "NotFound",
			transferID: transfer.ID,
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).
					Times(1).
					Return(repository.Transfer{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:       "InvalidID",
			transferID: 0,
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					GetTransfer(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/transfers/%d", tc.transferID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListTransfersAPI(t *testing.T) {
	n := int32(5)
	transfers := make([]repository.Transfer, n)
	for i := int32(0); i < n; i++ {
		transfers[i] = randomTransfer(int64(i+1), int64(i+2))
	}

	type Query struct {
		pageID   int32
		pageSize int32
	}

	testCases := []struct {
		name          string
		query         Query
		buildStubs    func(store *mockdb.MockQuerier)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				pageID:   1,
				pageSize: 5,
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				arg := repository.ListTransfersParams{
					Limit:  5,
					Offset: 0,
				}
				store.EXPECT().
					ListTransfers(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(transfers, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTransfers(t, recorder.Body, transfers)
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
					ListTransfers(gomock.Any(), gomock.Any()).
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
					ListTransfers(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "PageSizeTooLarge",
			query: Query{
				pageID:   1,
				pageSize: 11,
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				store.EXPECT().
					ListTransfers(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Page2",
			query: Query{
				pageID:   2,
				pageSize: 5,
			},
			buildStubs: func(store *mockdb.MockQuerier) {
				arg := repository.ListTransfersParams{
					Limit:  5,
					Offset: 5,
				}
				store.EXPECT().
					ListTransfers(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(transfers, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTransfers(t, recorder.Body, transfers)
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

			url := fmt.Sprintf("/transfers?page_id=%d&page_size=%d",
				tc.query.pageID, tc.query.pageSize)

			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func requireBodyMatchTransfer(t *testing.T, body *bytes.Buffer, transfer repository.Transfer) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotTransfer repository.Transfer
	err = json.Unmarshal(data, &gotTransfer)
	require.NoError(t, err)
	require.Equal(t, transfer.ID, gotTransfer.ID)
	require.Equal(t, transfer.FromAccountID, gotTransfer.FromAccountID)
	require.Equal(t, transfer.ToAccountID, gotTransfer.ToAccountID)
	require.Equal(t, transfer.Amount, gotTransfer.Amount)
}

func requireBodyMatchTransfers(t *testing.T, body *bytes.Buffer, transfers []repository.Transfer) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotTransfers []repository.Transfer
	err = json.Unmarshal(data, &gotTransfers)
	require.NoError(t, err)
	require.Len(t, gotTransfers, len(transfers))
	for i := range gotTransfers {
		require.Equal(t, transfers[i].ID, gotTransfers[i].ID)
		require.Equal(t, transfers[i].FromAccountID, gotTransfers[i].FromAccountID)
		require.Equal(t, transfers[i].ToAccountID, gotTransfers[i].ToAccountID)
		require.Equal(t, transfers[i].Amount, gotTransfers[i].Amount)
	}
}
