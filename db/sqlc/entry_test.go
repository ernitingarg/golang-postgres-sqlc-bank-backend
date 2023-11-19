package db

import (
	"context"
	"testing"
	"time"

	"github.com/ernitingarg/golang-postgres-sqlc-bank-backend/utils"
	"github.com/stretchr/testify/assert"
)

func createRandomEntry(t *testing.T) Entry {
	account := createRandomAccount(t)

	arg := CreateEntryParams{
		AccountID: account.ID,
		Amount:    utils.RandomDecimal(2, 1000),
	}

	entry, err := testQueries.CreateEntry(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotEmpty(t, entry)

	assert.Equal(t, arg.AccountID, entry.AccountID)
	assert.Equal(t, arg.Amount, entry.Amount)
	assert.NotZero(t, entry.ID)
	assert.NotEmpty(t, entry.CreatedAt)

	return entry
}

func TestCreateEntry(t *testing.T) {
	createRandomEntry(t)
}

func TestGetEntry(t *testing.T) {
	entry1 := createRandomEntry(t)

	entry2, err := testQueries.GetEntry(context.Background(), entry1.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, entry2)

	assert.Equal(t, entry1.ID, entry2.ID)
	assert.Equal(t, entry1.AccountID, entry2.AccountID)
	assert.Equal(t, entry1.Amount, entry2.Amount)
	assert.WithinDuration(t, entry1.CreatedAt.Time, entry2.CreatedAt.Time, time.Second)
}

func TestListEntries(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomEntry(t)
	}

	arg := ListEntriesParams{
		Limit:  5,
		Offset: 5,
	}

	entries, err := testQueries.ListEntries(context.Background(), arg)
	assert.NoError(t, err)
	assert.Len(t, entries, 5)

	for _, entry := range entries {
		assert.NotEmpty(t, entry)
	}
}
