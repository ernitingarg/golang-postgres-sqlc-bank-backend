package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/ernitingarg/golang-postgres-sqlc-bank-backend/db/mock"
	db "github.com/ernitingarg/golang-postgres-sqlc-bank-backend/db/sqlc"
	"github.com/ernitingarg/golang-postgres-sqlc-bank-backend/utils"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetAccountApi(t *testing.T) {
	account := createRandomAccount()

	testCases := []struct {
		name             string
		accountId        int64
		buildStubFunc    func(store *mockdb.MockStore)
		validateResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountId: account.ID,
			buildStubFunc: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
				assertAccount(t, account, recorder.Body)
			},
		},
		{
			name:      "NotFoundError",
			accountId: account.ID,
			buildStubFunc: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, pgx.ErrNoRows)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
				assertError(t, pgx.ErrNoRows, recorder.Body)
			},
		},
		{
			name:      "InternalServerError",
			accountId: account.ID,
			buildStubFunc: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, pgx.ErrTxClosed)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
				assertError(t, pgx.ErrTxClosed, recorder.Body)
			},
		},
		{
			name:      "BadRequest",
			accountId: 0,
			buildStubFunc: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			// build test stub
			tc.buildStubFunc(store)

			// start test server and send http request
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/accounts/%d", tc.accountId)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.validateResponse(recorder)
		})
	}
}

func TestCreateAccountApi(t *testing.T) {
	account := createRandomAccount()

	testCases := []struct {
		name             string
		account          db.Account
		buildStubFunc    func(store *mockdb.MockStore)
		validateResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:    "OK",
			account: account,
			buildStubFunc: func(store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    account.Owner,
					Currency: account.Currency,
					Balance:  decimal.NewFromInt(0),
				}

				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(account, nil)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
				assertAccount(t, account, recorder.Body)
			},
		},
		{
			name:    "InternalServerError",
			account: account,
			buildStubFunc: func(store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    account.Owner,
					Currency: account.Currency,
					Balance:  decimal.NewFromInt(0),
				}

				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Account{}, pgx.ErrTxClosed)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
				assertError(t, pgx.ErrTxClosed, recorder.Body)
			},
		},
		{
			name:    "BadRequest",
			account: db.Account{}, // empty account
			buildStubFunc: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubFunc(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := "/api/accounts"
			req := createAccountRequest{
				Owner:    tc.account.Owner,
				Currency: tc.account.Currency,
			}
			body, err := json.Marshal(req)
			assert.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			assert.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.validateResponse(recorder)
		})
	}
}

func TestUpdateAccountApi(t *testing.T) {
	account := createRandomAccount()

	testCases := []struct {
		name             string
		account          db.Account
		buildStubFunc    func(store *mockdb.MockStore)
		validateResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:    "OK",
			account: account,
			buildStubFunc: func(store *mockdb.MockStore) {
				arg := db.UpdateAccountParams{
					ID:      account.ID,
					Balance: account.Balance,
				}

				store.
					EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(account, nil)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
				assertAccount(t, account, recorder.Body)
			},
		},
		{
			name:    "InternalServerError",
			account: account,
			buildStubFunc: func(store *mockdb.MockStore) {
				arg := db.UpdateAccountParams{
					ID:      account.ID,
					Balance: account.Balance,
				}

				store.
					EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Account{}, pgx.ErrTxClosed)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
				assertError(t, pgx.ErrTxClosed, recorder.Body)
			},
		},
		{
			name:    "NotFound",
			account: account,
			buildStubFunc: func(store *mockdb.MockStore) {
				arg := db.UpdateAccountParams{
					ID:      account.ID,
					Balance: account.Balance,
				}

				store.
					EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Account{}, pgx.ErrNoRows)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
				assertError(t, pgx.ErrNoRows, recorder.Body)
			},
		},
		{
			name:    "BadRequest",
			account: db.Account{},
			buildStubFunc: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					UpdateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubFunc(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			req := updateAccountRequest{
				ID:      tc.account.ID,
				Balance: tc.account.Balance,
			}

			body, err := json.Marshal(req)
			assert.NoError(t, err)

			url := "/api/accounts"
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
			assert.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.validateResponse(recorder)
		})
	}
}

