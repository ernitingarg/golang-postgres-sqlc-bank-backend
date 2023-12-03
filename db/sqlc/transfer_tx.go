package db

import (
	"context"

	"github.com/shopspring/decimal"
)

// TransferTxParams contains the input parameters of the transfer transaction
type TransferTxParams struct {
	FromAccountId int64           `json:"from_account_id"`
	ToAccountId   int64           `json:"to_account_id"`
	Amount        decimal.Decimal `json:"amount"`
}

// TransferTxResult contains the out parameters of the transfer transaction
type TransferTxResult struct {
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	Transfer    Transfer `json:"transfer"`
}

// TransferTx tranfer amount from one account to another account.
// It creates transfer record, from/to entries and update balances of from/to accounts
func (store *SqlStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	amountToWithdraw := arg.Amount.Mul(decimal.NewFromFloat(-1))
	amountToDeposit := arg.Amount

	txErr := store.execTx(ctx, func(q *Queries) error {
		var err error

		// Create Transfer Record
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountId,
			ToAccountID:   arg.ToAccountId,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		// Create FromEntry record
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountId,
			Amount:    amountToWithdraw,
		})
		if err != nil {
			return err
		}

		// Create ToEntry record
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountId,
			Amount:    amountToDeposit,
		})
		if err != nil {
			return err
		}

		// To avoid deadlock, always perform transfer in a specific order
		// for eg: these 2 txs will endup in deadlock. t1: a1 -> a2 and t2: a2 -> a1.
		// To avoid deadlock, always update balance of a1 before a2.
		if arg.FromAccountId < arg.ToAccountId {
			result.FromAccount, result.ToAccount, err = addAmount(
				ctx,
				q,
				arg.FromAccountId,
				amountToWithdraw,
				arg.ToAccountId,
				amountToDeposit)
		} else {
			result.ToAccount, result.FromAccount, err = addAmount(
				ctx,
				q,
				arg.ToAccountId,
				amountToDeposit,
				arg.FromAccountId,
				amountToWithdraw)
		}

		return nil
	})

	return result, txErr
}

func addAmount(
	ctx context.Context,
	q *Queries,
	account1ID int64,
	account1Amount decimal.Decimal,
	account2ID int64,
	account2Amount decimal.Decimal) (account1 Account, account2 Account, err error) {

	// Add amount to the balance of account1
	account1, err = q.AddAccountBalance(
		ctx,
		AddAccountBalanceParams{
			ID:     account1ID,
			Amount: account1Amount,
		})
	if err != nil {
		return
	}

	// Add amount to the balance of account2
	account2, err = q.AddAccountBalance(
		ctx,
		AddAccountBalanceParams{
			ID:     account2ID,
			Amount: account2Amount,
		})

	return
}
