package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	mockdb "github.com/ernitingarg/golang-postgres-sqlc-bank-backend/db/mock"
	db "github.com/ernitingarg/golang-postgres-sqlc-bank-backend/db/sqlc"
	"github.com/ernitingarg/golang-postgres-sqlc-bank-backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(x any) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	err := utils.VerifyPassword(e.password, arg.HashPassword)
	if err != nil {
		return false
	}

	e.arg.HashPassword = arg.HashPassword

	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matched arg %v and password %v", e.arg, e.password)
}

func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
}

func TestCreateUserApi(t *testing.T) {
	user, password := createRandomUser()

	testCases := []struct {
		name             string
		body             gin.H
		buildStubFunc    func(store *mockdb.MockStore)
		validateResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"name":     user.Name,
				"email":    user.Email,
				"password": password,
			},
			buildStubFunc: func(store *mockdb.MockStore) {
				arg := db.CreateUserParams{
					Name:         user.Name,
					Email:        user.Email,
					HashPassword: user.HashPassword,
				}

				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
					Times(1).
					Return(user, nil)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
				assertUser(t, user, recorder.Body)
			},
		},
		{
			name: "InernalServerError",
			body: gin.H{
				"name":     user.Name,
				"email":    user.Email,
				"password": password,
			},
			buildStubFunc: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, pgx.ErrTxClosed)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
				assertError(t, pgx.ErrTxClosed, recorder.Body)
			},
		},
		{
			name: "InvalidUserName",
			body: gin.H{
				"name":     "@nitin",
				"email":    user.Email,
				"password": password,
			},
			buildStubFunc: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidPassword",
			body: gin.H{
				"name":     user.Name,
				"email":    user.Email,
				"password": utils.RandomString(7),
			},
			buildStubFunc: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidEmail",
			body: gin.H{
				"name":     user.Name,
				"email":    user.Name,
				"password": password,
			},
			buildStubFunc: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "DuplicateUserUniqueViolation",
			body: gin.H{
				"name":     user.Name,
				"email":    user.Email,
				"password": password,
			},
			buildStubFunc: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, getForbinddenError())
			},
			validateResponse: func(recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusForbidden, recorder.Code)
				assertError(t, getForbinddenError(), recorder.Body)
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

			url := "/api/users"
			data, err := json.Marshal(tc.body)
			assert.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			assert.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.validateResponse(recorder)
		})
	}
}

func createRandomUser() (db.User, string) {
	password := utils.RandomString(8)
	hashedPassword, _ := utils.HashPassword(password)

	user := db.User{
		Name:         utils.RandomString(6),
		Email:        utils.RandomEmail(),
		HashPassword: hashedPassword,
	}

	return user, password
}

func assertUser(t *testing.T, expectedUser db.User, body *bytes.Buffer) {
	data, err := io.ReadAll(body)
	assert.NoError(t, err)

	var actualUser db.User
	err = json.Unmarshal(data, &actualUser)
	assert.NoError(t, err)

	assert.Equal(t, expectedUser.Name, actualUser.Name)
	assert.Equal(t, expectedUser.Email, actualUser.Email)
	assert.Empty(t, actualUser.HashPassword)
}

func getForbinddenError() *pgconn.PgError {
	return pgconn.ErrorResponseToPgError(&pgproto3.ErrorResponse{Code: "23505"})
}
