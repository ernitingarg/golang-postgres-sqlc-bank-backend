package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/ernitingarg/golang-postgres-sqlc-bank-backend/db/mock"
	db "github.com/ernitingarg/golang-postgres-sqlc-bank-backend/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestTransferApi(t *testing.T) {

	fromAccount := createRandomAccount()
	fromAccount.Currency = db.CurrencyEUR

	toAccount := createRandomAccount()
	toAccount.Currency = db.CurrencyEUR

	toAccountDifferentCurrency := createRandomAccount()
	toAccountDifferentCurrency.Currency = db.CurrencyINR

	testCases := []struct {
		name             string
		body             gin.H
		buildStubFunc    func(store *mockdb.MockStore)
		validateResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          decimal.NewFromInt(1),
				"currency":        toAccount.Currency,
			},

			buildStubFunc: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).Times(1).Return(toAccount, nil)

				arg := db.TransferTxParams{
					FromAccountId: fromAccount.ID,
					ToAccountId:   toAccount.ID,
					Amount:        decimal.NewFromInt(1),
				}
				store.EXPECT().TransferTx(gomock.Any(), gomock.Eq(arg)).Times(1)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InternalServerError",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          decimal.NewFromInt(1),
				"currency":        toAccount.Currency,
			},

			buildStubFunc: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).Times(1).Return(toAccount, nil)

				arg := db.TransferTxParams{
					FromAccountId: fromAccount.ID,
					ToAccountId:   toAccount.ID,
					Amount:        decimal.NewFromInt(1),
				}
				store.EXPECT().TransferTx(gomock.Any(), gomock.Eq(arg)).Times(1).Return(db.TransferTxResult{}, pgx.ErrTxClosed)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
				assertError(t, pgx.ErrTxClosed, recorder.Body)
			},
		},
		{
			name: "NotFound",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          decimal.NewFromInt(1),
				"currency":        toAccount.Currency,
			},

			buildStubFunc: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).Times(1).Return(toAccount, nil)

				arg := db.TransferTxParams{
					FromAccountId: fromAccount.ID,
					ToAccountId:   toAccount.ID,
					Amount:        decimal.NewFromInt(1),
				}
				store.EXPECT().TransferTx(gomock.Any(), gomock.Eq(arg)).Times(1).Return(db.TransferTxResult{}, pgx.ErrNoRows)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
				assertError(t, pgx.ErrNoRows, recorder.Body)
			},
		},
		{
			name: "BadRequest",
			body: gin.H{
				"from_account_id": 0,
				"to_account_id":   0,
				"amount":          decimal.NewFromInt(1),
				"currency":        toAccount.Currency,
			},

			buildStubFunc: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRequestMisMatchFromCurrency",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccountDifferentCurrency.ID,
				"amount":          decimal.NewFromInt(1),
				"currency":        toAccountDifferentCurrency.Currency,
			},

			buildStubFunc: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRequestMisMatchToCurrency",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccountDifferentCurrency.ID,
				"amount":          decimal.NewFromInt(1),
				"currency":        fromAccount.Currency,
			},

			buildStubFunc: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccountDifferentCurrency.ID)).Times(1).Return(toAccountDifferentCurrency, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "FromAccountNotFound",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          decimal.NewFromInt(1),
				"currency":        toAccount.Currency,
			},

			buildStubFunc: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(db.Account{}, pgx.ErrNoRows)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
				assertError(t, pgx.ErrNoRows, recorder.Body)
			},
		},
		{
			name: "ToAccountNotFound",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          decimal.NewFromInt(1),
				"currency":        toAccount.Currency,
			},

			buildStubFunc: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).Times(1).Return(db.Account{}, pgx.ErrNoRows)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
				assertError(t, pgx.ErrNoRows, recorder.Body)
			},
		},
		{
			name: "FromAccountInternalServerError",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          decimal.NewFromInt(1),
				"currency":        toAccount.Currency,
			},

			buildStubFunc: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(db.Account{}, pgx.ErrTxClosed)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
				assertError(t, pgx.ErrTxClosed, recorder.Body)
			},
		},
		{
			name: "ToAccountInternalServerError",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          decimal.NewFromInt(1),
				"currency":        toAccount.Currency,
			},

			buildStubFunc: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).Times(1).Return(db.Account{}, pgx.ErrTxClosed)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
				assertError(t, pgx.ErrTxClosed, recorder.Body)
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

			url := "/api/transfers"

			data, err := json.Marshal(tc.body)
			assert.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
			assert.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.validateResponse(recorder)
		})
	}
}
