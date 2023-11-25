package db

import (
	"context"
	"fmt"
)

// execTx executes operations within a single transaction
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {

	tx, err := store.connPool.Begin(ctx)
	if err != nil {
		return err
	}

	err = fn(store.Queries)
	if err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("Tx err: %v, Rollback err: %v", err, rollbackErr)
		}

		return err
	}

	return tx.Commit(ctx)
}