func TestDeleteAccountApi(t *testing.T) {
	account := createRandomAccount()

	testCases := []struct {
		name             string
		accountId        int64
		buildStubFunc    func(store *mockdb.MockStore)
		validateResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountId: account.ID,
			buildStubFunc: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(nil)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:      "InternalServerError",
			accountId: account.ID,
			buildStubFunc: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(pgx.ErrTxClosed)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
				assertError(t, pgx.ErrTxClosed, recorder.Body)
			},
		},
		{
			name:      "NotFound",
			accountId: account.ID,
			buildStubFunc: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
				assertError(t, pgx.ErrNoRows, recorder.Body)
			},
		},
		{
			name:      "BadRequest",
			accountId: 0,
			buildStubFunc: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					DeleteAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubFunc(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/accounts/%d", tc.accountId)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			assert.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.validateResponse(recorder)
		})
	}
}

func TestListAccountsApi(t *testing.T) {
	n := 5

	accounts := make([]db.Account, n)
	for i := 0; i < 5; i++ {
		accounts[i] = createRandomAccount()
	}

	type Query struct {
		PageId   int
		PageSize int
	}

	testCases := []struct {
		name             string
		query            Query
		buildStubFunc    func(store *mockdb.MockStore)
		validateResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				PageId:   1,
				PageSize: 5,
			},
			buildStubFunc: func(store *mockdb.MockStore) {
				arg := db.ListAccountsParams{
					Limit:  5,
					Offset: 0,
				}
				store.
					EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(accounts, nil)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
				assertAccounts(t, accounts, recorder.Body)
			},
		},
		{
			name: "InternalServerError",
			query: Query{
				PageId:   1,
				PageSize: 5,
			},
			buildStubFunc: func(store *mockdb.MockStore) {
				arg := db.ListAccountsParams{
					Limit:  5,
					Offset: 0,
				}
				store.
					EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return([]db.Account{}, pgx.ErrTxClosed)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
				assertError(t, pgx.ErrTxClosed, recorder.Body)
			},
		},
		{
			name: "BadRequest",
			query: Query{
				PageId:   0,
				PageSize: 5,
			},
			buildStubFunc: func(store *mockdb.MockStore) {
				store.
					EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubFunc(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			//url := fmt.Sprintf("/api/accounts?page_id={%d}&page_size=%d", 1, 1)
			url := "/api/accounts"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			query := request.URL.Query()
			query.Add("page_id", fmt.Sprintf("%d", tc.query.PageId))
			query.Add("page_size", fmt.Sprintf("%d", tc.query.PageSize))
			request.URL.RawQuery = query.Encode()

			assert.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.validateResponse(recorder)
		})
	}
}

func createRandomAccount() db.Account {
	return db.Account{
		ID:       utils.RandomNumber(1, 1000),
		Owner:    utils.RandomString(5),
		Balance:  utils.RandomDecimal(10, 100),
		Currency: randomCurrency(),
	}
}

// RandomCurrency returns random currency among EUR, USD, INR
func randomCurrency() db.Currency {
	currencies := []db.Currency{db.CurrencyEUR, db.CurrencyUSD, db.CurrencyINR}
	index := rand.Intn(len(currencies))
	return currencies[index]
}

func assertAccount(t *testing.T, expectedAccount db.Account, body *bytes.Buffer) {
	data, err := io.ReadAll(body)
	assert.NoError(t, err)

	var actualAccount db.Account
	err = json.Unmarshal(data, &actualAccount)
	assert.NoError(t, err)

	assert.Equal(t, expectedAccount, actualAccount)
}

func assertAccounts(t *testing.T, expectedAccounts []db.Account, body *bytes.Buffer) {
	data, err := io.ReadAll(body)
	assert.NoError(t, err)

	var actualAccounts []db.Account
	err = json.Unmarshal(data, &actualAccounts)
	assert.NoError(t, err)

	assert.Equal(t, expectedAccounts, actualAccounts)
}

func assertError(t *testing.T, expectedErr error, body *bytes.Buffer) {
	data, err := io.ReadAll(body)
	assert.NoError(t, err)

	var actualErr map[string]interface{}
	err = json.Unmarshal(data, &actualErr)

	assert.Equal(t, errorResponse(expectedErr)["error"], actualErr["error"])
}
