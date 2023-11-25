package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestTranferTx(t *testing.T) {
	store := NewStore(connPool)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Println(">> before:", account1.Balance, account2.Balance)

	amountToTransfer := decimal.NewFromInt32(10)

	// Make 2 transfers from account1 to account2 for a given amount
	n := 2
	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountId: account1.ID,
				ToAccountId:   account2.ID,
				Amount:        amountToTransfer,
			})

			errs <- err
			results <- result
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		assert.NoError(t, err)

		result := <-results
		assert.NotEmpty(t, result)

		// Validate Transfer record
		transfer := result.Transfer
		assert.NotEmpty(t, transfer)
		assert.Equal(t, account1.ID, transfer.FromAccountID)
		assert.Equal(t, account2.ID, transfer.ToAccountID)
		assert.Equal(t, amountToTransfer, transfer.Amount)
		assert.NotZero(t, transfer.ID)
		assert.NotZero(t, transfer.CreatedAt)

		transfer, err = store.GetTransfer(context.Background(), transfer.ID)
		assert.NoError(t, err)
		assert.NotEmpty(t, transfer)

		// Validate FromEntry record
		fromEntry := result.FromEntry
		assert.NotEmpty(t, fromEntry)
		assert.Equal(t, account1.ID, fromEntry.AccountID)
		assert.Equal(t, amountToTransfer.Mul(decimal.NewFromFloat(-1)), fromEntry.Amount)
		assert.NotZero(t, fromEntry.ID)
		assert.NotZero(t, fromEntry.CreatedAt)

		fromEntry, err = store.GetEntry(context.Background(), fromEntry.ID)
		assert.NoError(t, err)
		assert.NotEmpty(t, fromEntry)

		// Validate ToEntry record
		toEntry := result.ToEntry
		assert.NotEmpty(t, toEntry)
		assert.Equal(t, account2.ID, toEntry.AccountID)
		assert.Equal(t, amountToTransfer, toEntry.Amount)
		assert.NotZero(t, toEntry.ID)
		assert.NotZero(t, toEntry.CreatedAt)

		toEntry, err = store.GetEntry(context.Background(), toEntry.ID)
		assert.NoError(t, err)
		assert.NotEmpty(t, toEntry)

		// Validate Account(s) record
		fromAccount := result.FromAccount
		assert.NotEmpty(t, fromAccount)
		assert.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		assert.NotEmpty(t, toAccount)
		assert.Equal(t, account2.ID, toAccount.ID)

		fmt.Println(">> tx:", fromAccount.Balance, toAccount.Balance)

		diff1 := account1.Balance.Sub(fromAccount.Balance) // initial - current
		diff2 := toAccount.Balance.Sub(account2.Balance)   // current - intial
		assert.Equal(t, diff1, diff2)
	}

	// Check the final updated balance
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, updatedAccount1)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, updatedAccount2)

	fmt.Println(">> after:", updatedAccount1.Balance, updatedAccount2.Balance)
	expectedTotalTransferredAmount := amountToTransfer.Mul(decimal.NewFromInt(int64(n)))

	assert.Equal(t, expectedTotalTransferredAmount, account1.Balance.Sub(updatedAccount1.Balance))
	assert.Equal(t, expectedTotalTransferredAmount, updatedAccount2.Balance.Sub(account2.Balance))
}

func TestTranferTxDeadlock(t *testing.T) {
	store := NewStore(connPool)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Println(">> before:", account1.Balance, account2.Balance)

	amountToTransfer := decimal.NewFromInt32(10)

	// Make 1 transfer from account1 to account2 for a given amount
	// Make 1 transfer from account2 to account1 for a given amount
	n := 2
	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccountId := account1.ID
		toAccountId := account2.ID

		if i%2 == 0 {
			fromAccountId = account2.ID
			toAccountId = account1.ID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountId: fromAccountId,
				ToAccountId:   toAccountId,
				Amount:        amountToTransfer,
			})

			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		assert.NoError(t, err)
	}

	// Check the final updated balance
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, updatedAccount1)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, updatedAccount2)

	fmt.Println(">> after:", updatedAccount1.Balance, updatedAccount2.Balance)

	// As there are equal transfer from a1 to a2 and from a2 to a1.
	// The final balance should be same as initial
	assert.Equal(t, account1.Balance, updatedAccount1.Balance)
	assert.Equal(t, account2.Balance, updatedAccount2.Balance)
}
