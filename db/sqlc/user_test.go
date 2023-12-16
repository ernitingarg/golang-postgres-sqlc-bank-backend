package db

import (
	"context"
	"testing"
	"time"

	"github.com/ernitingarg/golang-postgres-sqlc-bank-backend/utils"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

func createRandomUser(t *testing.T) User {
	hashedPassword, err := utils.HashPassword(utils.RandomString(8))
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)

	arg := CreateUserParams{
		Name:         utils.RandomString(6),
		Email:        utils.RandomEmail(),
		HashPassword: hashedPassword,
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotEmpty(t, user)
	assert.Equal(t, arg.Name, user.Name)
	assert.Equal(t, arg.Email, user.Email)
	assert.Equal(t, arg.HashPassword, user.HashPassword)
	assert.NotEmpty(t, user.CreatedAt)
	assert.NotEmpty(t, user.UpdatedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)

	user2, err := testQueries.GetUser(context.Background(), user1.Name)
	assert.NoError(t, err)
	assert.NotEmpty(t, user2)

	assert.Equal(t, user1.Name, user2.Name)
	assert.Equal(t, user1.Email, user2.Email)
	assert.Equal(t, user1.HashPassword, user2.HashPassword)
	assert.WithinDuration(t, user1.CreatedAt.Time, user2.CreatedAt.Time, time.Second)
	assert.WithinDuration(t, user1.UpdatedAt.Time, user2.UpdatedAt.Time, time.Second)
}

func TestUpdateUser(t *testing.T) {
	user1 := createRandomUser(t)

	hashedPasswordToBeUpdated, err := utils.HashPassword(utils.RandomString(8))
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPasswordToBeUpdated)

	arg := UpdateUserParams{
		Name:         user1.Name,
		HashPassword: hashedPasswordToBeUpdated,
	}

	user2, err := testQueries.UpdateUser(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotEmpty(t, user2)

	assert.Equal(t, user1.Name, user2.Name)
	assert.Equal(t, user1.Email, user2.Email)
	assert.Equal(t, arg.HashPassword, user2.HashPassword)
	assert.WithinDuration(t, user1.CreatedAt.Time, user2.CreatedAt.Time, time.Second)
	assert.WithinDuration(t, user1.UpdatedAt.Time, user2.UpdatedAt.Time, time.Second)
}

func TestDeleteUser(t *testing.T) {
	user1 := createRandomUser(t)

	err := testQueries.DeleteUser(context.Background(), user1.Name)
	assert.NoError(t, err)

	user2, err := testQueries.GetUser(context.Background(), user1.Name)
	assert.Error(t, err)
	assert.EqualError(t, err, pgx.ErrNoRows.Error())
	assert.Empty(t, user2)
}

func TestListUsers(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomUser(t)
	}

	arg := ListUsersParams{
		Limit:  5,
		Offset: 5,
	}

	users, err := testQueries.ListUsers(context.Background(), arg)
	assert.NoError(t, err)
	assert.Len(t, users, 5)

	for _, user := range users {
		assert.NotEmpty(t, user)
	}
}
